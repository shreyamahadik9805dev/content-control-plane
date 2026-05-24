# Content Control Plane (Podcasts)

**NOTE:** This README is a **living document**: it will pick up runbooks, API details, schema notes, and tradeoffs as the submission solidifies.

---

## What this is

This repo is a **content control plane** for **podcast shows** (show-level metadata in v1—not a full episode catalog). It pulls from Apple’s **iTunes Search** API (`media=podcast`), **normalizes** results into **PostgreSQL**, and exposes a **JSON HTTP API** (plus a **Vite + React** UI under **`frontend/`**) so you can **sync** on demand, **browse** the catalog, **pin** rows for curation, and read an **append-only audit** trail plus **sync run** history.

**Product AI (human-in-the-loop):** operators can request **structured metadata suggestions** (summary, operator tags, language, confidence) for a show. Suggestions land as **pending proposals** in Postgres; the catalog updates only after **Accept**—**Reject** leaves canonical rows unchanged. Mock mode works offline without API keys.

The same pattern applies anytime you’re integrating **someone else’s catalog**; podcasts are just a realistic, API-key-free example.

### What problem it solves

Teams often need to **surface third-party content** in a product, but the **vendor’s API is not your source of truth** for how you operate:

- **Shape and semantics differ** from what product, support, or compliance expect. You still need **your own schema**, stable internal keys, and room to attach **policy** (what’s allowed to appear, what’s promoted, what’s hidden).
- **Calling the vendor only at page-load time** means no durable answer to “what did we have yesterday?” if the API changes, throttles you, or returns bad data. A **stored, normalized copy** gives you something to diff, replay, and explain.
- **Curation and accountability** (“feature this show”, “who changed that flag?”) rarely belong in the vendor app. You want **your** workflows—**pin/unpin**, **audit events**, **sync runs** that record success/failure—on top of data **you** control.

This project is that **middle layer** in miniature: **ingest** from iTunes → **persist** under **`source_id` / your row shape** → **operator-facing** read and curation APIs. The same layout could later add scheduling, RBAC, or multi-tenant catalogs if scope grows.

---

## Run the project

**→ [docs/RUNNING.md](docs/RUNNING.md)** — includes Compose, ports, `curl` examples, Postgres GUI hints, and **`go test ./internal/... ./cmd/... ./tests/...`**. Open that file when you want to run or review the stack.

---

## Scope

### Current scope

- **Ingest** podcast *shows* (not full episode catalogs in v1) from a **public HTTP API**, map to an **internal schema** with a **stable external key** (`source_id`).
- **Persist** catalog + **sync run** history + **append-only audit** events in **PostgreSQL**.
- **Read path** with a **small in-process TTL cache** and explicit **cache invalidation** on writes that affect list/detail.
- **Curation:** at minimum **pin / unpin** with audit; **featured** reserved for follow-on UX if time allows.
- **Product AI:** **generate → review → accept/reject** proposals for operator **summary** and **operator_tags**; audit events `proposal.created`, `proposal.accepted`, `proposal.rejected`. Default **`AI_MOCK=true`** in Compose (no vendor key required).
- **Operator clarity:** how to run the stack and example API calls are in **[docs/RUNNING.md](docs/RUNNING.md)** (Docker-first). Deeper AI runbook: **[docs/PRODUCT_AI.md](docs/PRODUCT_AI.md)**.
- **Transparency:** tradeoffs stay in this doc; **how to run tests** is in **[docs/RUNNING.md](docs/RUNNING.md)**. See **AI usage** at the bottom for tooling disclosure.


---

## Why podcasts? 

- **Fits the challenge:** demonstrates a **real external API** over the network, **JSON transformation**, and **structured responses** without juggling API keys for v1.
- **iTunes Search (`media=podcast`)** returns **stable collection identifiers**, titles, publishers, **genres**, **feed URLs**, and **artwork**—enough to justify **normalization** and a **non-trivial internal row shape**.
- **Product-shaped story:** operators often need to **curate** what appears in a product surface; **pin + audit + sync runs** mirror how teams **govern** third-party catalogs.
- **Demo-friendly:** goal is for reviewers to **type a search query**, run sync, and **see rows land** in a DB-backed catalog—easy to narrate in a walkthrough.

---

## Problem?

Third-party catalog APIs are convenient but rarely match how **we** want to run the product internally. This repo is a **control plane** sketch: pull podcast show metadata, store a **normalized** copy under **your** keys, let operators **pin** rows, and keep a **light audit trail** and **sync run** history so we can explain what happened.

---

## Architecture

This matches what is in the repo today: **Vite + React** in **`frontend/`** and **`curl`** both hit the same Gin API.

### Request flow

```mermaid
flowchart TB
  subgraph clients [Clients]
    UI[React UI]
    CLI[curl]
  end

  subgraph go [Go service]
    H[HTTP handlers]
    S[Service layer]
    R[Repository]
    IT[iTunes client]
    AI[AI suggester]
    CA[TTL cache]
  end

  subgraph external [External]
    ITU[iTunes Search API]
    LLM[OpenAI optional]
  end

  subgraph storage [Storage]
    PG[(PostgreSQL)]
  end

  UI --> H
  CLI --> H
  H --> S
  S --> CA
  S --> R
  R --> PG
  S --> IT
  IT --> ITU
  S --> AI
  AI -.-> LLM
```

### Layering

```
HTTP  →  service  →  repository  →  PostgreSQL
           ↓              ↑
     iTunes client    proposals (pending → accept/reject)
           ↓
     AI suggester (mock or OpenAI; structured JSON only)
           ↓
     in-process TTL cache (read path)
```

