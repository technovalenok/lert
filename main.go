package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/technovalenok/lert/app"

	"github.com/technovalenok/lert/source"
)

type Response struct {
	Rates []app.Rate `json:"rates"`
}

type App struct {
	Sources []app.SourceInterface
}

func (a *App) AddSource(src app.SourceInterface) {
	a.Sources = append(a.Sources, src)
}

func (a *App) GetRatesHandler(res http.ResponseWriter, _ *http.Request) {
	header := res.Header()
	header.Set("Content-Type", "application/json")

	var result []app.Rate
	for _, src := range a.Sources {
		if rates, err := src.Rates(); err != nil {
			log.Printf("Source %s error: %v", src.Code(), err) // TODO log
			continue
		} else {
			result = append(result, rates...)
		}
	}

	response := &Response{result}
	responseJson, err := json.Marshal(response)
	if err != nil {
		log.Printf("Unable to marshall response: %s", err)
		res.WriteHeader(http.StatusInternalServerError)
		// TODO 5xx error response body
	} else {
		res.WriteHeader(http.StatusOK)
		fmt.Fprint(res, string(responseJson))
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file") // TODO logging
	}

	// init currencylayer.com source
	currencyLayerApiKey := os.Getenv("API_CURRENCYLAYER_KEY")
	currencyLayerSource := source.NewCurrencyLayerClient("currencylayer", currencyLayerApiKey)

	application := &App{}
	application.AddSource(currencyLayerSource)

	http.HandleFunc("/api/rate", application.GetRatesHandler)

	log.Fatal(http.ListenAndServe(":9000", nil))
}
