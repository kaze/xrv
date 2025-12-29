package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kaze/xrv/internal/cache"
	"github.com/kaze/xrv/internal/providers"
	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/service"
	"github.com/kaze/xrv/internal/visualization/browser"
	"github.com/kaze/xrv/internal/visualization/terminal"
)

var (
	vizBase        string
	vizCurrencies  string
	vizFrom        string
	vizTo          string
	vizNoCache     bool
	vizHeight      int
	vizWidth       int
	vizOutput      string
	vizPort        int
	vizInvert      bool
	vizInteractive bool
)

func NewVisualizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "visualize",
		Aliases: []string{"viz"},
		Short:   "Visualize exchange rate data",
		Long:    "Visualize historical exchange rate data with charts and statistics",
		RunE:    runVisualize,
	}

	cmd.Flags().StringVarP(&vizBase, "base", "b", "", "Base currency (optional in interactive mode)")
	cmd.Flags().StringVarP(&vizCurrencies, "currencies", "c", "", "Target currencies (optional in interactive mode)")
	cmd.Flags().StringVarP(&vizFrom, "from", "f", "", "Start date (YYYY-MM-DD) or relative (e.g., '1 year ago')")
	cmd.Flags().StringVarP(&vizTo, "to", "t", "", "End date (YYYY-MM-DD), defaults to today")
	cmd.Flags().StringVarP(&vizOutput, "output", "o", "terminal", "Output mode: terminal, browser")
	cmd.Flags().IntVar(&vizPort, "port", 8080, "Port for browser mode (default: 8080)")
	cmd.Flags().BoolVarP(&vizInteractive, "interactive", "i", false, "Interactive mode (browser with form)")
	cmd.Flags().BoolVar(&vizInvert, "invert", false, "Invert rates (show base in target currency)")
	cmd.Flags().BoolVar(&vizNoCache, "no-cache", false, "Disable caching")
	cmd.Flags().IntVar(&vizHeight, "height", 15, "Chart height (terminal mode)")
	cmd.Flags().IntVar(&vizWidth, "width", 80, "Chart width (terminal mode)")

	return cmd
}

func runVisualize(cmd *cobra.Command, args []string) error {
	apiClient := providers.NewFrankfurterClient("https://api.frankfurter.dev/v1", 30*time.Second, 3)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	cacheDir := filepath.Join(homeDir, ".xrv", "cache")
	badgerCache, err := cache.NewBadgerCache(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}
	defer badgerCache.Close()

	svc := service.NewService(apiClient, badgerCache)

	if vizInteractive || (vizOutput == "browser" && vizBase == "" && vizCurrencies == "") {
		server := browser.NewServer(vizPort, svc, apiClient)
		return server.Start()
	}

	if vizBase == "" {
		vizBase = "USD"
	}
	if vizCurrencies == "" {
		vizCurrencies = "EUR,GBP,JPY"
	}

	ctx := context.Background()

	endDate := time.Now()
	if vizTo != "" {
		var err error
		endDate, err = time.Parse("2006-01-02", vizTo)
		if err != nil {
			return fmt.Errorf("invalid end date format: %w", err)
		}
	}

	startDate := endDate.AddDate(-1, 0, 0) // Default to 1 year ago
	if vizFrom != "" {
		var err error
		startDate, err = time.Parse("2006-01-02", vizFrom)
		if err != nil {
			startDate, err = parseRelativeDate(vizFrom, endDate)
			if err != nil {
				return fmt.Errorf("invalid start date format: %w", err)
			}
		}
	}

	targets := strings.Split(vizCurrencies, ",")
	targetCurrencies := make([]domain.Currency, len(targets))
	for i, t := range targets {
		targetCurrencies[i] = domain.Currency(strings.TrimSpace(t))
	}

	fmt.Println("Fetching exchange rate data...")
	data, err := svc.FetchTimeSeriesData(ctx, service.FetchOptions{
		Base:      domain.Currency(vizBase),
		Targets:   targetCurrencies,
		StartDate: startDate,
		EndDate:   endDate,
		UseCache:  !vizNoCache,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	if vizInvert {
		data = invertRates(data)
	}

	stats := svc.CalculateStatistics(data)

	switch strings.ToLower(vizOutput) {
	case "browser":
		renderer := browser.NewRenderer(vizPort)
		return renderer.Render(data, stats)
	case "terminal":
		renderer := terminal.NewRenderer(vizHeight, vizWidth)
		return renderer.Render(data, stats)
	default:
		return fmt.Errorf("unsupported output mode: %s (use 'terminal' or 'browser')", vizOutput)
	}
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

func parseRelativeDate(relative string, base time.Time) (time.Time, error) {
	relative = strings.ToLower(strings.TrimSpace(relative))

	if strings.HasSuffix(relative, "year ago") || strings.HasSuffix(relative, "years ago") {
		var years int
		fmt.Sscanf(relative, "%d", &years)
		return base.AddDate(-years, 0, 0), nil
	}
	if strings.HasSuffix(relative, "month ago") || strings.HasSuffix(relative, "months ago") {
		var months int
		fmt.Sscanf(relative, "%d", &months)
		return base.AddDate(0, -months, 0), nil
	}
	if strings.HasSuffix(relative, "day ago") || strings.HasSuffix(relative, "days ago") {
		var days int
		fmt.Sscanf(relative, "%d", &days)
		return base.AddDate(0, 0, -days), nil
	}

	return base, fmt.Errorf("unsupported relative date format: %s", relative)
}
