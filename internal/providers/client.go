package providers

import (
	"context"
	"time"
)

type APIClient interface {
	GetTimeSeriesRates(ctx context.Context, startDate, endDate time.Time, base string, targets []string) (*TimeSeriesResponse, error)
	GetSupportedCurrencies(ctx context.Context) (CurrenciesResponse, error)
}

type APIError struct {
	StatusCode int
	Message    string
	URL        string
}

func (e *APIError) Error() string {
	return e.Message
}
