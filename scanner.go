package main

import (
	"context"

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

func NewScanner(client *client.VirusTotalClient, fileHandler *files.Handler, cfg *config.Config, logger *logger.Logger) *Scanner {
	return &Scanner{
		client:      client,
		fileHandler: fileHandler,
		config:      cfg,
		logger:      logger,
	}
}

func (s *Scanner) Run(ctx context.Context) error {
	var results []*client.DomainResult
	errorCount := 0

	s.logger.Info("Processing %d domains sequentially to comply with API rate limits", len(s.config.Domains))

	for i, domain := range s.config.Domains {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		s.logger.Info("Scanning domain %d/%d: %s", i+1, len(s.config.Domains), domain)

		result, err := s.client.QueryDomain(ctx, domain)
		if err != nil {
			s.logger.Error("Error querying domain %s: %v", domain, err)
			errorCount++
			continue
		}

		if result != nil {
			results = append(results, result)
			s.logger.Info("Successfully scanned domain: %s (%d undetected URLs)", result.Domain, len(result.UndetectedURLs))
		}
	}

	if len(results) > 0 && s.fileHandler.HasOutputFile() {
		if err := s.fileHandler.WriteResults(results); err != nil {
			s.logger.Warn("Failed to write results to file: %v", err)
		} else {
			s.logger.Info("Results written to output file")
		}
	}

	s.logger.Info("Scan completed: %d successful, %d errors", len(results), errorCount)
	return nil
}