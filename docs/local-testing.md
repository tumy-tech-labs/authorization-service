# Local End-to-End Testing

The following steps walk through a complete flow of running the service, creating users, obtaining tokens, and verifying policy decisions.

## 1. Start the stack

```sh
cp .env.example .env
# adjust values as needed
cp examples/rbac.yaml configs/policies.yaml

# build and launch the authorization service and Keycloak
 docker compose up --build
```

## 2. Reload the policy

In another terminal once the service is running:

```sh
curl -X POST http://localhost:8080/reload -d '{"tenantID":"acme"}'
```

Expected:

```text
policies reloaded
```

## 3. Obtain an admin token

Keycloak exposes demo users. Retrieve a token for `alice` (TenantAdmin) using the password grant:

```sh
curl -s -X POST \
  http://localhost:8081/realms/authz-service/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=password&client_id=authz-client&username=alice&password=alice'
```

Save the returned `access_token` in an environment variable for later use:

```sh
TOKEN=$(curl -s -X POST \
  http://localhost:8081/realms/authz-service/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=password&client_id=authz-client&username=alice&password=alice' | jq -r .access_token)
```

## 4. Create a new user in the service

Use the admin token to register a user `charlie` in tenant `acme` with the `admin` role defined in the sample policy:

```sh
curl -X POST http://localhost:8080/user/create \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","username":"charlie","roles":["admin"]}'
```

The service persists users under `configs/acme/users.yaml` when started with `--persist-users` (enabled in `docker-compose.yml`).

## 5. Add the user to Keycloak

Open [http://localhost:8081](http://localhost:8081) and log in with `admin`/`admin`. Create a user `charlie`, set a password, and assign the realm role `admin` so that issued tokens contain the matching `roles` claim.

## 6. Obtain a token for the new user

Request a token using `charlie`'s credentials:

```sh
CHARLIE_TOKEN=$(curl -s -X POST \
  http://localhost:8081/realms/authz-service/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=password&client_id=authz-client&username=charlie&password=charlie' | jq -r .access_token)
```

## 7. Verify policy decisions

Check that `charlie` can read `file1` according to the sample policy:

```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"charlie","resource":"file1","action":"read"}'
```

Expected response:

```json
{"allow":true}
```

Access checks do not require an authentication header, but other management endpoints do. Use the `Authorization: Bearer $CHARLIE_TOKEN` header with APIs that require authentication.

## 8. Clean up

Stop the services when finished:

```sh
docker compose down -v
```

This sequence demonstrates creating a user, acquiring tokens, and verifying policies in a local environment.
