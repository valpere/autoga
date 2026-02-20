package scraper

import (
	"context"

	"github.com/val/autoga/internal"
)

// Fetcher retrieves raw HTML content for a given URL.
type Fetcher interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}

// Extractor parses raw HTML and returns a populated ArticleResult.
type Extractor interface {
	Extract(url string, html []byte) (internal.ArticleResult, error)
}
