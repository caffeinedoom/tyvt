package files

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluckware/tyvt/internal/client"
)

type Handler struct {
	outputFile string
}


func NewHandler(outputFile string) *Handler {
	return &Handler{
		outputFile: outputFile,
	}
}

func (h *Handler) HasOutputFile() bool {
	return h.outputFile != ""
}

func (h *Handler) WriteResults(results []*client.DomainResult) error {
	if h.outputFile == "" {
		return fmt.Errorf("no output file specified")
	}

	dir := filepath.Dir(h.outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var urls []string
	var filteredResults []*client.DomainResult
	totalUndetectedURLs := 0

	for _, result := range results {
		if result.ResponseCode == 1 && len(result.UndetectedURLs) > 0 {
			filteredResults = append(filteredResults, result)
			totalUndetectedURLs += len(result.UndetectedURLs)

			// Extract URLs for plain text output
			for _, undetectedURL := range result.UndetectedURLs {
				urls = append(urls, undetectedURL.URL)
			}
		}
	}

	file, err := os.Create(h.outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write URLs in plain text format, one per line
	for _, url := range urls {
		if _, err := fmt.Fprintln(file, url); err != nil {
			return fmt.Errorf("failed to write URL to file: %w", err)
		}
	}

	fmt.Printf("✓ URLs written to %s (%d domains, %d undetected URLs)\n",
		h.outputFile, len(filteredResults), totalUndetectedURLs)

	return nil
}

func (h *Handler) AppendResult(result *client.DomainResult) error {
	if h.outputFile == "" {
		return nil
	}

	if result.ResponseCode != 1 || len(result.UndetectedURLs) == 0 {
		return nil
	}

	file, err := os.OpenFile(h.outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file for appending: %w", err)
	}
	defer file.Close()

	// Append URLs in plain text format, one per line
	for _, undetectedURL := range result.UndetectedURLs {
		if _, err := fmt.Fprintln(file, undetectedURL.URL); err != nil {
			return fmt.Errorf("failed to append URL to file: %w", err)
		}
	}

	return nil
}

