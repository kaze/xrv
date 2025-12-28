package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/service"
	"github.com/kaze/xrv/internal/statistics"
)

type Handlers struct {
	svc *service.Service
}

func NewHandlers(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) HandleChartUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	base := r.FormValue("base")
	if base == "" {
		base = "USD"
	}

	currenciesStr := r.FormValue("currencies")
	if currenciesStr == "" {
		http.Error(w, "Currencies required", http.StatusBadRequest)
		return
	}

	from := r.FormValue("from")
	to := r.FormValue("to")
	invert := r.FormValue("invert") == "on"

	startDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	targets := strings.Split(currenciesStr, ",")
	targetCurrencies := make([]domain.Currency, 0, len(targets))
	for _, t := range targets {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			targetCurrencies = append(targetCurrencies, domain.Currency(trimmed))
		}
	}

	if len(targetCurrencies) == 0 {
		http.Error(w, "At least one currency required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	data, err := h.svc.FetchTimeSeriesData(ctx, service.FetchOptions{
		Base:      domain.Currency(base),
		Targets:   targetCurrencies,
		StartDate: startDate,
		EndDate:   endDate,
		UseCache:  true,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch data: %v", err), http.StatusInternalServerError)
		return
	}

	if invert {
		data = invertRates(data)
	}

	stats := h.svc.CalculateStatistics(data)

	config, err := TransformToEChartsConfig(data, stats)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to transform data: %v", err), http.StatusInternalServerError)
		return
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal config: %v", err), http.StatusInternalServerError)
		return
	}

	type TemplateData struct {
		ChartConfigJSON template.JS
		Statistics      map[string]statistics.Statistics
	}

	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "chart", TemplateData{
		ChartConfigJSON: template.JS(configJSON),
		Statistics:      stats,
	})
}

func (h *Handlers) HandleStatisticsRefresh(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	if base == "" {
		base = "USD"
	}

	currenciesStr := r.URL.Query().Get("currencies")
	if currenciesStr == "" {
		http.Error(w, "Currencies required", http.StatusBadRequest)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	startDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	targets := strings.Split(currenciesStr, ",")
	targetCurrencies := make([]domain.Currency, 0, len(targets))
	for _, t := range targets {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			targetCurrencies = append(targetCurrencies, domain.Currency(trimmed))
		}
	}

	ctx := context.Background()
	data, err := h.svc.FetchTimeSeriesData(ctx, service.FetchOptions{
		Base:      domain.Currency(base),
		Targets:   targetCurrencies,
		StartDate: startDate,
		EndDate:   endDate,
		UseCache:  true,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch data: %v", err), http.StatusInternalServerError)
		return
	}

	stats := h.svc.CalculateStatistics(data)

	type TemplateData struct {
		Statistics map[string]statistics.Statistics
	}

	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "statistics", TemplateData{
		Statistics: stats,
	})
}

func invertRates(data *domain.TimeSeriesData) *domain.TimeSeriesData {
	inverted := &domain.TimeSeriesData{
		Base:       data.Base,
		Targets:    data.Targets,
		StartDate:  data.StartDate,
		EndDate:    data.EndDate,
		DataPoints: make([]domain.DataPoint, len(data.DataPoints)),
	}

	for i, dp := range data.DataPoints {
		invertedRates := make(map[domain.Currency]float64)
		for currency, rate := range dp.Rates {
			if rate != 0 {
				invertedRates[currency] = 1.0 / rate
			}
		}
		inverted.DataPoints[i] = domain.DataPoint{
			Date:  dp.Date,
			Rates: invertedRates,
		}
	}

	return inverted
}
