.PHONY: tidy run sqlc migrate-status migrate-up migrate-down docker-up docker-down

MODULE := github.com/zhazhahehedian/ai-fusion-video-go
HTTP_ADDR ?= :18080
DATABASE_URL ?= fusion:123456@tcp(127.0.0.1:43306)/ai_fusion_video?parseTime=true&loc=Local
REDIS_ADDR ?= 127.0.0.1:46379
REDIS_PASSWORD ?= 123456

tidy:
	go mod tidy

run:
	HTTP_ADDR=$(HTTP_ADDR) DATABASE_URL='$(DATABASE_URL)' REDIS_ADDR=$(REDIS_ADDR) REDIS_PASSWORD=$(REDIS_PASSWORD) go run ./cmd/server

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
