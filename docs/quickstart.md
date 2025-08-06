# Quickstart

1. **Clone the repository**
   ```sh
   git clone https://github.com/bradtumy/authorization-service.git
   cd authorization-service
   ```

2. **Create the environment file**
   ```sh
   cp .env.example .env
   # edit as needed
   ```

3. **Start the service**
   ```sh
   docker compose up --build
   ```

4. **Load the sample policy**
   ```sh
   cp examples/rbac.yaml configs/policies.yaml
   curl -X POST http://localhost:8080/reload -d '{"tenantID":"acme"}'
   ```
   Expected:
   ```text
   policies reloaded
   ```

5. **Run an access check**
   ```sh
   curl -s -X POST http://localhost:8080/check-access \
     -H 'Content-Type: application/json' \
     -d '{"tenantID":"acme","subject":"alice","resource":"file1","action":"read"}'
   ```
   Expected output:
   ```json
   {"allow":true}
   ```

Troubleshooting: if the service fails to start, ensure `.env` exists and Docker has access to the port.

## Next Steps
- [Tenants](tenants.md)
- [Policies](policies.md)
- [Context & Risk](context.md)
