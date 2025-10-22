package main

import (
	"context"
	"fmt"

	"github.com/pluckware/tyvt/internal/client"
	"github.com/pluckware/tyvt/pkg/config"
	"github.com/pluckware/tyvt/pkg/files"
	"github.com/pluckware/tyvt/pkg/logger"
)

type Scanner struct {
	client      *client.VirusTotalClient
	fileHandler *files.Handler
	config      *config.Config
	logger      *logger.Logger
}

// ScanError represents a single scan error with context
type ScanError struct {
	Domain string
	Err    error
}

func (e ScanError) Error() string {
	return fmt.Sprintf("domain %s: %v", e.Domain, e.Err)
}

func NewScanner(client *client.VirusTotalClient, fileHandler *files.Handler, cfg *config.Config, logger *logger.Logger) *Scanner {
	return &Scanner{
		client:      client,
		fileHandler: fileHandler,
		config:      cfg,
		logger:      logger,
	}
}

// Run processes all domains sequentially, respecting API rate limits.
// Returns an error if more than 50% of domains fail to scan.
func (s *Scanner) Run(ctx context.Context) error {
	var results []*client.DomainResult
	var errors []ScanError
	totalDomains := len(s.config.Domains)

	s.logger.Info("Processing %d domains sequentially to comply with API rate limits", totalDomains)

	for i, domain := range s.config.Domains {
		select {
		case <-ctx.Done():
			s.logger.Warn("Scan interrupted by context cancellation")
			return ctx.Err()
		default:
		}

		s.logger.Info("Scanning domain %d/%d: %s", i+1, totalDomains, domain)

		result, err := s.client.QueryDomain(ctx, domain)
		if err != nil {
			s.logger.Error("Error querying domain %s: %v", domain, err)
			errors = append(errors, ScanError{Domain: domain, Err: err})
			continue
		}

		if result != nil {
			results = append(results, result)
			s.logger.Info("Successfully scanned domain: %s (%d undetected URLs)", result.Domain, len(result.UndetectedURLs))
		}

		// Log progress every 10 domains
		if (i+1)%10 == 0 {
			s.logger.Info("Progress: %d/%d domains scanned, %d successful, %d errors", 
				i+1, totalDomains, len(results), len(errors))
		}
	}

	// Write results to file if configured
	if len(results) > 0 && s.fileHandler.HasOutputFile() {
		if err := s.fileHandler.WriteResults(results); err != nil {
			s.logger.Warn("Failed to write results to file: %v", err)
		} else {
			s.logger.Info("Results written to output file")
		}
	}

	// Calculate success rate
	successRate := float64(len(results)) / float64(totalDomains) * 100
	s.logger.Info("Scan completed: %d successful (%.1f%%), %d errors", len(results), successRate, len(errors))

	// Return error if more than 50% failed
	if len(errors) > totalDomains/2 {
		return fmt.Errorf("scan failed with %d/%d errors (>50%% failure rate)", len(errors), totalDomains)
	}

	// Warn if any errors occurred but still within acceptable threshold
	if len(errors) > 0 {
		s.logger.Warn("Completed with %d errors out of %d domains", len(errors), totalDomains)
	}

	return nil
}