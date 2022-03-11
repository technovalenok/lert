package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/technovalenok/lert/app"
	"go.uber.org/zap"

	"github.com/technovalenok/lert/source"
)

type Response struct {
	Rates []app.Rate `json:"rates"`
}

type App struct {
	Sources []app.SourceInterface
}

func (a *App) AddSource(src app.SourceInterface) *App {
	a.Sources = append(a.Sources, src)
	return a
}

func (a *App) GetRatesHandler(res http.ResponseWriter, _ *http.Request) {
	header := res.Header()
	header.Set("Content-Type", "application/json")

	var result []app.Rate
	for _, src := range a.Sources {
		if rates, err := src.Rates(); err != nil {
			zap.S().Errorf("Source %s error: %v", src.Code(), err)
			continue
		} else {
			result = append(result, rates...)
		}
	}

	response := &Response{result}
	responseJson, err := json.Marshal(response)
	if err != nil {
		zap.S().Errorf("Unable to marshall response: %s", err)
		res.WriteHeader(http.StatusInternalServerError)
		// TODO 5xx error response body
	} else {
		res.WriteHeader(http.StatusOK)
		fmt.Fprint(res, string(responseJson))
	}
}

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	logger.Info("Init server")
	err := godotenv.Load()
	if err != nil {
		logger.Info("Error loading .env file")
	}

	// init currencylayer.com source
	currencyLayerApiKey := os.Getenv("API_CURRENCYLAYER_KEY")
	currencyLayerSource := source.NewCurrencyLayerSource("currencylayer", currencyLayerApiKey)

	// init currencyapi.com source
	currencyApiApiKey := os.Getenv("API_CURRENCYAPI_KEY")
	currencyApiSource := source.NewCurrencyApiSource("currencyapi", currencyApiApiKey)

	application := &App{}
	application.
		AddSource(currencyLayerSource).
		AddSource(currencyApiSource)

	http.HandleFunc("/api/rate", application.GetRatesHandler)

	logger.Fatal("Server error", zap.Error(http.ListenAndServe(":9000", nil)))
}
