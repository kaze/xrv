package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kaze/xrv/internal/api"
	"github.com/kaze/xrv/internal/cache"
	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/service"
)

type mockAPIClient struct {
	timeSeriesResponse *api.TimeSeriesResponse
	currenciesResponse api.CurrenciesResponse
}

func (m *mockAPIClient) GetTimeSeriesRates(ctx context.Context, startDate, endDate time.Time, base string, targets []string) (*api.TimeSeriesResponse, error) {
	return m.timeSeriesResponse, nil
}

func (m *mockAPIClient) GetSupportedCurrencies(ctx context.Context) (api.CurrenciesResponse, error) {
	return m.currenciesResponse, nil
}

func TestServer_HandleIndex(t *testing.T) {
	mockAPI := &mockAPIClient{}
	memCache := cache.NewMemoryCache()
	defer memCache.Close()
	
	svc := service.NewService(mockAPI, memCache)

	server := NewServer(8080, svc, mockAPI)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Content-Type = %s, want text/html", contentType)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}

	if !contains(body, "XRV") {
		t.Error("Expected 'XRV' in response")
	}
	if !contains(body, "vizForm") {
		t.Error("Expected form with id 'vizForm' in response")
	}
}

func TestServer_HandleVisualize(t *testing.T) {
	mockAPI := &mockAPIClient{
		timeSeriesResponse: &api.TimeSeriesResponse{
			Base:      "USD",
			StartDate: "2023-01-01",
			EndDate:   "2023-01-31",
			Rates: map[string]map[string]float64{
				"2023-01-01": {"EUR": 0.85},
				"2023-01-02": {"EUR": 0.86},
			},
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()
	
	svc := service.NewService(mockAPI, memCache)

	server := NewServer(8080, svc, mockAPI)

	req := httptest.NewRequest("GET", "/visualize?base=USD&currencies=EUR&from=2023-01-01&to=2023-01-31&invert=false", nil)
	w := httptest.NewRecorder()

	server.handleVisualize(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}

	if !contains(body, "echarts") {
		t.Error("Expected echarts library in response")
	}
	if !contains(body, "initializeChart") {
		t.Error("Expected chart initialization in response")
	}
	if !contains(body, "stats-grid") {
		t.Error("Expected statistics section in response")
	}
}

func TestServer_HandleVisualize_WithInvert(t *testing.T) {
	mockAPI := &mockAPIClient{
		timeSeriesResponse: &api.TimeSeriesResponse{
			Base:      "USD",
			StartDate: "2023-01-01",
			EndDate:   "2023-01-02",
			Rates: map[string]map[string]float64{
				"2023-01-01": {"EUR": 0.5},
			},
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()
	
	svc := service.NewService(mockAPI, memCache)

	server := NewServer(8080, svc, mockAPI)

	req := httptest.NewRequest("GET", "/visualize?base=USD&currencies=EUR&from=2023-01-01&to=2023-01-02&invert=true", nil)
	w := httptest.NewRecorder()

	server.handleVisualize(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body := w.Body.String()
	if !contains(body, "echarts") {
		t.Error("Expected chart in response for inverted rates")
	}
}

func TestServer_HandleCurrencies(t *testing.T) {
	mockAPI := &mockAPIClient{
		currenciesResponse: api.CurrenciesResponse{
			"USD": "United States Dollar",
			"EUR": "Euro",
		},
	}

	memCache := cache.NewMemoryCache()
	defer memCache.Close()
	
	svc := service.NewService(mockAPI, memCache)

	server := NewServer(8080, svc, mockAPI)

	req := httptest.NewRequest("GET", "/currencies", nil)
	w := httptest.NewRecorder()

	server.handleCurrencies(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", contentType)
	}

	body := w.Body.String()
	if !contains(body, "USD") || !contains(body, "EUR") {
		t.Error("Expected currency codes in response")
	}
}

func TestInvertRates(t *testing.T) {
	server := &Server{}

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

	inverted := server.invertRates(data)

	if len(inverted.DataPoints) != 1 {
		t.Fatalf("Expected 1 data point, got %d", len(inverted.DataPoints))
	}

	invertedRate := inverted.DataPoints[0].Rates["EUR"]
	expected := 2.0
	if invertedRate != expected {
		t.Errorf("Inverted rate = %f, want %f", invertedRate, expected)
	}
}

func TestTemplates_Loaded(t *testing.T) {
	if templates == nil {
		t.Fatal("Templates not loaded")
	}

	indexTmpl := templates.Lookup("index.html")
	if indexTmpl == nil {
		t.Error("index.html template not found")
	}

	chartTmpl := templates.Lookup("chart.html")
	if chartTmpl == nil {
		t.Error("chart.html template not found")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
		(s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
