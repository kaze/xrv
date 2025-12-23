package statistics

import (
	"testing"
)

func TestCalculateBasic(t *testing.T) {
	tests := []struct {
		name  string
		rates []float64
		want  BasicStats
	}{
		{
			name:  "simple values",
			rates: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			want: BasicStats{
				Min:     1.0,
				Max:     5.0,
				Average: 3.0,
				Median:  3.0,
			},
		},
		{
			name:  "even number of values",
			rates: []float64{1.0, 2.0, 3.0, 4.0},
			want: BasicStats{
				Min:     1.0,
				Max:     4.0,
				Average: 2.5,
				Median:  2.5,
			},
		},
		{
			name:  "single value",
			rates: []float64{5.0},
			want: BasicStats{
				Min:     5.0,
				Max:     5.0,
				Average: 5.0,
				Median:  5.0,
			},
		},
		{
			name:  "negative values",
			rates: []float64{-2.0, -1.0, 0.0, 1.0, 2.0},
			want: BasicStats{
				Min:     -2.0,
				Max:     2.0,
				Average: 0.0,
				Median:  0.0,
			},
		},
		{
			name:  "decimal values",
			rates: []float64{0.85, 0.86, 0.87, 0.88, 0.89},
			want: BasicStats{
				Min:     0.85,
				Max:     0.89,
				Average: 0.87,
				Median:  0.87,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateBasic(tt.rates)

			if got.Min != tt.want.Min {
				t.Errorf("Min = %v, want %v", got.Min, tt.want.Min)
			}
			if got.Max != tt.want.Max {
				t.Errorf("Max = %v, want %v", got.Max, tt.want.Max)
			}
			if !floatsAlmostEqual(got.Average, tt.want.Average) {
				t.Errorf("Average = %v, want %v", got.Average, tt.want.Average)
			}
			if !floatsAlmostEqual(got.Median, tt.want.Median) {
				t.Errorf("Median = %v, want %v", got.Median, tt.want.Median)
			}
		})
	}
}

func TestCalculateBasic_Empty(t *testing.T) {
	rates := []float64{}
	got := CalculateBasic(rates)

	if got.Min != 0 || got.Max != 0 || got.Average != 0 || got.Median != 0 {
		t.Errorf("Expected zero values for empty input, got %+v", got)
	}
}

func TestCalculateBasic_UnsortedInput(t *testing.T) {
	rates := []float64{5.0, 2.0, 8.0, 1.0, 9.0}
	got := CalculateBasic(rates)

	if got.Min != 1.0 {
		t.Errorf("Min = %v, want 1.0", got.Min)
	}
	if got.Max != 9.0 {
		t.Errorf("Max = %v, want 9.0", got.Max)
	}
	if !floatsAlmostEqual(got.Average, 5.0) {
		t.Errorf("Average = %v, want 5.0", got.Average)
	}
	if got.Median != 5.0 {
		t.Errorf("Median = %v, want 5.0", got.Median)
	}
}

func floatsAlmostEqual(a, b float64) bool {
	tolerance := 0.0001
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}
