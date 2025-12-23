package statistics

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/stat"
)

type BasicStats struct {
	Min     float64
	Max     float64
	Average float64
	Median  float64
}

type VolatilityStats struct {
	StdDev           float64
	Variance         float64
	CoefficientOfVar float64
	AvgDailyReturn   float64
}

type TrendStats struct {
	Direction     string
	Slope         float64
	PercentChange float64
	SMA20         []float64
	SMA50         []float64
}

type Statistics struct {
	Basic      BasicStats
	Volatility VolatilityStats
	Trend      TrendStats
}

func Calculate(rates []float64) Statistics {
	return Statistics{
		Basic:      CalculateBasic(rates),
		Volatility: CalculateVolatility(rates),
		Trend:      CalculateTrend(rates),
	}
}

func CalculateBasic(rates []float64) BasicStats {
	if len(rates) == 0 {
		return BasicStats{}
	}

	data := make([]float64, len(rates))
	copy(data, rates)

	sort.Float64s(data)

	stats := BasicStats{
		Min:     data[0],
		Max:     data[len(data)-1],
		Average: stat.Mean(data, nil),
	}

	if len(data)%2 == 0 {
		mid := len(data) / 2
		stats.Median = (data[mid-1] + data[mid]) / 2.0
	} else {
		stats.Median = data[len(data)/2]
	}

	return stats
}

func CalculateVolatility(rates []float64) VolatilityStats {
	if len(rates) == 0 {
		return VolatilityStats{}
	}

	variance := stat.Variance(rates, nil)
	stdDev := math.Sqrt(variance)

	mean := stat.Mean(rates, nil)
	var coefficientOfVar float64
	if mean != 0 {
		coefficientOfVar = (stdDev / math.Abs(mean)) * 100
	}

	dailyReturns := calculateDailyReturns(rates)
	avgDailyReturn := 0.0
	if len(dailyReturns) > 0 {
		avgDailyReturn = stat.Mean(dailyReturns, nil)
	}

	return VolatilityStats{
		StdDev:           stdDev,
		Variance:         variance,
		CoefficientOfVar: coefficientOfVar,
		AvgDailyReturn:   avgDailyReturn,
	}
}

func CalculateTrend(rates []float64) TrendStats {
	if len(rates) == 0 {
		return TrendStats{}
	}

	slope := calculateSlope(rates)

	direction := "flat"
	if slope > 0.001 {
		direction = "upward"
	} else if slope < -0.001 {
		direction = "downward"
	}

	percentChange := 0.0
	if len(rates) >= 2 && rates[0] != 0 {
		percentChange = ((rates[len(rates)-1] - rates[0]) / rates[0]) * 100
	}

	sma20 := calculateSMA(rates, 20)
	sma50 := calculateSMA(rates, 50)

	return TrendStats{
		Direction:     direction,
		Slope:         slope,
		PercentChange: percentChange,
		SMA20:         sma20,
		SMA50:         sma50,
	}
}

func calculateDailyReturns(rates []float64) []float64 {
	if len(rates) < 2 {
		return []float64{}
	}

	returns := make([]float64, len(rates)-1)
	for i := 1; i < len(rates); i++ {
		if rates[i-1] != 0 {
			returns[i-1] = ((rates[i] - rates[i-1]) / rates[i-1]) * 100
		}
	}

	return returns
}

func calculateSlope(rates []float64) float64 {
	n := float64(len(rates))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range rates {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	numerator := n*sumXY - sumX*sumY
	denominator := n*sumX2 - sumX*sumX

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

func calculateSMA(rates []float64, period int) []float64 {
	if len(rates) < period || period <= 0 {
		return []float64{}
	}

	sma := make([]float64, len(rates)-period+1)
	for i := 0; i <= len(rates)-period; i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += rates[i+j]
		}
		sma[i] = sum / float64(period)
	}

	return sma
}
