package internal

// ScrapeRequest is the incoming payload for POST /scrape.
type ScrapeRequest struct {
	URLs []string `json:"urls"`
}

// ArticleResult holds the extracted content for a single URL.
type ArticleResult struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Byline   string `json:"byline"`
	Content  string `json:"content"`
	Excerpt  string `json:"excerpt"`
	SiteName string `json:"site_name"`
	Error    string `json:"error"`
}

// ScrapeResponse is the outgoing payload for POST /scrape.
type ScrapeResponse struct {
	Results []ArticleResult `json:"results"`
}
