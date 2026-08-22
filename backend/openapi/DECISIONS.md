# Contract decisions

`MaxAI_Crypto_Project_Document.md` and `MaxAI_Crypto_Backend_Spec.md` are the
architecture. Where the specification prescribes a rule, the rule wins. Where it
offers an *illustrative* example ("suggested", "recommended", "provisional") and
the already-implemented frontend sends a concrete wire value, the contract
follows the frontend, so `VITE_API_MODE` can switch from `mock` to `real`
without touching UI code. Every such case is recorded below.

## 1. Chain identifiers: `bnb`, `xrpl`

§21 illustrates chain IDs as `bnb_chain` and `xrp_ledger`. The frontend already
sends `bnb` and `xrpl` (`frontend/src/api/types/chain.ts`), and the rule §21
actually states is that identifiers must be stable machine values rather than
display names — which both spellings satisfy.

**Decision:** the wire values are `ethereum`, `bitcoin`, `bnb`, `solana`,
`litecoin`, `xrpl`, `tron`, `dogecoin`. They are seeded as reference data in
`migrations/000002_seed_chains.up.sql`, so renaming a chain later is a migration
rather than a code change.

## 2. Performance period `all`

§53 discusses an all-time period; the frontend enum uses the literal `all`
alongside `24h`, `7d`, `30d`.

**Decision:** the query parameter accepts `24h`, `7d`, `30d`, `all`. Internally
the domain uses `performance.PeriodAllTime`, whose lookback is the first valid
snapshot rather than a fixed window.

## 3. Google authentication is ID-token based

§15 requires that the backend never trusts a client-supplied identity. The
frontend performs the browser OAuth flow and posts `{ id_token }` to
`POST /auth/google`.

**Decision:** the contract accepts a Google **ID token**, which the backend
verifies against Google before creating or linking an identity. No authorization
code exchange and no client-supplied user ID is part of the MVP contract.

## 4. Error codes

§106 lists domain error codes; the frontend enumerates a slightly narrower set
and keeps the union open so unknown codes degrade gracefully
(`frontend/src/api/types/errors.ts`).

**Decision:** codes shared by both use the frontend spelling. Codes present only
in the specification are additive; each one is registered in
`internal/domain/apperr/codes.go` with a category the frontend already handles,
so an unrecognised code still routes to the right retry or re-auth behaviour.

## 5. Routes the frontend does not call yet

The specification requires wallet deletion (§17) and manual refresh (§63). The
frontend has no client method for either.

**Decision:** `DELETE /wallets/{walletId}` and `POST /wallets/{walletId}/sync`
are part of the contract. They are additive, so the frontend remains compatible
and can adopt them without a contract change.

## 6. `GET /ai/conversations/{conversationId}`

The frontend lists conversations and messages but never fetches a single
conversation.

**Decision:** the endpoint exists for deep links into a conversation. Additive,
same reasoning as above.

## 7. Money is always a string

§97 and §112 forbid floating point for monetary values. JSON numbers are
IEEE-754 doubles in every mainstream parser.

**Decision:** every monetary and percentage field is the `Decimal` schema — a
string. `openapi/contract_test.go` enforces this, and
`internal/domain/shared/decimal.go` rejects JSON numbers on the way in.

## 8. Pagination

§109 mandates cursor pagination.

**Decision:** collections accept only `limit` and `cursor`. `contract_test.go`
fails the build if a `page`, `offset` or `page_size` query parameter ever
appears.

## 9. SSE is limited to AI responses

§63 forbids WebSockets and real-time blockchain monitoring; §80 defines the AI
stream.

**Decision:** exactly one endpoint returns `text/event-stream`:
`POST /ai/conversations/{conversationId}/messages`. Wallet sync progress is read
by polling `GET /wallets/{walletId}`, whose `sync` object reports only stages the
backend has actually completed (§19).
