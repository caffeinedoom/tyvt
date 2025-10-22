package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Domain name validation pattern
// Matches valid DNS domain names (RFC 1035/1123 compliant)
var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// VirusTotal API key pattern (64 character hexadecimal string)
var apiKeyRegex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

// ValidateDomain checks if a domain name is valid according to DNS standards.
// Returns an error if the domain is invalid.
func ValidateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Trim whitespace
	domain = strings.TrimSpace(domain)

	// Check length (max 253 characters for FQDN)
	if len(domain) > 253 {
		return fmt.Errorf("domain too long (max 253 characters): %s", domain)
	}

	// Check if it matches domain pattern
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}

	// Additional check: no consecutive dots
	if strings.Contains(domain, "..") {
		return fmt.Errorf("invalid domain (consecutive dots): %s", domain)
	}

	return nil
}

// ValidateAPIKey checks if a string is a valid VirusTotal API key format.
// VirusTotal uses 64-character hexadecimal API keys.
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Trim whitespace
	apiKey = strings.TrimSpace(apiKey)

	// Check if it matches VirusTotal API key pattern
	if !apiKeyRegex.MatchString(apiKey) {
		return fmt.Errorf("invalid API key format (expected 64-char hex string)")
	}

	return nil
}

// ValidateProxyURL checks if a string is a valid proxy URL.
// Supports formats:
//   - http://proxy.com:8080
//   - https://proxy.com:8080
//   - socks5://proxy.com:1080
//   - http://username:password@proxy.com:8080
func ValidateProxyURL(proxyURL string) (*url.URL, error) {
	if proxyURL == "" {
		return nil, fmt.Errorf("proxy URL cannot be empty")
	}

	// Trim whitespace
	proxyURL = strings.TrimSpace(proxyURL)

	// Parse the URL
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL format: %w", err)
	}

	// Validate scheme
	validSchemes := map[string]bool{
		"http":   true,
		"https":  true,
		"socks5": true,
	}

	if !validSchemes[parsedURL.Scheme] {
		return nil, fmt.Errorf("unsupported proxy scheme '%s' (use http, https, or socks5)", parsedURL.Scheme)
	}

	// Validate host is present
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("proxy URL must include a host")
	}

	return parsedURL, nil
}

// ValidateDomains validates a slice of domains and returns all invalid ones
// along with their error messages.
func ValidateDomains(domains []string) (valid []string, errors []error) {
	for _, domain := range domains {
		if err := ValidateDomain(domain); err != nil {
			errors = append(errors, fmt.Errorf("domain '%s': %w", domain, err))
		} else {
			valid = append(valid, strings.TrimSpace(domain))
		}
	}
	return valid, errors
}

// ValidateAPIKeys validates a slice of API keys and returns all invalid ones
// along with their error messages.
func ValidateAPIKeys(keys []string) (valid []string, errors []error) {
	for _, key := range keys {
		if err := ValidateAPIKey(key); err != nil {
			errors = append(errors, fmt.Errorf("API key (***%s): %w", 
				maskAPIKey(key), err))
		} else {
			valid = append(valid, strings.TrimSpace(key))
		}
	}
	return valid, errors
}

// maskAPIKey returns a masked version of an API key for safe logging
// Shows only last 4 characters
func maskAPIKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return key[len(key)-4:]
}
