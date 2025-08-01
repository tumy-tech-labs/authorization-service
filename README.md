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
            "subject": "user1", 
            "resource": "file1",
            "action": "read",
            "conditions": {}
        }'
    ```

3. The service will respond with the policy decision:

    ```json
    {
        "allowed": true
    }
    ```

#### Modifying Policies

To modify the policies, edit the `policies.yaml` file located in the `configs` directory.

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

Open the configs/policies.yaml file.

Add a new policy to the file. For example, to allow user3 to write to file3:

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

Save the file and restart the authorization service to apply the changes:

```bash
go run cmd/main.go
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

To build and run the service using Docker:

1. Build the Docker image:

    ```sh
    docker build -t authorization-service .
    ```

2. Run the Docker container:

    ```sh
    docker run -d -p 8080:8080 --env-file .env authorization-service
    ```

### Contributing

Contributions are welcome! Please open an issue or submit a pull request.

### License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
