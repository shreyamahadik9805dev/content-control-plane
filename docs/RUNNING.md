# How to run the project

This document is the **runbook**: how to start the stack, use the operator UI, hit the API with `curl`, and run the **Go tests**.

**Currently:** **backend** (Go + Postgres + migrations), **operator UI** (`frontend/`), and Docker Compose wiring for both are in the repo.

---

## Prerequisites

- **Docker** and **Docker Compose** — recommended; runs Postgres, migrations, backend, and the Vite dev server together.
- **Without Compose:** **Go 1.25+** (see **`go.mod`**) and **PostgreSQL 16** for the API on the host, plus **Node.js 20+** and **npm** if you run **`npm run dev`** in **`frontend/`** (the Compose `frontend` image already includes Node 22).

---

## Run everything with Docker Compose (recommended)

From the repository root:

```bash
docker compose up --build
```

This starts, in order:

1. **Postgres** (`localhost:5432`, user / password / database: `ccp` / `ccp` / `ccp`)
2. **Migrations** (`migrate/migrate` against `./migrations`)
3. **Backend** HTTP API on **`http://localhost:8080`**
4. **Frontend** (Vite dev server) on **`http://localhost:5173`**

Default Compose settings use **`ITUNES_MOCK=true`** so sync works **without** calling Apple’s API (good for offline demos). To use the real iTunes Search API from the container, set `ITUNES_MOCK=false` in `docker-compose.yml` (or override via Compose env) and ensure the container has outbound network access.

The UI talks to the API from your browser. Because the page is served from port **5173** and the API from **8080**, the Go server enables **CORS** only when **`ENV=development`** (the Compose default for `backend`) so the browser allows those requests. Do not rely on that for production deployments.

Stop the stack:

```bash
docker compose down
```

To reset the database volume (wipes local data):

```bash
docker compose down -v
```

---

## Configuration (local / non-Compose)

Copy the example env file and adjust:

```bash
cp .env.example .env
```

Important variables (see `.env.example` for the full list):

| Variable | Purpose |
|----------|---------|
| `HTTP_ADDR` | Listen address (e.g. `:8080`) |
| `DATABASE_URL` | Postgres connection string |
| `ITUNES_BASE_URL` | iTunes Search base URL |
| `ITUNES_MOCK` | `true` = deterministic mock results; `false` = live HTTP to Apple |
| `CACHE_TTL_SECONDS` | In-memory cache TTL for list/detail |
| `ENV` | `development` allows a default `DATABASE_URL` when unset; use non-development when you want to require explicit config |

For the **operator UI** only (local `npm run dev`), copy **`frontend/.env.example`** to **`frontend/.env`** and set **`VITE_API_URL`** to wherever the Go API listens (default in the example is `http://localhost:8080`). Compose sets this for the `frontend` service so you usually do not need a file when using Docker.

---

## Operator UI (Vite)

With Compose running, open **`http://localhost:5173`**. You get sync, catalog filters, row detail + pin, and an audit tab—all backed by the same JSON routes documented below. **`VITE_API_URL`** on the `frontend` service is set to **`http://localhost:8080`** so the browser (on your machine) reaches the published backend port.

**Local UI without Docker** (API still running, e.g. `go run` or Compose backend only):

```bash
cd frontend
cp .env.example .env   # optional; defaults to http://localhost:8080
npm ci
npm run dev
```

Use the URL Vite prints (usually **`http://localhost:5173`**).

---

## Quick API checks

With the backend listening on port **8080**:

```bash
# Liveness
curl -s http://localhost:8080/health

# Sync podcasts (query is required). With mock iTunes, you still get persisted rows.
curl -s -X POST 'http://localhost:8080/sync/podcasts?query=news'

# List catalog
curl -s http://localhost:8080/podcasts

# Audit trail (optional limit)
curl -s 'http://localhost:8080/audit-logs?limit=50'
```

**Pin** example (replace `PODCAST_UUID` with a real `id` from `GET /podcasts`):

```bash
curl -s -X POST "http://localhost:8080/podcasts/PODCAST_UUID/pin" \
  -H 'Content-Type: application/json' \
  -d '{"pinned":true}'
```

---

## Database GUI (e.g. Beekeeper Studio)

Connect to the Compose Postgres from the host:

| Field | Value |
|-------|--------|
| Host | `localhost` |
| Port | `5432` |
| User | `ccp` |
| Password | `ccp` |
| Database | `ccp` |
| SSL | Off (local dev) |

Application tables live in the **`public`** schema (`podcasts`, `sync_runs`, `audit_logs`, plus `schema_migrations` for migration state).

---

## Tests

From the **repository root**, with **Go** installed (version in **`go.mod`**):

```bash
go test ./internal/... ./cmd/... ./tests/...
```

**Result:** every package with tests should show **`ok`**; packages without tests show **`? … [no test files]`**. If anything **`FAIL`**s, the command exits non-zero. For an explicit line after a successful run:

```bash
go test ./internal/... ./cmd/... ./tests/... && echo "Tests passed." || echo "Tests failed."
```

You do **not** need Docker or Postgres running for these tests; they use mocks and stubs. For a walkthrough-friendly list of scenario names, run **`go test ./tests/... -v`** (each test has a short note above it in **`tests/scenarios_test.go`**).
