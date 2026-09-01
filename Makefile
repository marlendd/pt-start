GOLANGCI_LINT_VERSION := v2.12.2

.PHONY: fmt test test-race vet lint check run up down logs ps config

fmt:
	go fmt ./...

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

lint:
	docker run --rm \
		--volume "$(CURDIR):/app:ro" \
		--workdir /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION)-alpine \
		golangci-lint run

check: fmt test-race vet lint

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
