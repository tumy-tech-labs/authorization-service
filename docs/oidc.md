# OIDC

## Overview
The service validates JSON Web Tokens issued by OpenID Connect providers to authenticate subjects.

## When to Use
Use OIDC when integrating with identity platforms like Okta or Auth0.

## Policy Example
See [examples/oidc.yaml](../examples/oidc.yaml) where the subject is derived from the `sub` claim.

## API Usage
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"tenantID":"acme","resource":"file:report","action":"read"}'
```

## CLI Usage
```sh
authzctl check-access --tenant acme --resource file:report --action read --token $TOKEN
```

## SDK Usage
Provide the JWT when constructing the client or on each request.

## Validation/Testing
Use a test identity provider and supply a valid token to verify rejection of expired or malformed JWTs.

## Observability
Token validation errors are logged and surfaced via `oidc_validation_fail_total` metrics.

## Notes & Caveats
Clock skew and issuer configuration must match your identity provider.
