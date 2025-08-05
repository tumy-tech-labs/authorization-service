# Delegation

## Overview
Delegation allows a principal to grant temporary privileges to another actor.

## When to Use
Use delegation for out-of-office access or break-glass scenarios.

## Policy Example
See [examples/delegation.yaml](../examples/delegation.yaml) for a simple delegation chain.

## API Usage
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"bob","resource":"document:report","action":"read"}'
```

## CLI Usage
```sh
authzctl check-access --tenant acme --subject bob --resource document:report --action read
```

## SDK Usage
The SDKs automatically honor delegation rules when evaluating requests.

## Validation/Testing
Create a delegation and verify that the delegate receives the intended access.

## Observability
Delegated decisions are tagged with `delegated=true` in metrics and logs.

## Notes & Caveats
Delegations include expiry timestamps; expired entries are ignored.
