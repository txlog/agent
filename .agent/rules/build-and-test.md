---
trigger: always_on
---

# Build & Test Commands

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

## Testing Patterns

Tests use Go's standard `testing` package. Key patterns:

- **Table-driven tests**: See `cmd/version_test.go` for version comparison examples
- **Viper reset**: Always `viper.Reset()` in test setup to avoid state pollution between tests
- Tests verify behavior, not internal state (e.g., checking headers set correctly in `util/main_test.go`)
