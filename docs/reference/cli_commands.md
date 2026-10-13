<!-- markdownlint-disable MD033 MD013 -->
# CLI Command Reference

This document provides a technical reference for the `txlog` command-line
interface.

## Global Flags

The following flags apply to all commands:

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--config` | string | `/etc/txlog.yaml` | Path to the configuration file. |

## Commands

### `txlog build`

Compiles transaction information from the local DNF history and synchronizes it
with the configured Txlog Server.

**Usage:**

```bash
txlog build [flags]
```

**Flags:**

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--force` | bool | `false` | Re-send every transaction, overwriting what is stored<br>on the server. Requires Txlog Server 1.38.0 or later. |

By default the agent asks the server which transactions it already has and sends
only the rest. With `--force` it sends all of them and the server replaces the
stored transaction and its item list. Use it after an agent upgrade that changed
how transaction data is parsed, instead of deleting the host from the server.

**Exit Codes:**

| Code | Description |
| :--- | :--- |
| `0` | Success. All transactions processed and sent<br>(or already up-to-date). |
| `1` | Error. Failed to retrieve transactions, connect to server,<br>or save data. |

### `txlog verify`

Verifies data integrity between the local DNF history and the server's records.
Checks for missing transactions and package mismatches.

**Usage:**

```bash
txlog verify [flags]
```

**Exit Codes:**

| Code | Description |
| :--- | :--- |
| `0` | Success. Data is fully synchronized and verified. |
| `1` | Failure. Integrity issues detected (missing transactions,<br>extra items, missing items) or execution error. |

### `txlog version`

Displays the current version of the Txlog Agent and the connected Txlog Server.
Also checks for available updates.

**Usage:**

```bash
txlog version [flags]
```

**Exit Codes:**

| Code | Description |
| :--- | :--- |
| `0` | Success. |
