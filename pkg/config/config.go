package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Domains          []string      `json:"domains"`
	APIKeys          []string      `json:"api_keys"`
	OutputFile       string        `json:"output_file,omitempty"`
	RotationInterval time.Duration `json:"rotation_interval"`
	ProxyList        []string      `json:"proxy_list,omitempty"`
}

func Load(domainsFile, keysFile, outputFile string) (*Config, error) {
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

	domains = filterEmptyStrings(domains)
	apiKeys = filterEmptyStrings(apiKeys)

	return &Config{
		Domains:          domains,
		APIKeys:          apiKeys,
		OutputFile:       outputFile,
		RotationInterval: 90 * time.Second,
		ProxyList:        []string{},
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