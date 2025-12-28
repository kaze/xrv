package browser

import (
	"context"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/kaze/xrv/internal/api"
	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/service"
)

var templates *template.Template

func init() {
	var err error
	templates, err = GetTemplates()
	if err != nil {
		panic(fmt.Sprintf("Failed to load templates: %v", err))
	}
}

type Server struct {
	port      int
	svc       *service.Service
	apiClient api.APIClient
}

func NewServer(port int, svc *service.Service, apiClient api.APIClient) *Server {
	if port <= 0 {
		port = 8080
	}
	return &Server{
		port:      port,
		svc:       svc,
		apiClient: apiClient,
	}
}

func (s *Server) Start() error {
	assetFS, err := LoadAssets()
	if err != nil {
		return fmt.Errorf("failed to load assets: %w", err)
	}

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(assetFS)))
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/visualize", s.handleVisualize)
	http.HandleFunc("/currencies", s.handleCurrencies)

	url := fmt.Sprintf("http://localhost:%d", s.port)
	fmt.Printf("\nStarting interactive browser visualization server...\n")
	fmt.Printf("Opening %s in your browser\n\n", url)
	fmt.Println("Press Ctrl+C to stop the server")

	go s.openBrowser(url)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "base.html", nil)
}

func (s *Server) handleVisualize(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	if base == "" {
		base = "USD"
	}

	currenciesStr := r.URL.Query().Get("currencies")
	if currenciesStr == "" {
		currenciesStr = "EUR,GBP,JPY"
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	invert := r.URL.Query().Get("invert") == "true"

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
	targetCurrencies := make([]domain.Currency, len(targets))
	for i, t := range targets {
		targetCurrencies[i] = domain.Currency(strings.TrimSpace(t))
	}

	ctx := context.Background()
	data, err := s.svc.FetchTimeSeriesData(ctx, service.FetchOptions{
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
		data = s.invertRates(data)
	}

	stats := s.svc.CalculateStatistics(data)

	renderer := NewRenderer(s.port)
	renderer.renderChartOnly(w, data, stats)
}

func (s *Server) handleCurrencies(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	currencies, err := s.apiClient.GetSupportedCurrencies(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch currencies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{")
	first := true
	for code, name := range currencies {
		if !first {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, `"%s":"%s"`, code, name)
		first = false
	}
	fmt.Fprintf(w, "}")
}

func (s *Server) invertRates(data *domain.TimeSeriesData) *domain.TimeSeriesData {
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
				inverted := 1.0 / rate
				invertedRates[currency] = math.Round(inverted*100) / 100
			}
		}
		inverted.DataPoints[i] = domain.DataPoint{
			Date:  dp.Date,
			Rates: invertedRates,
		}
	}

	return inverted
}

func (s *Server) openBrowser(url string) {
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
