package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	ErrNotFound          = "error: endpoint not found"
	ErrFetchingAPI       = "error fetching exchange rate from external API"
	ErrCreatingRequest   = "error creating HTTP request"
	ErrDecodingResponse  = "error decoding response body"
	ErrDatabaseOperation = "error saving to database"
)

const (
	ExchangeRateEndpoint = "/exchange-rate"
	DatabaseFile         = "./database.db"
)

var (
	ExternalAPIURL = os.Getenv("EXTERNAL_API_URL")
)

type ExchangeRateAPI struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type ExchangeRateDB struct {
	ID  int `gorm:"primaryKey"`
	Bid string
	gorm.Model
}

func main() {
	http.HandleFunc(ExchangeRateEndpoint, exchangeRateHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func exchangeRateHandler(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != ExchangeRateEndpoint {
		log.Println(ErrNotFound)
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte(ErrNotFound))
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)

	exchangeRate, err := fetchExchangeRate()
	if err != nil {
		log.Println(ErrFetchingAPI, ":", err)
		http.Error(response, ErrFetchingAPI, http.StatusInternalServerError)
		return
	}

	err = saveExchangeRateToDatabase(exchangeRate)
	if err != nil {
		log.Println(ErrDatabaseOperation, ":", err)
	}

	json.NewEncoder(response).Encode(exchangeRate.USDBRL)
}

func fetchExchangeRate() (*ExchangeRateAPI, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ExternalAPIURL, nil)
	if err != nil {
		return nil, errors.New(ErrCreatingRequest + ": " + err.Error())
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New(ErrFetchingAPI + ": " + err.Error())
	}
	defer res.Body.Close()

	var exchangeRate ExchangeRateAPI
	err = json.NewDecoder(res.Body).Decode(&exchangeRate)
	if err != nil {
		return nil, errors.New(ErrDecodingResponse + ": " + err.Error())
	}

	return &exchangeRate, nil
}

func saveExchangeRateToDatabase(exchangeRate *ExchangeRateAPI) error {
	db, err := gorm.Open(sqlite.Open(DatabaseFile), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&ExchangeRateDB{})

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	dbError := db.WithContext(ctx).Create(&ExchangeRateDB{Bid: exchangeRate.USDBRL.Bid})
	if dbError.Error != nil {
		return dbError.Error
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
