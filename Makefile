.PHONY: build run health scrape docker-build docker-up docker-down tidy

BIN := autoga
PORT ?= 8080

build:
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
