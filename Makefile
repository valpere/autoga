.DEFAULT_GOAL := help
.PHONY: build run health scrape docker-build docker-up docker-down tidy deploy-scenario gcloud-deploy help

BIN           := bin/autoga
PORT          ?= 8080
GCLOUD_REGION ?= europe-central2

-include .env
export

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Development"
	@echo "  build              Compile binary to $(BIN)"
	@echo "  run                Build and run locally (PORT=$(PORT))"
	@echo "  tidy               Run go mod tidy"
	@echo ""
	@echo "Testing"
	@echo "  health             GET /health on localhost:$(PORT)"
	@echo "  scrape             POST /scrape with a sample URL"
	@echo ""
	@echo "Docker"
	@echo "  docker-build       Build Docker image"
	@echo "  docker-up          Start via docker compose (detached)"
	@echo "  docker-down        Stop docker compose"
	@echo ""
	@echo "Deploy"
	@echo "  deploy-scenario    Create Make.com scenario (requires .env)"
	@echo "  gcloud-deploy      Deploy to Cloud Run (requires gcloud auth)"
	@echo ""
	@echo "Variables"
	@echo "  PORT               HTTP port for run/health/scrape (default: $(PORT))"
	@echo "  GCLOUD_REGION      Cloud Run region (default: $(GCLOUD_REGION))"

build:
	@mkdir -p bin
	go build -o $(BIN) ./cmd/autoga

run: build
	PORT=$(PORT) ./$(BIN)

tidy:
	go mod tidy

health:
	curl -s http://localhost:$(PORT)/health | jq .

scrape:
	curl -s -X POST http://localhost:$(PORT)/scrape \
		-H 'Content-Type: application/json' \
		-d '{"urls": ["https://go.dev/blog/go1.24"]}' | jq .

docker-build:
	docker build -f docker/Dockerfile -t autoga .

docker-up:
	docker compose -f docker/docker-compose.yml up --build -d

docker-down:
	docker compose -f docker/docker-compose.yml down

deploy-scenario:
	go run ./cmd/makesetup

gcloud-deploy:
	gcloud builds submit --config docker/cloudbuild.yaml .
	gcloud run deploy autoga \
		--image gcr.io/$$(gcloud config get-value project)/autoga \
		--platform managed \
		--region $(GCLOUD_REGION) \
		--allow-unauthenticated \
		--memory 256Mi \
		--cpu 1 \
		--min-instances 0 \
		--max-instances 2 \
		--set-env-vars "READ_TIMEOUT=5s,WRITE_TIMEOUT=60s,FETCH_TIMEOUT=15s,MAX_CONCURRENCY=5,MAX_URLS_PER_REQUEST=10" \
		--update-secrets "API_KEY=autoga-api-key:latest"
