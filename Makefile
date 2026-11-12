.PHONY: tidy run sqlc migrate-status migrate-up migrate-down docker-up docker-down

MODULE := github.com/zhazhahehedian/ai-fusion-video-go
HTTP_ADDR ?= :18080

tidy:
	go mod tidy

run:
	HTTP_ADDR=$(HTTP_ADDR) go run ./cmd/server

sqlc:
	sqlc generate -f db/sqlc.yaml

migrate-status:
	goose -dir db/migrations mysql "$(DATABASE_URL)" status

migrate-up:
	goose -dir db/migrations mysql "$(DATABASE_URL)" up

migrate-down:
	goose -dir db/migrations mysql "$(DATABASE_URL)" down

docker-up:
	docker compose up -d

docker-down:
	docker compose down
