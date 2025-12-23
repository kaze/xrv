.PHONY: all build test lint clean install run coverage

BINARY_NAME=xrv
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/xrv/main.go

all: lint test build

build:
	@echo "Building..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

test:
	@echo "Running tests..."
	@go test -v -race ./...

coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Linting..."
	@go vet ./...
	@gofmt -s -w .

clean:
	@echo "Cleaning..."
	@rm -rf ./bin
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

install:
	@echo "Installing..."
	@go install $(MAIN_PATH)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

run:
	@go run $(MAIN_PATH)

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
