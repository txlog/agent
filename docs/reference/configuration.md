<!-- markdownlint-disable MD033 MD013 -->
# Configuration File Reference

This document details the configuration parameters for the Txlog Agent.

**File Location:** `/etc/txlog.yaml`
**Format:** YAML

## Server Configuration (`server`)

Parameters related to the connection with the central Txlog Server.

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `server.url` | string | **Yes** | The base URL of the Txlog Server<br>(e.g., `https://txlog.example.com`). |
| `server.api_key` | string | No | API Key for authentication.<br>Requires Server version >= 1.14.0. |
| `server.username` | string | No | Username for Basic Authentication. |
| `server.password` | string | No | Password for Basic Authentication. |

## Agent Configuration (`agent`)

Parameters controlling the agent's internal behavior.

| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `agent.check_version` | boolean | `true` | If `true`, checks for newer agent versions<br>on execution. |

## Example Configuration

```yaml
server:
  url: https://txlog.internal:8080
  # Authentication (Choose one method)
  # Method 1: API Key (Recommended)
  api_key: "txlog_prod_..."
  # Method 2: Basic Auth
  # username: "admin"
  # password: "secure_password"

agent:
  check_version: true
```
