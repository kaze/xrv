package browser

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"time"

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
	config, err := TransformToEChartsConfig(data, stats)
	if err != nil {
		fmt.Fprintf(w, "Error transforming data: %v", err)
		return
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		fmt.Fprintf(w, "Error marshaling config: %v", err)
		return
	}

	type TemplateData struct {
		ChartConfigJSON template.JS
		Statistics      map[string]statistics.Statistics
	}

	templates.ExecuteTemplate(w, "chart-page", TemplateData{
		ChartConfigJSON: template.JS(configJSON),
		Statistics:      stats,
	})
}

func (r *Renderer) renderChart(w io.Writer, data *domain.TimeSeriesData, stats map[string]statistics.Statistics) {
	r.renderChartOnly(w, data, stats)
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
