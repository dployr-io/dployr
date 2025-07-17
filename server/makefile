# Makefile for dployr

.PHONY: build run swagger clean deps test

# Variables
BINARY_NAME=dployr
MAIN_PATH=./main.go
DOCS_PATH=./docs

# Default target
all: build

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install swag tool
install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger documentation
swagger: install-swag
	swag init -g $(MAIN_PATH) -o $(DOCS_PATH) --parseInternal --parseDependency

# Build the application
build: swagger
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application in development mode
run: swagger
	go run $(MAIN_PATH)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(DOCS_PATH)

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Hot reload for development (requires air)
dev:
	air

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application (includes swagger generation)"
	@echo "  run          - Run the application in development mode"
	@echo "  swagger      - Generate swagger documentation"
	@echo "  deps         - Install dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  dev          - Run with hot reload (requires air)"
	@echo "  help         - Show this help"