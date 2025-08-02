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
   ```

### Usage

#### API Endpoints

All requests must include a `tenantID` in the JSON body to scope operations.

| Method | Endpoint        | Description                                     |
| ------ | --------------- | ----------------------------------------------- |
| POST   | `/check-access` | Evaluate a tenant-scoped access request         |
| POST   | `/reload`       | Reload policies for a specific tenant from disk |
| POST   | `/compile`      | Convert natural language to YAML for a tenant   |

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
   go test ./...
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
