# RemoteDesk

Web interface for monitoring and managing office computers. Each machine runs a small
agent that reports its status, and this app shows the fleet on one dashboard, pairs new
devices with a shared token, and emulates firewall management, update rollouts, computer
groups, and audit logs. The agents and all device actions are simulated, nothing here
controls a real machine.

University web-dev coursework, built on the GoTTH stack (Go, templ, Tailwind, HTMX, SQLite).

## Quickstart

```bash
make dev
```

Builds the app and the agent, starts two paired demo agents (`lab-pc-1`, `lab-pc-2`),
and runs the site on http://localhost:4000. Register `admin@localhost` (any password)
and open the Computers page, both machines are already there, verified.

## Running it manually

```bash
go build -o bin/app ./cmd && go build -o bin/agent ./cmd/agent

AGENT_TOKEN=dev-token AGENT_SERVER=http://localhost:4000 AGENT_URL=http://localhost:4101 \
  PORT=:4101 AGENT_NAME=pc-1 ./bin/agent &

ADMIN_EMAIL=admin@localhost AGENT_TOKEN=dev-token SESSION_SECRET=dev-secret ./bin/app
```

Agents phone home on boot and every `AGENT_HEARTBEAT_SECS` (default 30), so they
appear on the admin page by themselves. Anything the server can reach works as a
device URL, local ports or public tunnel addresses.

## Environment variables

Web app:

| Variable            | Default    | Meaning                                              |
|---------------------|------------|------------------------------------------------------|
| `SERVERPORT`        | `:4000`    | Listen address                                       |
| `DATABASE_NAME`     | `goth.db`  | SQLite file (relative to working directory)          |
| `DATABASE_URL`      | —          | Postgres URL, used only with `ENV=production`        |
| `STATIC_DIR`        | `./static` | Static assets directory                              |
| `SESSION_SECRET`    | —          | Signs session cookies; empty means sessions die on restart |
| `ADMIN_EMAIL`       | —          | Login email of the admin; empty means nobody is admin |
| `AGENT_TOKEN`       | —          | Shared secret agents must present; empty disables auto-registration |

Agent:

| Variable              | Default              | Meaning                              |
|-----------------------|----------------------|--------------------------------------|
| `AGENT_NAME`          | `office-pc`          | Reported name                        |
| `AGENT_VERSION`       | `1.0.0`              | Reported version                     |
| `AGENT_OS`            | `NixOS 25.05`        | Reported OS                          |
| `PORT`                | `:4100`              | Listen address                       |
| `AGENT_URL`           | `http://localhost...`| Address the server uses to reach it  |
| `AGENT_SERVER`        | —                    | Web app base URL; empty disables registration |
| `AGENT_TOKEN`         | —                    | Must match the server's `AGENT_TOKEN`|
| `AGENT_HEARTBEAT_SECS`| `30`                 | Registration interval (min 5)        |

## Features

- Computers dashboard with live status, last-known version, and last-seen time.
- Token pairing: agents self-register, verified devices get a checkmark.
- Firewall rules per device, applied to the agent (emulated).
- Update tasks for devices or groups, immediate or scheduled, with live progress.
- Computer groups with dashboard filtering.
- Audit logs, account page (email/password change), admin user management.

## Nix

`flake.nix` builds both binaries (`nix build .#default` gives `bin/goth` and
`bin/agent`) and provides two NixOS modules: `services.goth` for the web app
(`port`, `adminEmail`, `agentToken`, `sessionSecret`) and `services.gothAgent`
to run a list of self-registering agent instances.

## Make targets

- `make dev` — build everything, run two demo agents plus the app.
- `make build` — build both binaries into `./bin/`.
- `make test` — full test suite (`go test -race`).
- `make vet`, `make staticcheck` — CI lint steps.
- `make templ-generate` — regenerate templates (needs `templ`).
- `make tailwind-build` — rebuild minified CSS (needs `./tailwindcss`).
