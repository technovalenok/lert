package app

import (
	"fmt"
)

const (
	CurrencyRUB = "RUB"
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
)

type ErrSourceDataUnavailable struct {
	Code    string
	Message string
}

func (esu ErrSourceDataUnavailable) Error() string {
	return fmt.Sprintf("Source %s unavailable: %s", esu.Code, esu.Message)
}

type Currency string

type Rate struct {
	Source    string
	UpdatedAt string
	From      Currency
	To        Currency
	Rate      float64
}

type SourceInterface interface {
	Rates() ([]Rate, error)
	Code() string
}
