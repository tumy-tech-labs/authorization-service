# Architecture

The authorization service is composed of a small set of components that work together to evaluate access decisions.

```mermaid
graph TD
    C[Client] -->|HTTP| API[HTTP API]
    API -->|Validate JWT| OIDC[OIDC Provider]
    API --> PolicyEngine
    PolicyEngine --> PolicyStore[(Policy Store)]
    PolicyEngine --> Graph[Relationship Graph]
    PolicyStore --> Store[(Persistence Backend)]
    API -->|Metrics & Traces| Telemetry[Prometheus / OTLP]
```

* **HTTP API** – exposes endpoints for policy management and access checks.
* **OIDC Middleware** – verifies JWTs against configured issuers.
* **Policy Engine** – evaluates CDL policies using a graph of relationships.
* **Policy Store** – caches policies for each tenant and persists them via pluggable backends (memory, SQLite, PostgreSQL).
* **Telemetry** – exports Prometheus metrics and OpenTelemetry traces for observability.
