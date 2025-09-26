package main

import (
	"context"
	"sync"

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
	var wg sync.WaitGroup
	resultsChan := make(chan *client.DomainResult, len(s.config.Domains))
	errorsChan := make(chan error, len(s.config.Domains))

	for _, domain := range s.config.Domains {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			result, err := s.client.QueryDomain(ctx, domain)
			if err != nil {
				s.logger.Error("Error querying domain %s: %v", domain, err)
				errorsChan <- err
				return
			}

			if result != nil {
				resultsChan <- result
			}
		}(domain)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	var results []*client.DomainResult
	errorCount := 0

	for {
		select {
		case result, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
				break
			}
			results = append(results, result)
			s.logger.Info("Successfully scanned domain: %s (%d undetected URLs)", result.Domain, len(result.UndetectedURLs))

		case err, ok := <-errorsChan:
			if !ok {
				errorsChan = nil
				break
			}
			errorCount++
			s.logger.Warn("Scan error: %v", err)

		case <-ctx.Done():
			return ctx.Err()
		}

		if resultsChan == nil && errorsChan == nil {
			break
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