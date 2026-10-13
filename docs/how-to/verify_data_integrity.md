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

### Case 3: Missing or Extra Items

If the output lists packages that differ between a stored transaction and the
local history:

```text
✗ Transaction #49 is missing 34 package(s) on server
⚠ Transaction #49 has 34 extra package(s) on server
```

Missing and extra counts that match, transaction after transaction, mean the
same packages are recorded on both sides but described differently. That is the
normal outcome of an agent upgrade that changed how package names are parsed.

**Action**: Overwrite what the server stores:

```bash
txlog build --force
```

A regular `txlog build` will not help here: the server already has these
transactions, so it skips them. `--force` sends them anyway and the server
replaces the stored item list. This requires Txlog Server 1.38.0 or later.

### Case 4: Unexplained Mismatches

If the differences do not line up with an agent upgrade, and the same package
appears with genuinely different versions on each side, treat it as possible
tampering with the local history:

1. Check your local logs.
2. Contact your system administrator if you suspect a security issue.
