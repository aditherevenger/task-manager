# Makefile for Go Task Manager

# Variables
BINARY_NAME=task_manager
MAIN_PACKAGE=./cmd/task_manager
BINARY_OUTPUT=./bin/$(BINARY_NAME)

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt

# Build flags
BUILD_FLAGS=-v

# Test flags
TEST_FLAGS=-v -cover

# Targets
.PHONY: all build run test clean fmt deps tidy help

all: clean fmt test build

build:
	@echo "Building..."
	@mkdir -p bin
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_OUTPUT) $(MAIN_PACKAGE)
	@echo "Build complete: $(BINARY_OUTPUT)"

run:
	@echo "Running..."
	$(BINARY_OUTPUT) -api -addr :8080

test:
	@echo "Running tests with coverage..."
	$(GOTEST) $(TEST_FLAGS) -coverprofile=coverage_all.out ./...
	findstr /V "cmd/task_manager/main.go" coverage_all.out > coverage.out
	@$(GOCMD) tool cover -func=coverage.out

clean:
	@echo "Cleaning..."
	@rm -rf bin
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

deps:
	@echo "Downloading dependencies..."
	$(GOGET) -v ./...

tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

help:
	@echo "Available targets:"
	@echo "  all     - Clean, format, test, and build"
	@echo "  build   - Build the application"
	@echo "  run     - Run the application"
	@echo "  test    - Run tests"
	@echo "  clean   - Remove build artifacts"
	@echo "  fmt     - Format code"
	@echo "  deps    - Download dependencies"
	@echo "  tidy    - Tidy go.mod file"
	@echo "  help    - Show this help message"
