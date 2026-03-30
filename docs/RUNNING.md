# How to run the project

This document is the **runbook**: how to start the app, hit the API, and (later) the UI and tests.

**Currently:** the **backend** (Go + Postgres + migrations) is in the repo. 

---

## Prerequisites

- **Docker** and **Docker Compose** (recommended path), or
- **Go 1.22+** and a local **PostgreSQL 16** instance if you run the API on the host instead of in a container.

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

Default Compose settings use **`ITUNES_MOCK=true`** so sync works **without** calling Apple’s API (good for offline demos). To use the real iTunes Search API from the container, set `ITUNES_MOCK=false` in `docker-compose.yml` (or override via Compose env) and ensure the container has outbound network access.

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

## Database GUI (I am using Beekeeper Studio)

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
