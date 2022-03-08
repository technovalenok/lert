package source

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/technovalenok/lert/app"
)

type CurrencyLayerSource struct {
	apiKey string
	code   string
}

type CurrencyLayerApiResponse struct {
	Success bool               `json:"success"`
	Source  string             `json:"source"`
	Quotes  map[string]float64 `json:"quotes"`
}

func NewCurrencyLayerClient(code, apiKey string) app.SourceInterface {
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

	log.Printf("Source %s response: %s", s.code, string(body)) // TODO log + interceptor

	var response CurrencyLayerApiResponse
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

	// TODO to database?
	rubRate := app.Rate{
		From: app.CurrencyUSD,
		To:   app.CurrencyRUB,
		Rate: response.Quotes["USDRUB"],
	}
	eurRate := app.Rate{
		From: app.CurrencyUSD,
		To:   app.CurrencyEUR,
		Rate: response.Quotes["USDEUR"],
	}

	return []app.Rate{rubRate, eurRate}, nil
}
