# Marquee

What's playing on your Jellyfin server, right now. A quiet, at-a-glance dashboard for a
side monitor or a phone: no management, no controls, just the current streams.

One Go binary serves the API and the React client.

## How it works

- The server polls Jellyfin's `GET /Sessions` every few seconds, keeps the latest
  snapshot, and pushes it to browsers over Server-Sent Events.
- Each card is built from a deliberately small subset of Jellyfin's DTOs: title,
  series/episode, artist, artwork, network (LAN / Tailscale / WAN), client + inferred OS,
  and the delivery method (direct / remux / transcode + hardware accelerator).
- Artwork is proxied through `/api/thumb/{id}?tag=…` and cropped to 16:9 in CSS.

## Requirements

- Go 1.26+
- Bun (client build and dev scripts)
- A Jellyfin server and an **admin** API key. Jellyfin blanks
  `HardwareAccelerationType` for non-admin keys, so software/hardware detection
  only works with an admin key.

## Configuration

Copy `.env.example` to `.env.local`:

| Variable | Default | Notes |
|---|---|---|
| `JELLYFIN_URL` | — | required, e.g. `http://jellyfin:8096` |
| `JELLYFIN_API_KEY` | — | required; admin key for HW detection |
| `LISTEN_ADDR` | `:3000` | |
| `POLL_INTERVAL` | `3s` | Go duration, minimum `500ms` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` or `text` |
| `TAILSCALE_RANGES` | `100.64.0.0/10` | comma-separated CIDRs treated as Tailscale |

## Development

```sh
bun install
bun run --cwd client install
bunx concurrently --version >/dev/null   # or: bun install at the root
bun run dev
```

`bun run dev` starts the Go server (with `.env.local` loaded) on `:3000` and Vite on
`:5173`; Vite proxies `/api` to the server.

The server script sources `.env.local` through `sh` rather than Bun's `--env-file`,
because `bun run --cwd server …` runs Bun's flag parser before the target command, so
the env-file values never reach `go run`. Bun also only auto-loads `.env*` from the
directory it starts in (the repo root), not the directory `go run` executes from.

```sh
bun run test    # go test ./...
bun run vet     # go vet ./...
bun run lint    # eslint (client)
```

## Build

```sh
bun run build   # client → server/internal/web/dist, then the binary ./marquee
JELLYFIN_URL=… JELLYFIN_API_KEY=… ./marquee
```

## API

| Route | Purpose |
|---|---|
| `GET /api/sessions` | current snapshot as JSON |
| `GET /api/sessions/stream` | SSE, event `snapshot` on every poll |
| `GET /api/thumb/{id}?tag=…` | artwork proxy (immutable, cached) |
| `GET /healthz` | liveness |
