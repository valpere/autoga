.DEFAULT_GOAL := help
.PHONY: build run health scrape docker-build docker-up docker-down tidy deploy-scenario help

BIN := bin/autoga
PORT ?= 8080

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
	@echo "Make.com"
	@echo "  deploy-scenario    Create scenario via Make.com API (requires .env)"
	@echo ""
	@echo "Variables"
	@echo "  PORT               HTTP port for run/health/scrape (default: $(PORT))"

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
	docker compose up --build -d

docker-down:
	docker compose down

deploy-scenario:
	go run ./cmd/makesetup
