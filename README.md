# autoga

Article scraper service for Google Alerts automation.

## Overview

```
Google Alerts RSS → Make.com → autoga /scrape → LLM API → Telegram
```

Make.com orchestrates the full flow. autoga's sole responsibility is receiving URLs and returning clean article text.

## API

### `POST /scrape`

```bash
curl -X POST http://localhost:8080/scrape \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer your-api-key' \
  -d '{"urls": ["https://example.com/article"]}'
```

```json
{
  "results": [
    {
      "url": "https://example.com/article",
      "title": "Article title",
      "byline": "Author Name",
      "content": "Full article text...",
      "excerpt": "Short summary...",
      "site_name": "Example",
      "error": ""
    }
  ]
}
```

Errors are per-URL — a failed URL does not affect others. The response is always `200`.

### `GET /health`

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

No authentication required.

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `API_KEY` | _(none)_ | Bearer token for `/scrape`. Auth disabled if empty |
| `READ_TIMEOUT` | `5s` | Server read timeout |
| `WRITE_TIMEOUT` | `60s` | Server write timeout |
| `FETCH_TIMEOUT` | `15s` | Per-URL fetch timeout |
| `MAX_CONCURRENCY` | `5` | Parallel scraping workers |
| `MAX_URLS_PER_REQUEST` | `10` | Max URLs per request |

## Running

```bash
# Local
make run

# Docker
make docker-up
```

See `.env.example` for all available variables.

## Make.com scenario

The `cmd/makesetup` CLI deploys the Make.com scenario via API:

```bash
cp .env.example .env
# fill in .env
source .env
make deploy-scenario
```

**Scenario flow:**

```
1. RSS watch          — polls Google Alerts RSS feed every 15 min
2. HTTP POST /scrape  — fetches full article text via autoga
3. HTTP POST LLM API  — generates Ukrainian digest (OpenAI-compatible API)
4. Telegram message   — publishes digest to channel
```

**One manual step:** after `deploy-scenario`, open the scenario in Make.com UI and set the Telegram bot connection.

## Development

```bash
make build            # compile
make run              # build + run on :8080
make health           # test /health
make scrape           # test /scrape with a sample URL
make tidy             # go mod tidy
```

## Project structure

```
cmd/
  autoga/       — scraper service entry point
  makesetup/    — Make.com scenario deploy CLI
internal/
  config/       — env-based configuration
  makecom/      — Make.com API client and blueprint builder
  scraper/      — fetcher, extractor, concurrent orchestrator
  server/       — HTTP server, router, middleware, handlers
  useragent/    — round-robin User-Agent rotation
  types.go      — shared request/response types
docker/
  Dockerfile    — multi-stage build
```
