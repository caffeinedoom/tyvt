# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TYVT is a sophisticated VirusTotal domain scanner built in Go that implements intelligent API key rotation, rate limiting, and IP rotation to avoid blocking while querying VirusTotal's domain API. The tool extracts `undetected_urls` data and saves results to JSON files.

## Essential Commands

### Build and Test
```bash
make build          # Build the binary
make test           # Run all tests
make test-coverage  # Run tests with HTML coverage report
go test ./pkg/config -v  # Run tests for specific package
```

### Run the Application
```bash
make run            # Build and run with sample data (requires valid API keys)
./tyvt -d domains.txt -k keys.txt -o results.json
```

### Development
```bash
make fmt            # Format code
make deps           # Install/update dependencies
make clean          # Clean build artifacts
```

## Architecture Overview

The application follows a modular architecture with clear separation of concerns:

### Core Components
- **Scanner (`scanner.go`)**: Main orchestrator that processes domains sequentially with rate limiting
- **VirusTotalClient (`internal/client/`)**: Handles VirusTotal API communication and response parsing
- **KeyRotator (`internal/rotator/`)**: Manages automatic API key rotation (configurable interval, currently 15s)
- **RateLimiter (`internal/limiter/`)**: Enforces VirusTotal quota limits (500/day, 15,500/month per key) and request intervals
- **IPRotator (`internal/rotator/`)**: Optional proxy rotation for IP-based blocking avoidance

### Key Architectural Decisions
- **Sequential Processing**: Domains are processed one-by-one to respect API rate limits, not concurrently
- **Per-Key Quota Tracking**: Each API key has individual daily/monthly quota tracking with automatic reset
- **Context-Aware Operations**: All long-running operations support context cancellation
- **Structured Logging**: Uses custom logger with multiple levels (DEBUG, INFO, WARN, ERROR)

### Data Flow
1. Configuration loading (`pkg/config/`) reads domains and API keys from files
2. KeyRotator automatically rotates keys on a timer (separate goroutine)
3. Scanner processes domains sequentially, checking rate limits per API key
4. Each request waits for rate limiter approval before proceeding
5. Results are collected and optionally written to JSON via FileHandler (`pkg/files/`)

## Important Configuration Details

- **Rotation Interval**: Currently set to 15 seconds (changed from original 90s)
- **API Limits**: 500 requests/day, 15,500 requests/month per key
- **Rate Limiting**: Minimum interval between requests plus quota enforcement
- **File Formats**: Support comments (#) in domains.txt and keys.txt files

## Testing Strategy

Tests focus on core components with time-sensitive functionality:
- Rate limiter timing and quota enforcement
- Key rotation mechanics and thread safety
- Configuration file parsing and validation
- Mock-friendly interfaces for external dependencies