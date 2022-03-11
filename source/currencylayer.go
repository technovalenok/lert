package source

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/technovalenok/lert/app"
	"go.uber.org/zap"
)

// CurrencyLayerSource is a source of https://currencylayer.com/
type CurrencyLayerSource struct {
	apiKey string
	code   string
}

type CurrencyLayerResponse struct {
	Success   bool               `json:"success"`
	Source    string             `json:"source"`
	Quotes    map[string]float64 `json:"quotes"`
	Timestamp int                `json:"timestamp"`
}

func NewCurrencyLayerSource(code, apiKey string) app.SourceInterface {
	return CurrencyLayerSource{code: code, apiKey: apiKey}
}

func (s CurrencyLayerSource) Code() string {
	return s.code
}

func (s *CurrencyLayerSource) ApiKey() string {
	return s.apiKey
}

func (s CurrencyLayerSource) Rates() ([]app.Rate, error) {
	params := "access_key=" + url.QueryEscape(s.apiKey) + "&currencies=EUR,RUB&source=USD"
	uri := fmt.Sprintf("http://apilayer.net/api/live?%s", params)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, &app.ErrSourceDataUnavailable{
			Code:    s.code,
			Message: fmt.Sprintf("Unable to get source data: %s", err),
		}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &app.ErrSourceDataUnavailable{
			Code:    s.code,
			Message: fmt.Sprintf("Unable to read source response body: %s", err),
		}
	}

	zap.S().Infof("Source %s response: %s", s.code, string(body))

	var response CurrencyLayerResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, &app.ErrSourceDataUnavailable{
			Code:    s.code,
			Message: fmt.Sprintf("Unable to decode response: %s", err),
		}
	}

	if response.Success == false {
		return nil, &app.ErrSourceDataUnavailable{
			Code:    s.code,
			Message: "Response failed with false status code",
		}
	}

	// TODO to database config?
	loc, _ := time.LoadLocation("UTC")
	updatedAt := time.Unix(int64(response.Timestamp), 0).In(loc)
	rubRate := app.Rate{
		Source:    s.code,
		UpdatedAt: updatedAt.Format(time.RFC3339),
		From:      app.CurrencyUSD,
		To:        app.CurrencyRUB,
		Rate:      response.Quotes["USDRUB"],
	}
	eurRate := app.Rate{
		Source:    s.code,
		UpdatedAt: updatedAt.Format(time.RFC3339),
		From:      app.CurrencyUSD,
		To:        app.CurrencyEUR,
		Rate:      response.Quotes["USDEUR"],
	}

	return []app.Rate{rubRate, eurRate}, nil
}
