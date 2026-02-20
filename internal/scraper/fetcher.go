package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

// Fetch performs an HTTP GET and returns the body, capped at maxBodyBytes.
func (f *HTTPFetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", useragent.Next())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return body, nil
}
