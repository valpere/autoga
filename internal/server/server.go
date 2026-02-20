package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/val/autoga/internal/config"
	"github.com/val/autoga/internal/scraper"
)

// New builds and returns an http.Server wired with all routes and middleware.
func New(cfg config.Config, sc *scraper.Scraper) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByIP(60, time.Minute))

	r.Get("/health", healthHandler)
	r.Post("/scrape", (&scrapeHandler{
		scraper:           sc,
		maxURLsPerRequest: cfg.MaxURLsPerRequest,
	}).ServeHTTP)

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
}
