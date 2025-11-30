# Getting Started with Txlog Agent

This tutorial will guide you through the process of installing, configuring, and
running the Txlog Agent for the first time. By the end of this guide, you will
have a working agent syncing your system's RPM transaction history to a central
Txlog Server.

## Prerequisites

Before you begin, ensure you have:

1. **A Linux system using RPM/DNF** (e.g., Fedora, RHEL, CentOS, AlmaLinux).
2. **Root or sudo access** to the system.
3. **A running Txlog Server** instance (you will need its URL).
    * *Note: If you don't have a server yet, ask your team lead for the
        `https://...` endpoint.*

## Step 1: Install the Agent

First, we need to add the repository and install the package. Open your terminal
and run:

```bash
# Add the repository
sudo dnf localinstall -y https://rpm.rda.run/repo.rpm

# Install the txlog agent
sudo dnf install -y txlog
```

You can verify the installation by checking the version:

```bash
txlog version
```

You should see output similar to: `txlog version 1.2.3 ...`

## Step 2: Configure the Agent

The agent needs to know where to send the data. The configuration file is
located at `/etc/txlog.yaml`.

1. Open the file with your preferred editor (e.g., `nano` or `vi`):

    ```bash
    sudo nano /etc/txlog.yaml
    ```

2. Locate the `server` section. It will look something like this:

    ```yaml
    server:
      url: https://txlog-server.example.com:8080
      # api_key: ...
    ```

3. **Update the URL** to point to your actual Txlog Server.

4. (Optional) If your server requires authentication, uncomment and set the
    `api_key` or `username`/`password`.

    *Example with API Key:*

    ```yaml
    server:
      url: https://my-txlog.internal:443
      api_key: txlog_prod_123456
    ```

5. Save and exit the file.

## Step 3: Sync Your First Transactions

Now that the agent is configured, let's send your current RPM history to the
server.

Run the `build` command:

```bash
txlog build
```

**What happens next?**
The agent will read your local DNF history database, compile the transaction
logs, and upload them to the configured server.

If successful, you will see output indicating the synchronization status.

## Step 4: Verify the Data

To ensure everything is in sync, use the `verify` command. This checks if your
local state matches what the server has recorded.

```bash
txlog verify
```

* **Success**: If the output is clean or says "All transactions verified",
    you are good to go!
* **Failure**: If it lists missing transactions, run `txlog build` again to
    sync them.

## Next Steps

Congratulations! You have successfully set up the Txlog Agent.

* Check out the **How-to Guides** to learn about configuring authentication
    or running in CI/CD.
* Read the **Explanation** section to understand how the agent ensures data
    integrity.
