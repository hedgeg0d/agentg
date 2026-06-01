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
- **Access control** — declarative whitelist policy plus runtime management:
  admins, allowed Telegram IDs and `@usernames`, an optional shared password, and
  password sessions with a configurable TTL. Admins manage the whitelist from inside
  Telegram via the *Users* menu — no redeploy needed.

## Access control

Policy is seeded from `config.json` and combined with runtime state persisted to
`data/access.json`:

```json
"access": {
  "admins": [111111111],
  "allowed_users": [222222222],
  "allowed_usernames": ["alice"],
  "password": "s3cret",
  "session_ttl_minutes": 1440
}
```

- **admins** — full access, plus the *Users* menu to grant/revoke others.
- **allowed_users / allowed_usernames** — granted access (usernames match
  case-insensitively, with or without the leading `@`).
- **password** — if set, unknown users are prompted for it; a correct answer opens a
  session valid for `session_ttl_minutes` (`0` = never expires). Leave empty to deny
  everyone not on the whitelist. Store a **bcrypt hash** rather than the plaintext —
  generate one with `./agentg -hashpw` and paste the `$2a$…` string into the config.
  Plaintext is still accepted (with a startup warning); both are checked in constant
  time, and the password is never written to `access.json`.
- **Bootstrap** — if no admin is configured *and* no password is set, the first user
  to message the bot is registered as admin. When using a password, configure at
  least one admin explicitly.

Everyone else gets a clear *access denied* reply showing their ID and username so an
admin can add them in two taps.

## Architecture

```
main.go                 entry point, wiring
internal/config         declarative configuration
internal/auth           access policy, password sessions, whitelist persistence
internal/store          pinned services (JSON persistence)
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

`/start` · `/shell` · `/monitor` · `/services` · `/users` (admin) · `/id`
