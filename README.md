# MaxAI Crypto


<img width="1728" height="960" alt="image" src="https://github.com/user-attachments/assets/ab238bbd-f2d6-4659-89e2-a974250b5071" />

AI financial intelligence for crypto portfolios. Users connect a wallet address, the backend syncs on-chain data and market prices, and the app shows portfolio facts plus an AI assistant that explains holdings, transactions and scenarios.

> **Dashboard = Facts. AI = Intelligence.**

The frontend is a **client of the MaxAI domain API**, not a second implementation of it. All financial calculations live in the Go backend; the UI formats and visualises what the API returns.

## Repository structure

```
MaxAICrypto/
├── frontend/          # React SPA (Vite)
├── backend/           # Go REST API + background worker
│   ├── cmd/api/       # HTTP server
│   ├── cmd/worker/    # Asynq job worker (sync, snapshots, prices)
│   ├── internal/      # domain, application, infrastructure, transport
│   ├── migrations/    # PostgreSQL schema
│   ├── queries/       # sqlc SQL queries
│   ├── openapi/       # OpenAPI contract (source of truth for the API)
│   └── docker-compose.yml
└── README.md
```

## Tech stack

| Layer | Technologies |
| ----- | ------------ |
| **Frontend** | React 19, TypeScript, Vite, Tailwind CSS 4, Zustand, TanStack Query, React Router, Recharts, Vitest |
| **Backend** | Go 1.26, chi, pgx, sqlc, Redis, Asynq, JWT |
| **Data** | PostgreSQL 16, Redis 7 |
| **External providers** | Zerion, Tatum, CoinGecko, OpenAI (server-side only) |

## Core principles

- **Read-only.** No private keys, seed phrases, signing, swapping or trading.
- **Backend owns finance.** Monetary values are decimal strings; the frontend never does arithmetic.
- **Unknown ≠ zero.** `null` means unknown and must render as `—`, never `$0`.
- **Guests are real users.** Anonymous accounts are persisted; upgrading to email/Google keeps the same user ID and all data.
- **Async sync.** Wallet import enqueues a background job; the UI polls sync stages reported by the backend.
- **Contract-first API.** Routes are defined in `backend/openapi/openapi.yaml` before implementation. The frontend mock adapter implements the same contract as the real API.
- **Cursor pagination only.** No `page` / `offset` parameters.
- **SSE only for AI.** Wallet sync progress is polled; AI responses stream over Server-Sent Events.

## Supported chains (MVP)

`ethereum`, `bitcoin`, `bnb`, `solana`, `litecoin`, `xrpl`, `tron`, `dogecoin`

Seeded in `backend/migrations/000002_seed_chains.up.sql`.

## Prerequisites

