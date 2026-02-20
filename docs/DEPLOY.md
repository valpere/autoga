# AutoGA — Deployment & Configuration Guide

## Overview

AutoGA has two binaries:

| Binary | Purpose |
|--------|---------|
| `cmd/autoga` | HTTP scraper service — receives URLs, returns clean article text |
| `cmd/makesetup` | One-shot CLI — deploys Make.com scenarios from `.env` + `.rss_feeds.csv` |

The full automation pipeline:

```
Google Alerts RSS
  → Make.com: RSS module (every N minutes)
  → Make.com: HTTP POST → autoga /scrape (fetch + extract article text)
  → Make.com: HTTP POST → Ollama Cloud (LLM digest in Ukrainian)
  → Make.com: Telegram Bot (publish to channel)
```

---

## Prerequisites

- Go 1.25+
- `gcloud` CLI authenticated (`gcloud auth login`)
- A Google Cloud project with billing enabled
- Make.com account (free tier works)
- Ollama Cloud account with API key
- Telegram Bot with a channel

---

## 1. Environment Configuration

Copy the example and fill in all values:

```bash
cp .env.example .env
```

### `.env` reference

```bash
# ── autoga scraper service ─────────────────────────────────────────────────────
PORT=8080
API_KEY=<random hex string>        # Bearer token for /scrape. Generate with:
                                   # openssl rand -hex 24

READ_TIMEOUT=5s
WRITE_TIMEOUT=60s
FETCH_TIMEOUT=15s                  # per-URL fetch timeout

MAX_CONCURRENCY=5                  # parallel fetch workers
MAX_URLS_PER_REQUEST=10            # max URLs in a single /scrape call

# ── Make.com scenario deploy ───────────────────────────────────────────────────
MAKE_API_TOKEN=                    # Make.com → Profile → API Access
MAKE_TEAM_ID=                      # Make.com URL: /team/<ID>/...
MAKE_ZONE=eu1                      # eu1 or us1
MAKE_FOLDER_ID=                    # optional: place scenario in a folder
MAKE_INTERVAL_SEC=900              # polling interval (seconds). Default: 15 min

SCENARIO_NAME=AutoGA Digest        # prefix for scenario names

RSS_FEEDS=.rss_feeds.csv           # path to RSS feeds CSV

AUTOGA_URL=https://your-service.run.app   # public URL of the deployed scraper
AUTOGA_API_KEY=<same as API_KEY>

# ── Ollama Cloud LLM ───────────────────────────────────────────────────────────
LLM_API_URL=https://ollama.com/api/chat
LLM_API_KEY=                       # from https://ollama.com/settings/keys
LLM_MODEL=glm-5                    # cloud models: glm-5, qwen3.5, kimi-k2.5

# ── Telegram ───────────────────────────────────────────────────────────────────
TELEGRAM_CHAT_ID=                  # channel ID, e.g. -1001234567890
                                   # (add bot as admin, send a message, check
                                   # https://api.telegram.org/bot<token>/getUpdates)
TELEGRAM_CONNECTION_ID=            # Make.com connection ID (see step 4 below)
```

### `.rss_feeds.csv`

One Google Alerts feed per row — `name, url`:

```csv
# name, url
Artificial Intelligence,https://www.google.com/alerts/feeds/<id>/<feed>
Штучний Інтелект,https://www.google.com/alerts/feeds/<id>/<feed>
```

