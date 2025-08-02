MVP F Story Pack – Enterprise Readiness & Observability

Story F1 – Structured Audit Logging
Context: We need structured, traceable audit logs for every access check and tenant/policy action.
Tasks for Codex:
Implement JSON structured logging with fields: timestamp, correlation_id, tenant_id, subject, action, resource, decision, policy_id, reason.
Generate a correlation_id (UUID) for each incoming request and propagate it through logs.
Ensure sensitive values (secrets, raw tokens) are never logged.
Add log levels (info, warn, error) for appropriate events.
Update README with log format, sample outputs, and guidance on enabling/disabling debug logging.
Acceptance Criteria:
All requests produce structured logs.
Each log entry contains a correlation ID.
No sensitive data is logged.
README documents usage and examples.

Story F2 – OpenTelemetry Integration
Context: Enterprise environments require observability through metrics and distributed tracing.
Tasks for Codex:
Integrate OpenTelemetry SDK for Go.
Create spans for request lifecycle and policy evaluation.
Add /metrics endpoint exposing Prometheus metrics (http_requests_total, latency histograms, policy_eval_count`).
Update README with instructions for enabling tracing and scraping metrics.
Acceptance Criteria:
OTEL traces visible in a local Jaeger/Tempo setup.
/metrics endpoint returns Prometheus-compatible data.
README explains observability setup.

Story F3 – OIDC & JWKS Validation
Context: Current JWT validation is static. We need OIDC-compatible validation with JWKS rotation.
Tasks for Codex:
Support configuration of one or more OIDC providers (via .env or config file).
Fetch JWKS from provider(s) and cache with expiry.
Validate iss and aud claims against config.
Reject tokens failing signature or claim validation.
Update README with OIDC/JWKS setup instructions (sample Keycloak config).
Acceptance Criteria:
Tokens signed by configured OIDC issuers are validated.
Expired/invalid tokens are rejected with 401 Unauthorized.
README documents OIDC integration.

Story F4 – CLI (authzctl)
Context: Admins and developers need a CLI to interact with the service.
Tasks for Codex:
Build authzctl CLI with subcommands:
tenant create (wraps API)
tenant delete
policy validate (offline validation)
check-access (dry-run access check)
Ensure CLI reads config from .env or flags.
Package CLI as part of build.
Update README with CLI install instructions and usage examples.
Acceptance Criteria:
CLI supports all listed subcommands.
CLI returns exit codes (0 for success, non-zero for errors).
README documents CLI usage.

Story F5 – Helm Chart for Kubernetes
Context: Enterprises deploy in Kubernetes. We need a Helm chart.
Tasks for Codex:
Create helm/authorization-service chart with configurable values: image, tag, replicas, resources, env, secrets.
Provide example values.yaml.
Support overrides for port, OIDC config, and policy backend.
Update README with Helm deployment instructions.
Acceptance Criteria:
helm install deploys service successfully on minikube/kind.
Configurable values work as expected.
README documents Helm deployment.

Story F6 – Persistent Policy Backend
Context: Current policy storage is file-based. Enterprises need persistent, dynamic storage.
Tasks for Codex:
Add Postgres backend for storing tenant policies.
Support file-based backend as fallback.
Implement hot reload of policies from DB without restart.
Add migration script (migrate up/down) for DB schema.
Update README with backend configuration instructions and migration usage.
Acceptance Criteria:
Policies persist across container restarts.
Switching backends works via config.
README documents backend setup.

Story F7 – SDKs & Developer Docs
Context: Developers should easily integrate authorization-service into apps.
Tasks for Codex:
Create sdk/go and sdk/python folders.
Implement SDK functions: CheckAccess(), CompileRule(), ValidatePolicy().
Add examples in both languages.
Expand Postman collection to cover new endpoints.
Update README with SDK usage guides (Go & Python examples).
Acceptance Criteria:
SDKs work in simple client test apps.
SDKs have unit tests.
README has integration examples. 