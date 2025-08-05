# Context & Risk

## Overview
Context providers enrich requests with environmental data such as time, location or risk scores.

## When to Use
Leverage context to make adaptive decisions based on runtime signals.

## Policy Example
See [examples/context.yaml](../examples/context.yaml) for a policy denying high-risk requests.

## API Usage
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file:secret","action":"read","context":{"risk":"high"}}'
```

## CLI Usage
```sh
authzctl check-access --tenant acme --subject alice --resource file:secret --action read --context risk=high
```

## SDK Usage
Pass a context map when calling `CheckAccess` in Go or Python.

## Validation/Testing
Simulate various contexts to ensure policies react appropriately.

## Observability
Context keys are recorded as attributes on evaluation traces.

## Notes & Caveats
Untrusted context sources should be validated to avoid spoofing.
