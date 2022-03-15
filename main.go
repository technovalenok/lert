package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/technovalenok/lert/app"
	"github.com/technovalenok/lert/client"
	"github.com/technovalenok/lert/handler"
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

func (a *App) GetAllRates(workers ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(workers))
	defer wg.Wait()

	for _, f := range workers {
		go func(w func()) {
			defer wg.Done()
			w()
		}(f)
	}
}

func (a *App) GetRatesHandler(res http.ResponseWriter, _ *http.Request) {
	header := res.Header()
	header.Set("Content-Type", "application/json")

	result := make(map[string][]app.Rate)
	var w []func()
	for _, src := range a.Sources {
		s := src
		w = append(w, func() {
			rates, err := s.Rates()
			if err != nil {
				zap.S().Errorf("Source %s error: %v", src.Code(), err)
			} else {
				result[s.Code()] = rates
			}
		})
	}

	a.GetAllRates(w...)

	var rs []app.Rate
	for _, v := range result {
		rs = append(rs, v...)
	}

	response := &Response{rs}
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
		logger.Fatal("Error loading .env file")
	}

	httpClient := client.NewHTTPClient()

	// init currencylayer.com source
	currencyLayerApiKey := os.Getenv("API_CURRENCYLAYER_KEY")
	currencyLayerSource := source.NewCurrencyLayerSource("currencylayer", currencyLayerApiKey, *httpClient)

	// init currencyapi.com source
	currencyApiApiKey := os.Getenv("API_CURRENCYAPI_KEY")
	currencyApiSource := source.NewCurrencyApiSource("currencyapi", currencyApiApiKey, *httpClient)

	application := &App{}
	application.
		AddSource(currencyLayerSource).
		AddSource(currencyApiSource)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/rate", application.GetRatesHandler)
	wrappedMux := handler.NewHttpHandler(mux)

	logger.Fatal("Server error", zap.Error(http.ListenAndServe(":9000", wrappedMux)))
}
