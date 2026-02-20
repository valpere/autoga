package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/val/autoga/internal/useragent"
)

const maxBodyBytes = 5 * 1024 * 1024 // 5 MB

// HTTPFetcher fetches URLs using a shared http.Client with configurable timeout.
type HTTPFetcher struct {
	client *http.Client
}

// NewHTTPFetcher creates an HTTPFetcher with the given per-request timeout.
func NewHTTPFetcher(timeout time.Duration) *HTTPFetcher {
	return &HTTPFetcher{
		client: &http.Client{Timeout: timeout},
	}
}

// unwrapGoogleURL extracts the real destination from Google redirect URLs
// (https://www.google.com/url?...&url=<target>&...). Returns the input unchanged
// if it is not a Google redirect.
func unwrapGoogleURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.Host != "www.google.com" || u.Path != "/url" {
		return raw
	}
	if target := u.Query().Get("url"); target != "" {
		return target
	}
	return raw
}

// Fetch performs an HTTP GET and returns the body, capped at maxBodyBytes.
func (f *HTTPFetcher) Fetch(ctx context.Context, rawURL string) ([]byte, error) {
	target := unwrapGoogleURL(rawURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", useragent.Next())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", target, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, target)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return body, nil
}
