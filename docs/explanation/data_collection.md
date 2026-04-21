# Data Collection Explanation

This document explains what data the txlog agent collects and why.

## What We Collect

### System Identification

- **Machine ID**: Read from `/etc/machine-id`. This persistent identifier allows tracking the same server across reinstalls and hostname changes.
- **Hostname**: Obtained via the `hostname` command. Used for human-friendly server identification.

### Package Data

- **Transaction History**: All DNF/yum transactions including:
  - Transaction ID and timestamp
  - User who executed the command
  - Command line used
  - Packages affected (name, version, architecture, action)
  - Scriptlet output from package installation

### System State

- **OS Information**: Pretty name from `/etc/os-release`
- **Restart Status**: Whether the system needs restarting after updates
- **Agent Version**: For compatibility tracking

## Why We Collect It

The txlog system is designed for:

- **Compliance Auditing**: Track who made changes and when
- **Security Monitoring**: Identify outdated or vulnerable packages across the fleet
- **Change Management**: Understand the history of system modifications
- **Troubleshooting**: Correlate system issues with recent package changes

## Privacy Considerations

The collected data may be considered personally identifiable information (PII) under certain regulations:

- Machine IDs can uniquely identify systems
- Hostnames often contain organizational information
- Command lines may reveal application names or configurations
- Usernames identify individuals who made changes

Organizations should review their data protection obligations before deploying txlog.
