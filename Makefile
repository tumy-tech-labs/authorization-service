.PHONY: build run test up down logs oidc-token

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

oidc-token:
	curl -s -X POST \
	  http://localhost:8081/realms/authz-service/protocol/openid-connect/token \
	  -H 'Content-Type: application/x-www-form-urlencoded' \
	  -d 'grant_type=password&client_id=authz-client&username=$(USER)&password=$(PASS)' | jq -r .access_token
