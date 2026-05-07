# refreSSH

A cross-platform (Linux, Windows, macOS) persistent terminal session manager, optimized for long-running AI CLI agents.

[![GitHub](https://img.shields.io/github/v/release/strickdd/refreSSH?include_prereleases)](https://github.com/strickdd/refreSSH)
[![Go Report Card](https://goreportcard.com/badge/github.com/strickdd/refreSSH)](https://goreportcard.com/report/github.com/strickdd/refreSSH)

## Overview

**refreSSH** allows you to maintain persistent shell sessions that survive SSH disconnections. Unlike traditional multiplexers, it is designed with a background daemon architecture that supports both a TUI and a local-only Web UI for parallel session management.

## Features

- **Persistence:** Sessions stay alive even after you disconnect from SSH.
- **Cross-Platform:** Full support for Windows, Linux, and macOS.
- **Dual Interface:** Manage sessions via CLI/TUI or a local-only Web UI.
- **AI Agent Optimized:** Specifically designed to host long-running AI CLI agents that require stable environments.
- **Local-Only Web UI:** Built-in terminal emulation (xterm.js) accessible only from `127.0.0.1` for security.

## Prerequisites

### For Development
- **Go:** 1.22 or higher
- **Git**
- **Beads (`bd`):** Used for issue tracking and project management.

### For Running
- A terminal emulator with PTY support.

## Installation

### From Source
```bash
git clone https://github.com/strickdd/refreSSH.git
cd refreSSH
go mod tidy
go build -o refressh .
```

## Usage

*Note: CLI documentation will be updated as features are implemented.*

```bash
# Start the daemon and open the TUI
refressh

# List active sessions
refressh list

# Attach to a specific session
refressh attach <session-id>
```

## Contributing

We use **Beads** for task tracking. Before starting work, please check for open issues:

```bash
bd ready
```

---
Project Repository: [https://github.com/strickdd/refreSSH](https://github.com/strickdd/refreSSH)
