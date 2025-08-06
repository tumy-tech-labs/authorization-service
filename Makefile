.PHONY: build run test up down logs

APP=authorization-service
CLI=authzctl

build:
	go build -o $(APP) ./cmd
	go build -o $(CLI) ./cmd/$(CLI)

run:
	go run ./cmd

test:
	go test ./...

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f
