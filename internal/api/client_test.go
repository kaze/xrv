package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://api.example.com", 30*time.Second, 3)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.baseURL != "https://api.example.com" {
		t.Errorf("baseURL = %v, want https://api.example.com", client.baseURL)
	}

	if client.timeout != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", client.timeout)
	}

	if client.retryAttempts != 3 {
		t.Errorf("retryAttempts = %v, want 3", client.retryAttempts)
	}
}

func TestGetTimeSeriesRates_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2023-01-01..2023-01-31" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("from") != "USD" {
			t.Errorf("from = %s, want USD", query.Get("from"))
		}
		if query.Get("to") != "EUR,GBP" {
			t.Errorf("to = %s, want EUR,GBP", query.Get("to"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"amount": 1.0,
			"base": "USD",
			"start_date": "2023-01-01",
			"end_date": "2023-01-31",
			"rates": {
				"2023-01-01": {"EUR": 0.85, "GBP": 0.73},
				"2023-01-02": {"EUR": 0.86, "GBP": 0.74}
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, 1)
	ctx := context.Background()

	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
	targets := []string{"EUR", "GBP"}

	resp, err := client.GetTimeSeriesRates(ctx, startDate, endDate, "USD", targets)

	if err != nil {
		t.Fatalf("GetTimeSeriesRates() error = %v", err)
	}

	if resp.Base != "USD" {
		t.Errorf("Base = %s, want USD", resp.Base)
	}

	if resp.StartDate != "2023-01-01" {
		t.Errorf("StartDate = %s, want 2023-01-01", resp.StartDate)
	}

	if resp.EndDate != "2023-01-31" {
		t.Errorf("EndDate = %s, want 2023-01-31", resp.EndDate)
	}

	if len(resp.Rates) != 2 {
		t.Errorf("Rates length = %d, want 2", len(resp.Rates))
	}

	if resp.Rates["2023-01-01"]["EUR"] != 0.85 {
		t.Errorf("EUR rate = %f, want 0.85", resp.Rates["2023-01-01"]["EUR"])
	}
}

func TestGetTimeSeriesRates_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, 1)
	ctx := context.Background()

	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.GetTimeSeriesRates(ctx, startDate, endDate, "USD", []string{"EUR"})

	if err == nil {
		t.Error("Expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Errorf("Expected APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestGetSupportedCurrencies_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/currencies" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"USD": "United States Dollar",
			"EUR": "Euro",
			"GBP": "British Pound"
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, 1)
	ctx := context.Background()

	resp, err := client.GetSupportedCurrencies(ctx)

	if err != nil {
		t.Fatalf("GetSupportedCurrencies() error = %v", err)
	}

	if len(resp) != 3 {
		t.Errorf("Currencies length = %d, want 3", len(resp))
	}

	if resp["USD"] != "United States Dollar" {
		t.Errorf("USD = %s, want United States Dollar", resp["USD"])
	}

	if resp["EUR"] != "Euro" {
		t.Errorf("EUR = %s, want Euro", resp["EUR"])
	}

	if resp["GBP"] != "British Pound" {
		t.Errorf("GBP = %s, want British Pound", resp["GBP"])
	}
}

func TestGetTimeSeriesRates_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, 1)
	ctx := context.Background()

	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.GetTimeSeriesRates(ctx, startDate, endDate, "USD", []string{"EUR"})

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetTimeSeriesRates_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.GetTimeSeriesRates(ctx, startDate, endDate, "USD", []string{"EUR"})

	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
}
