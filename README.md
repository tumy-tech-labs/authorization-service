# Authorization Service

[![CI](https://github.com/bradtumy/authorization-service/actions/workflows/ci.yml/badge.svg)](https://github.com/bradtumy/authorization-service/actions/workflows/ci.yml)
[![Release](https://github.com/bradtumy/authorization-service/actions/workflows/release.yml/badge.svg)](https://github.com/bradtumy/authorization-service/actions/workflows/release.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

[![Go Reference](https://pkg.go.dev/badge/github.com/bradtumy/authorization-service.svg)](https://pkg.go.dev/github.com/bradtumy/authorization-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/bradtumy/authorization-service)](https://goreportcard.com/report/github.com/bradtumy/authorization-service)

Authorization Service is an open-source authorization service that reads policies in a simple CDL and provides authorization decisions based on the information provided.

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Flows](docs/flows.md)
- [API Reference](docs/api.md)
- [Roadmap](ROADMAP.md)

## Quickstart

Run the service locally with Go:

```sh
git clone https://github.com/bradtumy/authorization-service.git
cd authorization-service
go run cmd/main.go
```

Or launch with Docker:

```sh
docker compose up --build
```

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Docker (optional for containerized deployment)

### Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/bradtumy/authorization-service.git
   cd authorization-service
   ```

2. Set up the `.env` file in the project root with the following variables:

   ```sh
   CLIENT_ID=my-client-id
   CLIENT_SECRET=my-client-secret
   PORT=8080
   STORE_BACKEND=memory # or sqlite
   OIDC_ISSUERS=https://issuer.example.com
   OIDC_AUDIENCES=my-client-id
   LOG_LEVEL=info
  ```

## Usage

### authzctl CLI

The repository ships with a small CLI for interacting with the service. Run `make build`
to compile both the server and the CLI. The `authzctl` binary can read configuration
from a `.env` file or via flags.

Examples:

```sh
# create and delete tenants
./authzctl tenant create my-tenant
./authzctl tenant delete my-tenant

# validate a policy file offline
./authzctl policy validate path/to/policy.yaml

# dry-run access check
./authzctl check-access --tenant default --subject user1 --resource file1 --action read
```

Set `AUTHZCTL_ADDR` and `AUTHZCTL_TOKEN` in the environment or use the
`--addr` and `--token` flags to point the CLI at a remote service.

### API Endpoints

All requests must include a `tenantID` in the JSON body to scope operations.

| Method | Endpoint        | Description                                     |
| ------ | --------------- | ----------------------------------------------- |
| POST   | `/check-access` | Evaluate a tenant-scoped access request         |
| POST   | `/reload`       | Reload policies for a specific tenant from disk |
| POST   | `/compile`      | Convert natural language to YAML for a tenant   |
| POST   | `/tenant/create`| Register a new tenant                            |
| POST   | `/tenant/delete`| Remove an existing tenant                        |
| GET    | `/tenant/list`  | List all tenants                                |

### Context Providers

Runtime context can influence policy decisions. The service includes a pluggable
framework where each provider implements:

```go
type ContextProvider interface {
    GetContext(req *http.Request) (map[string]string, error)
}
```

The following providers are enabled by default:

- **TimeProvider** – sets a `business_hours` flag based on the current time.
- **GeoIPProvider** – extracts the remote IP and returns a stubbed country code.
- **RiskProvider** – reads a static score from the `X-Risk-Score` header.

Context from all providers is merged with request conditions, passed into policy
evaluations, and recorded as tracing attributes. Additional providers can be
added by implementing the interface and registering them in `api/api.go`.

### Multi-Tenant Example

Each tenant references its own policy file and access checks are scoped by the `tenantID` field. The following example creates two tenants and demonstrates isolated access evaluations:

```sh
# create tenants
./authzctl tenant create acme
./authzctl tenant create globex

# load or reload policies for each tenant
curl -H "Authorization: Bearer <TOKEN>" -H "Content-Type: application/json" \
  -d '{"tenantID":"acme"}'   http://localhost:8080/reload
curl -H "Authorization: Bearer <TOKEN>" -H "Content-Type: application/json" \
  -d '{"tenantID":"globex"}' http://localhost:8080/reload

# check access within each tenant
./authzctl check-access --tenant acme --subject alice --resource file1 --action read
./authzctl check-access --tenant globex --subject bob --resource file1 --action read
```

Requests for one tenant will not evaluate policies from another tenant.

### SDK Usage

#### Go

```go
client := sdk.NewClient("http://localhost:8080")
decision, err := client.CheckAccess(sdk.AccessRequest{
    TenantID: "default",
    Subject:  "alice",
    Resource: "file1",
    Action:   "read",
})
if err != nil {
    panic(err)
}
fmt.Println(decision.Allow)
```

#### Python

```python
from authorization import Client

client = Client("http://localhost:8080")
decision = client.check_access("default", "alice", "file1", "read")
print(decision["allow"])
```

### Logging

The service emits structured JSON logs for every request. Each log entry contains the
following fields:

- `timestamp` – time the log entry was recorded (UTC)
- `correlation_id` – UUID generated per request
- `tenant_id` – tenant associated with the action
- `subject` – authenticated user or subject
- `action` – operation being performed
- `resource` – resource acted upon
- `decision` – allow/deny/success/error outcome
- `policy_id` – policy involved in the decision
- `reason` – additional context

Example access log:

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "info",
  "correlation_id": "123e4567-e89b-12d3-a456-426614174000",
  "tenant_id": "default",
  "subject": "user1",
  "action": "read",
  "resource": "file1",
  "decision": "allow",
  "policy_id": "policy1",
  "reason": "matched rule"
}
```

Example error log:

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "error",
  "correlation_id": "123e4567-e89b-12d3-a456-426614174000",
  "action": "tenant_list",
  "reason": "failed to list tenants"
}
```

#### Metrics

Prometheus metrics are exposed at the `/metrics` endpoint. Configure Prometheus with a
scrape job pointing at the service's port:

```yaml
scrape_configs:
  - job_name: "authz"
    static_configs:
      - targets: ["localhost:8080"]
```

The exporter publishes `http_requests_total`, `http_request_duration_seconds`, and
`policy_eval_count` metrics.

#### Validating Metrics Locally

1. Start the service:

```sh
POLICY_FILE=configs/policies.yaml go run cmd/main.go
```

2. In another terminal, issue a request and inspect the metrics:

```sh
curl -H 'Authorization: Bearer test' \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"default","subject":"user1","resource":"file1","action":"read","conditions":{}}' \
  http://localhost:8080/check-access
curl -H 'Authorization: Bearer test' http://localhost:8080/metrics
```

The metrics output will show counters such as `http_requests_total` and
`policy_eval_count`, along with latency histograms for each path.

#### Tracing

Distributed traces are emitted via OpenTelemetry. Run a local Jaeger instance and point
the service at it using the OTLP endpoint:

```sh
docker run -d -p 4318:4318 -p 16686:16686 jaegertracing/all-in-one
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 go run cmd/main.go
```

Traces will be visible at `http://localhost:16686`. Any OTLP-compatible backend such as
Grafana Tempo can be used by adjusting the endpoint.

Set `LOG_LEVEL` to control verbosity (`debug`, `info`, `warn`, `error`). The default level
is `info`. Sensitive values such as secrets or raw tokens are deliberately omitted from
all logs.

### OIDC Configuration

Tokens presented to the service are validated against one or more OpenID Connect issuers. Providers can be configured via environment variables or a YAML file.

**Environment variables**

```
OIDC_ISSUERS=http://localhost:8080/realms/master
OIDC_AUDIENCES=account
```

**YAML file**

```
providers:
  - issuer: http://localhost:8080/realms/master
    audience: account
```

The middleware fetches each issuer's JWKS and caches the keys with automatic expiry, allowing key rotation without service restarts. JWKS documents are refreshed automatically at a configurable interval and whenever tokens with new key IDs are encountered, so new signing keys are picked up without downtime.

#### Keycloak example

For Keycloak running locally:

1. Create a client in your realm and note its `client_id`.
2. Obtain a token using the client credentials flow:

```
curl -d 'client_id=account' -d 'client_secret=<secret>' \
  http://localhost:8080/realms/master/protocol/openid-connect/token
```

Use the returned `access_token` as the Bearer token when calling the service.

```sh
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/metrics
```

#### Request Policy Decision

Use an access token issued by your OIDC provider to request a policy decision from the authorization service.

1. Start the server:

   ```sh
   go run cmd/main.go
   ```

2. Send a POST request to the `/check-access` endpoint:

   ```sh
   curl -X POST http://localhost:8080/check-access \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer <JWT>" \
       -d '{
           "tenantID": "default",
           "subject": "user1",
           "resource": "file1",
           "action": "read",
           "conditions": {}
       }'
   ```

3. The service will respond with the policy decision:

   ```json
   {
     "allow": true,
     "policy_id": "policy1",
     "reason": "allowed by policy",
     "context": {
       "subject": "user1",
       "resource": "file1",
       "action": "read"
     }
   }
   ```

#### Testing Tenant-Aware Checks

Include a `tenantID` with each request to scope policy evaluations. Different tenants can
maintain separate policy files. To verify isolation between tenants:

```sh
# Tenant A
curl -X POST http://localhost:8080/check-access \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <JWT>" \
    -d '{"tenantID":"tenantA","subject":"alice","resource":"file1","action":"read"}'

# Tenant B
curl -X POST http://localhost:8080/check-access \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <JWT>" \
    -d '{"tenantID":"tenantB","subject":"alice","resource":"file1","action":"read"}'
```

Each tenant receives a decision based solely on its own policies.

#### Tenant Lifecycle Management

Use the API or the `policyctl` CLI to create, list, and delete tenants.

**API examples:**

```sh
curl -X POST http://localhost:8080/tenant/create \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <JWT>" \
    -d '{"tenantID":"acme","name":"Acme Inc"}'

curl -H "Authorization: Bearer <JWT>" http://localhost:8080/tenant/list

curl -X POST http://localhost:8080/tenant/delete \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <JWT>" \
    -d '{"tenantID":"acme"}'
```

**CLI examples:**

```sh
export POLICYCTL_TOKEN=<JWT>
policyctl tenant create acme
policyctl tenant list
policyctl tenant delete acme
```

### Quickstart: Multi-Tenant Example

Run two isolated tenants locally.

#### Using Docker

```sh
docker compose up -d
export POLICYCTL_TOKEN=<JWT>
policyctl tenant create acme
policyctl tenant create globex
curl -H "Authorization: Bearer $POLICYCTL_TOKEN" -H "Content-Type: application/json" \
  -d '{"tenantID":"acme","subject":"alice","resource":"file1","action":"read"}' \
  http://localhost:8080/check-access
curl -H "Authorization: Bearer $POLICYCTL_TOKEN" -H "Content-Type: application/json" \
  -d '{"tenantID":"globex","subject":"alice","resource":"file1","action":"read"}' \
  http://localhost:8080/check-access
```

#### Using the CLI

```sh
go run cmd/main.go &
export POLICYCTL_TOKEN=<JWT>
policyctl tenant create acme
policyctl tenant create globex
```

#### Modifying Policies

To modify the policies, edit the `policies.yaml` file located in the `configs` directory.
After saving your changes, trigger a reload without restarting the service:

```sh
curl -X POST http://localhost:8080/reload \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <JWT>" \
    -d '{"tenantID":"default"}'
```

On success the service logs a message indicating that policies were reloaded.

#### Compile Natural Language Policy

You can convert an English rule into a YAML policy using either the HTTP API or the CLI.

**API example:**

```sh
curl -X POST http://localhost:8080/compile \
    -H "Content-Type: application/json" \
    -d '{"tenantID": "default", "rule": "Mary can approve invoices"}'
```

**CLI example:**

```sh
go run cmd/policyctl/main.go compile "Mary can approve invoices"
```

### Graph Relationships

The authorization service can map relationships between users, groups, and resources. These
relationships form a directed graph that the evaluator expands when checking access.

#### CLI examples

```
policyctl graph add user:alice group:managers
policyctl graph add group:managers resource:server1
policyctl graph delegate alice mary
policyctl graph list
```

### Delegation

The relationship graph also supports user-to-user delegation. When a user delegates to another,
the delegate can act on behalf of the delegator through a chain of delegation edges.

Common use cases include:

- **Vacation coverage**: Mary grants Alice temporary rights to approve invoices while she is away.
- **Tiered support**: First-line support engineers can delegate complex changes to higher-tier members.
- **Emergency access**: On-call engineers can delegate access to incident responders when needed.

```sh
# allow alice to act as mary
policyctl graph delegate alice mary
```

When `alice` makes a request the evaluator will consider `mary`'s policies if `alice` does not
have direct access. The resulting decision includes the `delegator` field indicating which user
granted the effective permission.

If no applicable policies exist along the chain, the request is denied and the `delegator` field
is omitted from the response.

**Sample request via delegation:**

```sh
curl -X POST http://localhost:8080/check-access \
    -H "Content-Type: application/json" \
    -d '{"tenantID":"default","subject":"alice","resource":"file1","action":"read"}'
```

**Sample response:**

```json
{
  "allow": true,
  "policy_id": "policy1",
  "reason": "allowed by policy",
  "delegator": "mary",
  "context": {
    "subject": "alice",
    "resource": "file1",
    "action": "read"
  }
}
```

#### Sample policy

```yaml
policies:
  - id: "partner-access"
    description: "Partners can view the dashboard during business hours if risk is low"
    subjects:
      - role: "partner"
    resource:
      - "dashboard"
    action:
      - "view"
    effect: "allow"
    when:
      - context.time == "business-hours"
      - context.risk < "medium"
```

#### Example `policies.yaml`

```yaml
policies:
  - id: "policy1"
    description: "Allow admin to read any file"
    subjects:
      - role: "admin"
    resource:
      - "*"
    action:
      - "read"
    effect: "allow"

  - id: "policy2"
    description: "Allow admin to write any file"
    subjects:
      - role: "admin"
    resource:
      - "*"
    action:
      - "write"
    effect: "allow"

  - id: "policy3"
    description: "Allow editor to read any file"
    subjects:
      - role: "editor"
    resource:
      - "*"
    action:
      - "read"
    effect: "allow"

  - id: "policy4"
    description: "Allow editor to edit own files"
    subjects:
      - role: "editor"
    resource:
      - "file2"
    action:
      - "edit"
    effect: "allow"
```

#### Adding a New Policy

Open the configs/policies.yaml file and add a new policy. For example, to allow user3 to write to file3:

```yaml
policies:
  - id: "policy5"
    description: "Allow editor to execute own files"
    subjects:
      - role: "editor"
    resource:
      - "file2"
    action:
      - "execute"
    effect: "allow"
```

Save the file and call the `/reload` endpoint to apply the changes:

```sh
curl -X POST http://localhost:8080/reload \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <JWT>" \
    -d '{"tenantID":"default"}'
```

### Development

To develop and test the service, follow these steps:

1. Install dependencies:

   ```sh
   go mod tidy
   ```

2. Run tests:

   Integration tests exercise the OIDC middleware against a real JWKS endpoint.
   Provide the policy file path and run all tests:

   ```sh
   POLICY_FILE=$(pwd)/configs/policies.yaml go test ./...
   ```

### Persistence Backends

The service stores tenants and policies using a pluggable backend. The backend is
selected with the `STORE_BACKEND` environment variable:

* `memory` (default) – stores all data in memory.
* `sqlite` – persists data in a SQLite database using `STORE_SQLITE_DSN` for the
  connection string (defaults to `file:authorization.db?_foreign_keys=on`).
* `postgres` – persists data in a PostgreSQL database using `STORE_PG_DSN` for
  the connection string (defaults to
  `postgres://postgres:postgres@localhost:5432/authz?sslmode=disable`).

When using SQLite, run the provided migration to create the required tables:

```sh
sqlite3 authorization.db < migrations/001_init.sql
```

For PostgreSQL, apply the migration scripts:

```sh
psql "$STORE_PG_DSN" -f migrations/001_init.up.sql    # migrate up
psql "$STORE_PG_DSN" -f migrations/001_init.down.sql  # migrate down
```

#### Local PostgreSQL Testing

For a quick local environment you can run Postgres in Docker and execute the
integration tests against it:

```sh
docker run --rm -d --name authz-postgres -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=authz -p 5432:5432 postgres:16-alpine
psql "postgres://postgres:postgres@localhost:5432/authz?sslmode=disable" \
  -f migrations/001_init.up.sql
STORE_BACKEND=postgres STORE_PG_DSN=postgres://postgres:postgres@localhost:5432/authz?sslmode=disable \
  go test ./tests/integration -run PostgresPersistence
docker stop authz-postgres
```

Developers can run the service with a specific backend by setting:

```sh
export STORE_BACKEND=postgres # or sqlite
export STORE_PG_DSN=postgres://postgres:postgres@localhost:5432/authz?sslmode=disable
go run cmd/main.go
```

### Policy Backend

Policies can be sourced either from the filesystem (default) or from the
configured database. Select the behaviour with `POLICY_BACKEND`:

* `file` – load policies from YAML files referenced by `POLICY_FILE`.
* `db` – load policies from the database and reload them periodically without a
  service restart.

## Observability

The service includes built-in instrumentation for metrics and tracing.

- **Metrics**: Prometheus metrics are exposed at `/metrics` and include `http_requests_total`, `http_request_duration_seconds`, and `policy_eval_count` counters.
- **Tracing**: OpenTelemetry traces are emitted for each request. Configure `OTEL_EXPORTER_OTLP_ENDPOINT` to point to your collector (defaults to `http://localhost:4318`).

## Docker Deployment

The project ships with a `Dockerfile` and a `docker-compose.yml` for running the
service in a containerized environment.

1. Create a `.env` file in the project root with the required variables
   (`CLIENT_ID`, `CLIENT_SECRET`, `PORT`, `OIDC_ISSUERS`, `OIDC_AUDIENCES`).
2. Start the service:

   ```sh
   docker compose up --build
   ```

3. Stop the service:

   ```sh
   docker compose down
   ```

For convenience, a `Makefile` is provided:

```sh
make up     # Start services using docker compose
make down   # Stop services
make test   # Run unit tests
```

## Helm Deployment

A Helm chart is available in `helm/authorization-service` for deploying the service to Kubernetes clusters such as Minikube or Kind.

1. Copy the provided `helm/authorization-service/example-values.yaml` and adjust it for your environment.
2. Install the chart:

   ```sh
   helm install authz helm/authorization-service -f values.yaml
   ```

   Image repository, tag, replica count, service port, OIDC settings, policy backend, resources, and additional environment variables or secrets can be overridden via the values file or `--set` flags.

Example inline overrides:

```sh
helm install authz helm/authorization-service \
  --set image.repository=myrepo/authorization-service \
  --set image.tag=v1.0.0 \
  --set service.port=9090
```

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
New contributors can start with our [good first issues](https://github.com/bradtumy/authorization-service/issues?q=is%3Aopen+label%3A%22good+first+issue%22).

## How to file issues

If you encounter a bug or have a feature request, please [open an issue](https://github.com/bradtumy/authorization-service/issues/new/choose).
Use the provided templates to ensure we have the details needed to address your report.

## Governance

Project roles, maintainer responsibilities, and review processes are documented in [GOVERNANCE.md](GOVERNANCE.md).

## Security

For information on reporting vulnerabilities, please see [SECURITY.md](SECURITY.md).

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) to understand the expectations for participants in this project.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
