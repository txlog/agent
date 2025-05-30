---
title: txlog
section: 1
header: Transaction Log - Agent
footer: txlog 1.6.0
date: May 27, 2025
---

# NAME

**txlog**: Compile data on the number of updates and installs using
yum / dnf transaction info.

# SYNOPSIS

**txlog** [*COMMAND*]

# DESCRIPTION

The **txlog** command is a tool for compiling and sending transaction data from
RPM-based systems to the Txlog server. It collects information about package
installations, updates, and removals, providing a comprehensive view of system
changes over time. This data can be used for monitoring, analytics, and
troubleshooting purposes. The agent operates by reading transaction logs
generated by package managers like `yum` or `dnf`, processing the information,
and then sending it to a specified Txlog server for storage and analysis. The
agent is designed to be lightweight and efficient, minimizing its impact on
system performance while ensuring accurate and timely data collection.

# COMMANDS

**build**
: Compile transaction info

**help**
: You know what this option does

**version**
: Show agent and server version number

## FLAGS

**--config**
: config file (default is /etc/txlog.yaml)

**--help**
: help for txlog. Use "txlog [command] --help" for more information about a command

# CONFIGURATION FILE

**/etc/txlog.yaml**
Normally `txlog` uses sane defaults, but if you want to activate any option or
integration, go to this file, uncomment the section and modify it. Useful during
development, since you can set another parameters for this environment.

# CONFIGURATION OPTIONS

All data is sent to the Transaction Log server.

## Agent section

**check_version** (boolean)
: Controls whether the agent checks for new versions when running txlog version.
When enabled, the agent will contact https://txlog.rda.run/agent/version to
verify if a new version is available. Default: true

## Server section

**url** (string)
: Specifies the URL of the txlog server where logs will be sent. Must include
protocol (http/https) and port if not using defaults. Default: http://localhost:8080

**username** (string, optional)
: Username for basic authentication with the txlog server. Must be uncommented to
enable authentication. Default: not set

**password** (string, optional)
: Password for basic authentication with the txlog server. Must be uncommented to
enable authentication. Should be used in conjunction with username. Default: not set

# BUGS

Submit bug reports online at
<https://github.com/txlog/agent/issues>

# SEE ALSO

Full documentation and sources at
<https://txlog.rda.run>
