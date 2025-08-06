# Policies

## Overview
Policies express RBAC and ABAC rules in a concise YAML DSL consumed by the service.

## When to Use
Define who can perform which action on a resource, optionally constrained by attributes.

## Policy Example
An RBAC policy is provided in [examples/rbac.yaml](../examples/rbac.yaml) and an ABAC variant in [examples/abac.yaml](../examples/abac.yaml).

## API Usage
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"file1","action":"read"}'
```

## CLI Usage
```sh
authzctl policy validate examples/rbac.yaml
authzctl check-access --tenant acme --subject alice --resource file1 --action read
```

## SDK Usage
Use the Go or Python SDK `CheckAccess` call after loading policies for a tenant.

## Validation/Testing
Run `authzctl policy validate` and unit tests to ensure policies compile.

## Observability
Policy evaluation counters are exported as `policy_eval_count{decision,reason}`.

## Notes & Caveats
Malformed policies will be rejected at load time; use `policy validate` to detect issues early.

## Managing Users
Roles referenced in policies are assigned to users dynamically. Manage users and their roles via the [User API](users.md).
