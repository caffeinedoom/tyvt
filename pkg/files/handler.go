package files

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pluckware/tyvt/internal/client"
)

type Handler struct {
	outputFile string
}

type OutputData struct {
	Metadata ScanMetadata            `json:"metadata"`
	Results  []*client.DomainResult  `json:"results"`
}

type ScanMetadata struct {
	ScanTime    time.Time `json:"scan_time"`
	TotalDomains int      `json:"total_domains"`
	SuccessCount int      `json:"success_count"`
	ErrorCount   int      `json:"error_count"`
	Version     string    `json:"version"`
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

	var filteredResults []*client.DomainResult
	totalUndetectedURLs := 0

	for _, result := range results {
		if result.ResponseCode == 1 && len(result.UndetectedURLs) > 0 {
			filteredResults = append(filteredResults, result)
			totalUndetectedURLs += len(result.UndetectedURLs)
		}
	}

	outputData := OutputData{
		Metadata: ScanMetadata{
			ScanTime:     time.Now(),
			TotalDomains: len(results),
			SuccessCount: len(filteredResults),
			ErrorCount:   len(results) - len(filteredResults),
			Version:      "1.0.0",
		},
		Results: filteredResults,
	}

	file, err := os.Create(h.outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(outputData); err != nil {
		return fmt.Errorf("failed to encode results to JSON: %w", err)
	}

	fmt.Printf("✓ Results written to %s (%d domains, %d undetected URLs)\n",
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

	var existingData OutputData

	if _, err := os.Stat(h.outputFile); err == nil {
		file, err := os.Open(h.outputFile)
		if err != nil {
			return fmt.Errorf("failed to open existing file: %w", err)
		}

		if err := json.NewDecoder(file).Decode(&existingData); err != nil {
			file.Close()
			existingData = OutputData{
				Metadata: ScanMetadata{
					Version: "1.0.0",
				},
				Results: []*client.DomainResult{},
			}
		} else {
			file.Close()
		}
	} else {
		existingData = OutputData{
			Metadata: ScanMetadata{
				Version: "1.0.0",
			},
			Results: []*client.DomainResult{},
		}
	}

	existingData.Results = append(existingData.Results, result)
	existingData.Metadata.ScanTime = time.Now()
	existingData.Metadata.TotalDomains = len(existingData.Results)
	existingData.Metadata.SuccessCount = len(existingData.Results)

	return h.writeData(existingData)
}

func (h *Handler) writeData(data OutputData) error {
	dir := filepath.Dir(h.outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(h.outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}