package validation

import (
	"strings"
	"testing"
)

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		wantError bool
	}{
		// Valid domains
		{"valid simple domain", "example.com", false},
		{"valid subdomain", "www.example.com", false},
		{"valid long domain", "subdomain.example.co.uk", false},
		{"valid with numbers", "test123.example.com", false},
		{"valid with hyphen", "my-site.example.com", false},
		{"valid multi-level", "a.b.c.example.com", false},

		// Invalid domains
		{"empty domain", "", true},
		{"no TLD", "example", true},
		{"starts with dot", ".example.com", true},
		{"ends with dot", "example.com.", true},
		{"consecutive dots", "example..com", true},
		{"starts with hyphen", "-example.com", true},
		{"ends with hyphen", "example-.com", true},
		{"only TLD", ".com", true},
		{"whitespace only", "   ", true},
		{"special characters", "example!.com", true},
		{"underscore", "my_site.example.com", true},
		{"too long", strings.Repeat("a", 250) + ".com", true},
		{"spaces", "my site.com", true},
		{"invalid chars", "example@site.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDomain(tt.domain)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateDomain(%q) error = %v, wantError %v", 
					tt.domain, err, tt.wantError)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		wantError bool
	}{
		// Valid API keys
		{"valid key lowercase", "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add", false},
		{"valid key uppercase", "C9A9CFEA8329CDF114760ED36FC8468DD1A1CB826D4ADAB9FEE96BAD9EC74ADD", false},
		{"valid key mixed case", "C9a9cFEa8329CdF114760Ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add", false},

		// Invalid API keys
		{"empty key", "", true},
		{"too short", "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74ad", true},
		{"too long", "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add1", true},
		{"non-hex chars", "g9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add", true},
		{"with spaces", "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74ad ", true},
		{"special chars", "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74ad!", true},
		{"whitespace only", "                                                                ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKey(tt.apiKey)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateAPIKey() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateProxyURL(t *testing.T) {
	tests := []struct {
		name      string
		proxyURL  string
		wantError bool
		wantHost  string // Expected host after parsing
	}{
		// Valid proxy URLs
		{"http proxy", "http://proxy.com:8080", false, "proxy.com:8080"},
		{"https proxy", "https://proxy.com:8080", false, "proxy.com:8080"},
		{"socks5 proxy", "socks5://proxy.com:1080", false, "proxy.com:1080"},
		{"with auth", "http://user:pass@proxy.com:8080", false, "proxy.com:8080"},
		{"with complex password", "http://user:p@ss:w0rd@proxy.com:8080", false, "proxy.com:8080"},
		{"no port", "http://proxy.com", false, "proxy.com"},
		{"with path", "http://proxy.com:8080/path", false, "proxy.com:8080"},
		{"IP address", "http://192.168.1.1:8080", false, "192.168.1.1:8080"},
		{"with whitespace", "  http://proxy.com:8080  ", false, "proxy.com:8080"},

		// Invalid proxy URLs
		{"empty", "", true, ""},
		{"no scheme", "proxy.com:8080", true, ""},
		{"invalid scheme", "ftp://proxy.com:8080", true, ""},
		{"no host", "http://", true, ""},
		{"whitespace only", "   ", true, ""},
		{"malformed URL", "http://[invalid", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := ValidateProxyURL(tt.proxyURL)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateProxyURL(%q) error = %v, wantError %v", 
					tt.proxyURL, err, tt.wantError)
			}

			if !tt.wantError && parsedURL != nil {
				if parsedURL.Host != tt.wantHost {
					t.Errorf("ValidateProxyURL(%q) host = %q, want %q",
						tt.proxyURL, parsedURL.Host, tt.wantHost)
				}
			}
		})
	}
}

func TestValidateDomains(t *testing.T) {
	tests := []struct {
		name        string
		domains     []string
		wantValid   int
		wantInvalid int
	}{
		{
			name:        "all valid",
			domains:     []string{"example.com", "google.com", "github.com"},
			wantValid:   3,
			wantInvalid: 0,
		},
		{
			name:        "all invalid",
			domains:     []string{"invalid", "bad..domain.com", "-nope.com"},
			wantValid:   0,
			wantInvalid: 3,
		},
		{
			name:        "mixed valid and invalid",
			domains:     []string{"example.com", "invalid", "github.com"},
			wantValid:   2,
			wantInvalid: 1,
		},
		{
			name:        "empty slice",
			domains:     []string{},
			wantValid:   0,
			wantInvalid: 0,
		},
		{
			name:        "with whitespace",
			domains:     []string{"  example.com  ", "github.com"},
			wantValid:   2,
			wantInvalid: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errors := ValidateDomains(tt.domains)
			if len(valid) != tt.wantValid {
				t.Errorf("ValidateDomains() got %d valid, want %d", len(valid), tt.wantValid)
			}
			if len(errors) != tt.wantInvalid {
				t.Errorf("ValidateDomains() got %d errors, want %d", len(errors), tt.wantInvalid)
			}
		})
	}
}

func TestValidateAPIKeys(t *testing.T) {
	validKey := "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add"
	invalidKey := "invalid_key"

	tests := []struct {
		name        string
		keys        []string
		wantValid   int
		wantInvalid int
	}{
		{
			name:        "all valid",
			keys:        []string{validKey, validKey},
			wantValid:   2,
			wantInvalid: 0,
		},
		{
			name:        "all invalid",
			keys:        []string{invalidKey, "short", ""},
			wantValid:   0,
			wantInvalid: 3,
		},
		{
			name:        "mixed",
			keys:        []string{validKey, invalidKey},
			wantValid:   1,
			wantInvalid: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errors := ValidateAPIKeys(tt.keys)
			if len(valid) != tt.wantValid {
				t.Errorf("ValidateAPIKeys() got %d valid, want %d", len(valid), tt.wantValid)
			}
			if len(errors) != tt.wantInvalid {
				t.Errorf("ValidateAPIKeys() got %d errors, want %d", len(errors), tt.wantInvalid)
			}
		})
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{"long key", "c9a9cfea8329cdf114760ed36fc8468dd1a1cb826d4adab9fee96bad9ec74add", "4add"},
		{"short key", "abc", "****"},
		{"empty key", "", "****"},
		{"exactly 4 chars", "abcd", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskAPIKey(tt.key)
			if got != tt.want {
				t.Errorf("maskAPIKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
