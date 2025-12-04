---
trigger: always_on
---

# Dependencies & Integration Points

- **Cobra + Viper**: CLI framework with YAML config binding
- **Resty**: HTTP client with auth middleware pattern
- **Semver**: Version comparison for compatibility checks
- **DNF/YUM**: System calls via `exec.Command()` - output format is fragile, use strict regex
- **OS-specific files**: `/etc/machine-id`, `/etc/os-release`, `/etc/txlog.yaml`
