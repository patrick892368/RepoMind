package ai

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const providerHTTPTimeout = 60 * time.Second

type providerHTTPError struct {
	Provider   string
	StatusCode int
	Status     string
	Body       string
}

func (e providerHTTPError) Error() string {
	provider := strings.TrimSpace(e.Provider)
	if provider == "" {
		provider = "provider"
	}
	return fmt.Sprintf("%s api returned %s: %s", provider, e.Status, e.Body)
}

func providerHTTPClient(envPath string) (*http.Client, error) {
	proxyValue := lookupEnvValue(envPath, "HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy", "HTTP_PROXY", "http_proxy")
	if proxyValue == "" {
		return &http.Client{Timeout: providerHTTPTimeout}, nil
	}

	proxyURL, err := url.Parse(proxyValue)
	if err != nil {
		return nil, fmt.Errorf("parse proxy URL: %w", err)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyURL(proxyURL)
	return &http.Client{Timeout: providerHTTPTimeout, Transport: transport}, nil
}

func doProviderRequest(client HTTPDoer, providerName string, req *http.Request) ([]byte, error) {
	if client == nil {
		client = &http.Client{Timeout: providerHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", providerName, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, providerHTTPError{
			Provider:   providerName,
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       truncateForError(string(raw), 500),
		}
	}
	return raw, nil
}
