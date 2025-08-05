# Deployment

## Overview
The service ships with Docker and Helm assets for local and Kubernetes deployments.

## When to Use
Deploy with Docker for evaluation or Helm for production clusters.

## Policy Example
Any policy from the [examples](../examples) directory can be mounted into the container at `/configs/policies.yaml`.

## API Usage
After deployment, interact with the REST API at the exposed service address.

## CLI Usage
Configure `AUTHZCTL_ADDR` to point to your deployed instance and run normal commands.

## SDK Usage
Initialize clients with the external URL of your deployment.

## Validation/Testing
Run health checks and a sample `check-access` to confirm the service is reachable.

## Observability
Expose `/metrics` and OTLP endpoints through your ingress for centralized monitoring.

## Notes & Caveats
Ensure secrets and policy files are mounted securely in production environments.
