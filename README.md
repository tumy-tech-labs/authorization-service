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
    go run main.go
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
            "conditions": []
        }'
    ```

3. The service will respond with the policy decision:

    ```json
    {
        "allowed": true
    }
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
