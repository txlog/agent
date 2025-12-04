---
trigger: always_on
---

# Key Files & Their Roles

- **`cmd/build.go`**: Core business logic - transaction diffing, DNF parsing, HTTP submission
- **`cmd/root.go`**: Config loading (expects `/etc/txlog.yaml`), version validation on startup
- **`util/main.go`**: Shared utilities (auth, package name parsing, machine-id reading)
- **`util/needsRestarting.go`**: Checks if system needs reboot via `dnf needs-restarting -r`
- **`nfpm.yaml`**: RPM package definition - version MUST match `cmd/version.go`
- **`Makefile`**: All build targets use specific Go flags (`-ldflags="-s -w -extldflags=-static"` for static binaries)
