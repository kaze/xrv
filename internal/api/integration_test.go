// +build integration

package api

import (
	"context"
	"testing"
	"time"
)

// Run with: go test -tags=integration ./internal/api/...

func TestRealAPI_GetSupportedCurrencies(t *testing.T) {
	client := NewClient("https://api.frankfurter.dev/v1", 30*time.Second, 3)
	ctx := context.Background()

	resp, err := client.GetSupportedCurrencies(ctx)
	if err != nil {
		t.Fatalf("GetSupportedCurrencies() error = %v", err)
	}

	// Check some common currencies exist
	commonCurrencies := []string{"USD", "EUR", "GBP", "JPY"}
	for _, currency := range commonCurrencies {
		if _, exists := resp[currency]; !exists {
			t.Errorf("Expected currency %s not found in response", currency)
		}
	}

	t.Logf("Found %d supported currencies", len(resp))
}

func TestRealAPI_GetTimeSeriesRates_ShortRange(t *testing.T) {
	client := NewClient("https://api.frankfurter.dev/v1", 30*time.Second, 3)
	ctx := context.Background()

	// Test with a recent short range (last 7 days)
	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -7)

	resp, err := client.GetTimeSeriesRates(ctx, startDate, endDate, "USD", []string{"EUR", "GBP"})
	if err != nil {
		t.Fatalf("GetTimeSeriesRates() error = %v", err)
	}

	if resp.Base != "USD" {
		t.Errorf("Base = %s, want USD", resp.Base)
	}

	if len(resp.Rates) == 0 {
		t.Error("Expected rates data, got empty map")
	}

	t.Logf("Received %d data points", len(resp.Rates))

	// Check that we have EUR and GBP rates
	for date, rates := range resp.Rates {
		if _, hasEUR := rates["EUR"]; !hasEUR {
			t.Errorf("Missing EUR rate for date %s", date)
		}
		if _, hasGBP := rates["GBP"]; !hasGBP {
			t.Errorf("Missing GBP rate for date %s", date)
		}
	}
}

func TestRealAPI_GetTimeSeriesRates_LongRange(t *testing.T) {
	client := NewClient("https://api.frankfurter.dev/v1", 30*time.Second, 3)
	ctx := context.Background()

	// Test with a long historical range (1 year)
	endDate := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	resp, err := client.GetTimeSeriesRates(ctx, startDate, endDate, "USD", []string{"EUR"})
	if err != nil {
		t.Fatalf("GetTimeSeriesRates() error = %v", err)
	}

	if resp.Base != "USD" {
		t.Errorf("Base = %s, want USD", resp.Base)
	}

	// Should have approximately 365 data points (excluding weekends/holidays)
	if len(resp.Rates) < 250 {
		t.Errorf("Expected at least 250 data points for a full year, got %d", len(resp.Rates))
	}

	t.Logf("Received %d data points for year 2023", len(resp.Rates))
}

func TestRealAPI_GetTimeSeriesRates_Historical1999(t *testing.T) {
	client := NewClient("https://api.frankfurter.dev/v1", 30*time.Second, 3)
	ctx := context.Background()

	// Test with the earliest available date (1999-01-04)
	startDate := time.Date(1999, 1, 4, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(1999, 1, 31, 0, 0, 0, 0, time.UTC)

	resp, err := client.GetTimeSeriesRates(ctx, startDate, endDate, "EUR", []string{"USD"})
	if err != nil {
		t.Fatalf("GetTimeSeriesRates() error = %v", err)
	}

	if resp.Base != "EUR" {
		t.Errorf("Base = %s, want EUR", resp.Base)
	}

	if len(resp.Rates) == 0 {
		t.Error("Expected rates data from 1999, got empty map")
	}

	t.Logf("Successfully fetched historical data from 1999: %d data points", len(resp.Rates))
}
