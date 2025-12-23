package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type APIClient interface {
	GetTimeSeriesRates(ctx context.Context, startDate, endDate time.Time, base string, targets []string) (*TimeSeriesResponse, error)
	GetSupportedCurrencies(ctx context.Context) (CurrenciesResponse, error)
}

type Client struct {
	baseURL       string
	timeout       time.Duration
	retryAttempts int
	httpClient    *http.Client
}

func NewClient(baseURL string, timeout time.Duration, retryAttempts int) *Client {
	return &Client{
		baseURL:       strings.TrimSuffix(baseURL, "/"),
		timeout:       timeout,
		retryAttempts: retryAttempts,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) GetTimeSeriesRates(ctx context.Context, startDate, endDate time.Time, base string, targets []string) (*TimeSeriesResponse, error) {
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	url := fmt.Sprintf("%s/%s..%s?from=%s&to=%s",
		c.baseURL,
		startStr,
		endStr,
		base,
		strings.Join(targets, ","),
	)

	var resp *TimeSeriesResponse
	var lastErr error

	for attempt := 0; attempt < c.retryAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt*attempt) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		httpResp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(httpResp.Body)
			lastErr = &APIError{
				StatusCode: httpResp.StatusCode,
				Message:    fmt.Sprintf("API request failed: %s", string(body)),
				URL:        url,
			}
			if httpResp.StatusCode >= 500 && attempt < c.retryAttempts-1 {
				continue // Retry on server errors
			}
			return nil, lastErr
		}

		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		resp = &TimeSeriesResponse{}
		if err := json.Unmarshal(body, resp); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

func (c *Client) GetSupportedCurrencies(ctx context.Context) (CurrenciesResponse, error) {
	url := fmt.Sprintf("%s/currencies", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, &APIError{
			StatusCode: httpResp.StatusCode,
			Message:    fmt.Sprintf("API request failed: %s", string(body)),
			URL:        url,
		}
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var resp CurrenciesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp, nil
}
