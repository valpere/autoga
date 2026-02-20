package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/val/autoga/internal/config"
	"github.com/val/autoga/internal/scraper"
	"github.com/val/autoga/internal/server"
)

func main() {
	cfg := config.Load()

	fetcher := scraper.NewHTTPFetcher(cfg.FetchTimeout)
	extractor := scraper.NewReadabilityExtractor()
	sc := scraper.New(fetcher, extractor, cfg.MaxConcurrency)

	srv := server.New(cfg, sc)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("autoga listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
