# Contributing to dployr

Thank you for your interest in contributing to **dployr**.  
This guide outlines the setup, architecture conventions, and contribution process for maintaining consistency across the project.

---

## Development Setup

### Prerequisites
- Go **1.24+**
- Git
- Make

### Local Development
```bash
git clone https://github.com/dployr-io/dployr.git
cd dployr
make build
````

### Running Tests

```bash
make test                     # Run all tests
go test ./pkg/...             # Test public packages only
go test ./internal/...        # Test internal packages only
```

### Development Workflow

```bash
make build-daemon             # Build daemon for local testing
./dist/dployrd                # Run daemon locally
./dist/dployr --help          # Test CLI commands
```

### Pre‑flight checks (must pass)

Run local checks that mirror CI:

```bash
make ci                       # full local CI parity

# or run individually
gofmt -s -l .                 # must output nothing
go vet ./...
staticcheck ./...
go build ./...
go test -race -count=1 ./...
```

---

## Architecture Guidelines

### Package Structure

* `cmd/` – Application entry points (main packages)
* `pkg/` – Public library code (importable by external projects)
* `internal/` – Private application code (not importable)
* `api/` – OpenAPI specifications and API contracts

### Layer Separation

* **Handlers** (`pkg/*/handlers.go`) – HTTP request/response logic
* **Services** (`internal/*/service.go`) – Business logic orchestration
* **Stores** [`internal/store/`](.) – Data persistence implementations
* **Models** [`pkg/store/`](.) – Data structure definitions

---

## Code Standards

### Go Conventions

* Use `gofmt` and `goimports`
* Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
* Keep interfaces small and focused

### Error Handling

```go
// [✓] Good: wrap errors with context
return fmt.Errorf("failed to create deployment %s: %w", id, err)

// [X] Bad: generic error message
return err
```

### Context Usage

* Always accept `context.Context` as the first parameter
* Use context for cancellation and timeouts
* Pass request or trace IDs through the context

### Testing

* Unit tests for core business logic
* Integration tests for database and I/O
* Use table-driven tests for multiple scenarios
* Mock external dependencies where possible

---

## API Changes

### OpenAPI First

1. Update the `api/openapi.yaml` specification
2. Implement handlers according to the spec
3. Update CLI commands if required
4. Add integration tests for new endpoints

### Backward Compatibility

* Avoid breaking changes in stable APIs
* Introduce **API versioning** for major updates
* Deprecate before removing functionality

---

## Runtime Support

### Adding New Runtimes

1. Add runtime constant to `pkg/store/service.go`
2. Implement detection logic in `internal/deploy/utils.go`
3. Add installation steps in `internal/deploy/deploy.go`
4. Update OpenAPI enum in `api/openapi.yaml`
5. Add CLI help text and usage examples

### Runtime Requirements

* Must support automated installation
* Should work across Windows, Linux, and macOS
* Must include version detection
* Should support dependency management

---

## Submission Guidelines

### Pull Request Process

1. Check for existing PRs or issues related to your change
2. Create an issue to discuss the feature/fix before implementation
3. Fork and create a feature branch from `main`
4. Make focused, atomic commits
5. Add tests for all new functionality
6. Update documentation if needed
7. Ensure CI checks pass
8. Request a review from maintainers

### Commit Messages

Follow the **Conventional Commits** for ordered changelog:

```
feat(runtime): add Python 3.12 support
fix(deploy): resolve timeout on large repositories
docs(api): update deployment endpoint examples
test(store): add integration tests for user roles
```

**Types:** `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `chore`
**Scopes:** `runtime`, `deploy`, `proxy`, `store`, `api`, `cli`, `daemon`

**Commit Quality Guidelines:**
- Keep changes small and focused
- Squash and merge PRs to single commits
- Write commit messages for users, not developers
- Good commits = good release notes

### Code Review Criteria

* Functionality works as intended
* Code follows project conventions
* Tests cover new functionality
* Documentation is updated
* No security vulnerabilities ([Go Security Guide](https://go.dev/security/best-practices))
* Performance impact considered

---

## Release Process

### Version Bumping

```bash
./scripts/release.sh patch --beta   # Beta release
./scripts/release.sh patch          # Bug fixes
./scripts/release.sh minor          # New features
./scripts/release.sh major          # Breaking changes
```

### Release Checklist

* [ ] All tests pass
* [ ] Docs polished—every package should have a clear `doc.go` explaining what it does.
* [ ] Commit messages are clean (users will see these)
* [ ] Consider beta release first: `./scripts/release.sh patch --beta`
* [ ] Version bumped appropriately
* [ ] Git tag created and pushed
* [ ] GitHub Actions build succeeds
* [ ] Package managers updated

---

## Getting Help

* **Issues:** Report bugs or request features
* **Discussions:** Ask questions and share ideas
* **Discord:** Real-time community chat (link in README)
* **Email:** Contact maintainers for technical questions

---

## License

By contributing, you agree that your contributions will be licensed under the **Apache License, Version 2.0**.

---
