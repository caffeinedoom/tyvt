.PHONY: build test clean run help

# Build the binary
build:
	go build -o tyvt .

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f tyvt coverage.out coverage.html

# Run with sample data (requires valid API keys in keys.txt)
run: build
	./tyvt -d domains.txt -k keys.txt -o results.json

# Install dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o tyvt-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o tyvt-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o tyvt-windows-amd64.exe .

# Show help
help:
	@echo "Available commands:"
	@echo "  build        - Build the binary"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Build and run with sample data"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  help         - Show this help message"