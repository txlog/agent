<!-- markdownlint-disable MD033 MD013 -->
# Environment Variables Reference

This document lists the environment variables that affect the Txlog Agent's
behavior.

| Variable | Description | Values |
| :--- | :--- | :--- |
| `NO_COLOR` | Disables ANSI color output in the terminal. | Any non-empty string<br>(e.g., `1`, `true`). |

## Usage Example

```bash
# Run build without color output
NO_COLOR=1 txlog build
```