- **Node.js** 20+ and npm
- **Go** 1.26+
- **Docker** and Docker Compose
- CLI tools for backend development:
  - [golang-migrate](https://github.com/golang-migrate/migrate) — database migrations
  - [sqlc](https://sqlc.dev/) — generate type-safe SQL access code

## Quick start (frontend only, mock backend)

The fastest way to explore the UI without running the Go backend:

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Open [http://localhost:5173](http://localhost:5173). By default `VITE_API_MODE=mock` — an in-memory adapter with `localStorage` persistence.

## Full local development

### 1. Start infrastructure

```bash
cd backend
docker compose up -d --wait
```

This starts:

| Service | Port | Credentials |
| ------- | ---- | ----------- |
| PostgreSQL | `5432` | `maxai` / `maxai`, database `maxai` |
| Redis | `6379` | no password |

### 2. Configure backend

```bash
cd backend
cp .env.example .env
```

Edit `.env` and set at minimum:

- `AUTH_JWT_SECRET` — at least 32 bytes (`openssl rand -base64 48`)
- Provider keys as needed: `ZERION_API_KEY`, `TATUM_API_KEY`, `COINGECKO_API_KEY`, `OPENAI_API_KEY`
- `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` for Google sign-in

### 3. Run migrations

```bash
cd backend
make migrate-up
```

### 4. Start backend processes

In separate terminals:

```bash
cd backend
make run-api      # HTTP API on :8080
make run-worker   # background jobs (wallet sync, snapshots, prices)
```

Health checks:

- `GET http://localhost:8080/health` — liveness
- `GET http://localhost:8080/ready` — readiness (Postgres + Redis)

### 5. Start frontend against real API

```bash
cd frontend
cp .env.example .env
```

Set in `.env`:

```env
VITE_API_MODE=real
VITE_DEV_API_PROXY_TARGET=http://localhost:8080
```

```bash
npm install
npm run dev
```

The Vite dev server proxies `/api/v1` to the backend.

## Environment variables

### Frontend (`frontend/.env`)

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `VITE_API_BASE_URL` | _(empty)_ | Backend origin; `/api/v1` is appended by the client |
| `VITE_API_MODE` | `mock` | `mock` or `real` |
| `VITE_API_TIMEOUT_MS` | `15000` | Request timeout |
| `VITE_DEV_API_PROXY_TARGET` | `http://localhost:8080` | Dev proxy target when `VITE_API_MODE=real` |
| `VITE_ANALYTICS_ENABLED` | `false` | Product analytics toggle |
| `VITE_ANALYTICS_DEBUG` | `false` | Log analytics events to console |

Provider credentials **never** belong in the frontend.

### Backend (`backend/.env`)

See `backend/.env.example` for the full list. Key groups:

- **Application** — `APP_ENV`, `HTTP_PORT`, `LOG_LEVEL`, `CORS_ALLOWED_ORIGINS`
- **Postgres** — `DATABASE_URL`
- **Redis** — `REDIS_URL`
- **Auth** — `AUTH_JWT_SECRET`, token TTLs, Google OAuth
- **Providers** — Zerion, Tatum, CoinGecko, OpenAI
- **Sync & cache** — sync interval, price/portfolio cache TTLs, freshness thresholds
- **Rate limits** — per-IP and per-user limits
- **Worker** — `WORKER_CONCURRENCY`

## API overview

Base path: `/api/v1`

| Group | Endpoints |
| ----- | --------- |
| **Auth** | `POST /auth/guest`, `/auth/email/register`, `/auth/email/login`, `/auth/google`, `/auth/refresh`, `/auth/upgrade`, `/auth/logout`, `GET /auth/session` |
| **Wallets** | `GET/POST /wallets`, `GET/DELETE /wallets/{walletId}`, `POST /wallets/{walletId}/sync` |
| **Portfolio** | `GET /wallets/{walletId}/portfolio` |
| **Performance** | `GET /wallets/{walletId}/performance?period=24h\|7d\|30d\|all` |
| **Transactions** | `GET /wallets/{walletId}/transactions`, `GET .../transactions/{transactionId}` |
| **AI** | `GET /ai/usage`, `POST /ai/scenarios`, `GET/POST /ai/conversations`, messages + SSE stream |

Full contract: [`backend/openapi/openapi.yaml`](backend/openapi/openapi.yaml)  
Contract decisions: [`backend/openapi/DECISIONS.md`](backend/openapi/DECISIONS.md)

## Authentication model

| Kind | Description |
| ---- | ----------- |
| `GUEST` | Anonymous account created via `POST /auth/guest`. No email or password. |
| `REGISTERED` | Account with email/password or Google identity. |

- Identities live in `auth_identities` (providers: `guest`, `google`, `email`).
- Passwords are stored as bcrypt hashes only — plaintext passwords cannot be recovered from the database.
- Refresh tokens are server-side sessions with rotation and reuse detection.
- Upgrading a guest (`POST /auth/upgrade`) keeps the same `users.id`.

### Inspect users in the database

```bash
docker exec maxai-crypto-postgres-1 psql -U maxai -d maxai \
  -c "SELECT u.id, u.kind, u.email, i.provider, i.subject FROM users u LEFT JOIN auth_identities i ON i.user_id = u.id;"
```

## Database schema

Main tables (see `backend/migrations/000001_init.up.sql`):

| Table | Purpose |
| ----- | ------- |
| `users`, `auth_identities`, `refresh_sessions` | Accounts and authentication |
| `chains`, `assets`, `prices` | Reference data and market prices |
| `wallets`, `wallet_sync_states`, `wallet_sync_runs`, `wallet_positions` | Wallet import and sync |
| `transactions` | Normalised on-chain activity |
| `portfolio_snapshots`, `portfolio_snapshot_positions` | Historical portfolio state |
| `conversations`, `conversation_messages`, `ai_usage` | AI chat and rate limits |
| `subscriptions`, `scenario_calculations` | Plans and what-if scenarios |

## Backend architecture

```
cmd/api, cmd/worker
    └── internal/app/bootstrap       # dependency wiring
        ├── application/             # use cases (auth, wallets, sync, portfolio, ai, …)
        ├── domain/                  # entities, repository interfaces, business rules
        ├── infrastructure/          # postgres, redis, jwt, external auth
        ├── providers/               # zerion, tatum, coingecko, openai
        ├── jobs/                    # asynq task definitions and handlers
        └── transport/http/          # chi router, handlers, middleware
```

API and worker are **separate processes** sharing the same config, Postgres and Redis. A slow wallet sync never blocks HTTP requests.

## Frontend architecture

```
frontend/src/
  api/          contract, HTTP client, errors, SSE, endpoints, mock adapter
  app/          config, providers, router, analytics
  features/     auth, onboarding, wallets, portfolio, performance, assets,
                transactions, ai, scenarios
  components/   ui, layout, feedback, finance, data-quality
  stores/       Zustand: UI, preferences, onboarding draft
  lib/          formatting, dates, validation, copy, errors, query
  pages/        landing, analyze, sign-in, wallet dashboard
```

Data flow:

```
React → feature hook (TanStack Query) → api/ → REST /api/v1 → Go backend
```

More detail: [`frontend/README.md`](frontend/README.md)

### Mock backend scenarios

When `VITE_API_MODE=mock`, wallet address substrings trigger test conditions:

| Address contains | Simulated condition |
| ---------------- | ------------------- |
| `partial` | Asset without price (PARTIAL valuation) |
| `stale` / `verystale` | Stale / very stale data |
| `unavailable` | Portfolio valuation unavailable |
| `fail` | Initial sync fails |
| `syncpartial` | Sync completes as PARTIAL |
| `empty` | Empty wallet |
| `nohistory` | No snapshots → performance UNAVAILABLE |
| `slow` | Long-running initial sync |

## Make targets (backend)

```bash
cd backend
make help            # list all targets
make up              # start Postgres + Redis
make down            # stop containers
make reset           # stop and delete volumes
make migrate-up      # apply migrations
make migrate-down    # roll back one migration
make sqlc            # regenerate sqlc code
make run-api         # start API
make run-worker      # start worker
make test            # unit tests
make test-integration # integration tests (needs live DB + Redis)
make openapi         # validate OpenAPI contract
make check           # fmt + vet + test
```

## npm scripts (frontend)

```bash
cd frontend
npm run dev          # dev server (http://localhost:5173)
npm run build        # typecheck + production build
npm run typecheck    # tsc -b --noEmit
npm run lint         # oxlint
npm test             # vitest
npm run test:watch   # vitest in watch mode
```

## Testing

**Frontend:**

```bash
cd frontend && npm test
```

Covers mock adapter contract, AI stream reducer, formatting, data-quality rendering, onboarding validation, and an end-to-end journey (landing → wallet → sync → portfolio → AI).

**Backend:**

```bash
cd backend && make test
cd backend && make test-integration   # requires docker compose up
cd backend && make openapi          # contract validation
```

## Troubleshooting

| Problem | Check |
| ------- | ----- |
| API won't start | `.env` exists, `AUTH_JWT_SECRET` is set, Postgres/Redis are healthy (`make up`) |
| Migrations fail | `DATABASE_URL` matches docker-compose credentials |
| Frontend 401/403 | `VITE_API_MODE=real`, backend running, guest session created |
| Wallet stuck syncing | Worker process running (`make run-worker`), provider API keys set |
| CORS errors | `CORS_ALLOWED_ORIGINS` includes `http://localhost:5173` |

Reset local database (destructive):

```bash
cd backend
make reset
make up
make migrate-up
```

## Security notes

- Never commit `.env` files — they are gitignored.
- Provider API keys (Zerion, Tatum, CoinGecko, OpenAI) and `AUTH_JWT_SECRET` stay server-side.
- The app is read-only: it analyses public wallet addresses only.

## License

Private project. All rights reserved unless stated otherwise.
