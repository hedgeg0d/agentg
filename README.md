# agentg

A Telegram bot for managing a Linux machine — built with native Telegram UX in mind:
reply keyboards, inline buttons, and live auto-editing messages.

Pure Go, **no cgo**, so it cross-compiles cleanly to targets like RISC-V.

## Features

- **Shell** — persistent bash session per chat (working directory and environment
  survive between commands), with command timeouts and auto-recovery.
- **Monitor** — a single message that auto-edits every few seconds with CPU, RAM,
  swap, disk, load and uptime. A *Stop* button under it ends the stream.
- **Services** — inspect `systemd` units, start/stop/restart them, and pin favorites
  to a quick-access list for one-tap control.
- **Status** — one-shot host snapshot (hostname, kernel, resources).
- **Single-owner security** — the first user to `/start` claims the machine; everyone
  else is denied.

## Architecture

```
main.go                 entry point, wiring
internal/config         config loading
internal/store          owner + pinned services (JSON persistence)
internal/shell          persistent bash sessions
internal/monitor        /proc-based resource sampling + rendering
internal/services       systemd control
internal/bot            Telegram handlers, keyboards, routing
```

Resource metrics are read directly from `/proc` and `statfs(2)` — no external
dependencies beyond the Telegram client.

## Setup

```sh
cp config.example.json config.json   # then put your bot token in it
go build -o agentg .
./agentg
```

Cross-compile for RISC-V:

```sh
CGO_ENABLED=0 GOARCH=riscv64 GOOS=linux go build -o agentg .
```

`config.json` and `data/` are git-ignored and never committed.

## Commands

`/start` · `/shell` · `/monitor` · `/services` · `/id`
