package browser

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kaze/xrv/internal/api"
	"github.com/kaze/xrv/internal/cache"
	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/service"
)

func TestHandleChartUpdate(t *testing.T) {
	mockAPI := &mockAPIClient{
		timeSeriesResponse: &api.TimeSeriesResponse{
			Base:      "USD",
			StartDate: "2024-01-01",
			EndDate:   "2024-01-31",
			Rates: map[string]map[string]float64{
				"2024-01-01": {"EUR": 0.85, "GBP": 0.75},
				"2024-01-02": {"EUR": 0.86, "GBP": 0.76},
			},
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	svc := service.NewService(mockAPI, memCache)
	handlers := NewHandlers(svc)

	form := url.Values{}
	form.Set("base", "USD")
	form.Set("currencies", "EUR,GBP")
	form.Set("from", "2024-01-01")
	form.Set("to", "2024-01-31")
	form.Set("invert", "off")

	req := httptest.NewRequest(http.MethodPost, "/htmx/chart", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handlers.HandleChartUpdate(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}

	if !contains(body, "chartCanvas") {
		t.Error("Expected chartCanvas div in response")
	}

	if !contains(body, "initializeChart") {
		t.Error("Expected initializeChart call in response")
	}

	if !contains(body, "stats-grid") {
		t.Error("Expected statistics section in response")
	}
}

func TestHandleChartUpdate_WithInvert(t *testing.T) {
	mockAPI := &mockAPIClient{
		timeSeriesResponse: &api.TimeSeriesResponse{
			Base:      "USD",
			StartDate: "2024-01-01",
			EndDate:   "2024-01-02",
			Rates: map[string]map[string]float64{
				"2024-01-01": {"EUR": 0.5},
			},
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	svc := service.NewService(mockAPI, memCache)
	handlers := NewHandlers(svc)

	form := url.Values{}
	form.Set("base", "USD")
	form.Set("currencies", "EUR")
	form.Set("from", "2024-01-01")
	form.Set("to", "2024-01-02")
	form.Set("invert", "on")

	req := httptest.NewRequest(http.MethodPost, "/htmx/chart", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handlers.HandleChartUpdate(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body := w.Body.String()
	if !contains(body, "chartCanvas") {
		t.Error("Expected chart in response for inverted rates")
	}
}

func TestHandleChartUpdate_InvalidMethod(t *testing.T) {
	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	svc := service.NewService(&mockAPIClient{}, memCache)
	handlers := NewHandlers(svc)

	req := httptest.NewRequest(http.MethodGet, "/htmx/chart", nil)
	w := httptest.NewRecorder()

	handlers.HandleChartUpdate(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestHandleChartUpdate_MissingCurrencies(t *testing.T) {
	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	svc := service.NewService(&mockAPIClient{}, memCache)
	handlers := NewHandlers(svc)

	form := url.Values{}
	form.Set("base", "USD")
	form.Set("from", "2024-01-01")
	form.Set("to", "2024-01-31")

	req := httptest.NewRequest(http.MethodPost, "/htmx/chart", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handlers.HandleChartUpdate(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleStatisticsRefresh(t *testing.T) {
	mockAPI := &mockAPIClient{
		timeSeriesResponse: &api.TimeSeriesResponse{
			Base:      "EUR",
			StartDate: "2024-01-01",
			EndDate:   "2024-01-31",
			Rates: map[string]map[string]float64{
				"2024-01-01": {"USD": 1.10, "GBP": 0.85},
				"2024-01-02": {"USD": 1.11, "GBP": 0.86},
			},
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	svc := service.NewService(mockAPI, memCache)
	handlers := NewHandlers(svc)

	req := httptest.NewRequest(http.MethodGet, "/htmx/statistics?base=EUR&currencies=USD,GBP&from=2024-01-01&to=2024-01-31", nil)
	w := httptest.NewRecorder()

	handlers.HandleStatisticsRefresh(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}

	if !contains(body, "stats-grid") {
		t.Error("Expected stats-grid in response")
	}

	if !contains(body, "USD") {
		t.Error("Expected USD in statistics")
	}

	if !contains(body, "GBP") {
		t.Error("Expected GBP in statistics")
	}
}

func TestInvertRates_Function(t *testing.T) {
	data := &domain.TimeSeriesData{
		Base:    "USD",
		Targets: []domain.Currency{"EUR"},
		DataPoints: []domain.DataPoint{
			{
				Date: time.Now(),
				Rates: map[domain.Currency]float64{
					"EUR": 0.5,
				},
			},
		},
	}

	inverted := invertRates(data)

	if len(inverted.DataPoints) != 1 {
		t.Fatalf("Expected 1 data point, got %d", len(inverted.DataPoints))
	}

	invertedRate := inverted.DataPoints[0].Rates["EUR"]
	expected := 2.0
	if invertedRate != expected {
		t.Errorf("Inverted rate = %f, want %f", invertedRate, expected)
	}
}

func TestHandleExportCSV(t *testing.T) {
	mockAPI := &mockAPIClient{
		timeSeriesResponse: &api.TimeSeriesResponse{
			Base:      "USD",
			StartDate: "2024-01-01",
			EndDate:   "2024-01-02",
			Rates: map[string]map[string]float64{
				"2024-01-01": {"EUR": 0.85, "GBP": 0.75},
				"2024-01-02": {"EUR": 0.86, "GBP": 0.76},
			},
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	svc := service.NewService(mockAPI, memCache)
	handlers := NewHandlers(svc)

	req := httptest.NewRequest(http.MethodGet, "/export/csv?base=USD&currencies=EUR,GBP&from=2024-01-01&to=2024-01-02", nil)
	w := httptest.NewRecorder()

	handlers.HandleExportCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/csv" {
		t.Errorf("Content-Type = %s, want text/csv", contentType)
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if !contains(contentDisposition, "attachment") || !contains(contentDisposition, ".csv") {
		t.Errorf("Content-Disposition = %s, want attachment with .csv filename", contentDisposition)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty CSV data")
	}

	if !contains(body, "Date") || !contains(body, "EUR") || !contains(body, "GBP") {
		t.Error("Expected CSV header with Date and currency columns")
	}

	if !contains(body, "2024-01-01") {
		t.Error("Expected date in CSV data")
	}
}

func TestHandleExportJSON(t *testing.T) {
	mockAPI := &mockAPIClient{
		timeSeriesResponse: &api.TimeSeriesResponse{
			Base:      "USD",
			StartDate: "2024-01-01",
			EndDate:   "2024-01-02",
			Rates: map[string]map[string]float64{
				"2024-01-01": {"EUR": 0.85},
				"2024-01-02": {"EUR": 0.86},
			},
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	svc := service.NewService(mockAPI, memCache)
	handlers := NewHandlers(svc)

	req := httptest.NewRequest(http.MethodGet, "/export/json?base=USD&currencies=EUR&from=2024-01-01&to=2024-01-02", nil)
	w := httptest.NewRecorder()

	handlers.HandleExportJSON(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", contentType)
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if !contains(contentDisposition, "attachment") || !contains(contentDisposition, ".json") {
		t.Errorf("Content-Disposition = %s, want attachment with .json filename", contentDisposition)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty JSON data")
	}

	if !contains(body, "base") || !contains(body, "data") || !contains(body, "statistics") {
		t.Error("Expected JSON with base, data, and statistics fields")
	}
}

func TestHandleExportCSV_MissingParams(t *testing.T) {
	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	svc := service.NewService(&mockAPIClient{}, memCache)
	handlers := NewHandlers(svc)

	req := httptest.NewRequest(http.MethodGet, "/export/csv", nil)
	w := httptest.NewRecorder()

	handlers.HandleExportCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}
