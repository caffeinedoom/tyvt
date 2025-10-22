package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoad_ValidFiles(t *testing.T) {
	domainsContent := "example.com\ngoogle.com\n# This is a comment\namazon.com"
	// Use valid 64-char hex keys for testing
	keysContent := "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add\n3b3febd37b5f774837bdb9fa4d6cfc78d022ab2f35b1bba18b152d779a77cbb9\nABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789"

	domainsFile := createTempFile(t, "domains.txt", domainsContent)
	keysFile := createTempFile(t, "keys.txt", keysContent)

	defer os.Remove(domainsFile)
	defer os.Remove(keysFile)

	cfg, err := Load(domainsFile, keysFile, "output.json", "")
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

	expectedKeyCount := 3
	if len(cfg.APIKeys) != expectedKeyCount {
		t.Errorf("Expected %d keys, got %d", expectedKeyCount, len(cfg.APIKeys))
	}

	if cfg.RotationInterval != 15*time.Second {
		t.Errorf("Expected 15s rotation interval, got %v", cfg.RotationInterval)
	}

	if cfg.ProxyURL != nil {
		t.Errorf("Expected no proxy, got %v", cfg.ProxyURL)
	}
}

func TestLoad_EmptyFiles(t *testing.T) {
	domainsFile := createTempFile(t, "empty_domains.txt", "")
	keysFile := createTempFile(t, "empty_keys.txt", "")

	defer os.Remove(domainsFile)
	defer os.Remove(keysFile)

	_, err := Load(domainsFile, keysFile, "", "")
	if err == nil {
		t.Error("Expected error for empty domains file")
	}

	if !strings.Contains(err.Error(), "no domains found") {
		t.Errorf("Expected 'no domains found' error, got: %v", err)
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	_, err := Load("nonexistent_domains.txt", "nonexistent_keys.txt", "", "")
	if err == nil {
		t.Error("Expected error for nonexistent files")
	}
}

func TestLoad_InvalidDomains(t *testing.T) {
	// Mix of valid and invalid domains
	domainsContent := "example.com\ninvalid_domain\ngoogle.com\n..bad.com"
	keysContent := "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add"

	domainsFile := createTempFile(t, "mixed_domains.txt", domainsContent)
	keysFile := createTempFile(t, "keys.txt", keysContent)

	defer os.Remove(domainsFile)
	defer os.Remove(keysFile)

	cfg, err := Load(domainsFile, keysFile, "", "")
	if err != nil {
		t.Fatalf("Load should succeed with some valid domains: %v", err)
	}

	// Should have filtered out invalid domains
	if len(cfg.Domains) != 2 {
		t.Errorf("Expected 2 valid domains, got %d", len(cfg.Domains))
	}
}

func TestLoad_InvalidAPIKeys(t *testing.T) {
	domainsContent := "example.com"
	// Mix of valid and invalid keys
	keysContent := "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add\ninvalid_key\ntoo_short"

	domainsFile := createTempFile(t, "domains.txt", domainsContent)
	keysFile := createTempFile(t, "mixed_keys.txt", keysContent)

	defer os.Remove(domainsFile)
	defer os.Remove(keysFile)

	cfg, err := Load(domainsFile, keysFile, "", "")
	if err != nil {
		t.Fatalf("Load should succeed with some valid keys: %v", err)
	}

	// Should have filtered out invalid keys
	if len(cfg.APIKeys) != 1 {
		t.Errorf("Expected 1 valid API key, got %d", len(cfg.APIKeys))
	}
}

func TestLoad_WithValidProxy(t *testing.T) {
	domainsContent := "example.com"
	keysContent := "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add"

	domainsFile := createTempFile(t, "domains.txt", domainsContent)
	keysFile := createTempFile(t, "keys.txt", keysContent)

	defer os.Remove(domainsFile)
	defer os.Remove(keysFile)

	proxyURL := "http://proxy.example.com:8080"
	cfg, err := Load(domainsFile, keysFile, "", proxyURL)
	if err != nil {
		t.Fatalf("Load failed with valid proxy: %v", err)
	}

	if cfg.ProxyURL == nil {
		t.Fatal("Expected proxy URL to be set, got nil")
	}

	if cfg.ProxyURL.Host != "proxy.example.com:8080" {
		t.Errorf("Expected proxy host proxy.example.com:8080, got %s", cfg.ProxyURL.Host)
	}
}

func TestLoad_WithInvalidProxy(t *testing.T) {
	domainsContent := "example.com"
	keysContent := "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add"

	domainsFile := createTempFile(t, "domains.txt", domainsContent)
	keysFile := createTempFile(t, "keys.txt", keysContent)

	defer os.Remove(domainsFile)
	defer os.Remove(keysFile)

	// Invalid proxy URL (no scheme)
	_, err := Load(domainsFile, keysFile, "", "proxy.com:8080")
	if err == nil {
		t.Error("Expected error for invalid proxy URL")
	}

	if !strings.Contains(err.Error(), "invalid proxy URL") {
		t.Errorf("Expected 'invalid proxy URL' error, got: %v", err)
	}
}

func TestLoad_WithAuthenticatedProxy(t *testing.T) {
	domainsContent := "example.com"
	keysContent := "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add"

	domainsFile := createTempFile(t, "domains.txt", domainsContent)
	keysFile := createTempFile(t, "keys.txt", keysContent)

	defer os.Remove(domainsFile)
	defer os.Remove(keysFile)

	proxyURL := "http://user:password@proxy.example.com:8080"
	cfg, err := Load(domainsFile, keysFile, "", proxyURL)
	if err != nil {
		t.Fatalf("Load failed with authenticated proxy: %v", err)
	}

	if cfg.ProxyURL == nil {
		t.Fatal("Expected proxy URL to be set, got nil")
	}

	if cfg.ProxyURL.User == nil {
		t.Error("Expected proxy credentials to be preserved")
	}

	username := cfg.ProxyURL.User.Username()
	if username != "user" {
		t.Errorf("Expected username 'user', got %s", username)
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