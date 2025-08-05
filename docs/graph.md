# Graph

## Overview
The graph engine models relationships such as ownership or group membership to drive authorization decisions.

## When to Use
Use graph-based policies to represent hierarchies or transitive access relationships.

## Policy Example
See [examples/graph.yaml](../examples/graph.yaml) for a policy leveraging graph relations.

## API Usage
```sh
curl -s -X POST http://localhost:8080/check-access \
  -H 'Content-Type: application/json' \
  -d '{"tenantID":"acme","subject":"alice","resource":"document:q1","action":"read"}'
```

## CLI Usage
```sh
authzctl check-access --tenant acme --subject alice --resource document:q1 --action read
```

## SDK Usage
Go/Python SDKs automatically traverse the graph when evaluating resources.

## Validation/Testing
Create graph edges and issue access checks to verify traversal.

## Observability
Graph lookups are instrumented with timing metrics and traces.

## Notes & Caveats
Ensure graphs remain acyclic to prevent evaluation loops.
