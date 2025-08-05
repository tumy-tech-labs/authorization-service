# Simulation

## Overview
The simulation endpoint performs dry-run evaluations to preview decisions without enforcing them.

## When to Use
Use simulation during policy development or troubleshooting to understand outcomes.

## Policy Example
See [examples/simulation.yaml](../examples/simulation.yaml).

## API Usage
```sh
curl -s -X POST http://localhost:8080/simulate \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file:test","action":"read"}'
```

## CLI Usage
```sh
authzctl simulate --tenant acme --subject alice --resource file:test --action read
```

## SDK Usage
Call the `Simulate` method in the Go or Python SDK to retrieve a hypothetical decision.

## Validation/Testing
Compare simulation results to actual `check-access` responses when policies change.

## Observability
Simulation requests are labeled `simulation=true` in metrics and traces.

## Notes & Caveats
Simulation does not persist any state; context providers still run as in a real request.
