package service

import (
	"context"
	"testing"
	"time"

	"github.com/kaze/xrv/internal/cache"
	"github.com/kaze/xrv/internal/providers"
	"github.com/kaze/xrv/internal/domain"
)

type mockAPIClient struct {
	timeSeriesResponse *providers.TimeSeriesResponse
	currenciesResponse providers.CurrenciesResponse
	err                error
}

func (m *mockAPIClient) GetTimeSeriesRates(ctx context.Context, startDate, endDate time.Time, base string, targets []string) (*providers.TimeSeriesResponse, error) {
	return m.timeSeriesResponse, m.err
}

func (m *mockAPIClient) GetSupportedCurrencies(ctx context.Context) (providers.CurrenciesResponse, error) {
	return m.currenciesResponse, m.err
}

func TestService_FetchTimeSeriesData(t *testing.T) {
	mockAPI := &mockAPIClient{
		timeSeriesResponse: &providers.TimeSeriesResponse{
			Base:      "USD",
			StartDate: "2023-01-01",
			EndDate:   "2023-01-03",
			Rates: map[string]map[string]float64{
				"2023-01-01": {"EUR": 0.85, "GBP": 0.73},
				"2023-01-02": {"EUR": 0.86, "GBP": 0.74},
				"2023-01-03": {"EUR": 0.87, "GBP": 0.75},
			},
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	

	svc := NewService(mockAPI, memCache)

	ctx := context.Background()
	opts := FetchOptions{
		Base:      "USD",
		Targets:   []domain.Currency{"EUR", "GBP"},
		StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
		UseCache:  true,
	}

	data, err := svc.FetchTimeSeriesData(ctx, opts)
	if err != nil {
		t.Fatalf("FetchTimeSeriesData() error = %v", err)
	}

	if data.Base != "USD" {
		t.Errorf("Base = %v, want USD", data.Base)
	}

	if len(data.Targets) != 2 {
		t.Errorf("Targets length = %d, want 2", len(data.Targets))
	}

	if len(data.DataPoints) != 3 {
		t.Errorf("DataPoints length = %d, want 3", len(data.DataPoints))
	}

	data2, err := svc.FetchTimeSeriesData(ctx, opts)
	if err != nil {
		t.Fatalf("FetchTimeSeriesData() second call error = %v", err)
	}

	if len(data2.DataPoints) != 3 {
		t.Errorf("Cached DataPoints length = %d, want 3", len(data2.DataPoints))
	}
}

func TestService_FetchTimeSeriesData_NoCache(t *testing.T) {
	mockAPI := &mockAPIClient{
		timeSeriesResponse: &providers.TimeSeriesResponse{
			Base:      "USD",
			StartDate: "2023-01-01",
			EndDate:   "2023-01-02",
			Rates: map[string]map[string]float64{
				"2023-01-01": {"EUR": 0.85},
			},
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	
	svc := NewService(mockAPI, memCache)

	ctx := context.Background()
	opts := FetchOptions{
		Base:      "USD",
		Targets:   []domain.Currency{"EUR"},
		StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		UseCache:  false,
	}

	data, err := svc.FetchTimeSeriesData(ctx, opts)
	if err != nil {
		t.Fatalf("FetchTimeSeriesData() error = %v", err)
	}

	if data.Base != "USD" {
		t.Errorf("Base = %v, want USD", data.Base)
	}
}

func TestService_CalculateStatistics(t *testing.T) {
	mockAPI := &mockAPIClient{}
	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	
	svc := NewService(mockAPI, memCache)

	data := &domain.TimeSeriesData{
		Base:    "USD",
		Targets: []domain.Currency{"EUR"},
		DataPoints: []domain.DataPoint{
			{Date: time.Now(), Rates: map[domain.Currency]float64{"EUR": 0.85}},
			{Date: time.Now(), Rates: map[domain.Currency]float64{"EUR": 0.86}},
			{Date: time.Now(), Rates: map[domain.Currency]float64{"EUR": 0.87}},
		},
	}

	stats := svc.CalculateStatistics(data)

	if len(stats) != 1 {
		t.Fatalf("Expected 1 currency stats, got %d", len(stats))
	}

	eurStats, exists := stats["EUR"]
	if !exists {
		t.Fatal("EUR stats not found")
	}

	if eurStats.Basic.Min != 0.85 {
		t.Errorf("Min = %v, want 0.85", eurStats.Basic.Min)
	}
}

func TestService_GetSupportedCurrencies(t *testing.T) {
	mockAPI := &mockAPIClient{
		currenciesResponse: providers.CurrenciesResponse{
			"USD": "United States Dollar",
			"EUR": "Euro",
			"GBP": "British Pound",
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	
	svc := NewService(mockAPI, memCache)

	ctx := context.Background()
	currencies, err := svc.GetSupportedCurrencies(ctx)
	if err != nil {
		t.Fatalf("GetSupportedCurrencies() error = %v", err)
	}

	if len(currencies) != 3 {
		t.Errorf("Currencies length = %d, want 3", len(currencies))
	}

	if currencies["USD"] != "United States Dollar" {
		t.Errorf("USD = %v, want United States Dollar", currencies["USD"])
	}
}
