package scraper

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-shiori/go-readability"

	"github.com/val/autoga/internal"
)

// ReadabilityExtractor uses go-readability to extract article content from HTML.
type ReadabilityExtractor struct{}

// NewReadabilityExtractor creates a ReadabilityExtractor.
func NewReadabilityExtractor() *ReadabilityExtractor {
	return &ReadabilityExtractor{}
}

// sanitize removes characters that break JSON string embedding in Make.com templates:
// double-quotes become single-quotes, backslashes are dropped.
var sanitizeReplacer = strings.NewReplacer(`"`, `'`, `\`, ``)

func sanitize(s string) string { return sanitizeReplacer.Replace(s) }

// Extract parses the HTML and returns an ArticleResult populated with readable content.
func (e *ReadabilityExtractor) Extract(rawURL string, html []byte) (internal.ArticleResult, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return internal.ArticleResult{URL: rawURL}, fmt.Errorf("parse URL: %w", err)
	}

	article, err := readability.FromReader(bytes.NewReader(html), parsed)
	if err != nil {
		return internal.ArticleResult{URL: rawURL}, fmt.Errorf("readability: %w", err)
	}

	return internal.ArticleResult{
		URL:      rawURL,
		Title:    sanitize(article.Title),
		Byline:   sanitize(article.Byline),
		Content:  sanitize(strings.Join(strings.Fields(article.TextContent), " ")),
		Excerpt:  sanitize(article.Excerpt),
		SiteName: sanitize(article.SiteName),
	}, nil
}
