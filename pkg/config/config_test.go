package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoad_ValidFiles(t *testing.T) {
	domainsContent := "example.com\ngoogle.com\n# This is a comment\namazon.com"
	keysContent := "key1\nkey2\nkey3"

	domainsFile := createTempFile(t, "domains.txt", domainsContent)
	keysFile := createTempFile(t, "keys.txt", keysContent)

	defer os.Remove(domainsFile)
	defer os.Remove(keysFile)

	cfg, err := Load(domainsFile, keysFile, "output.json")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expectedDomains := []string{"example.com", "google.com", "amazon.com"}
	if len(cfg.Domains) != len(expectedDomains) {
		t.Errorf("Expected %d domains, got %d", len(expectedDomains), len(cfg.Domains))
	}

	for i, domain := range expectedDomains {
		if cfg.Domains[i] != domain {
			t.Errorf("Expected domain %s, got %s", domain, cfg.Domains[i])
		}
	}

	expectedKeys := []string{"key1", "key2", "key3"}
	if len(cfg.APIKeys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(cfg.APIKeys))
	}

	if cfg.RotationInterval != 90*time.Second {
		t.Errorf("Expected 90s rotation interval, got %v", cfg.RotationInterval)
	}
}

func TestLoad_EmptyFiles(t *testing.T) {
	domainsFile := createTempFile(t, "empty_domains.txt", "")
	keysFile := createTempFile(t, "empty_keys.txt", "")

	defer os.Remove(domainsFile)
	defer os.Remove(keysFile)

	_, err := Load(domainsFile, keysFile, "")
	if err == nil {
		t.Error("Expected error for empty domains file")
	}

	if !strings.Contains(err.Error(), "no domains found") {
		t.Errorf("Expected 'no domains found' error, got: %v", err)
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	_, err := Load("nonexistent_domains.txt", "nonexistent_keys.txt", "")
	if err == nil {
		t.Error("Expected error for nonexistent files")
	}
}

func createTempFile(t *testing.T, name, content string) string {
	file, err := os.CreateTemp("", name)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := file.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return file.Name()
}