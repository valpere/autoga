# CLAUDE.md

Project-specific instructions for Claude Code.

## What this project is

A Go HTTP service that receives URLs and returns clean article text (using Mozilla Readability). It is one component in a Make.com automation: Google Alerts RSS → autoga → LLM → Telegram.

autoga has no knowledge of LLM or Telegram — that is Make.com's responsibility.

## Commands

```bash
make build            # compile ./autoga binary
make run              # build + run on :8080
make tidy             # go mod tidy
make health           # curl /health
make scrape           # curl /scrape with sample URL
make docker-up        # start via docker compose
make deploy-scenario  # deploy Make.com scenario (requires .env)
```

## Architecture

Two binaries:
- `cmd/autoga` — the scraper HTTP service
- `cmd/makesetup` — one-shot CLI that deploys the Make.com scenario

Key packages:
- `internal/scraper` — `Fetcher` and `Extractor` interfaces with `HTTPFetcher` + `ReadabilityExtractor` implementations; concurrent orchestration via semaphore pattern
- `internal/server` — Chi router; auth middleware in `middleware.go`; handlers in `handlers.go`
- `internal/makecom` — Make.com API client + blueprint builder for the RSS→scrape→LLM→Telegram scenario
- `internal/config` — all config from env vars, no config files
- `internal/useragent` — atomic round-robin UA rotation, no mutex

## Key decisions

- **Per-URL errors**: failed URLs return `error` field, never break the whole request
- **API key auth**: disabled when `API_KEY` env is empty (dev-friendly); `/health` is always open
- **LLM agnostic**: Go service doesn't call LLM — Make.com does
- **Blueprint template vars**: Make.com `{{moduleId.field}}` syntax is embedded as literal strings in the blueprint JSON; `results[1]` is 1-indexed per Make.com convention
- **Body size cap**: 5 MB via `io.LimitReader` in fetcher

## Dependencies

| Module | Purpose |
|--------|---------|
| `github.com/go-chi/chi/v5` | HTTP router + middleware |
| `github.com/go-chi/httprate` | Rate limiting |
| `github.com/go-shiori/go-readability` | Article text extraction |

No external dependencies in `internal/makecom` — uses stdlib `net/http` only.

## Known limitations

- Make.com `{{...}}` template vars in the LLM request body are interpolated as raw strings. Article content containing `"` or `\` may produce malformed JSON. Mitigation: add a `json:CreateJSON` module in Make.com UI between scraper and LLM steps.
- Telegram connection must be set manually in Make.com UI after `deploy-scenario` — connections cannot be created via Make.com API.
