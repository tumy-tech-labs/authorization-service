# Policy as Code

This document describes how to manage policies for the authorization service using a dedicated Git repository.  Policies and tests live next to application code so that changes can be reviewed, linted and deployed through a familiar workflow.

## Repository Layout

```
customer-policies/
├── policies/            # YAML policy files loaded by the engine
│   ├── rbac.yaml        # role based examples
│   ├── abac.yaml        # attribute based rules
│   └── context.yaml     # context aware policies
├── tests/               # `authzctl test` inputs
│   ├── allow.yaml       # example allow test
│   └── deny.yaml        # example deny test
├── bundles/             # optional packaged policies for offline apply
├── workflows/           # CI/CD automation
│   └── deploy.yaml      # lint → test → simulate → deploy
├── README.md            # usage instructions
└── .gitignore
```

Clone the template repository and place your policy files under `policies/`.  The default examples show RBAC, ABAC and context aware rules and can be used as a starting point.

## Local Development

The [`authzctl`](../cmd/authzctl) command line utility ships with the service and can be used to work with policies locally.

```bash
# lint policy syntax
authzctl policy validate policies/rbac.yaml

# run policy tests
authzctl test tests/

# simulate decisions without mutating the server
authzctl simulate --bundle policies/

# apply policies to the running engine
authzctl apply-bundle policies/
```

`authzctl test` expects files under `tests/` that describe an input and the expected allow/deny result.  The command runs each case against the local policy bundle.

## CI/CD Workflow

Automation is handled through a GitHub Actions workflow in `.github/workflows/deploy.yaml` of the policy repository:

1. **Lint** – run `authzctl policy validate` on every policy file.
2. **Test** – execute `authzctl test` to verify behaviour.
3. **Simulate** – dry‑run a deployment using `authzctl simulate`.
4. **Deploy** – apply the policies to the configured authorization engine.

The deploy step requires the following repository secrets:

- `TENANT_ID` – tenant identifier used by the authorization service.
- `AUTHZ_SERVER` – base URL of the running authorization service.

## Runtime Versioning

Policies are pulled from Git and the commit SHA of the currently applied revision is surfaced by the service.  The value can be retrieved via the `GET /policies/version` API and is also included in access‑check responses under the `commit` field of a decision.  This allows operators to trace every decision back to the exact set of policies in effect.

## Further Reading

For a high level walkthrough, see the [Policy as Code Quickstart](../README.md#policy-as-code-quickstart).  The rest of the service documentation is available under the [docs/](./) directory.

