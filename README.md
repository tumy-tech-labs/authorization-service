# Authorization Service

[![CI](https://github.com/bradtumy/authorization-service/actions/workflows/ci.yml/badge.svg)](https://github.com/bradtumy/authorization-service/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bradtumy/authorization-service)](https://goreportcard.com/report/github.com/bradtumy/authorization-service)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Authorization-Service is an open-source, multi-tenant, context-aware, risk-adaptive authorization engine for modern applications.

## Quickstart

```sh
git clone https://github.com/bradtumy/authorization-service.git
cd authorization-service
docker compose up --build
```

### Policy as Code Quickstart

The service can source policies from a Git repository so that they can be versioned and tested like application code.

1. Clone the [customer-policies](https://github.com/your-org/customer-policies) template.
2. Write YAML policies under `policies/` and add tests under `tests/`.
3. Run the following commands locally:

   ```bash
   authzctl policy validate policies/rbac.yaml   # lint
   authzctl test tests/                          # run tests
   authzctl simulate --bundle policies/          # dry‑run
   authzctl apply-bundle policies/               # deploy
   ```

4. CI/CD executes the same lint → test → simulate → deploy workflow via `.github/workflows/deploy.yaml`.
5. After deployment, verify the active policy version:

   ```bash
   curl -s http://localhost:8080/policies/version
   ```

The returned commit SHA is also echoed in `/check-access` responses under the `commit` field of the decision.

## Feature Highlights

- **RBAC** – role-based access control policies.
- **ABAC** – attribute-based rules.
- **Graph** – relationship graphs for complex hierarchies.
- **Delegation** – temporary privileges with delegation chains.
- **Context & Risk** – adapt decisions to runtime signals.
- **Simulation** – dry-run policy evaluation.
- **OIDC** – identity tokens via OpenID Connect.
- **Observability** – metrics, logs and traces for every decision.

## Getting Started Example

`examples/rbac.yaml`:

```yaml
roles:
  - name: admin
    policies:
      - allow-read-all
users:
  - username: alice
    roles: [admin]
policies:
  - id: allow-read-all
    subjects:
      - role: admin
    resources:
      - "*"
    actions:
      - read
    effect: allow
```

Run a check:

```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file1","action":"read"}'
```

## Documentation

- [Quickstart](docs/quickstart.md)
- [Tenants](docs/tenants.md)
- [Policies](docs/policies.md)
- [Graph](docs/graph.md)
- [Delegation](docs/delegation.md)
- [Context & Risk](docs/context.md)
- [Remediation](docs/remediation.md)
- [Simulation](docs/simulation.md)
- [OIDC](docs/oidc.md)
- [Observability](docs/observability.md)
- [Deployment](docs/deployment.md)
- [Contributing](docs/contributing.md)
- [Architecture](docs/architecture.md) · [Flows](docs/flows.md)

Examples live in [`examples/`](examples) and the Postman collection in [`postman/`](postman).

## License

Apache 2.0 licensed. See [LICENSE](LICENSE).
