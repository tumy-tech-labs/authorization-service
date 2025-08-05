# Observability

## Overview
The service emits Prometheus metrics, structured logs and OpenTelemetry traces for every request.

## When to Use
Enable observability to monitor authorization performance and troubleshoot issues.

## Policy Example
See [examples/observability.yaml](../examples/observability.yaml) for a sample policy used in metrics demos.

## API Usage
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file:metrics","action":"read"}'
# scrape metrics
curl -s http://localhost:8080/metrics | head
```

## CLI Usage
```sh
authzctl check-access --tenant acme --subject alice --resource file:metrics --action read
```

## SDK Usage
SDK requests automatically generate traces and metrics when the environment is configured.

## Validation/Testing
Use `curl /metrics` and an OTLP collector to confirm telemetry is emitted.

## Observability
Metrics: `http_requests_total`, `policy_eval_count`; logs include decision reasons; traces show timing.

## Notes & Caveats
High-volume telemetry can impact performance; sample or filter as needed.
