package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pluckware/tyvt/internal/limiter"
	"github.com/pluckware/tyvt/internal/rotator"
)

const (
	VirusTotalAPIURL = "https://virustotal.com/vtapi/v2/domain/report"
	DefaultTimeout   = 30 * time.Second
)

type VirusTotalClient struct {
	httpClient  *http.Client
	keyRotator  *rotator.KeyRotator
	ipRotator   *rotator.IPRotator
	rateLimiter *limiter.RateLimiter
}

type DomainResult struct {
	Domain         string                 `json:"domain"`
	ResponseCode   int                    `json:"response_code"`
	UndetectedURLs []UndetectedURL        `json:"undetected_urls,omitempty"`
	RawResponse    map[string]interface{} `json:"raw_response,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

type UndetectedURL struct {
	URL          string    `json:"url"`
	Positives    int       `json:"positives"`
	Total        int       `json:"total"`
	ScanDate     string    `json:"scan_date"`
	LastModified time.Time `json:"last_modified"`
}

func NewVirusTotalClient(keyRotator *rotator.KeyRotator, ipRotator *rotator.IPRotator, rateLimiter *limiter.RateLimiter) *VirusTotalClient {
	transport := &http.Transport{}

	if ipRotator != nil {
		transport.Proxy = ipRotator.ProxyFunc()
	}

	return &VirusTotalClient{
		httpClient: &http.Client{
			Timeout:   DefaultTimeout,
			Transport: transport,
		},
		keyRotator:  keyRotator,
		ipRotator:   ipRotator,
		rateLimiter: rateLimiter,
	}
}

func (c *VirusTotalClient) QueryDomain(ctx context.Context, domain string) (*DomainResult, error) {
	apiKey := c.keyRotator.CurrentKey()
	if apiKey == "" {
		return nil, fmt.Errorf("no API key available")
	}

	if err := c.rateLimiter.Wait(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	reqURL := fmt.Sprintf("%s?apikey=%s&domain=%s", VirusTotalAPIURL, url.QueryEscape(apiKey), url.QueryEscape(domain))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "tyvt/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	result := &DomainResult{
		Domain:      domain,
		RawResponse: rawResponse,
		Timestamp:   time.Now(),
	}

	if responseCode, ok := rawResponse["response_code"].(float64); ok {
		result.ResponseCode = int(responseCode)
	}

	if result.ResponseCode != 1 {
		return result, nil
	}

	if err := c.parseUndetectedURLs(rawResponse, result); err != nil {
		return result, fmt.Errorf("failed to parse undetected URLs: %w", err)
	}

	return result, nil
}

func (c *VirusTotalClient) parseUndetectedURLs(rawResponse map[string]interface{}, result *DomainResult) error {
	undetectedInterface, exists := rawResponse["undetected_urls"]
	if !exists {
		return nil
	}

	undetectedArray, ok := undetectedInterface.([]interface{})
	if !ok {
		return nil
	}

	for _, item := range undetectedArray {
		urlData, ok := item.([]interface{})
		if !ok || len(urlData) < 4 {
			continue
		}

		url, ok := urlData[0].(string)
		if !ok {
			continue
		}

		positives, ok := urlData[1].(float64)
		if !ok {
			continue
		}

		total, ok := urlData[2].(float64)
		if !ok {
			continue
		}

		scanDate, ok := urlData[3].(string)
		if !ok {
			continue
		}

		undetectedURL := UndetectedURL{
			URL:          url,
			Positives:    int(positives),
			Total:        int(total),
			ScanDate:     scanDate,
			LastModified: time.Now(),
		}

		result.UndetectedURLs = append(result.UndetectedURLs, undetectedURL)
	}

	return nil
}