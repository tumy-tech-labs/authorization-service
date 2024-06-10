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

### Usage

#### Set Up Environment Variables

Create a `.env` file in the project root with the following content:

```dotenv
JWT_SECRET=your-256-bit-secret
```

#### Generate JWT

1. Navigate to the scripts directory and run the JWT generation script:

    ```bash
    cd scripts
    go run generate_jwt.go
    ```

    Response:

    ```bash
    Generated JWT Token:  eyJhbGciOiJIUzI1NiIsInR5cCI6xxxxxxxxxxx
    ```

#### Start the Server

1. Start the authorization service:

    ```bash
    go run cmd/main.go
    ```

2. Alternatively, you can build and run the Docker container:

    ```bash
    docker build -t authorization-service .
    docker run -p 8080:8080 authorization-service
    ```

#### Request Policy Decision

1. Use the generated JWT token in the Authorization header to make a policy decision request:

    ```bash
    curl -X POST http://localhost:8080/check-access \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6xxxxxxxxxxx" \
        -d '{                  
            "subject": "user1", 
            "resource": "file1",
            "action": "read",
            "conditions": []
        }'
    ```

    Response:

    ```json
    {
      "allowed": true
    }
    ```

## Configuration

Policies are defined in `configs/policies.yaml`. You can modify this file to add or update policies. Here is an example of the `policies.yaml` file:

```yaml
- id: "1"
  resource: "file1"
  action: "read"
  effect: "allow"
  conditions: []

- id: "2"
  resource: "file2"
  action: "write"
  effect: "deny"
  conditions: []
