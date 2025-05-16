---
title: txlog
section: 1
header: Transaction Log - Agent
footer: txlog 1.5.1
date: May 16, 2025
---

# NAME

**txlog**: Compile data on the number of updates and installs using
yum / dnf transaction info.

# SYNOPSIS

**txlog** [*COMMAND*]

# DESCRIPTION

**txlog** aims to track package transactions on RPM systems, compiling data on the
number of updates and installs. Designed to enhance system reliability, this
initiative collects and centralizes information, providing valuable insights
into the evolution of packages.

# COMMANDS

**build**
: Compile transaction info

**executions**
: List build executions

**help**
: You know what this option does

**items**
: List transactions items

**machine_id**
: List the machine_id of the given hostname

**transactions**
: List compiled transactions

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

**url**
: URL address of a Transaction Log server instance

# BUGS

Submit bug reports online at
<https://github.com/txlog/agent/issues>

# SEE ALSO

Full documentation and sources at
<https://txlog.rda.run>
