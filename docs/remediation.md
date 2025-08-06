# Remediation

## Overview
Remediation actions guide users to resolve denied requests, such as performing multi-factor authentication.

## When to Use
Use remediation to offer just-in-time elevation instead of outright denial.

## Policy Example
See [examples/remediation.yaml](../examples/remediation.yaml).

## API Usage
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file:secret","action":"read","context":{"risk":"medium"}}'
```

## CLI Usage
```sh
authzctl check-access --tenant acme --subject alice --resource file:secret --action read --context risk=medium
```

## SDK Usage
Evaluate access and inspect the remediation instructions returned alongside a deny decision.

## Validation/Testing
Trigger a remediation flow and confirm the client performs the required action.

## Observability
Remediation attempts increment `remediation_attempt_total` metrics.

## Notes & Caveats
Clients must interpret remediation instructions; the service does not enforce them automatically.
