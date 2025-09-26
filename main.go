package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pluckware/tyvt/internal/client"
	"github.com/pluckware/tyvt/internal/limiter"
	"github.com/pluckware/tyvt/internal/rotator"
	"github.com/pluckware/tyvt/pkg/config"
	"github.com/pluckware/tyvt/pkg/files"
	"github.com/pluckware/tyvt/pkg/logger"
)

func main() {
	var (
		domainsFile = flag.String("d", "", "Path to domains file (required)")
		keysFile    = flag.String("k", "", "Path to API keys file (required)")
		outputFile  = flag.String("o", "", "Output file for results (optional)")
	)
	flag.Parse()

	if *domainsFile == "" || *keysFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -d domains.txt -k keys.txt [-o output.json]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg, err := config.Load(*domainsFile, *keysFile, *outputFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	appLogger := logger.New(logger.LevelInfo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		appLogger.Info("Received shutdown signal, stopping...")
		cancel()
	}()

	rateLimiter := limiter.New(cfg.RotationInterval)
	keyRotator := rotator.NewKeyRotator(cfg.APIKeys, cfg.RotationInterval)
	ipRotator := rotator.NewIPRotator(cfg.ProxyList)

	vtClient := client.NewVirusTotalClient(keyRotator, ipRotator, rateLimiter)
	fileHandler := files.NewHandler(*outputFile)

	scanner := NewScanner(vtClient, fileHandler, cfg, appLogger)

	appLogger.Info("Starting scan of %d domains with %d API keys", len(cfg.Domains), len(cfg.APIKeys))
	if len(cfg.ProxyList) > 0 {
		appLogger.Info("Using %d proxy servers for IP rotation", len(cfg.ProxyList))
	}

	if err := scanner.Run(ctx); err != nil {
		appLogger.Error("Scanner failed: %v", err)
		os.Exit(1)
	}

	appLogger.Info("Scan completed successfully")
}