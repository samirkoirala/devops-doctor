## What

<!-- Short description of the change -->

## Why

<!-- Motivation / issue link -->

## How to test

```bash
go vet ./...
go test ./...
go build -o devops-doctor ./cmd/devops-doctor
./devops-doctor check …
```

## Checklist

- [ ] `go vet ./...` and `go test ./...` pass
- [ ] README updated if user-facing behaviour changed
- [ ] Not a security fix meant for private disclosure (see SECURITY.md)
