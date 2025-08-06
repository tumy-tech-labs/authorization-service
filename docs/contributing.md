# Contributing

## Overview
We welcome community contributions to improve the authorization service and its documentation.

## When to Use
Refer to this guide before opening issues or pull requests.

## Policy Example
See [examples/rbac.yaml](../examples/rbac.yaml) when adding tests or docs.

## API Usage
Use the API examples in this documentation to reproduce issues and verify fixes.

## CLI Usage
Run `make build` and the `authzctl` commands to validate changes locally.

## SDK Usage
Update Go or Python SDK snippets as part of feature work.

## Validation/Testing
Execute `go test ./...` and ensure linting passes before submitting a PR.

## Observability
Check logs and metrics during development to diagnose failures.

## Local Development

Run with hot reload using [air](https://github.com/cosmtrek/air):

```sh
go install github.com/cosmtrek/air@latest
air
```

Run tests:

```sh
make test
```

Run lint:

```sh
make lint
```

Run security scan:

```sh
make sec
```

Run integration tests:

```sh
make integration-test
```

## Notes & Caveats
Follow the [Code of Conduct](../CODE_OF_CONDUCT.md) and sign your commits if required.
