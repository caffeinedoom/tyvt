package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pluckware/tyvt/pkg/validation"
)

type Config struct {
	Domains          []string      `json:"domains"`
	APIKeys          []string      `json:"api_keys"`
	OutputFile       string        `json:"output_file,omitempty"`
	ProxyURL         *url.URL      `json:"-"` // Optional proxy URL (not serialized to JSON)
	RotationInterval time.Duration `json:"rotation_interval"`
}

// Load reads configuration from files and validates all inputs.
// proxyURL is optional - pass empty string for no proxy.
func Load(domainsFile, keysFile, outputFile, proxyURL string) (*Config, error) {
	domains, err := readLines(domainsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read domains file: %w", err)
	}

	apiKeys, err := readLines(keysFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read API keys file: %w", err)
	}

	if len(domains) == 0 {
		return nil, fmt.Errorf("no domains found in %s", domainsFile)
	}

	if len(apiKeys) == 0 {
		return nil, fmt.Errorf("no API keys found in %s", keysFile)
	}

	// Filter empty strings before validation
	domains = filterEmptyStrings(domains)
	apiKeys = filterEmptyStrings(apiKeys)

	// Validate domains
	validDomains, domainErrors := validation.ValidateDomains(domains)
	if len(domainErrors) > 0 {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Found %d invalid domain(s):\n", len(domainErrors))
		for _, err := range domainErrors {
			fmt.Fprintf(os.Stderr, "   - %v\n", err)
		}
	}

	if len(validDomains) == 0 {
		return nil, fmt.Errorf("no valid domains found after validation")
	}

	// Validate API keys
	validKeys, keyErrors := validation.ValidateAPIKeys(apiKeys)
	if len(keyErrors) > 0 {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Found %d invalid API key(s):\n", len(keyErrors))
		for _, err := range keyErrors {
			fmt.Fprintf(os.Stderr, "   - %v\n", err)
		}
	}

	if len(validKeys) == 0 {
		return nil, fmt.Errorf("no valid API keys found after validation")
	}

	// Log validation summary
	if len(validDomains) < len(domains) || len(validKeys) < len(apiKeys) {
		fmt.Fprintf(os.Stderr, "✓ Validation complete: %d/%d domains valid, %d/%d API keys valid\n\n", 
			len(validDomains), len(domains), len(validKeys), len(apiKeys))
	}

	// Validate proxy URL if provided
	var parsedProxyURL *url.URL
	if proxyURL != "" {
		parsedProxyURL, err = validation.ValidateProxyURL(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		fmt.Fprintf(os.Stderr, "✓ Using proxy: %s://%s\n\n", parsedProxyURL.Scheme, parsedProxyURL.Host)
	}

	return &Config{
		Domains:          validDomains,
		APIKeys:          validKeys,
		OutputFile:       outputFile,
		ProxyURL:         parsedProxyURL,
		RotationInterval: 15 * time.Second,
	}, nil
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}

func filterEmptyStrings(slice []string) []string {
	var result []string
	for _, s := range slice {
		if strings.TrimSpace(s) != "" {
			result = append(result, strings.TrimSpace(s))
		}
	}
	return result
}