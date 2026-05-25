# Contributing

Thank you for your interest in contributing to this project.

## Getting started

1. Fork the repository and clone your fork.
2. Install prerequisites: Go `1.24+`, Terraform CLI, and a running `dotnet-ipam` API endpoint.
3. Install dependencies:

```bash
make tidy
```

4. Run the test suite to confirm your setup:

```bash
make test
```

## Making changes

- Open an issue first for significant changes so the approach can be agreed on before work begins.
- Keep commits focused — one logical change per commit.
- Run `make test` before pushing. All tests must pass.
- If your change affects public behaviour (new resource, schema change, bug fix), add or update the relevant acceptance test in `internal/provider/resources_acc_test.go`.

## Running acceptance tests

Acceptance tests require a live IPAM API instance:

```bash
IPAM_ACC=1 \
IPAM_BASE_URL=http://localhost:8080 \
IPAM_USERNAME=admin \
IPAM_PASSWORD=Admin1234! \
make testacc
```

## Pull requests

- Target the `main` branch.
- Fill in the PR description with what changed and why.
- Link any related issues.

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Prefer editing existing files over creating new ones.
- Keep provider schema changes consistent with the conventions in existing resources.

## Licence

By contributing you agree that your contributions will be licensed under the [MIT License](LICENSE).
