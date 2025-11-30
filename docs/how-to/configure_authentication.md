# How to Configure Authentication

This guide explains how to configure the Txlog Agent to authenticate with a
secured Txlog Server.

## Configure API Key Authentication

If your server uses API Key authentication (recommended for service accounts):

1. Open the configuration file:

    ```bash
    sudo nano /etc/txlog.yaml
    ```

2. Find the `server` section.

3. Uncomment the `api_key` line and paste your key:

    ```yaml
    server:
      url: https://your-server.com
      api_key: your_secret_api_key_here
    ```

4. Save the file.

## Configure Basic Authentication

If your server uses Username/Password authentication:

1. Open the configuration file:

    ```bash
    sudo nano /etc/txlog.yaml
    ```

2. Find the `server` section.

3. Uncomment the `username` and `password` lines and set your credentials:

    ```yaml
    server:
      url: https://your-server.com
      username: your_username
      password: your_secure_password
    ```

4. Save the file.

## Verify the Configuration

To check if authentication is working, run a build:

```bash
txlog build
```

If authentication fails, the agent will exit with an error message indicating
"Unauthorized" or "Forbidden".
