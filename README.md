# XRV - eXchange Rate Visualizer 📊

A powerful Go-based CLI application for visualizing historical exchange rates with beautiful terminal charts and comprehensive statistical analysis.

## Features

- **Historical Data**: Access exchange rates back to 1999 (via Frankfurter API)
- **Multiple Currencies**: Compare multiple currency pairs simultaneously
- **Dual Visualization Modes**:
  - Beautiful ASCII charts for terminal
  - Interactive browser charts with go-echarts
- **Interactive Mode**: Dynamic browser interface with form controls
- **Export Capabilities**: Export data as CSV, JSON, or PNG
- **Comprehensive Statistics**:
  - Basic stats (min, max, average, median)
  - Volatility metrics (standard deviation, coefficient of variation)
  - Trend analysis (direction, percentage change, moving averages)
- **Rate Inversion**: View base currency in terms of targets
- **Lightning Fast**: Persistent caching with BadgerDB
- **Fully Tested**: Comprehensive test coverage with TDD approach

## Quick Start

### Build from source

```bash
git clone https://github.com/kaze/xrv
cd xrv
make build
./bin/xrv viz --base USD --currencies EUR --from "30 days ago"
```

## Usage Examples

### Basic visualization

```bash
# Visualize USD to EUR over the last 30 days
./bin/xrv viz --base USD --currencies EUR --from "30 days ago"
```

### Multiple currencies

```bash
# Compare EUR against USD, GBP, and JPY over 90 days
./bin/xrv viz --base EUR --currencies USD,GBP,JPY --from "90 days ago"
```

### Long historical data

```bash
# View 10 years of USD to EUR data
./bin/xrv viz --base USD --currencies EUR --from "10 years ago"

# Historical data from 1999
./bin/xrv viz --base EUR --currencies USD --from 1999-01-04
```

### Custom chart size

```bash
# Larger chart for better visualization
./bin/xrv viz --base USD --currencies EUR --from "1 year ago" --height 20 --width 100
```

### Browser visualization

```bash
# Open visualization in browser (static chart)
./bin/xrv viz --base USD --currencies EUR --from "90 days ago" --output browser

# Interactive mode with form and dynamic updates
./bin/xrv viz --interactive
# or
./bin/xrv viz -i
```

### Rate inversion

```bash
# Show inverted rates (base in terms of target)
./bin/xrv viz --base USD --currencies EUR --from "30 days ago" --invert
```

### Disable caching

```bash
# Force fresh data fetch
./bin/xrv viz --base USD --currencies EUR --from "7 days ago" --no-cache
```

## Available Commands

### visualize (viz)

Visualize historical exchange rate data with charts and statistics.

**Flags:**
- `--base, -b`: Base currency (default: USD)
- `--currencies, -c`: Target currencies, comma-separated (default: EUR,GBP,JPY)
- `--from, -f`: Start date (YYYY-MM-DD) or relative (e.g., "30 days ago", "1 year ago")
- `--to, -t`: End date (YYYY-MM-DD), defaults to today
- `--output, -o`: Output mode: terminal, browser (default: terminal)
- `--interactive, -i`: Interactive browser mode with form
- `--invert`: Invert rates (show base in target currency)
- `--port`: Port for browser mode (default: 8080)
- `--height`: Chart height in lines (default: 15, terminal mode only)
- `--width`: Chart width in characters (default: 80, terminal mode only)
- `--no-cache`: Disable caching and fetch fresh data

## Sample Output

```
Fetching exchange rate data...

📊 EUR to USD, GBP, JPY
📅 2025-09-24 to 2025-12-23

━━━ USD ━━━
 1.18 ┤                                                              ╭
 1.17 ┼─╮  ╭╮╭─╮                                             ╭──╮  ╭╯
 1.16 ┤ ╰──╯╰╯ ╰─╮                                          ╭╯  ╰──╯
 1.15 ┤          ╰────────────────────────────────────────╯
                                EUR/USD

📈 Statistics:
  Min:     1.1491
  Max:     1.1786
  Average: 1.1633
  Median:  1.1630

📊 Volatility:
  StdDev:  0.0076
  Coeff:   0.65%

📉 Trend:
  Direction: flat
  Change:    0.26%
```

## Architecture

XRV follows clean architecture principles with TDD:

```
xrv/
├── cmd/xrv/              # Application entry point
├── internal/
│   ├── domain/           # Core domain models
│   ├── providers/        # Exchange rate data providers (Frankfurter)
│   ├── cache/            # Caching layer (BadgerDB)
│   ├── statistics/       # Statistical calculations
│   ├── service/          # Business logic orchestration
│   ├── visualization/    # Terminal and browser rendering
│   └── cli/              # CLI commands (Cobra)
└── configs/              # Configuration files
```

## Development

### Run tests

```bash
# All tests
make test

# With coverage
make coverage

# Integration tests (requires internet)
go test -tags=integration ./internal/providers/...
```

### Build

```bash
# Build binary
make build

# Build and run
make build && ./bin/xrv viz --help
```

### Code quality

```bash
# Lint and format
make lint
```

## Caching

XRV uses BadgerDB for persistent caching:
- **Historical data** (dates before today): Cached indefinitely
- **Current day data**: Cached for 1 hour
- **Cache location**: `~/.xrv/cache/`

Cache provides significant performance improvements:
- First fetch: ~300ms
- Cached fetch: ~30ms (10x faster!)

## Dependencies

- **API**: Frankfurter (ECB exchange rate data)
- **CLI**: Cobra (command structure)
- **Terminal Charts**: asciigraph (ASCII visualization)
- **Browser Charts**: go-echarts (interactive web charts)
- **Cache**: BadgerDB (persistent storage)
- **Stats**: Gonum (statistical calculations)

## Supported Currencies

XRV supports 31+ currencies via the Frankfurter API, including:
- USD, EUR, GBP, JPY, CHF, CAD, AUD, NZD, and many more!

## Features Implemented

- ✅ Browser-based visualization with go-echarts
- ✅ Interactive mode with dynamic form controls
- ✅ CSV/JSON/PNG export functionality
- ✅ Rate inversion support

## Future Enhancements

Potential improvements:
- Interactive Bubbletea UI with keyboard navigation
- Custom cache management commands
- More trend indicators (RSI, MACD, Bollinger Bands)
- Support for additional exchange rate providers

## License

MIT

## Acknowledgments

- Exchange rate data: [Frankfurter API](https://www.frankfurter.app/)
- Built following TDD principles
- Created as a learning project for Go development
