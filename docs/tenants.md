# Tenants

## Overview
Tenants isolate policies and data so multiple organizations can share a single deployment safely.

## When to Use
Adopt multi-tenancy when building SaaS platforms or environments with strict data separation.

## Policy Example
See [examples/tenants.yaml](../examples/tenants.yaml) for a tenant-scoped policy.

## API Usage
Create and list tenants:
```sh
curl -s -X POST http://localhost:8080/tenant/create -d '{"tenantID":"acme"}'
curl -s http://localhost:8080/tenant/list
```

## CLI Usage
```sh
authzctl tenant create acme
authzctl tenant list
```

## SDK Usage
Go and Python SDKs accept `TenantID` on every request to scope evaluations.

## Validation/Testing
Issue a `check-access` for each tenant to confirm isolation.

## Observability
Metrics and logs are tagged with the tenant ID for filtering.

## Notes & Caveats
Ensure tenant identifiers are globally unique and validated to prevent injection.
