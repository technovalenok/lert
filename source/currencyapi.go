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

// CurrencyApiSource is a source of https://currencyapi.com
type CurrencyApiSource struct {
	apiKey string
	code   string
	url    string
	client http.Client
}

type CurrencyApiData struct {
	Code  string  `json:"code"`
	Value float64 `json:"value"`
}

type CurrencyApiResponseMeta struct {
	UpdatedAt string `json:"last_updated_at"`
}

type CurrencyApiResponse struct {
	Meta CurrencyApiResponseMeta    `json:"meta,omitempty"`
	Data map[string]CurrencyApiData `json:"data"`
}

func NewCurrencyApiSource(code, apiKey string, client http.Client) app.SourceInterface {
	return CurrencyApiSource{
		code:   code,
		apiKey: apiKey,
		url:    "https://api.currencyapi.com/v3/latest",
		client: client,
	}
}

func (s CurrencyApiSource) Code() string {
	return s.code
}

func (s CurrencyApiSource) Rates() ([]app.Rate, error) {
	params := "apikey=" + url.QueryEscape(s.apiKey) + "&currencies=EUR,RUB&base_currency=USD"
	uri := fmt.Sprintf("%s?%s", s.url, params)
	resp, err := s.client.Get(uri)

	if resp.StatusCode != http.StatusOK {
		return nil, &app.ErrSourceDataUnavailable{
			Code:    s.code,
			Message: fmt.Sprintf("Invalid response code (%d)", resp.StatusCode),
		}
	}
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

	var response CurrencyApiResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, &app.ErrSourceDataUnavailable{
			Code:    s.code,
			Message: fmt.Sprintf("Unable to decode response: %s", err),
		}
	}

	// TODO to database?
	lastUpdatedAt, err := time.Parse(time.RFC3339, response.Meta.UpdatedAt)
	if err != nil {
		return nil, &app.ErrSourceDataUnavailable{
			Code:    s.code,
			Message: fmt.Sprintf("Unable to parse last update time: %s", err),
		}
	}
	rubRate := app.Rate{
		Source:    s.code,
		UpdatedAt: lastUpdatedAt.Format(time.RFC3339),
		From:      app.CurrencyUSD,
		To:        app.CurrencyRUB,
		Rate:      response.Data["RUB"].Value,
	}
	eurRate := app.Rate{
		Source:    s.code,
		UpdatedAt: lastUpdatedAt.Format(time.RFC3339),
		From:      app.CurrencyUSD,
		To:        app.CurrencyEUR,
		Rate:      response.Data["EUR"].Value,
	}

	return []app.Rate{rubRate, eurRate}, nil
}
