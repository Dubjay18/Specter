# Specter
<p align="center">
  <img src="https://akns-images.eonline.com/eol_images/Entire_Site/20241013/cr_1024x759-241113062653-GettyImages-1088051728.jpg?fit=around%7C1024:759&output-quality=90&crop=1024:759;center,top" width="380" alt="Harvey Specter" />
</p>

<p align="center">
  <em>"I don't shadow test. I shadow WIN."</em><br/>
  — Harvey Specter, probably
</p>

<h1 align="center">Specter</h1>

<p align="center">
  A distributed shadow-mode traffic mirror with divergence analysis.<br/>
  Because your rewrite <em>says</em> it works. Specter makes it prove it.
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/Dubjay/specter"><img src="https://goreportcard.com/badge/github.com/Dubjay/specter" alt="Go Report Card" /></a>
  <a href="https://github.com/Dubjay/specter/actions/workflows/ci.yml"><img src="https://github.com/Dubjay/specter/actions/workflows/ci.yml/badge.svg" alt="Build" /></a>
  <img src="https://img.shields.io/badge/status-in%20progress-yellow" />
  <img src="https://img.shields.io/badge/built%20with-Go-00ADD8?logo=go" />
  <img src="https://img.shields.io/badge/inspired%20by-Uber%20Ringpop-black" />
</p>

---

## What is Specter?

Specter sits in front of your services and plays the long game.

Every request that comes in gets forwarded to your **live** service as normal.
Simultaneously, a silent copy gets fired at your **shadow** service — a canary,
a rewrite, a new version, whatever you're testing. The shadow response never
reaches the client. Instead, Specter compares the two, logs every divergence,
and builds a statistical profile of how your new service behaves under *real*
production traffic.

No fake load tests. No synthetic data. The real thing — with zero risk.

> **"Anyone can do it with perfect data. You want to be great?**
> **Test against the messy stuff."**
> — Harvey Specter (we're paraphrasing)

---

## Why Specter?

Every team doing a service rewrite, database migration, or language port faces
the same problem: *you can't know if the new thing is correct until real traffic
hits it — but you can't risk real traffic hitting it until you know it's correct.*

That's the catch-22. Specter breaks it.

| Without Specter | With Specter |
|---|---|
| "It passed staging" 🤞 | "It matched 99.97% of production traffic" ✅ |
| Find bugs after cutover | Find bugs before cutover |
| Blind confidence | Evidence-based confidence |
| Sleepless deploy nights | Boring deploy afternoons |

---

## Quick Start

Three commands to run Specter locally with Docker Compose:

```bash
git clone https://github.com/Dubjay/specter.git && cd specter
docker compose -f docker/docker-compose.yaml up --build
curl -s -H "X-User-ID: user-123" http://127.0.0.1:8080/profile
```

Optional: inspect aggregated divergence stats:

```bash
curl -s http://127.0.0.1:8080/api/stats
```

## TUI Dashboard

The terminal UI polls `GET /api/stats` every second and shows live divergence metrics.

For a one-command local demo (starts mock upstreams + Specter proxy + TUI):

```bash
make tui-demo
```

### Run TUI against a running Specter instance

If Specter is already running on `:8080` (for example via Docker Compose):

```bash
TERM=xterm-256color go run ./cmd/specter --config internal/config/specter.yaml --ui tui
```

### Local demo flow (without Docker)

Start the two mock upstreams and Specter proxy:

```bash
go run ./cmd/testserver --port 3000 --mode live
go run ./cmd/testserver --port 3001 --mode shadow
go run ./cmd/specter --config internal/config/specter.yaml --ui proxy
```

Generate some traffic in another terminal:

```bash
for i in $(seq 1 20); do
  curl -s -H "X-User-ID: user-$i" http://127.0.0.1:8080/profile > /dev/null
done
```

Then open the TUI:

```bash
TERM=xterm-256color go run ./cmd/specter --config internal/config/specter.yaml --ui tui
```

### Controls

- `j`/`k` or `↑`/`↓`: move selection/scroll
- `Enter`: open selected divergence drill-down
- `Esc` or `b`: return to dashboard from drill-down
- `q` (or `Ctrl+C`): quit

## Config File Reference

Specter config lives in YAML (examples: `internal/config/specter.yaml`, `internal/config/specter-1.yaml`, `internal/config/specter-2.yaml`).

### `specter`

- `listen` (string): Address and port for Specter to bind to, for example `":8080"`. Default: `":8080"`.
- `live_target` (string): Base URL for the live/primary upstream service that serves client traffic.
- `shadow_target` (string): Base URL for the shadow/candidate upstream service used for mirrored requests.
- `routing_key` (string): HTTP header name used for deterministic routing/ring ownership, e.g. `"X-User-ID"`. Default: `"X-User-ID"`.

### `cluster`

- `node_name` (string): Unique node identifier in the Specter cluster.
- `bind_addr` (string): Gossip/memberlist bind address (host:port), for example `"0.0.0.0:7946"`.
- `peers` (array of strings): Seed peers to join on startup, e.g. `["10.0.0.10:7946", "10.0.0.11:7946"]`.

### `store`

- `backend` (string): Storage backend. Supported values are `"badger"` and `"postgres"`. Default: `"badger"`.
- `badger_path` (string): Filesystem path for local BadgerDB data. Used when `backend: "badger"`. Default: `"./data/specter"`.
- `postgres_dsn` (string): DSN/connection string for Postgres. Used when `backend: "postgres"`.

### `sampling`

- `rate` (float): Mirror percentage from `0.0` to `1.0` (`1.0` means mirror all requests). Default: `1.0`.
- `divergence_only` (bool): If `true`, only stores events where live vs shadow responses diverge. Default: `false`.

## How the Ring Works

Specter uses consistent hashing with virtual nodes: each cluster node is inserted many times on a hash ring, and each request key (from `specter.routing_key`, such as `X-User-ID`) is hashed to the nearest clockwise position. That determines request ownership consistently across the cluster, minimizes remapping when nodes join/leave, and keeps traffic distribution smoother than single-slot-per-node hashing.

## Contributing

Contributions are welcome.

1. Fork and clone the repo.
2. Create a feature branch.
3. Run checks locally: `go test ./...`.
4. Open a pull request with a clear summary of the change and any behavior/UX impact.

If your change adds or modifies behavior, include corresponding tests under `internal/...` where appropriate.
