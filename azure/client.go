package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"time"
)

const (
	calculatorPageURL = "https://azure.microsoft.com/en-us/pricing/calculator/"
	saveEstimateURL   = "https://azure.microsoft.com/api/v2/calculator/shared-estimates/save/"
	userAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

var (
	tokenRegex1 = regexp.MustCompile(`name="__RequestVerificationToken"\s+type="hidden"\s+value="([^"]+)"`)
	tokenRegex2 = regexp.MustCompile(`value="([^"]+)"\s+name="__RequestVerificationToken"`)
)

// Client handles interaction with the Azure Pricing Calculator API
type Client struct {
	httpClient *http.Client
}

// NewClient initializes a new Azure Calculator API client with cookie jar
func NewClient() (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &Client{
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 20 * time.Second,
		},
	}, nil
}

// fetchVerificationToken retrieves the anti-CSRF token and establishes the session
func (c *Client) fetchVerificationToken() (string, error) {
	req, err := http.NewRequest("GET", calculatorPageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GET request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to load calculator page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("calculator page returned unexpected status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	bodyStr := string(bodyBytes)

	match := tokenRegex1.FindStringSubmatch(bodyStr)
	if len(match) > 1 {
		return match[1], nil
	}

	match = tokenRegex2.FindStringSubmatch(bodyStr)
	if len(match) > 1 {
		return match[1], nil
	}

	return "", fmt.Errorf("failed to find __RequestVerificationToken in calculator page HTML")
}

// CreateEstimate generates a shared Azure estimate and returns the official share URL
func (c *Client) CreateEstimate(opts RequestOptions) (*EstimateResult, error) {
	token, err := c.fetchVerificationToken()
	if err != nil {
		return nil, fmt.Errorf("session initialization failed: %w", err)
	}

	payload := NewEstimatePayload(opts.Name, opts.Currency)

	for _, res := range opts.Resources {
		instance := BuildInstance(res)
		payload.Instances = append(payload.Instances, instance)
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload: %w", err)
	}

	req, err := http.NewRequest("POST", saveEstimateURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create save request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("RequestVerificationToken", token)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", calculatorPageURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call save estimate endpoint: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read save estimate response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("save estimate returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	rawID := strings.Trim(string(respBytes), "\"\r\n\t ")
	if rawID == "" {
		return nil, fmt.Errorf("empty estimate ID received from Azure API")
	}

	shortHash := strings.ReplaceAll(rawID, "-", "")

	return &EstimateResult{
		ID:            rawID,
		Hash:          shortHash,
		ShareURL:      fmt.Sprintf("https://azure.com/e/%s", shortHash),
		CalculatorURL: fmt.Sprintf("https://azure.microsoft.com/en-us/pricing/calculator/?shared-estimate=%s", shortHash),
		Name:          payload.Name,
		Currency:      strings.ToUpper(payload.Currency),
		ItemCount:     len(payload.Instances),
	}, nil
}
