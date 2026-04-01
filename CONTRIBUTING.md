# Contributing

Thanks for helping improve **devops-doctor**.

## Before you start

- Check [existing issues](https://github.com/samirkoirala/devops-doctor/issues) and [PRs](https://github.com/samirkoirala/devops-doctor/pulls).
- For **security-sensitive** reports, see [SECURITY.md](SECURITY.md).

## Development setup

```bash
git clone https://github.com/samirkoirala/devops-doctor.git
cd devops-doctor
go mod download
go vet ./...
go test ./...
go build -o devops-doctor ./cmd/devops-doctor
./devops-doctor check --verbose
```

## Pull requests

1. Fork the repo and create a branch from `main` (e.g. `fix/docker-check-timeout` or `feat/add-check`).
2. Keep changes focused; match existing code style and structure.
3. Update **README** if you change user-visible behaviour or flags.
4. Ensure `go vet ./...` and `go test ./...` pass locally (CI runs the same on PRs).

Open a PR with a clear description: **what** changed, **why**, and how to verify.

## Releases (maintainers)

- Tag with semantic versioning (e.g. `v0.1.0`).
- Add release notes on the [Releases](https://github.com/samirkoirala/devops-doctor/releases) page summarizing changes.
- `go install github.com/samirkoirala/devops-doctor/cmd/devops-doctor@latest` follows the default branch; pinned installs use `@vX.Y.Z`.

## Going public on GitHub (maintainers)

1. **Settings → General → Danger zone → Change repository visibility** → Public.
2. **Settings → Code security and analysis**: turn on **Dependency graph**, **Dependabot alerts**, and **Code scanning** (this repo includes a CodeQL workflow).
3. Optional: enable **Discussions** under repository **Settings → General → Features**.
4. Suggested issue labels (create if missing): `bug`, `enhancement`, `dependencies`, `documentation`, `good first issue`.
5. After the first push to `main`, confirm **Actions** runs **CI** and **CodeQL** successfully.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). Participating means you agree to uphold it.
