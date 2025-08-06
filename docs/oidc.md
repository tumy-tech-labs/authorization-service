# OIDC

## Overview
The service validates JSON Web Tokens (JWTs) issued by OpenID Connect providers to authenticate subjects.

## Local Keycloak Setup
Run the services with Docker:

```sh
docker compose up --build
```

Keycloak will be available at [http://localhost:8081](http://localhost:8081). Log in with the demo users:

| User  | Password | Role        |
|-------|----------|-------------|
| alice | alice    | TenantAdmin |
| bob   | bob      | User        |

The authorization service loads matching demo users from `configs/acme/users.yaml` so that `alice` already has the `TenantAdmin` role required to manage users.

## Getting a Token
Use the password grant to obtain a token via `curl`:

```sh
curl -s -X POST \
  http://localhost:8081/realms/authz-service/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=password&client_id=authz-client&username=alice&password=alice'
```

The response contains an `access_token` that includes the `roles` and `tenantID` claims.

## Calling the API
Pass the token in the `Authorization` header when calling protected endpoints:

```sh
TOKEN=$(curl -s -X POST \
  http://localhost:8081/realms/authz-service/protocol/openid-connect/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=password&client_id=authz-client&username=alice&password=alice' | jq -r .access_token)

curl -X POST http://localhost:8080/user/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"tenantID":"acme","username":"charlie","roles":["User"]}'
```

## Notes
- `roles` claim lists the user's realm roles.
- `tenantID` claim is hard coded to `acme` for demo purposes.