Get feed URLs: [alerts.google.com](https://alerts.google.com) → edit alert → **Deliver to** → **RSS feed** → copy URL.

Each row in this file becomes one Make.com scenario named `<SCENARIO_NAME>: <name>`.

---

## 2. Local Development & Testing

### Run the service

```bash
# No API key = auth disabled (dev-friendly)
API_KEY= make run
# or with auth:
make run
```

Service starts on `PORT` from `.env` (default 8080).

### Health check

```bash
make health
# {"status":"ok"}
```

### Scrape a URL

```bash
make scrape
# tests with https://go.dev/blog/go1.24
```

Custom URL:

```bash
curl -s -X POST http://localhost:8080/scrape \
  -H 'Content-Type: application/json' \
  -d '{"urls": ["https://example.com/article"]}' | jq .
```

With auth enabled:

```bash
curl -s -X POST http://localhost:8080/scrape \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <API_KEY>' \
  -d '{"urls": ["https://example.com/article"]}' | jq .
```

### Response format

```json
{
  "results": [
    {
      "url":       "https://example.com/article",
      "title":     "Article Title",
      "byline":    "Author Name",
      "content":   "Full plain text of the article...",
      "excerpt":   "Short excerpt...",
      "site_name": "Example Site",
      "error":     ""
    }
  ]
}
```

Errors are per-URL — a failed URL returns `error` field, the rest still succeed.

### Run via Docker

```bash
make docker-up    # build + start (detached)
make docker-down  # stop
```

---

## 3. Deploy to Google Cloud Run

### First-time setup

```bash
# Authenticate
gcloud auth login
gcloud config set project <YOUR_PROJECT_ID>

# Enable required APIs
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  secretmanager.googleapis.com

# Store API key in Secret Manager
echo -n "<your API_KEY value>" | \
  gcloud secrets create autoga-api-key --data-file=-

# Grant Cloud Run access to the secret
PROJECT_NUMBER=$(gcloud projects describe $(gcloud config get-value project) \
  --format='value(projectNumber)')
gcloud secrets add-iam-policy-binding autoga-api-key \
  --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### Deploy

```bash
make gcloud-deploy
# or with a different region:
GCLOUD_REGION=us-central1 make gcloud-deploy
```

This runs two steps:

```bash
# 1. Build image via Cloud Build (uses docker/Dockerfile, config in docker/cloudbuild.yaml)
gcloud builds submit --config docker/cloudbuild.yaml .

# 2. Deploy the built image to Cloud Run
gcloud run deploy autoga \
  --image gcr.io/<PROJECT_ID>/autoga \
  --platform managed \
  --region europe-central2 \
  --allow-unauthenticated \
  --memory 256Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 2 \
  --set-env-vars "READ_TIMEOUT=5s,WRITE_TIMEOUT=60s,FETCH_TIMEOUT=15s,MAX_CONCURRENCY=5,MAX_URLS_PER_REQUEST=10" \
  --update-secrets "API_KEY=autoga-api-key:latest"
```

The built image is stored in Container Registry as `gcr.io/<PROJECT_ID>/autoga:latest`.

After deploy, copy the **Service URL** into `AUTOGA_URL` in `.env`.

### Update the API key secret

```bash
echo -n "<new key>" | gcloud secrets versions add autoga-api-key --data-file=-
# then redeploy so the new version is picked up
make gcloud-deploy
```

### Notes

- `--min-instances 0` scales to zero when idle (free tier friendly). Cold starts add ~2–3 s latency — the Make.com HTTP module timeout is set to 90 s to compensate.
- `/health` is always unauthenticated; `/scrape` requires `Authorization: Bearer <API_KEY>` when `API_KEY` is set.
- Logs: `gcloud run logs tail autoga --region europe-central2`

---

## 4. Set Up Make.com

### Get credentials

| Value | Where to find |
|-------|--------------|
| `MAKE_API_TOKEN` | Make.com → avatar → **Profile** → **API Access** → Generate token |
| `MAKE_TEAM_ID` | Make.com URL bar: `eu1.make.com/team/<ID>/...` |
| `MAKE_ZONE` | `eu1` or `us1` depending on your account region |

### Create a Telegram connection

Make.com connections cannot be created via API — do it manually once:

1. Make.com → **Connections** → **Add connection** → Telegram Bot
2. Paste your bot token
3. After saving, open the connection and note the numeric ID from the URL:
   `eu1.make.com/connections/<CONNECTION_ID>/edit`
4. Put that ID in `TELEGRAM_CONNECTION_ID` in `.env`

### Deploy scenarios

```bash
make deploy-scenario
```

This creates one Make.com scenario per row in `.rss_feeds.csv`, each named `<SCENARIO_NAME>: <feed name>`.

Output:

```
created: id=1234567 name="AutoGA Digest: Artificial Intelligence"
  edit: https://eu1.make.com/scenario/1234567/edit
```

### Activate scenarios

Newly created scenarios are inactive. For each scenario:

1. Open the edit link printed by `make deploy-scenario`
2. On the Telegram module, select your Telegram Bot connection from the dropdown
3. Click the **ON/OFF** toggle at the bottom of the screen to activate

The scenario will now run every `MAKE_INTERVAL_SEC` seconds (default 15 minutes).

### Re-deploying scenarios

`make deploy-scenario` always creates new scenarios — it does not update existing ones. Delete old scenarios in Make.com UI before re-running if you want to avoid duplicates.

---

## 5. Ollama Cloud

1. Sign up at [ollama.com](https://ollama.com)
2. Go to **Settings → API Keys** → create a key
3. Set in `.env`:
   ```
   LLM_API_URL=https://ollama.com/api/chat
   LLM_API_KEY=<your key>
   LLM_MODEL=glm-5
   ```

Available cloud models (as of Feb 2026): `glm-5`, `qwen3.5`, `kimi-k2.5`.

---

## 6. Makefile Quick Reference

```
make build            Compile binary to bin/autoga
make run              Build and run locally (PORT from .env)
make tidy             go mod tidy
make health           GET /health on localhost:PORT
make scrape           POST /scrape with sample URL
make docker-build     Build Docker image
make docker-up        Start via docker compose (detached)
make docker-down      Stop docker compose
make deploy-scenario  Create Make.com scenarios (requires .env)
make gcloud-deploy    Deploy to Cloud Run (requires gcloud auth)
```

---

## 7. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `HTTP 401` from `/scrape` | Missing or wrong `Authorization` header | Add `Authorization: Bearer <API_KEY>` header |
| Empty `content` in response | Google Alerts URL not unwrapped | Should be automatic; check `error` field for fetch errors |
| `DataError: invalid character` in Make.com LLM step | Article content contained `"` or `\` | Fixed in scraper — redeploy with `make gcloud-deploy` |
| `timeout of 40000ms exceeded` in Make.com | Cloud Run cold start + slow article | HTTP module timeout is set to 90 s; consider `--min-instances 1` for always-warm |
| `missing separator` in Makefile | Text accidentally pasted into `.env` | Check `.env` for stray lines at the end |
| Telegram module shows no connection | Connection must be set manually in Make.com UI | See step 4 above |
| `required env var X is not set` | Missing `.env` value | Fill in all required fields in `.env` |
