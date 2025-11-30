# How to Secure the Configuration

This guide explains how to protect the sensitive credentials stored in your
`/etc/txlog.yaml` file.

## Set File Permissions

Since the configuration file contains secrets (API keys or passwords), it should
only be readable by the root user.

1. Open a terminal.

2. Change the file ownership to root (if not already):

    ```bash
    sudo chown root:root /etc/txlog.yaml
    ```

3. Set the permissions to `600` (read/write for owner only):

    ```bash
    sudo chmod 600 /etc/txlog.yaml
    ```

## Verify Permissions

Check the permissions to confirm they are correct:

```bash
ls -l /etc/txlog.yaml
```

**Expected Output**:

```text
-rw------- 1 root root ... /etc/txlog.yaml
```

If you see `-rw-r--r--` or similar, run the `chmod` command again.
