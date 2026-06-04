# agentg

A Telegram bot for managing a Linux machine — built with native Telegram UX in mind:
reply keyboards, inline buttons, and live auto-editing messages.

Pure Go, **no cgo**, so it cross-compiles cleanly to targets like RISC-V.

## Features

- **Shell** — persistent bash session per chat (working directory and environment
  survive between commands). Output streams live: the bot edits one message every
  couple of seconds as the command produces output, so long runs like `apt upgrade`
  show progress instead of going silent. An inactivity timeout restarts the session
  if a command stalls (e.g. one waiting on stdin). Frequently used commands can be
  saved as favorites and run with one tap from the shell screen.
- **Monitor** — a single message that auto-edits every few seconds with CPU, RAM,
  swap, disk, load and uptime. A *Stop* button under it ends the stream.
- **Services** — inspect `systemd` units, start/stop/restart them, and pin favorites
  to a quick-access list for one-tap control.
- **Status** — one-shot host snapshot (hostname, kernel, resources).
- **Power** — admin-only reboot / power off with an inline confirmation step.
- **D-Bus notifications** — local programs can push notifications to Telegram over
  D-Bus, and the bot can optionally pose as the freedesktop notification daemon so
  plain `notify-send` is forwarded too.
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

## D-Bus notifications

Enable the service in `config.json`:

```json
"notifications": {
  "dbus_enabled": true,
  "system_bus": false,
  "replace_notify_send": true
}
```

The bot exposes `com.agentg.Notifier` on the session bus (or the system bus when
`system_bus` is `true`). Any local program can deliver a notification to every
recipient (admins and allowed numeric IDs):

```sh
gdbus call --session \
  --dest com.agentg.Notifier \
  --object-path /com/agentg/Notifier \
  --method com.agentg.Notifier.Notify "Backup" "nightly backup finished"
```

With `replace_notify_send`, the bot also claims `org.freedesktop.Notifications`, so
standard tooling is forwarded to Telegram with no code changes:

```sh
notify-send "Disk" "root volume is 85% full"
```

This best-effort claim is skipped (with a log line) if a desktop notification daemon
already owns that name, so it is intended for headless machines.

## Shell output

```json
"shell": {
  "stream_output": true
}
```

When `stream_output` is enabled (the default — omit the block or set `true`), command
output is streamed by editing a single Telegram message at most once every two
seconds. Set it to `false` to instead post the full output once the command
completes. Output longer than a Telegram message is shown as a live tail while
running, with the complete log posted in chunks at the end.

`command_timeout_seconds` is an **inactivity** timeout: it fires only when a command
produces no output for that long, so steadily-printing commands run to completion
while a command blocked on stdin is cut off and the session restarted.

### Favorite commands

Save a command with `/fav <command>` (e.g. `/fav apt update`). On entering shell mode
the bot lists your favorites as one-tap buttons; running one prints the command above
its output so the result is self-explanatory. Send `/fav` with no argument to open the
list and remove entries.

## Architecture

```
main.go                 entry point, wiring
internal/config         declarative configuration
internal/auth           access policy, password sessions, whitelist persistence
internal/store          pinned services (JSON persistence)
internal/shell          persistent bash sessions
internal/monitor        /proc-based resource sampling + rendering
internal/services       systemd control
internal/notify         D-Bus notification service
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

`/start` · `/shell` · `/monitor` · `/services` · `/fav` · `/users` (admin) · `/power` (admin) · `/id`
