package statistics

import (
	"testing"
)

func TestCalculator_Calculate(t *testing.T) {
	

	rates := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	stats := Calculate(rates)

	if stats.Basic.Min != 1.0 {
		t.Errorf("Basic.Min = %v, want 1.0", stats.Basic.Min)
	}
	if stats.Basic.Max != 5.0 {
		t.Errorf("Basic.Max = %v, want 5.0", stats.Basic.Max)
	}

	if stats.Volatility.StdDev == 0 {
		t.Error("Expected non-zero standard deviation")
	}

	if stats.Trend.Direction != "upward" {
		t.Errorf("Trend.Direction = %s, want upward", stats.Trend.Direction)
	}
}

func TestCalculateTrend_Direction(t *testing.T) {
	tests := []struct {
		name  string
		rates []float64
		want  string
	}{
		{"upward", []float64{1.0, 2.0, 3.0, 4.0, 5.0}, "upward"},
		{"downward", []float64{5.0, 4.0, 3.0, 2.0, 1.0}, "downward"},
		{"flat", []float64{3.0, 3.0, 3.0, 3.0, 3.0}, "flat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend := CalculateTrend(tt.rates)
			if trend.Direction != tt.want {
				t.Errorf("Direction = %s, want %s", trend.Direction, tt.want)
			}
		})
	}
}

func TestCalculateSMA(t *testing.T) {
	rates := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	sma := calculateSMA(rates, 3)

	expected := []float64{2.0, 3.0, 4.0}

	if len(sma) != len(expected) {
		t.Fatalf("SMA length = %d, want %d", len(sma), len(expected))
	}

	for i, v := range expected {
		if !floatsAlmostEqual(sma[i], v) {
			t.Errorf("SMA[%d] = %v, want %v", i, sma[i], v)
		}
	}
}
