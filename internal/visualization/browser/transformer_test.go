package browser

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/statistics"
)

func TestTransformToEChartsConfig(t *testing.T) {
	tests := []struct {
		name          string
		data          *domain.TimeSeriesData
		stats         map[string]statistics.Statistics
		wantTitle     string
		wantSubtitle  string
		wantDatesLen  int
		wantSeriesLen int
	}{
		{
			name: "single currency",
			data: &domain.TimeSeriesData{
				Base:      "USD",
				Targets:   []domain.Currency{"EUR"},
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
				DataPoints: []domain.DataPoint{
					{
						Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						Rates: map[domain.Currency]float64{
							"EUR": 0.85,
						},
					},
					{
						Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						Rates: map[domain.Currency]float64{
							"EUR": 0.86,
						},
					},
					{
						Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
						Rates: map[domain.Currency]float64{
							"EUR": 0.87,
						},
					},
				},
			},
			stats: map[string]statistics.Statistics{
				"EUR": {
					Basic: statistics.BasicStats{
						Min:     0.85,
						Max:     0.87,
						Average: 0.86,
						Median:  0.86,
					},
					Volatility: statistics.VolatilityStats{
						StdDev:           0.01,
						Variance:         0.0001,
						CoefficientOfVar: 1.16,
						AvgDailyReturn:   0.58,
					},
					Trend: statistics.TrendStats{
						Direction:     "Upward",
						Slope:         0.01,
						PercentChange: 2.35,
					},
				},
			},
			wantTitle:     "USD Exchange Rates",
			wantSubtitle:  "2024-01-01 to 2024-01-03",
			wantDatesLen:  3,
			wantSeriesLen: 1,
		},
		{
			name: "multiple currencies",
			data: &domain.TimeSeriesData{
				Base:      "EUR",
				Targets:   []domain.Currency{"USD", "GBP", "JPY"},
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				DataPoints: []domain.DataPoint{
					{
						Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						Rates: map[domain.Currency]float64{
							"USD": 1.10,
							"GBP": 0.85,
							"JPY": 130.0,
						},
					},
					{
						Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						Rates: map[domain.Currency]float64{
							"USD": 1.11,
							"GBP": 0.86,
							"JPY": 131.0,
						},
					},
				},
			},
			stats: map[string]statistics.Statistics{
				"USD": {
					Basic: statistics.BasicStats{Average: 1.105},
					Trend: statistics.TrendStats{Direction: "Upward"},
				},
				"GBP": {
					Basic: statistics.BasicStats{Average: 0.855},
					Trend: statistics.TrendStats{Direction: "Upward"},
				},
				"JPY": {
					Basic: statistics.BasicStats{Average: 130.5},
					Trend: statistics.TrendStats{Direction: "Upward"},
				},
			},
			wantTitle:     "EUR Exchange Rates",
			wantSubtitle:  "2024-01-01 to 2024-01-02",
			wantDatesLen:  2,
			wantSeriesLen: 3,
		},
		{
			name: "empty data points",
			data: &domain.TimeSeriesData{
				Base:       "USD",
				Targets:    []domain.Currency{"EUR"},
				StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				DataPoints: []domain.DataPoint{},
			},
			stats:         map[string]statistics.Statistics{},
			wantTitle:     "USD Exchange Rates",
			wantSubtitle:  "2024-01-01 to 2024-01-01",
			wantDatesLen:  0,
			wantSeriesLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := TransformToEChartsConfig(tt.data, tt.stats)
			if err != nil {
				t.Fatalf("TransformToEChartsConfig() error = %v", err)
			}

			if config.Title.Text != tt.wantTitle {
				t.Errorf("Title.Text = %s, want %s", config.Title.Text, tt.wantTitle)
			}

			if config.Title.Subtext != tt.wantSubtitle {
				t.Errorf("Title.Subtext = %s, want %s", config.Title.Subtext, tt.wantSubtitle)
			}

			if len(config.XAxis.Data) != tt.wantDatesLen {
				t.Errorf("XAxis.Data length = %d, want %d", len(config.XAxis.Data), tt.wantDatesLen)
			}

			if len(config.Series) != tt.wantSeriesLen {
				t.Errorf("Series length = %d, want %d", len(config.Series), tt.wantSeriesLen)
			}

			for i, series := range config.Series {
				if len(series.Data) != tt.wantDatesLen {
					t.Errorf("Series[%d].Data length = %d, want %d", i, len(series.Data), tt.wantDatesLen)
				}

				if series.Type != "line" {
					t.Errorf("Series[%d].Type = %s, want line", i, series.Type)
				}

				if !series.Smooth {
					t.Errorf("Series[%d].Smooth = false, want true", i)
				}
			}

			jsonBytes, err := json.Marshal(config)
			if err != nil {
				t.Errorf("Failed to marshal config to JSON: %v", err)
			}

			if len(jsonBytes) == 0 {
				t.Error("Marshaled JSON is empty")
			}
		})
	}
}

func TestTransformToEChartsConfig_SeriesNames(t *testing.T) {
	data := &domain.TimeSeriesData{
		Base:    "USD",
		Targets: []domain.Currency{"EUR", "GBP"},
		DataPoints: []domain.DataPoint{
			{
				Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Rates: map[domain.Currency]float64{
					"EUR": 0.85,
					"GBP": 0.75,
				},
			},
		},
	}

	stats := map[string]statistics.Statistics{
		"EUR": {
			Basic: statistics.BasicStats{Average: 0.85},
			Trend: statistics.TrendStats{Direction: "Stable"},
		},
		"GBP": {
			Basic: statistics.BasicStats{Average: 0.75},
			Trend: statistics.TrendStats{Direction: "Downward"},
		},
	}

	config, err := TransformToEChartsConfig(data, stats)
	if err != nil {
		t.Fatalf("TransformToEChartsConfig() error = %v", err)
	}

	eurFound := false
	gbpFound := false

	for _, series := range config.Series {
		if contains(series.Name, "EUR") && contains(series.Name, "0.8500") && contains(series.Name, "Stable") {
			eurFound = true
		}
		if contains(series.Name, "GBP") && contains(series.Name, "0.7500") && contains(series.Name, "Downward") {
			gbpFound = true
		}
	}

	if !eurFound {
		t.Error("EUR series name should contain currency code, average, and trend")
	}

	if !gbpFound {
		t.Error("GBP series name should contain currency code, average, and trend")
	}
}
