# Release Process

## Trigger

- Create and push a semantic version tag: `vMAJOR.MINOR.PATCH`
- Example: `v0.1.0`

The GitHub workflow `.github/workflows/release.yml` runs automatically on matching tags.

## Workflow Steps

1. Checkout with full git history (`fetch-depth: 0`)
2. `make tidy`
3. Run GoReleaser with `release --clean`
4. Publish archives and `checksums.txt` to GitHub Releases

## Packaging

- Packaging config is in `.goreleaser.yaml`
- Binary name: `terraform-provider-dotnet-ipam`
- Targets: Linux/macOS/Windows for amd64 + arm64

## Requirements

- `GITHUB_TOKEN` is provided by GitHub Actions (workflow uses `contents: write`).
