package server

import (
	"encoding/json"
	"net/http"

	"github.com/val/autoga/internal"
	"github.com/val/autoga/internal/scraper"
)

type scrapeHandler struct {
	scraper           *scraper.Scraper
	maxURLsPerRequest int
}

func (h *scrapeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req internal.ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if len(req.URLs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "urls must not be empty"})
		return
	}

	if len(req.URLs) > h.maxURLsPerRequest {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "too many URLs",
		})
		return
	}

	results := h.scraper.Scrape(r.Context(), req.URLs)
	writeJSON(w, http.StatusOK, internal.ScrapeResponse{Results: results})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
