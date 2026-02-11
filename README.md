<div align="center">

# dipt

**Docker Image Pull Tool**

Pull and save Docker images as tar files — no Docker daemon required.

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

[English](README.md) · [中文](README_zh.md)

</div>

---

## Why dipt?

Sometimes you need Docker images but can't run `docker pull` — no daemon, restricted networks, air-gapped environments. **dipt** talks directly to container registries via [go-containerregistry](https://github.com/google/go-containerregistry) and saves images as standard tar files you can `docker load` anywhere.

## Highlights

- **Interactive TUI** — Powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea), with real-time progress and log viewer
- **Mirror Registries** — Auto-detect, health-check, and fallback across multiple mirrors
- **Multi-Platform** — linux / windows / darwin × amd64 / arm64 / arm / 386
- **Private Registries** — Username/password auth with secure password input
- **Smart Retry** — Exponential backoff with jitter on transient failures
- **Three-Tier Config** — Environment variables > project config > user config

## Quick Start

```bash
# Build
git clone <repo-url> && cd dipt
go build -o dipt ./cmd/dipt

# Launch TUI
./dipt
```

The first run walks you through a setup wizard. After that, the main menu gives you three options:

| Menu | What it does |
|------|-------------|
| **Pull Image** | Enter image name, pick platform, download as `.tar` |
| **Settings** | Default OS, arch, save dir, registry credentials |
| **Mirrors** | Add, remove, test mirror registries |

## Keyboard

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Navigate |
| `Enter` | Confirm |
| `Tab` / `Shift+Tab` | Next / previous field |
| `←→` | Cycle options |
| `Esc` | Back |
| `q` / `Ctrl+C` | Quit |

## Configuration

### User config `~/.dipt_config`

Auto-created by the setup wizard:

```json
{
  "default_os": "linux",
  "default_arch": "amd64",
  "default_save_dir": "./images",
  "registry": {
    "mirrors": ["https://mirror.example.com"],
    "username": "",
    "password": ""
  }
}
```

### Project config `./config.json`

Optional override per project. Same schema, higher priority.

### Environment variables

| Variable | Description |
|----------|-------------|
| `DIPT_DEFAULT_OS` | Default OS |
| `DIPT_DEFAULT_ARCH` | Default architecture |
| `DIPT_DEFAULT_SAVE_DIR` | Default save directory |
| `DIPT_REGISTRY_MIRRORS` | Comma-separated mirror URLs |
| `DIPT_REGISTRY_USERNAME` | Registry username |
| `DIPT_REGISTRY_PASSWORD` | Registry password |
| `DIPT_CUSTOM_MIRROR` | Prepend a custom mirror |
| `DIPT_TIMEOUT` | Timeout in seconds (default `120`) |
| `DIPT_NO_INTERACTIVE=1` | Skip setup wizard |
| `DIPT_DRY_RUN=1` | Dry-run mode |

> Priority: env vars > `./config.json` > `~/.dipt_config`

## License

[MIT](LICENSE)
