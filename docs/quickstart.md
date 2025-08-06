# Quickstart

## Overview
Spin up the authorization service locally using Docker Compose for rapid experimentation.

## When to Use
Ideal for developers evaluating features without installing dependencies.

## Policy Example
Use [examples/rbac.yaml](../examples/rbac.yaml) to allow an admin to read any file.

## API Usage
```sh
# Start the stack
Docker compose up --build

# Load policies and check access
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file1","action":"read"}'
```
Postman: import `postman/authorization-service.postman_collection.json` and run the **Check Access** request.

## CLI Usage
```sh
make build
./authzctl check-access --tenant acme --subject alice --resource file1 --action read
```

## SDK Usage
Go and Python SDKs can target the local instance at `http://localhost:8080`.

## Validation/Testing
Run `go test ./...` to ensure the code compiles and basic tests pass.

## Observability
Metrics are exposed at `http://localhost:8080/metrics`; traces are sent to the configured OTLP endpoint.

## Notes & Caveats
The quickstart uses in-memory stores and is not intended for production workloads.
