package domain

import "time"

type Currency string

type ExchangeRate struct {
	Date   time.Time
	Base   Currency
	Target Currency
	Rate   float64
}

type DataPoint struct {
	Date  time.Time
	Rates map[Currency]float64
}

type TimeSeriesData struct {
	Base       Currency
	Targets    []Currency
	StartDate  time.Time
	EndDate    time.Time
	DataPoints []DataPoint
}
