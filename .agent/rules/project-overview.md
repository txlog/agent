---
trigger: always_on
---

# Project Overview

This is a Go-based RPM transaction monitoring agent for RHEL/CentOS systems. It collects DNF/YUM transaction history and sends it to a central Txlog server for monitoring and analytics.

**Architecture**: CLI tool (Cobra) → parses DNF history → sends to REST API server
**Key flow**: `txlog build` → fetch saved transactions → parse local DNF history → send unsent transactions → save execution metadata
