# refreSSH

A cross-platform (Linux, Windows, macOS) persistent terminal session manager, optimized for long-running AI CLI agents.

[![CI](https://github.com/strickdd/refreSSH/actions/workflows/ci.yml/badge.svg)](https://github.com/strickdd/refreSSH/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/strickdd/refreSSH)](https://goreportcard.com/report/github.com/strickdd/refreSSH)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Overview

**refreSSH** allows you to maintain persistent shell sessions that survive SSH disconnections. Unlike traditional multiplexers, it is designed with a background daemon architecture that supports both a TUI and a local-only Web UI for parallel session management.

## Features

- **Persistence:** Sessions stay alive even after you disconnect from SSH.
- **Cross-Platform:** Full support for Windows, Linux, and macOS.
- **TUI Management:** Manage sessions via a rich Terminal User Interface.
- **AI Agent Optimized:** Specifically designed to host long-running AI CLI agents that require stable environments.
- **Web UI:** A local-only web interface with terminal emulation (xterm.js), accessible via `refressh ui`.
- **Sandbox Deployments:** Spin up completely isolated, containerized daemon environments using Docker.

## Prerequisites

### For Development
- **Go:** 1.24 or higher
- **Git**
- **Beads (`bd`):** Used for issue tracking and project management.

### For Running
- A terminal emulator with PTY support.
- **Docker & Docker Compose** (Optional, required only for Sandbox deployments).

## Installation

### From Source
```bash
git clone https://github.com/strickdd/refreSSH.git
cd refreSSH
go mod tidy
go build -o refressh ./cmd/refressh
```

## Usage

### Standard Mode
```bash
# Start the daemon
refressh daemon start

# Create a new session
refressh create my-session bash

# List active sessions
refressh list

# Attach to a specific session
refressh attach my-session

# Stop a session
refressh stop my-session
```

### Sandbox Mode (Isolated Container)
You can run the refreSSH daemon inside an isolated Docker container for maximum security. The CLI automatically handles building and deploying the container, securely mapping the API back to your local machine.

```bash
# Start an isolated Sandbox container
refressh sandbox start

# The sandbox is now running on 127.0.0.1:9000.
# You can connect to it using the standard UI or CLI commands.
refressh ui

# Stop the Sandbox container
refressh sandbox stop
```

## Contributing

We use **Beads** for task tracking. Before starting work, please check for open issues:

```bash
bd ready
```

---
Project Repository: [https://github.com/strickdd/refreSSH](https://github.com/strickdd/refreSSH)
