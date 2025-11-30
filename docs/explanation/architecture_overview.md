# Architecture Overview

This document explains the high-level architecture of the Txlog Agent and how it
integrates with the broader Txlog ecosystem.

## System Context

The Txlog Agent acts as a bridge between the local operating system's package
manager (DNF/RPM) and the central Txlog Server.

* **Local System**: The source of truth. The agent reads the DNF history
    database directly using the `dnf history` command.
* **Txlog Agent**: The intermediary. It parses local data, normalizes it, and
    securely transmits it.
* **Txlog Server**: The destination. It stores transaction logs for analysis,
    auditing, and visualization.

## Data Flow

1. **Extraction**: The agent executes `dnf history list` and `dnf history info`
    to extract raw transaction data.
2. **Parsing**: The raw text output is parsed using regular expressions to
    build structured Go objects (`TransactionDetail`, `Package`).
3. **Synchronization**:
    * The agent queries the server for a list of already saved transaction IDs
        for the current machine.
    * It compares the local list with the server list to identify missing
        transactions.
4. **Transmission**: Missing transactions are sent one by one via a REST API
    (`POST /v1/transactions`).
5. **Verification**: The `verify` command performs a two-way check to ensure
    data consistency (checksums, package lists).

## Design Principles

* **Idempotency**: The agent is designed to run repeatedly without creating
    duplicate data. It always checks the server state before sending data.
* **Zero-Dependency on Server State**: The agent does not maintain a local
    database of "sent" items. The server is the single source of truth for what
    has been archived.
* **Fail-Safe**: If a transaction fails to send, the process stops or logs the
    error, but data is never lost because it remains in the local DNF history.
