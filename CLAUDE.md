# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TYVT is a sophisticated VirusTotal domain scanner built in Go that implements intelligent API key rotation, rate limiting, and IP rotation to avoid blocking while querying VirusTotal's domain API. The tool extracts `undetected_urls` data and saves results to plain text files.

**Recent Improvements (Oct 2025):**
- Input validation for domains and API keys
- Deduplicated rate limiter code (removed 35 lines)
- Enhanced error handling with failure thresholds
- Improved test coverage (31.1% → 42.6%)
- Added comprehensive validation package with 96.6% coverage

## Essential Commands

### Build and Test
```bash
make build          # Build the binary
make test           # Run all tests
make test-coverage  # Run tests with HTML coverage report
go test ./pkg/config -v  # Run tests for specific package
```

### Installation
```bash
make install        # Install to /usr/local/bin (requires sudo)
make install-user   # Install to ~/.local/bin (no sudo required)
make uninstall      # Remove from system
```

### Run the Application
```bash
make run            # Build and run with sample data (requires valid API keys)
./tyvt -d domains.txt -k keys.txt -o results.txt
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
- **KeyRotator (`internal/rotator/`)**: Manages automatic API key rotation (15s interval)
- **RateLimiter (`internal/limiter/`)**: Enforces per-key quota tracking (500/day, 15,500/month) with intelligent rate limiting
- **IPRotator (`internal/rotator/`)**: Optional proxy rotation for IP-based blocking avoidance (currently inactive)
- **Validator (`pkg/validation/`)**: Validates domain names and API keys before processing

### Key Architectural Decisions
- **Sequential Processing**: Domains are processed one-by-one to respect API rate limits
- **Per-Key Quota Tracking**: Each API key has individual daily/monthly quota tracking with automatic reset
- **Context-Aware Operations**: All long-running operations support context cancellation
- **Input Validation**: All inputs validated before API calls to prevent wasted requests
- **Structured Logging**: Uses custom logger with multiple levels (DEBUG, INFO, WARN, ERROR)
- **Error Aggregation**: Scanner returns error if >50% of domains fail

### Data Flow
1. Configuration loading (`pkg/config/`) reads and validates domains and API keys from files
2. Invalid entries are filtered out with warnings logged to stderr
3. KeyRotator automatically rotates keys on a timer (separate goroutine)
4. Scanner processes domains sequentially, checking rate limits per API key
5. Each request waits for rate limiter approval before proceeding
6. Results are collected and written to plain text file (one URL per line)

## Important Configuration Details

- **Rotation Interval**: 15 seconds (configured in `pkg/config/config.go`)
- **API Limits**: 500 requests/day, 15,500 requests/month per key
- **Rate Limiting**: Minimum 15s interval between requests plus per-key quota enforcement
- **Current Request Rate**: ~0.067 req/s (1 request every 15 seconds)
- **File Formats**: Support comments (#) in domains.txt and keys.txt files
- **Validation**: Domain names must be valid DNS format; API keys must be 64-char hex strings

## Testing Strategy

Tests focus on core components with comprehensive coverage:
- **Rate limiter**: Timing, quota enforcement, thread safety (89.6% coverage)
- **Key rotation**: Mechanics and thread safety (58.2% coverage)  
- **Configuration**: File parsing and validation (91.1% coverage)
- **Validation**: Domain and API key format validation (96.6% coverage)
- **Overall coverage**: 42.6% (up from 31.1%)

### Testing Best Practices
- Table-driven tests for comprehensive scenario coverage
- Mock-friendly interfaces for external dependencies
- Context cancellation testing
- Thread-safety validation with concurrent access tests

## Code Quality Improvements

### Completed
- ✅ Removed empty directories (`cmd/`, `internal/parser/`)
- ✅ Removed unused struct fields (`maxRequests`, `requestCount`)
- ✅ Deduplicated rate limiter quota checking logic
- ✅ Removed unused methods (`CanMakeRequest()`, `GetCurrentIndex()`)
- ✅ Added comprehensive input validation
- ✅ Improved error handling with failure thresholds
- ✅ Added `.gitignore` for security
- ✅ Enhanced scanner with progress reporting

### Security Considerations
- Input validation prevents malformed API calls
- API keys masked in error messages (shows only last 4 chars)
- Domain format validation prevents potential injection issues
- `.gitignore` configured to prevent accidental key commits

## Dependencies

**Zero external dependencies** - uses only Go standard library (1.23.0+)