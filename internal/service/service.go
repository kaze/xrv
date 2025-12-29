package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kaze/xrv/internal/cache"
	"github.com/kaze/xrv/internal/providers"
	"github.com/kaze/xrv/internal/domain"
	"github.com/kaze/xrv/internal/statistics"
)

type APIClient interface {
	GetTimeSeriesRates(ctx context.Context, startDate, endDate time.Time, base string, targets []string) (*providers.TimeSeriesResponse, error)
	GetSupportedCurrencies(ctx context.Context) (providers.CurrenciesResponse, error)
}

type FetchOptions struct {
	Base      domain.Currency
	Targets   []domain.Currency
	StartDate time.Time
	EndDate   time.Time
	UseCache  bool
}

type Service struct {
	apiClient APIClient
	cache     cache.Cache
	
}

func NewService(apiClient APIClient, cache cache.Cache, ) *Service {
	return &Service{
		apiClient: apiClient,
		cache:     cache,
		
	}
}

func (s *Service) FetchTimeSeriesData(ctx context.Context, opts FetchOptions) (*domain.TimeSeriesData, error) {
	cacheKey := s.generateCacheKey(opts)

	if opts.UseCache {
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			var data domain.TimeSeriesData
			if err := json.Unmarshal(cached, &data); err == nil {
				return &data, nil
			}
		}
	}

	targetsStr := make([]string, len(opts.Targets))
	for i, t := range opts.Targets {
		targetsStr[i] = string(t)
	}

	resp, err := s.apiClient.GetTimeSeriesRates(
		ctx,
		opts.StartDate,
		opts.EndDate,
		string(opts.Base),
		targetsStr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}

	data := s.transformToTimeSeriesData(resp, opts.Base, opts.Targets)

	if opts.UseCache {
		if cached, err := json.Marshal(data); err == nil {
			ttl := s.calculateTTL(opts.EndDate)
			s.cache.Set(ctx, cacheKey, cached, ttl)
		}
	}

	return data, nil
}

func (s *Service) CalculateStatistics(data *domain.TimeSeriesData) map[string]statistics.Statistics {
	result := make(map[string]statistics.Statistics)

	for _, target := range data.Targets {
		rates := make([]float64, 0, len(data.DataPoints))
		for _, dp := range data.DataPoints {
			if rate, exists := dp.Rates[target]; exists {
				rates = append(rates, rate)
			}
		}

		if len(rates) > 0 {
			result[string(target)] = statistics.Calculate(rates)
		}
	}

	return result
}

func (s *Service) GetSupportedCurrencies(ctx context.Context) (providers.CurrenciesResponse, error) {
	return s.apiClient.GetSupportedCurrencies(ctx)
}

func (s *Service) generateCacheKey(opts FetchOptions) string {
	targets := make([]string, len(opts.Targets))
	for i, t := range opts.Targets {
		targets[i] = string(t)
	}
	sort.Strings(targets)

	key := fmt.Sprintf("timeseries:%s:%s:%s:%s",
		opts.Base,
		strings.Join(targets, "-"),
		opts.StartDate.Format("2006-01-02"),
		opts.EndDate.Format("2006-01-02"),
	)

	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash[:16])
}

func (s *Service) calculateTTL(endDate time.Time) time.Duration {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	if endDay.Before(today) {
		return 0
	}

	return 1 * time.Hour
}

func (s *Service) transformToTimeSeriesData(resp *providers.TimeSeriesResponse, base domain.Currency, targets []domain.Currency) *domain.TimeSeriesData {
	data := &domain.TimeSeriesData{
		Base:    base,
		Targets: targets,
	}

	if startDate, err := time.Parse("2006-01-02", resp.StartDate); err == nil {
		data.StartDate = startDate
	}
	if endDate, err := time.Parse("2006-01-02", resp.EndDate); err == nil {
		data.EndDate = endDate
	}

	dates := make([]string, 0, len(resp.Rates))
	for date := range resp.Rates {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	data.DataPoints = make([]domain.DataPoint, 0, len(dates))
	for _, dateStr := range dates {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			rates := make(map[domain.Currency]float64)
			for currency, rate := range resp.Rates[dateStr] {
				rates[domain.Currency(currency)] = rate
			}

			data.DataPoints = append(data.DataPoints, domain.DataPoint{
				Date:  date,
				Rates: rates,
			})
		}
	}

	return data
}
