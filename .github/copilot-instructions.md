# Txlog Agent - AI Coding Assistant Instructions

## Project Overview

This is a Go-based RPM transaction monitoring agent for RHEL/CentOS systems. It collects DNF/YUM transaction history and sends it to a central Txlog server for monitoring and analytics.

**Architecture**: CLI tool (Cobra) → parses DNF history → sends to REST API server
**Key flow**: `txlog build` → fetch saved transactions → parse local DNF history → send unsent transactions → save execution metadata

## Critical Build & Test Commands

```bash
# Build production binary (CGO disabled, static linking, trimmed)
make build              # → bin/txlog

# Development workflow
make fmt               # Format all packages
make vet               # Check for issues
go test ./...          # Run tests
make man               # Build manpage (requires pandoc)
make rpm               # Build RPM package (requires nfpm)

# Run locally
go run main.go version
go run main.go build --config /path/to/txlog.yaml
```

## Project-Specific Conventions

### Version Management
- **Single source of truth**: `agentVersion` constant in `cmd/version.go` (currently "1.9.0")
- When bumping version: update `cmd/version.go`, `nfpm.yaml`, and commit together
- Server version compatibility checks are critical (e.g., API key auth requires server ≥1.14.0)

### Authentication Pattern
API key takes precedence over basic auth. All server requests use `util.SetAuthentication(request)`:
```go
// Prefer API key if set, fallback to basic auth
if apiKey := viper.GetString("server.api_key"); apiKey != "" {
    request.SetHeader("X-API-Key", apiKey)
} else if username := viper.GetString("server.username"); username != "" {
    request.SetBasicAuth(username, viper.GetString("server.password"))
}
```

### Package Manager Abstraction
Use `util.PackageBinary()` not hardcoded `yum`/`dnf`. It auto-detects based on OS version (RHEL/CentOS ≥8 uses DNF).

### Transaction Parsing Pattern
DNF output parsing uses regex with strict input validation:
```go
validInput := regexp.MustCompile(`^[a-zA-Z0-9_\-\./\\]+$`)
if !validInput.MatchString(transaction_id) {
    return TransactionDetail{}, fmt.Errorf("invalid input")
}
```
Always sanitize before exec.Command() to prevent injection.

### Date Handling
Use `util.DateConversion()` to convert DNF's varied date formats to RFC3339. Relies on `dateparse` library for fuzzy parsing.

### Error Handling in HTTP Clients
Check both network errors AND status codes:
```go
response, err := request.Post(url)
if err != nil {
    return 0, 0, err  // Network error
}
if response.StatusCode() != 200 {
    return 0, 0, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
}
```

## Key Files & Their Roles

- **`cmd/build.go`**: Core business logic - transaction diffing, DNF parsing, HTTP submission
- **`cmd/root.go`**: Config loading (expects `/etc/txlog.yaml`), version validation on startup
- **`util/main.go`**: Shared utilities (auth, package name parsing, machine-id reading)
- **`util/needsRestarting.go`**: Checks if system needs reboot via `dnf needs-restarting -r`
- **`nfpm.yaml`**: RPM package definition - version MUST match `cmd/version.go`
- **`Makefile`**: All build targets use specific Go flags (`-ldflags="-s -w -extldflags=-static"` for static binaries)

## Testing Patterns

Tests use Go's standard `testing` package. Key patterns:
- **Table-driven tests**: See `cmd/version_test.go` for version comparison examples
- **Viper reset**: Always `viper.Reset()` in test setup to avoid state pollution between tests
- Tests verify behavior, not internal state (e.g., checking headers set correctly in `util/main_test.go`)

## Documentation Standards

- **Markdown files**: All markdown files (*.md) MUST be valid according to markdownlint rules
- Use `.editorconfig` settings for consistent formatting
- Man pages are generated from `doc/txlog.1.md` using pandoc - keep markdown source clean

## Dependencies & Integration Points

- **Cobra + Viper**: CLI framework with YAML config binding
- **Resty**: HTTP client with auth middleware pattern
- **Semver**: Version comparison for compatibility checks
- **DNF/YUM**: System calls via `exec.Command()` - output format is fragile, use strict regex
- **OS-specific files**: `/etc/machine-id`, `/etc/os-release`, `/etc/txlog.yaml`

## Common Pitfalls

1. **Don't assume DNF**: Use `util.PackageBinary()` - older systems have YUM
2. **Config file path**: Default is `/etc/txlog.yaml`, not in repo root
3. **API key compatibility**: Server < 1.14.0 will reject API keys - check in `initConfig()`
4. **RPM package building**: Requires `nfpm` and `pandoc` installed (see README dev setup)
5. **Static binary requirement**: Production needs CGO disabled for portability across RHEL versions
