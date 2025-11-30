# How to Verify Data Integrity

This guide explains how to ensure that the transaction history on your local
machine matches exactly what is stored on the Txlog Server.

## Run a Verification Check

To compare your local DNF history with the server's records, run:

```bash
txlog verify
```

## Interpret the Results

### Case 1: All Clear

If the output says:

```text
All transactions verified
```

Then your data is perfectly in sync. No further action is needed.

### Case 2: Missing Transactions

If the output lists transactions that exist locally but not on the server:

```text
Missing transactions: [100, 101, 102]
```

**Action**: Run a build to sync the missing data:

```bash
txlog build
```

### Case 3: Integrity Errors

If the output indicates a mismatch in package data (e.g., checksums don't
match):

```text
Transaction 50: Package mismatch ...
```

**Action**: This indicates a potential data corruption or a modification of the
local history.

1. Check your local logs for tampering.
2. Contact your system administrator if you suspect a security issue.
