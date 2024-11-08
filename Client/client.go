package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	ErrFetchingExchangeRate = "error fetching dollar exchange rate"
	ErrCreatingHTTPRequest  = "error creating HTTP request"
	ErrSendingHTTPRequest   = "error sending HTTP request"
	ErrDecodingResponse     = "error decoding response body"
	ErrCreatingFile         = "error creating file"
	ErrWritingToFile        = "error writing to file"
)

const (
	FileName           = "exchange_rate.txt"
	ExchangeRatePrefix = "Exchange Rate (USD): "
)

var (
	ExchangeRateAPIURL = os.Getenv("EXCHANGE_RATE_API_URL")
)

type ExchangeRate struct {
	USDValue string `json:"bid"`
}

func main() {
	if ExchangeRateAPIURL == "" {
		log.Println("Environment variable EXCHANGE_RATE_API_URL is not set.")
		return
	}

	exchangeRate, err := fetchDollarExchangeRate()
	if err != nil {
		log.Println(ErrFetchingExchangeRate, ":", err)
		return
	}

	displayExchangeRate(exchangeRate.USDValue)
	saveExchangeRateToFile(exchangeRate.USDValue)
}

func fetchDollarExchangeRate() (*ExchangeRate, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, ExchangeRateAPIURL, nil)
	if err != nil {
		return nil, errors.New(ErrCreatingHTTPRequest + ": " + err.Error())
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.New(ErrSendingHTTPRequest + ": " + err.Error())
	}
	defer response.Body.Close()

	var exchangeRate ExchangeRate
	err = json.NewDecoder(response.Body).Decode(&exchangeRate)
	if err != nil {
		return nil, errors.New(ErrDecodingResponse + ": " + err.Error())
	}

	return &exchangeRate, nil
}

func displayExchangeRate(value string) {
	println(ExchangeRatePrefix + value)
}

func saveExchangeRateToFile(value string) {
	file, err := os.Create(FileName)
	if err != nil {
		log.Println(ErrCreatingFile, ":", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(ExchangeRatePrefix + value + "\n")
	if err != nil {
		log.Println(ErrWritingToFile, ":", err)
	}
}
