package browser

import (
	"fmt"

	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/statistics"
)

type EChartsTitle struct {
	Text    string `json:"text"`
	Subtext string `json:"subtext"`
}

type EChartsTooltip struct {
	Trigger string `json:"trigger"`
	Show    bool   `json:"show"`
}

type EChartsLegend struct {
	Data []string `json:"data"`
	Show bool     `json:"show"`
	Top  string   `json:"top"`
}

type EChartsXAxis struct {
	Type string   `json:"type"`
	Name string   `json:"name"`
	Data []string `json:"data"`
}

type EChartsYAxis struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type EChartsSeries struct {
	Name   string    `json:"name"`
	Type   string    `json:"type"`
	Data   []float64 `json:"data"`
	Smooth bool      `json:"smooth"`
}

type EChartsToolboxFeatureSaveAsImage struct {
	Show  bool   `json:"show"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type EChartsToolboxFeatureDataZoom struct {
	Show  bool              `json:"show"`
	Title map[string]string `json:"title"`
}

type EChartsToolboxFeature struct {
	SaveAsImage *EChartsToolboxFeatureSaveAsImage `json:"saveAsImage"`
	DataZoom    *EChartsToolboxFeatureDataZoom    `json:"dataZoom"`
}

type EChartsToolbox struct {
	Show    bool                   `json:"show"`
	Feature *EChartsToolboxFeature `json:"feature"`
}

type EChartsDataZoom struct {
	Type  string `json:"type"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type EChartsConfig struct {
	Title    EChartsTitle      `json:"title"`
	Tooltip  EChartsTooltip    `json:"tooltip"`
	Legend   EChartsLegend     `json:"legend"`
	XAxis    EChartsXAxis      `json:"xAxis"`
	YAxis    EChartsYAxis      `json:"yAxis"`
	Series   []EChartsSeries   `json:"series"`
	Toolbox  EChartsToolbox    `json:"toolbox"`
	DataZoom []EChartsDataZoom `json:"dataZoom"`
}

func TransformToEChartsConfig(data *domain.TimeSeriesData, stats map[string]statistics.Statistics) (*EChartsConfig, error) {
	if data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}

	dates := make([]string, len(data.DataPoints))
	for i, dp := range data.DataPoints {
		dates[i] = dp.Date.Format("2006-01-02")
	}

	legendData := make([]string, 0, len(data.Targets))
	series := make([]EChartsSeries, 0, len(data.Targets))

	for _, target := range data.Targets {
		rates := make([]float64, len(data.DataPoints))
		for i, dp := range data.DataPoints {
			if rate, exists := dp.Rates[target]; exists {
				rates[i] = rate
			}
		}

		seriesName := string(target)
		if stat, exists := stats[string(target)]; exists {
			seriesName = fmt.Sprintf("%s (Avg: %.4f, Trend: %s)", 
				target, stat.Basic.Average, stat.Trend.Direction)
		}

		legendData = append(legendData, seriesName)

		series = append(series, EChartsSeries{
			Name:   seriesName,
			Type:   "line",
			Data:   rates,
			Smooth: true,
		})
	}

	config := &EChartsConfig{
		Title: EChartsTitle{
			Text:    fmt.Sprintf("%s Exchange Rates", data.Base),
			Subtext: fmt.Sprintf("%s to %s", 
				data.StartDate.Format("2006-01-02"), 
				data.EndDate.Format("2006-01-02")),
		},
		Tooltip: EChartsTooltip{
			Trigger: "axis",
			Show:    true,
		},
		Legend: EChartsLegend{
			Data: legendData,
			Show: true,
			Top:  "10%",
		},
		XAxis: EChartsXAxis{
			Type: "category",
			Name: "Date",
			Data: dates,
		},
		YAxis: EChartsYAxis{
			Type: "value",
			Name: "Exchange Rate",
		},
		Series: series,
		Toolbox: EChartsToolbox{
			Show: true,
			Feature: &EChartsToolboxFeature{
				SaveAsImage: &EChartsToolboxFeatureSaveAsImage{
					Show:  true,
					Type:  "png",
					Title: "Save",
				},
				DataZoom: &EChartsToolboxFeatureDataZoom{
					Show: true,
					Title: map[string]string{
						"zoom": "Zoom",
						"back": "Reset",
					},
				},
			},
		},
		DataZoom: []EChartsDataZoom{
			{
				Type:  "slider",
				Start: 0,
				End:   100,
			},
		},
	}

	return config, nil
}
