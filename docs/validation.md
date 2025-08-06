# Validation Checklist

## Tenant lifecycle
1. Create tenant
   ```sh
   curl -s -X POST http://localhost:8080/tenant/create -d '{"id":"acme"}'
   ```
   Expected:
   ```json
   {"status":"ok"}
   ```
2. List tenants
   ```sh
   curl -s http://localhost:8080/tenant/list
   ```
   Expected:
   ```json
   ["acme"]
   ```
3. Delete tenant
   ```sh
   curl -s -X POST http://localhost:8080/tenant/delete -d '{"id":"acme"}'
   ```
   Expected:
   ```json
   {"status":"ok"}
   ```

## Policy lifecycle
```sh
cp examples/rbac.yaml configs/policies.yaml
curl -X POST http://localhost:8080/reload -d '{"tenantID":"acme"}'
```
Expected:
```text
policies reloaded
```

## Access checks
Allow:
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file1","action":"read"}'
```
Expected:
```json
{"allow":true}
```

Deny:
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file1","action":"delete"}'
```
Expected:
```json
{"allow":false}
```

Context / risk:
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file:secret","action":"read","context":{"risk":"high"}}'
```
Expected:
```json
{"allow":false}
```

## Metrics
```sh
curl -s http://localhost:8080/metrics | grep http_requests_total
```
Expected: a line containing `http_requests_total`.

## OIDC token validation with Keycloak
```sh
# Obtain a token
TOKEN=$(curl -s -X POST http://localhost:8081/realms/master/protocol/openid-connect/token \
  -d 'client_id=my-client-id' \
  -d 'client_secret=my-client-secret' \
  -d 'grant_type=client_credentials' | jq -r .access_token)

# Use the token
curl -s -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/check-access \
  -d '{"tenantID":"acme","subject":"alice","resource":"file1","action":"read"}'
```
Expected:
```json
{"allow":true}
```

## Tracing
After running checks, open Jaeger at `http://localhost:16686` and confirm spans for `authorization-service` are visible.
