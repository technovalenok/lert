package source

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/technovalenok/lert/app"
)

// CurrencyLayerSource is a source of https://currencylayer.com/
type CurrencyLayerSource struct {
	apiKey string
	code   string
	url    string
	client http.Client
}

type CurrencyLayerResponse struct {
	Success   bool               `json:"success"`
	Source    string             `json:"source"`
	Quotes    map[string]float64 `json:"quotes"`
	Timestamp int                `json:"timestamp"`
}

func NewCurrencyLayerSource(code, apiKey string, client http.Client) app.SourceInterface {
	return CurrencyLayerSource{
		code:   code,
		apiKey: apiKey,
		url:    "http://apilayer.net/api/live",
		client: client,
	}
}

func (s CurrencyLayerSource) Code() string {
	return s.code
}

func (s CurrencyLayerSource) Rates() ([]app.Rate, error) {
	params := "access_key=" + url.QueryEscape(s.apiKey) + "&currencies=EUR,RUB&source=USD"
	uri := fmt.Sprintf("%s?%s", s.url, params)
	resp, err := s.client.Get(uri)

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
