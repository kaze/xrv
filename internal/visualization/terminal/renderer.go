package terminal

import (
	"fmt"
	"strings"

	"github.com/guptarohit/asciigraph"
	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/statistics"
)

type Renderer struct {
	height int
	width  int
}

func NewRenderer(height, width int) *Renderer {
	if height <= 0 {
		height = 20
	}
	if width <= 0 {
		width = 80
	}
	return &Renderer{
		height: height,
		width:  width,
	}
}

func (r *Renderer) Render(data *domain.TimeSeriesData, stats map[string]statistics.Statistics) error {
	fmt.Println()
	fmt.Printf("📊 %s to %s\n", data.Base, strings.Join(r.currenciesToStrings(data.Targets), ", "))
	fmt.Printf("📅 %s to %s\n", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02"))
	fmt.Println()

	for _, target := range data.Targets {
		rates := r.extractRates(data, target)
		if len(rates) == 0 {
			continue
		}

		fmt.Printf("━━━ %s ━━━\n", target)

		graph := asciigraph.Plot(rates,
			asciigraph.Height(r.height),
			asciigraph.Width(r.width),
			asciigraph.Caption(fmt.Sprintf("%s/%s", data.Base, target)),
		)
		fmt.Println(graph)
		fmt.Println()

		if stat, exists := stats[string(target)]; exists {
			r.displayStats(stat, string(target))
		}
		fmt.Println()
	}

	return nil
}

func (r *Renderer) displayStats(stat statistics.Statistics, currency string) {
	fmt.Println("📈 Statistics:")
	fmt.Printf("  Min:     %.4f\n", stat.Basic.Min)
	fmt.Printf("  Max:     %.4f\n", stat.Basic.Max)
	fmt.Printf("  Average: %.4f\n", stat.Basic.Average)
	fmt.Printf("  Median:  %.4f\n", stat.Basic.Median)
	fmt.Println()
	fmt.Println("📊 Volatility:")
	fmt.Printf("  StdDev:  %.4f\n", stat.Volatility.StdDev)
	fmt.Printf("  Coeff:   %.2f%%\n", stat.Volatility.CoefficientOfVar)
	fmt.Println()
	fmt.Println("📉 Trend:")
	fmt.Printf("  Direction: %s\n", stat.Trend.Direction)
	fmt.Printf("  Change:    %.2f%%\n", stat.Trend.PercentChange)
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

func (r *Renderer) currenciesToStrings(currencies []domain.Currency) []string {
	result := make([]string, len(currencies))
	for i, c := range currencies {
		result[i] = string(c)
	}
	return result
}
