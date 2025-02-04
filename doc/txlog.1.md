---
title: txlog
section: 1
header: Package Transaction Log - Agent
footer: txlog 0.1.3
date: February 03, 2025
---

# NAME

**txlog**: Compile data on the number of updates and installs using
yum / dnf transaction info.

# SYNOPSIS

**txlog** [*OPTION*]

# DESCRIPTION

**txlog** aims to track package transactions on RPM systems, compiling data on the
number of updates and installs. Designed to enhance system reliability, this
initiative collects and centralizes information, providing valuable insights
into the evolution of packages.

# OPTIONS

**-b**, **\--build**
: Compile transaction info

**-q**, **\--query**
: List compiled transactions

**-v**, **\--version**
: Show version number

**-h**, **\--help**
: You know what this option does

# CONFIGURATION FILE

**/etc/txlog.yaml**
Normally `txlog` uses sane defaults, but if you want to activate any option or
integration, go to this file, uncomment the section and modify it. Useful during
development, since you can set another parameters for this environment.

# CONFIGURATION OPTIONS

All data is sent to the Transaction Log server.

**url**
: URL address of a Transaction Log server instance

# QUERY FILTERS

You can pass filters to the **query** option to display the compiled
transactions for a given host or transaction.

**-q**
: returns transaction list from the current machine

**-q "machine_id: 123456789"**
: returns transaction list from the machine in question

**-q "machine_id: 123456789, transaction_id: 1"**
: returns information from transaction `1`

**-q "machine_id: 123456789, transaction_id: 1-5"**
: returns information from transactions `1` through `5`

# BUGS

Submit bug reports online at
<https://github.com/txlog/agent/issues>

# SEE ALSO

Full documentation and sources at
<https://github.com/txlog>
