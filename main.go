package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	_ "github.com/mattn/go-sqlite3"

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
	DB      *sql.DB
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

func (a *App) SaveRates(rates []app.Rate) error {
	for _, rate := range rates {
		statement, err := a.DB.Prepare(" INSERT INTO rates (source, currency_from, currency_to, rate, updated_at) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			return err
		}
		_, err = statement.Exec(rate.Source, rate.From, rate.To, rate.Rate, rate.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
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
			err = a.SaveRates(rates)
			if err != nil {
				zap.S().Errorf("Error saving source %s to database: %v", src.Code(), err)
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
	// Init logger
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer logger.Sync()
	logger.Info("Init server")

	// Load env
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	// Init database
	db, err := sql.Open("sqlite3", os.Getenv("DATABASE_DSN"))
	if err != nil {
		zap.S().Fatalf("DB init error: %s", err)
	}
	defer db.Close()

	portEnv := os.Getenv("APP_SERVER_PORT")
	if portEnv == "" {
		zap.S().Fatalf("APP_SERVER_PORT variable must be set")
	}
	port, _ := strconv.Atoi(portEnv)

	// Init application
	httpClient := client.NewHTTPClient()

	// Init currencylayer.com source
	currencyLayerApiKey := os.Getenv("API_CURRENCYLAYER_KEY")
	currencyLayerSource := source.NewCurrencyLayerSource("currencylayer", currencyLayerApiKey, *httpClient)

	// Init currencyapi.com source
	currencyApiApiKey := os.Getenv("API_CURRENCYAPI_KEY")
	currencyApiSource := source.NewCurrencyApiSource("currencyapi", currencyApiApiKey, *httpClient)

	application := &App{}
	application.DB = db
	application.
		AddSource(currencyLayerSource).
		AddSource(currencyApiSource)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/rate", application.GetRatesHandler)
	wrappedMux := handler.NewHttpHandler(mux)

	logger.Fatal("Server error", zap.Error(http.ListenAndServe(fmt.Sprintf(":%d", port), wrappedMux)))
}
