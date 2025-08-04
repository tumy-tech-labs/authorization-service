# Flows

## Policy Evaluation

```mermaid
sequenceDiagram
    participant C as Client
    participant S as Authorization Service
    participant PE as Policy Engine
    C->>S: /check-access request
    S->>PE: Evaluate policy
    PE-->>S: Decision (allow/deny)
    S-->>C: Response
```

1. Client sends a `check-access` request with a JWT and tenant ID.
2. Service validates the request and forwards it to the policy engine.
3. Policy engine evaluates CDL rules and returns a decision.
4. Service responds to the client with the result.

## OIDC Authentication

```mermaid
sequenceDiagram
    participant U as User
    participant OP as OIDC Provider
    participant S as Authorization Service
    U->>OP: Authenticate
    OP-->>U: JWT
    U->>S: API call with token
    S->>OP: Fetch JWKS / validate token
    S-->>U: Response
```

The service verifies incoming JWTs against configured issuers using their JWKS endpoints.

## Observability

```mermaid
sequenceDiagram
    participant S as Authorization Service
    participant Prom as Prometheus
    participant Otel as OTLP Collector
    Prom-->>S: Scrape /metrics
    S-->>Prom: Metrics
    S-->>Otel: Traces
```

Metrics are exported in Prometheus format and traces are sent via OTLP to the configured collector.
