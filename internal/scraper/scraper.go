package scraper

import (
	"context"
	"sync"

	"github.com/val/autoga/internal"
)

// Scraper orchestrates concurrent fetching and extraction of articles.
type Scraper struct {
	fetcher    Fetcher
	extractor  Extractor
	maxWorkers int
}

// New creates a Scraper with the given fetcher, extractor, and concurrency limit.
func New(fetcher Fetcher, extractor Extractor, maxWorkers int) *Scraper {
	return &Scraper{
		fetcher:    fetcher,
		extractor:  extractor,
		maxWorkers: maxWorkers,
	}
}

// Scrape processes urls concurrently and returns one ArticleResult per URL.
// Errors are captured per-URL and never cause the whole operation to fail.
func (s *Scraper) Scrape(ctx context.Context, urls []string) []internal.ArticleResult {
	results := make([]internal.ArticleResult, len(urls))
	sem := make(chan struct{}, s.maxWorkers)
	var wg sync.WaitGroup

	for i, u := range urls {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results[idx] = s.scrapeOne(ctx, url)
		}(i, u)
	}

	wg.Wait()
	return results
}

func (s *Scraper) scrapeOne(ctx context.Context, url string) internal.ArticleResult {
	clean := unwrapGoogleURL(url)

	html, err := s.fetcher.Fetch(ctx, url)
	if err != nil {
		return internal.ArticleResult{URL: clean, Error: err.Error()}
	}

	result, err := s.extractor.Extract(clean, html)
	if err != nil {
		return internal.ArticleResult{URL: clean, Error: err.Error()}
	}

	return result
}
