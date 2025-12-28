package browser

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/statistics"
)

type Renderer struct {
	port int
}

func NewRenderer(port int) *Renderer {
	if port <= 0 {
		port = 8080
	}
	return &Renderer{port: port}
}

func (r *Renderer) Render(data *domain.TimeSeriesData, stats map[string]statistics.Statistics) error {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		r.renderChart(w, data, stats)
	})

	url := fmt.Sprintf("http://localhost:%d", r.port)

	fmt.Printf("\n🌐 Starting browser visualization server...\n")
	fmt.Printf("📊 Opening %s in your browser\n\n", url)
	fmt.Println("Press Ctrl+C to stop the server")

	go r.openBrowser(url)

	return http.ListenAndServe(fmt.Sprintf(":%d", r.port), nil)
}

func (r *Renderer) renderChartOnly(w io.Writer, data *domain.TimeSeriesData, stats map[string]statistics.Statistics) {
	line := r.createChart(data, stats)

	line.Render(w)

	type TemplateData struct {
		Statistics map[string]statistics.Statistics
	}

	templates.ExecuteTemplate(w, "chart", TemplateData{
		Statistics: stats,
	})
}

func (r *Renderer) renderChart(w io.Writer, data *domain.TimeSeriesData, stats map[string]statistics.Statistics) {
	line := r.createChart(data, stats)
	line.Render(w)
}

func (r *Renderer) createChart(data *domain.TimeSeriesData, stats map[string]statistics.Statistics) *charts.Line {
	line := charts.NewLine()

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "600px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    fmt.Sprintf("%s Exchange Rates", data.Base),
			Subtitle: fmt.Sprintf("%s to %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")),
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(true),
			Top:  "10%",
		}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: opts.Bool(true),
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  opts.Bool(true),
					Type:  "png",
					Title: "Save",
				},
				DataZoom: &opts.ToolBoxFeatureDataZoom{
					Show:  opts.Bool(true),
					Title: map[string]string{"zoom": "Zoom", "back": "Reset"},
				},
			},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Date",
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Exchange Rate",
			Type: "value",
		}),
	)

	dates := make([]string, len(data.DataPoints))
	for i, dp := range data.DataPoints {
		dates[i] = dp.Date.Format("2006-01-02")
	}
	line.SetXAxis(dates)

	for _, target := range data.Targets {
		rates := r.extractRates(data, target)
		lineData := make([]opts.LineData, len(rates))
		for i, rate := range rates {
			lineData[i] = opts.LineData{Value: rate}
		}

		statLabel := ""
		if stat, exists := stats[string(target)]; exists {
			statLabel = fmt.Sprintf(" (Avg: %.4f, Trend: %s)", stat.Basic.Average, stat.Trend.Direction)
		}

		line.AddSeries(string(target)+statLabel, lineData).
			SetSeriesOptions(
				charts.WithLineChartOpts(opts.LineChart{
					Smooth: opts.Bool(true),
				}),
				charts.WithLabelOpts(opts.Label{
					Show: opts.Bool(false),
				}),
			)
	}

	return line
}

func (r *Renderer) extractRates(data *domain.TimeSeriesData, currency domain.Currency) []float64 {
	rates := make([]float64, 0, len(data.DataPoints))
	for _, dp := range data.DataPoints {
		if rate, exists := dp.Rates[currency]; exists {
			rates = append(rates, rate)
		}
	}
	return rates
}

func (r *Renderer) openBrowser(url string) {
	time.Sleep(500 * time.Millisecond)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Could not open browser automatically. Please open: %s\n", url)
	}
}
