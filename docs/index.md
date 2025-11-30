# Txlog Agent Documentation

Welcome to the documentation for the Txlog Agent. This agent compiles
transactional data from your RPM-based system and sends it to a central server
for monitoring and analytics.

## 📚 Documentation Structure

The documentation is organized into four sections:

### 1. [Tutorials](tutorials/getting_started.md)

*Learning-oriented lessons for beginners.*

* **[Getting Started](tutorials/getting_started.md)**: A hands-on guide to
    installing, configuring, and running the agent for the first time.

### 2. [How-to Guides](how-to/)

*Task-oriented guides for specific goals.*

* **[Configure Authentication](how-to/configure_authentication.md)**: Set up
    API Keys or Basic Auth.
* **[Verify Data Integrity](how-to/verify_data_integrity.md)**: Ensure local
    and server data match.
* **[Run in CI/CD](how-to/run_in_cicd.md)**: Use the agent in automated
    pipelines with `NO_COLOR`.
* **[Secure Configuration](how-to/secure_configuration.md)**: Protect your
    credentials.

### 3. [Reference](reference/)

*Information-oriented technical descriptions.*

* **[CLI Commands](reference/cli_commands.md)**: `build`, `verify`,
    `version`.
* **[Configuration](reference/configuration.md)**: `txlog.yaml` parameters.
* **[Environment Variables](reference/environment_variables.md)**: `NO_COLOR`
    and others.

### 4. [Explanation](explanation/)

*Understanding-oriented background knowledge.*

* **[Architecture Overview](explanation/architecture_overview.md)**: How the
    system works.
* **[Data Synchronization](explanation/data_synchronization.md)**: The logic
    behind the sync process.
* **[Design Choices](explanation/design_choices.md)**: Why we chose Go,
    Cobra, and Regex.

---

## Quick Links

* **Project Repository**: [github.com/txlog/agent](https://github.com/txlog/agent)
* **Issue Tracker**: [github.com/txlog/agent/issues](https://github.com/txlog/agent/issues)