- **Handlers:** transport only (status codes, binding, thin).
- **Service:** sync orchestration, **proposal** lifecycle, cache invalidation, audit/sync_run writes.
- **Repository:** interface + Postgres implementation (`pgx`) for test seams.
- **iTunes package:** timeouts, retries, optional **mock** for offline runs.
- **`internal/ai`:** `Suggester` interface; **mock** when `AI_MOCK=true`, else **OpenAI** when configured. The model never writes Postgres directly.

---

## Repository structure

Current layout:

```
content-control-plane/
├── cmd/server/              # main()
├── internal/
│   ├── config/              # env / .env loading
│   ├── domain/              # shared models (JSON tags)
│   ├── handler/             # Gin routes
│   ├── service/             # podcasts + proposals (AI apply path)
│   ├── repository/          # Store interface + postgres
│   ├── ai/                  # mock + OpenAI structured suggestions
│   ├── client/itunes/       # external API + mock
│   └── cache/               # TTL wrapper
├── migrations/              # SQL (golang-migrate compatible); includes 000003_product_ai
├── tests/                   # cross-package scenario tests (presenter-oriented names)
├── frontend/                # Vite + React + TS (detail panel: AI proposals)
├── docs/
│   ├── RUNNING.md           # runbook: Compose, UI, API checks, tests
│   └── PRODUCT_AI.md        # Product AI scope, env, demo steps
├── docker-compose.yml
├── Dockerfile
├── .env.example
└── README.md
```

**`tests/`** holds cross-package **flow** scenarios (mock iTunes, sync, HTTP codes) with names that read well under **`go test ./tests/... -v`**. **`internal/service/normalize_test.go`** covers **ingest mapping** (`normalizeShow` is unexported, so those tests stay in-package).

---

## Data model

Implemented in **`migrations/`** (tables below).

- **`podcasts`** — internal `id`, unique **`source_id`** (e.g. iTunes `collectionId`), title, publisher/author, categories (`jsonb`), feed + artwork URLs, optional episode count, **`summary`** and **`operator_tags`** (operator-facing enrichment, set on proposal accept), **`pinned` / `featured`**, timestamps. iTunes upserts **refresh vendor fields** without clobbering curation flags or accepted AI enrichment on conflict.
- **`ai_proposals`** — pending / accepted / rejected suggestions per podcast: structured **`payload`**, input **`context`**, model/provider/latency metadata, timestamps.
- **`sync_runs`** — one row per sync attempt: query string, status, counts, errors, start/end times.
- **`audit_logs`** — append-only events (e.g. pin/unpin, sync completed/failed, `proposal.created` / `accepted` / `rejected`) with small JSON metadata.

---

## HTTP API

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/health` | Liveness |
| `POST` | `/sync/podcasts?query=…` | Run ingest for search term |
| `GET` | `/podcasts` | List catalog |
| `GET` | `/podcasts/:id` | Detail by UUID |
| `POST` | `/podcasts/:id/pin` | Body `{"pinned": bool}`; writes audit |
| `GET` | `/audit-logs?limit=…` | Recent audit rows |
| `POST` | `/podcasts/:id/suggestions` | Create **pending** AI proposal (mock or OpenAI) |
| `GET` | `/podcasts/:id/suggestions` | List proposals; `?status=pending` optional |
| `POST` | `/suggestions/:id/accept` | Apply summary/tags to podcast; audit |
| `POST` | `/suggestions/:id/reject` | Reject proposal; optional `{"note":"…"}` |

**Examples:** **[docs/RUNNING.md](docs/RUNNING.md)** (`curl` snippets). **Product AI demo:** sync a query → select a show in the UI → **Generate suggestion** → Accept or Reject → check **Audit**.

### Product AI configuration

| Variable | Default (Compose) | Purpose |
|----------|-------------------|---------|
| `AI_MOCK` | `true` | Deterministic suggestions; no API key |
| `AI_PROVIDER` | `openai` | Live provider when mock is off |
| `OPENAI_API_KEY` | — | Required when `AI_MOCK=false` |
| `AI_MODEL` | `gpt-4o-mini` | Chat model for live calls |
| `AI_HTTP_TIMEOUT_SECONDS` | `60` | Provider HTTP timeout |

See **`.env.example`** and **[docs/PRODUCT_AI.md](docs/PRODUCT_AI.md)**.

---

## Tech stack

| Area | Choice | 
|------|--------|
| Runtime | Go 1.25+ (see `go.mod`) |
| HTTP | Gin | 
| DB | PostgreSQL | 
| Driver | pgx/v5 | 
| Migrations | golang-migrate (CLI) | 
| Cache | go-cache (memory) | 
| External API | iTunes Search | 
| Product AI | Mock (default) or OpenAI Chat Completions (structured JSON) |
| UI | Vite + React + TS |
| Run | Docker Compose | 

---

## Design tradeoffs (early)

- **In-memory cache:** easy locally; not shared across replicas (Redis deferred as a future scope).
- **On-demand sync:** no scheduler in v1; reduces moving parts as I prioritize the scope.
- **iTunes:** subject to network and vendor behavior; mock mode for CI/offline.
- **Human-in-the-loop AI:** proposals are stored separately from canonical catalog rows; **accept** is the only path that mutates `summary` / `operator_tags`, with full audit. No autonomous curation in v1.

---

## AI usage

**In the product:** optional **metadata suggestions** with operator approval (see **Product AI** above). The LLM returns structured JSON; the Go service validates and persists proposals—never silent writes to production catalog fields.

**In development:** AI-assisted tools (e.g. ChatGPT/Claude/Cursor) helped with **boilerplate**, **documentation drafting**, and **design iteration**. **Architecture decisions and review-ready quality** remain human-owned.

---

## Third-party attribution

**iTunes Search API** is owned by Apple; this repo is an independent exercise and not affiliated with Apple.
