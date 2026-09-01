.PHONY: fmt test test-race vet check run up down logs ps config

fmt:
	go fmt ./...

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

check: fmt test-race vet

run:
	set -a; . ./.env; set +a; go run ./cmd/shortener

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps -a

config:
	docker compose config --quiet
