# How to Run in CI/CD Pipelines

This guide explains how to run the Txlog Agent in automated environments (CI/CD,
scripts) where cleaner logs are preferred.

## Disable Colored Output

The agent produces colored output and emojis by default, which can clutter logs
in Jenkins, GitHub Actions, or GitLab CI.

To disable this, set the `NO_COLOR` environment variable.

### Option 1: Inline Command

Run the command with the variable set for just that execution:

```bash
NO_COLOR=1 txlog build
```

### Option 2: Export Variable

Set it for the entire script session:

```bash
export NO_COLOR=1

txlog build
txlog verify
```

## Example: GitHub Actions Workflow

Here is a snippet for a GitHub Actions step:

```yaml
- name: Sync RPM History
  run: |
    sudo txlog build
  env:
    NO_COLOR: 1
```
