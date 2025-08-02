# Authorization Service

Authorization Service is an open-source authorization service that reads policies in a simple CDL and provides authorization decisions based on the information provided.

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
   JWT_SECRET=my-jwt-secret
   PORT=8080
   STORE_BACKEND=memory # or sqlite
   OIDC_ISSUER=https://issuer.example.com
   OIDC_AUDIENCE=my-client-id
   LOG_LEVEL=info
   ```

### Usage

#### API Endpoints

All requests must include a `tenantID` in the JSON body to scope operations.

| Method | Endpoint        | Description                                     |
| ------ | --------------- | ----------------------------------------------- |
| POST   | `/check-access` | Evaluate a tenant-scoped access request         |
| POST   | `/reload`       | Reload policies for a specific tenant from disk |
| POST   | `/compile`      | Convert natural language to YAML for a tenant   |
| POST   | `/tenant/create`| Register a new tenant                            |
| POST   | `/tenant/delete`| Remove an existing tenant                        |
| GET    | `/tenant/list`  | List all tenants                                |

### Observability

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

Prometheus metrics and OpenTelemetry traces are also exposed. To scrape metrics, point
Prometheus at the `/metrics` endpoint. Traces can be exported by configuring the standard
`OTEL_EXPORTER_OTLP_ENDPOINT` environment variable.

Set `LOG_LEVEL` to control verbosity (`debug`, `info`, `warn`, `error`). The default level
is `info`. Sensitive values such as secrets or raw tokens are deliberately omitted from
all logs.

### OIDC Configuration

Tokens presented to the service are validated using an OpenID Connect issuer. Configure
`OIDC_ISSUER` and `OIDC_AUDIENCE` in `.env` to enable JWKS validation. Example with
Keycloak:

```
OIDC_ISSUER=http://localhost:8080/realms/master
OIDC_AUDIENCE=account
```

The middleware will automatically fetch and cache the JWKS for signature verification.

#### Generate JWT

To generate a client credential JWT token:

1. Navigate to the `scripts` directory:

   ```sh
   cd scripts
   ```

2. Run the `generate_jwt.go` script:

   ```sh
   go run generate_jwt.go
   ```

3. The script will output a JWT token:

   ```sh
   Generated JWT Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

#### Request Policy Decision

Use the generated JWT token to request a policy decision from the authorization service.

1. Start the server:

   ```sh
   go run cmd/main.go
   ```

2. Send a POST request to the `/check-access` endpoint:

   ```sh
   curl -X POST http://localhost:8080/check-access \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
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

```
policies:
  - id: "manager-read"
    description: "Managers can read server1"
    subjects:
      - role: "managers"
    resource:
      - "server1"
    action:
      - "read"
    effect: "allow"
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

   ```sh
   POLICY_FILE=../configs/policies.yaml JWT_SECRET=secret go test ./...
   ```

### Persistence Backends

The service stores tenants and policies using a pluggable backend. The backend is
selected with the `STORE_BACKEND` environment variable:

* `memory` (default) – stores all data in memory.
* `sqlite` – persists data in a SQLite database using `STORE_SQLITE_DSN` for the
  connection string (defaults to `file:authorization.db?_foreign_keys=on`).

When using SQLite, run the provided migration to create the required tables:

```sh
sqlite3 authorization.db < migrations/001_init.sql
```

Developers can run the service with SQLite by setting:

```sh
export STORE_BACKEND=sqlite
export STORE_SQLITE_DSN=authorization.db
go run cmd/main.go
```

### Docker Deployment

The project ships with a `Dockerfile` and a `docker-compose.yml` for running the
service in a containerized environment.

1. Create a `.env` file in the project root with the required variables
   (`CLIENT_ID`, `CLIENT_SECRET`, `JWT_SECRET`, `PORT`).
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

### Contributing

Contributions are welcome! Please open an issue or submit a pull request.

### License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
