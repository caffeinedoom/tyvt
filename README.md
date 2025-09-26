# TYVT - VirusTotal Domain Scanner

A sophisticated Go tool for querying VirusTotal's domain API with intelligent key rotation, rate limiting, and IP rotation capabilities.

## Features

- **API Key Rotation**: Automatically rotates through multiple VirusTotal API keys every 90 seconds
- **Rate Limiting**: Respects VirusTotal's API rate limits to prevent blocking
- **IP Rotation**: Support for proxy rotation to avoid IP-based blocking
- **Batch Processing**: Process multiple domains from a file
- **JSON Output**: Save `undetected_urls` data to JSON files
- **Robust Error Handling**: Comprehensive logging and error recovery
- **Modular Design**: Clean, testable, and maintainable code architecture

## Installation

```bash
git clone <repository-url>
cd tyvt
go build -o tyvt .
```

## Usage

### Basic Usage
```bash
./tyvt -d domains.txt -k keys.txt
```

### With Output File
```bash
./tyvt -d domains.txt -k keys.txt -o results.json
```

### Command Line Options
- `-d`: Path to domains file (required)
- `-k`: Path to API keys file (required)
- `-o`: Output file for results (optional)

## File Formats

### Domains File (`domains.txt`)
```
# Comments are supported
example.com
google.com
github.com
stackoverflow.com
```

### API Keys File (`keys.txt`)
```
# One API key per line
your_api_key_1_here
your_api_key_2_here
your_api_key_3_here
```

## Output Format

When using the `-o` flag, results are saved in JSON format:

```json
{
  "metadata": {
    "scan_time": "2023-XX-XXTXX:XX:XXZ",
    "total_domains": 5,
    "success_count": 4,
    "error_count": 1,
    "version": "1.0.0"
  },
  "results": [
    {
      "domain": "example.com",
      "response_code": 1,
      "undetected_urls": [
        {
          "url": "http://example.com/path",
          "positives": 0,
          "total": 67,
          "scan_date": "2023-XX-XX XX:XX:XX",
          "last_modified": "2023-XX-XXTXX:XX:XXZ"
        }
      ],
      "timestamp": "2023-XX-XXTXX:XX:XXZ"
    }
  ]
}
```

## Features in Detail

### Key Rotation
- Automatically rotates API keys every 90 seconds
- Supports single or multiple API keys
- Prevents API quota exhaustion

### Rate Limiting
- Built-in rate limiting to respect VirusTotal's API limits
- Configurable minimum intervals between requests
- Context-aware cancellation support

### IP Rotation
- Optional proxy support for IP rotation
- Helps avoid IP-based blocking
- Configurable proxy list

### Error Handling
- Comprehensive logging at multiple levels (DEBUG, INFO, WARN, ERROR)
- Graceful handling of API errors
- Retry logic for transient failures
- Clean shutdown on interrupt signals

## Testing

Run the test suite:
```bash
go test ./...
```

## Architecture

The project follows a clean, modular architecture:

```
├── main.go              # CLI entry point
├── scanner.go           # Main scanning orchestrator
├── internal/
│   ├── client/          # VirusTotal API client
│   ├── limiter/         # Rate limiting
│   └── rotator/         # Key and IP rotation
└── pkg/
    ├── config/          # Configuration management
    ├── files/           # File I/O handlers
    └── logger/          # Logging utilities
```

## Security Considerations

- Never commit real API keys to version control
- Use environment variables or secure key management for production
- Be mindful of VirusTotal's terms of service
- Implement appropriate delays to avoid overwhelming the API

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

[Add your license information here]
