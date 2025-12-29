# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

XRV is a Go-based CLI application for visualizing historical exchange rates. It fetches data from the Frankfurter API, provides comprehensive statistical analysis, and supports both terminal and browser-based visualizations. The project follows clean architecture principles with Test-Driven Development (TDD).

## Essential Commands

### Building and Running
```bash
# Build binary to ./bin/xrv
make build

# Run the application
./bin/xrv viz --base USD --currencies EUR --from "30 days ago"

# Run directly without building
go run ./cmd/xrv/main.go viz --help
```

### Testing
```bash
# Run all unit tests with race detector
make test

# Run with coverage report (generates coverage.html)
make coverage

# Run integration tests (requires internet connection)
go test -tags=integration ./internal/providers/...

# Run tests for a specific package
go test -v ./internal/service/...

# Run a specific test
go test -v -run TestServiceFetchTimeSeriesData ./internal/service/...
```

### Code Quality
```bash
# Run linter and formatter
make lint

# Update dependencies
make deps
```

### Visualization Modes
```bash
# Terminal mode (default)
./bin/xrv viz --base EUR --currencies USD,GBP,JPY --from "1 year ago"

# Browser mode (static chart)
./bin/xrv viz --base USD --currencies EUR --from "90 days ago" --output browser --port 8080

# Interactive browser mode (with form for dynamic queries)
./bin/xrv viz --interactive
```

## Architecture

The codebase follows clean architecture with clear separation of concerns:

### Layer Structure

**Domain Layer** (`internal/domain/`)
- Core business models: `Currency`, `ExchangeRate`, `DataPoint`, `TimeSeriesData`
- No external dependencies
- Pure Go types representing the business domain

**Providers Layer** (`internal/providers/`)
- `APIClient`: Interface for exchange rate data providers
- `FrankfurterClient`: HTTP client for Frankfurter API with retry logic and exponential backoff
- Implements provider pattern for testability and extensibility
- Handles two endpoints: time series rates and supported currencies
- Default: 30s timeout, 3 retry attempts

**Cache Layer** (`internal/cache/`)
- `Cache` interface with two implementations:
  - `BadgerCache`: Persistent storage using BadgerDB (production)
  - `MemoryCache`: In-memory map (testing only)
- Smart TTL strategy:
  - Historical data (before today): cached indefinitely (TTL=0)
  - Current day data: cached for 1 hour
- Cache location: `~/.xrv/cache/`

**Service Layer** (`internal/service/`)
- `Service`: Orchestrates API, cache, and statistics
- Handles cache key generation (SHA256 hash of query parameters)
- Coordinates data fetching with cache-first strategy
- Transforms API responses to domain models

**Statistics Layer** (`internal/statistics/`)
- `Calculator`: Computes comprehensive statistics
  - Basic: min, max, average, median
  - Volatility: standard deviation, coefficient of variation
  - Trends: direction, percentage change, moving averages
- Uses Gonum for statistical calculations

**Visualization Layer** (`internal/visualization/`)
- Two renderers:
  - `terminal.Renderer`: ASCII charts using asciigraph
  - `browser.Renderer`: HTML charts using go-echarts
- `browser.Server`: HTTP server for interactive mode with Go templates

**CLI Layer** (`internal/cli/`)
- Built with Cobra framework
- `visualize` command (alias: `viz`) with extensive flags
- Handles date parsing (absolute: YYYY-MM-DD, relative: "30 days ago")
- Supports rate inversion with `--invert` flag

### Data Flow

1. **CLI** parses user input and creates `FetchOptions`
2. **Service** checks cache using hashed key
3. On cache miss: **Service** → **API Client** → Frankfurter API
4. **Service** transforms API response to domain models and caches it
5. **Service** calculates statistics using **Calculator**
6. **Renderer** (terminal or browser) visualizes data + statistics

### Key Interfaces

```go
// Service dependencies
type APIClient interface {
    GetTimeSeriesRates(...) (*TimeSeriesResponse, error)
    GetSupportedCurrencies(...) (CurrenciesResponse, error)
}

type Cache interface {
    Get(ctx, key) ([]byte, error)
    Set(ctx, key, value, ttl) error
    Delete(ctx, key) error
    Clear(ctx) error
    Close() error
}
```

These interfaces enable dependency injection and comprehensive testing with mocks.

## Testing Strategy

- **TDD approach**: Tests written first, then implementation
- **Table-driven tests**: Common pattern for testing multiple scenarios
- **Interface mocking**: Service layer tests mock APIClient and Cache
- **Integration tests**: Separate build tag for real API calls
- **Coverage**: Use `make coverage` to ensure high test coverage

## Code Style and Conventions

- **TDD is mandatory**: Always write tests before implementation
- **No emojis**: Keep code and output clean and professional
- **No comments or docstrings**: Code should be self-documenting through clear naming and structure

## Important Implementation Details

### Cache Key Generation
Cache keys are SHA256 hashes (first 16 bytes) of:
`timeseries:{base}:{sorted-targets}:{startDate}:{endDate}`

Targets are sorted alphabetically to ensure consistent cache hits.

### Date Parsing
Supports two formats:
- Absolute: `YYYY-MM-DD` (e.g., `2024-01-01`)
- Relative: `N days/months/years ago` (e.g., `30 days ago`)

### Rate Inversion
When `--invert` flag is used, all rates are inverted (1/rate) to show the base currency's value in terms of the target currency.

### Browser Mode
- Static mode (`--output browser`): Generates and opens HTML chart
- Interactive mode (`--interactive` or `-i`): Starts HTTP server with form for dynamic queries
- Server includes both form and chart on the same page using Go templates

## Dependencies

- **CLI Framework**: `github.com/spf13/cobra`
- **Terminal Charts**: `github.com/guptarohit/asciigraph`
- **Browser Charts**: `github.com/go-echarts/go-echarts/v2`
- **Persistent Cache**: `github.com/dgraph-io/badger/v4`
- **Statistics**: `gonum.org/v1/gonum`
- **Configuration**: `github.com/spf13/viper`

## Common Development Patterns

### Adding a New Statistic
1. Add calculation function to `internal/statistics/`
2. Add field to `Statistics` struct in `internal/statistics/models.go`
3. Update `Calculator.Calculate()` to compute the new metric
4. Update renderer to display the new statistic

### Adding a New Visualization Mode
1. Create new renderer in `internal/visualization/{mode}/`
2. Implement `Render(data, stats)` method
3. Add mode to `--output` flag options in `internal/cli/visualize.go`
4. Add case to output mode switch in `runVisualize()`

### Modifying Provider Implementation
1. Update models in `internal/providers/frankfurter.go`
2. Add/modify client method in `internal/providers/frankfurter.go`
3. Write integration test with `// +build integration` tag
4. Update service layer to use new endpoint

### Adding a New Provider
1. Create new file `internal/providers/{provider}.go`
2. Implement `APIClient` interface with provider-specific logic
3. Add provider-specific response models
4. Write unit and integration tests
5. Update service initialization to use new provider
