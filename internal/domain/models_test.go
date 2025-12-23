package domain

import (
	"testing"
	"time"
)

func TestCurrency_String(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		want     string
	}{
		{"USD", "USD", "USD"},
		{"EUR", "EUR", "EUR"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.currency); got != tt.want {
				t.Errorf("Currency.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExchangeRate_Valid(t *testing.T) {
	now := time.Now()
	rate := ExchangeRate{
		Date:   now,
		Base:   "USD",
		Target: "EUR",
		Rate:   0.85,
	}

	if rate.Base != "USD" {
		t.Errorf("Base = %v, want USD", rate.Base)
	}
	if rate.Target != "EUR" {
		t.Errorf("Target = %v, want EUR", rate.Target)
	}
	if rate.Rate != 0.85 {
		t.Errorf("Rate = %v, want 0.85", rate.Rate)
	}
	if !rate.Date.Equal(now) {
		t.Errorf("Date = %v, want %v", rate.Date, now)
	}
}

func TestDataPoint_GetRate(t *testing.T) {
	dp := DataPoint{
		Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Rates: map[Currency]float64{
			"EUR": 0.85,
			"GBP": 0.73,
		},
	}

	tests := []struct {
		name     string
		currency Currency
		want     float64
		exists   bool
	}{
		{"EUR exists", "EUR", 0.85, true},
		{"GBP exists", "GBP", 0.73, true},
		{"JPY not exists", "JPY", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := dp.Rates[tt.currency]
			if exists != tt.exists {
				t.Errorf("Rate exists = %v, want %v", exists, tt.exists)
			}
			if exists && got != tt.want {
				t.Errorf("Rate = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeSeriesData_Valid(t *testing.T) {
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	ts := TimeSeriesData{
		Base:      "USD",
		Targets:   []Currency{"EUR", "GBP"},
		StartDate: start,
		EndDate:   end,
		DataPoints: []DataPoint{
			{
				Date: start,
				Rates: map[Currency]float64{
					"EUR": 0.85,
					"GBP": 0.73,
				},
			},
		},
	}

	if ts.Base != "USD" {
		t.Errorf("Base = %v, want USD", ts.Base)
	}
	if len(ts.Targets) != 2 {
		t.Errorf("Targets length = %d, want 2", len(ts.Targets))
	}
	if !ts.StartDate.Equal(start) {
		t.Errorf("StartDate = %v, want %v", ts.StartDate, start)
	}
	if !ts.EndDate.Equal(end) {
		t.Errorf("EndDate = %v, want %v", ts.EndDate, end)
	}
	if len(ts.DataPoints) != 1 {
		t.Errorf("DataPoints length = %d, want 1", len(ts.DataPoints))
	}
}

func TestTimeSeriesData_Empty(t *testing.T) {
	ts := TimeSeriesData{
		Base:       "USD",
		Targets:    []Currency{},
		StartDate:  time.Now(),
		EndDate:    time.Now(),
		DataPoints: []DataPoint{},
	}

	if len(ts.Targets) != 0 {
		t.Errorf("Expected empty Targets, got %d", len(ts.Targets))
	}
	if len(ts.DataPoints) != 0 {
		t.Errorf("Expected empty DataPoints, got %d", len(ts.DataPoints))
	}
}
