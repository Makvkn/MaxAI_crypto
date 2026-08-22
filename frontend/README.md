# MaxAI Crypto — Frontend

AI financial intelligence for crypto portfolios. This app is a **client of the
MaxAI domain API**, not a second implementation of it: the Go backend owns every
financial calculation, and the frontend displays, formats and visualises what it
returns.

> Dashboard = Facts. AI = Intelligence.

## Stack

React 19 · TypeScript · Vite · Tailwind CSS 4 · Zustand · TanStack Query ·
React Router · Recharts · Vitest.

## Getting started

```bash
npm install
npm run dev          # http://localhost:5173, mock backend by default
```

| Script              | Purpose                                    |
| ------------------- | ------------------------------------------ |
| `npm run dev`       | Dev server                                 |
| `npm run build`     | Typecheck + production build               |
| `npm run typecheck` | `tsc -b --noEmit`                          |
| `npm run lint`      | Oxlint                                     |
| `npm test`          | Vitest (unit, component, journey)          |
| `npm run test:watch`| Vitest in watch mode                       |

## Environment

Copy `.env.example` to `.env`. No provider keys (OpenAI, Zerion, Tatum,
CoinGecko) ever belong in this app — they stay behind the backend.

| Variable                    | Meaning                                        |
| --------------------------- | ---------------------------------------------- |
| `VITE_API_BASE_URL`         | Backend origin; requests go to `/api/v1`       |
| `VITE_API_MODE`             | `mock` (default) or `real`                     |
| `VITE_API_TIMEOUT_MS`       | Request timeout                                |
| `VITE_ANALYTICS_ENABLED`    | Enables the analytics abstraction              |
| `VITE_ANALYTICS_DEBUG`      | Logs analytics events to the console           |
| `VITE_DEV_API_PROXY_TARGET` | Dev proxy target when `VITE_API_MODE=real`     |

Switching `VITE_API_MODE` is the only change needed to move between the mock
adapter and the real backend: both implement the same `MaxAIApi` contract.

## Architecture

```
src/
  api/          contract, HTTP client, errors, SSE, endpoints, mock adapter
  app/          config, providers, router, analytics bootstrap
  features/     auth, onboarding, wallets, portfolio, performance, assets,
                transactions, ai, scenarios
  components/   ui, layout, feedback, finance, data-quality
  stores/       Zustand: UI, preferences, onboarding draft
  lib/          formatting, dates, validation, copy, errors, query, analytics
  pages/        landing, analyze, sign-in, wallet dashboard
```

The data path is always:

```
React → feature hook (TanStack Query) → api/ → REST /api/v1 → Go backend
```

Rules the code holds itself to:

- **No financial arithmetic.** Values arrive as fixed-point decimal strings and
  are only formatted (`lib/formatting`). Unknown values render `—`, never `$0`.
- **No `fetch` in components.** Everything goes through `api/client.ts`, which
  handles auth headers, single-flight token refresh, timeouts, cancellation and
  error normalisation.
- **Server state in TanStack Query, client state in Zustand.** Query keys are
  centralised in `lib/query/queryKeys.ts`; lists use cursor pagination only.
- **Domain-level errors.** `ApiError` codes map to user copy in
  `lib/errors/messages.ts`; provider names and stack traces never reach the UI.
- **Data quality is first-class.** `ValuationStatus`, `DataQuality` and
  `DataFreshness` drive banners, badges and footnotes rather than being ignored.
- **Async sync is respected.** Creating a wallet enqueues a backend job; the UI
  polls wallet state and renders only the stages the backend reports. No timers
  invent progress.
- **AI is structured.** SSE parsing lives in `features/ai/streaming` (client,
  types, pure reducer). Responses carry claims, references, data quality and
  intent — including `UNSUPPORTED`.
- **Read-only.** No keys, seed phrases, signing, swapping or trading anywhere.

## API contract

`src/api/types/` is a **provisional** hand-written contract that mirrors the
documented backend DTOs, and `src/api/contract.ts` is the full API surface.
When the backend publishes its OpenAPI document, generated types replace
`api/types/` and the endpoint modules keep their signatures; components import
domain types only, so the change stays inside `src/api/`.

## Mock backend

`src/api/mock/` implements the same contract in memory (persisted to
`localStorage`), including asynchronous sync stages, cursor pagination, AI
streaming, usage limits and error states. The wallet address selects the
scenario, so every state is reachable without touching code:

| Address contains | Simulated condition                       |
| ---------------- | ----------------------------------------- |
| `partial`        | A visible asset has no price (PARTIAL)    |
| `stale`          | Data is stale                             |
| `verystale`      | Data is very stale                        |
| `unavailable`    | Portfolio valuation unavailable           |
| `fail`           | Initial sync fails                        |
| `syncpartial`    | Initial sync completes as PARTIAL         |
| `empty`          | Wallet holds nothing                      |
| `nohistory`      | No snapshots, so performance is UNAVAILABLE |
| `slow`           | Long-running initial sync                 |

Anything else behaves as a healthy wallet.

## Tests

```bash
npm test
```

Coverage spans the mock adapter as a contract (async sync, unknown prices,
cursor pagination, streaming, usage limits), the AI stream reducer, the
formatting layer, data-quality rendering, onboarding validation, and one
end-to-end journey through landing → network → address → sync → portfolio →
Ask AI → transaction explanation → scenario simulation.
