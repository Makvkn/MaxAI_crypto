# MaxAI Crypto

> **AI Financial Intelligence for Crypto**

> **Understand your crypto, not just track it.**

---

# 1. Product Definition

**MaxAI Crypto** — AI-сервис для анализа криптовалютного портфеля.

Продукт не является ещё одним crypto wallet, exchange, trading terminal или generic chatbot.

Его задача — взять сложные данные:

- blockchain;
- balances;
- transactions;
- market prices;
- historical portfolio data;

и превратить их в понятные ответы:

> **What is happening?**

> **Why is it happening?**

> **What happens if...?**

Главная ценность продукта находится на уровне **financial intelligence**, а не простого отображения blockchain data.

### Product principle

> **Zerion показывает, что у тебя есть. MaxAI Crypto объясняет, что это значит.**

---

# 2. Positioning

### Product

**MaxAI Crypto**

### Positioning

**AI Financial Intelligence for Crypto**

### Main message

**Understand your crypto, not just track it.**

### Product category

AI-powered crypto portfolio intelligence.

### Product philosophy

```text
Blockchain Data
      ↓
Market Data
      ↓
Portfolio Data
      ↓
Analysis
      ↓
AI
      ↓
What is happening?
      ↓
Why?
      ↓
What happens if?
```

MaxAI Crypto находится выше существующей blockchain infrastructure.

Мы не строим собственный blockchain indexer и не конкурируем с blockchain data providers.

Providers являются инфраструктурным слоем и должны быть заменяемыми adapters.

---

# 3. Target User

Основной пользователь:

**Crypto holder / investor**

Пользователь уже владеет криптоактивами, но не хочет самостоятельно разбираться в большом количестве blockchain-, market- и portfolio-данных.

Типичный пользователь:

- имеет несколько криптоактивов;
- может использовать несколько wallets;
- понимает базовые принципы криптовалют;
- не является профессиональным трейдером;
- хочет понимать состояние своего портфеля;
- хочет получать объяснения на человеческом языке.

### Main user question

> **"What is happening with my crypto portfolio and why?"**

---

# 4. Main Problem

MaxAI Crypto решает одну основную проблему:

> **Превратить сложные данные криптокошелька и рынка в понятные человеку ответы и объяснения.**

Вместо самостоятельного анализа:

- десятков токенов;
- изменения цен;
- transactions;
- balances;
- historical data;
- portfolio allocation;

пользователь получает готовое объяснение.

### Example

Вместо:

```text
Portfolio:
$24,850

24h:
-4.2%

ETH:
52%

BTC:
21%

SOL:
9%
...
```

MaxAI Crypto может сказать:

> Your portfolio is down 4.2% today.

> ETH caused approximately 70% of the decline because it represents 52% of your portfolio.

### Core principle

```text
Dashboard = Facts

AI = Intelligence
```

Dashboard отвечает:

> **What do I have?**

AI отвечает:

> **What does it mean?**

---

# 5. MVP User Journey

MVP должен поддерживать complete user journey:

```text
Landing
   ↓
Select blockchain
   ↓
Enter wallet
   ↓
Initial synchronization
   ↓
Portfolio
   ↓
Historical performance
   ↓
AI Insight
   ↓
Ask AI
   ↓
Transaction explanation
   ↓
Scenario simulation
```

MVP считается готовым, когда этот сценарий работает надёжно и выдаёт достоверные данные.

---

# 6. Supported Blockchain Networks

MVP поддерживает:

1. Ethereum
2. Bitcoin
3. BNB Chain
4. Solana
5. Litecoin
6. XRP Ledger
7. TRON
8. Dogecoin

Важно:

> **TRON — blockchain. TRX — его native token.**

Архитектура должна позволять добавлять новые сети без изменения domain/business logic.

---

# 7. Wallet Model

### MVP

Пользователь анализирует один wallet.

```text
Select network
      ↓
Wallet address
      ↓
Analyze
      ↓
Portfolio
```

На уровне UI:

> **1 wallet → 1 analysis**

Но database и backend architecture сразу поддерживают:

```text
1 User
  ↓
N Wallets
```

Это позволяет в будущем добавить:

- несколько wallets;
- wallet groups;
- portfolio aggregation;
- connected wallets.

### Wallet lifecycle

Wallet имеет состояния:

```text
ACTIVE
SYNCING
ERROR
PAUSED
DELETED
```

`DELETED` реализуется как soft delete на domain уровне.

---

# 8. Wallet Authentication

MVP поддерживает:

- Guest mode;
- Google;
- Email.

Guest является обычным anonymous user account.

### Guest flow

```text
Anonymous User
      ↓
Temporary Account
      ↓
Wallet
      ↓
Portfolio Data
      ↓
Google / Email
      ↓
Account Upgrade
      ↓
Same user_id
      ↓
Existing data preserved
```

Wallet-based authentication не входит в MVP.

### Authentication architecture

Используется:

```text
Access Token
+
Refresh Token
```

Auth logic находится на backend.

---

# 9. Guest Abuse Protection

Anonymous users должны иметь отдельные ограничения.

MVP:

- IP rate limiting;
- anonymous user rate limiting;
- AI daily limit;
- wallet creation limit;
- provider request throttling.

CAPTCHA не является обязательной частью MVP и добавляется только при необходимости.

Frontend не является механизмом защиты.

---

# 10. Dashboard

Основной интерфейс состоит из двух логических областей:

```text
┌───────────────────────────────────────────┐
│                                           │
│              Portfolio                    │
│                                           │
│              $24,850                      │
│              ↓ 4.2%                       │
│                                           │
│              Chart                        │
│                                           │
│              Assets                       │
│                                           │
├───────────────────────────┬───────────────┤
│                           │               │
│                           │   AI Insight  │
│                           │               │
│                           │   Why is my   │
│                           │   portfolio   │
│                           │   down?       │
│                           │               │
│                           │   [Analyze]   │
│                           │               │
└───────────────────────────┴───────────────┘
```

Основные элементы:

- total portfolio value;
- performance;
- historical chart;
- asset allocation;
- assets;
- AI Insight;
- Ask AI;
- transaction explanations;
- scenario simulation.

AI-информация должна находиться рядом с основной portfolio information.

---

# 11. Initial Synchronization

Initial sync выполняется асинхронно.

```text
Wallet created
      ↓
InitialWalletSyncJob
      ↓
Fetch balances
      ↓
Fetch transactions
      ↓
Normalize
      ↓
Fetch prices
      ↓
Calculate valuation
      ↓
Create snapshot
      ↓
Wallet READY
```

HTTP request не ждёт завершения тяжёлой blockchain synchronization.

---

# 12. Synchronization State Machine

Wallet sync использует отдельный `sync_status`:

```text
PENDING
SYNCING
READY
PARTIAL
FAILED
```

Переходы:

```text
PENDING
   ↓
SYNCING
   ├──→ READY
   ├──→ PARTIAL
   └──→ FAILED
```

Freshness является отдельным состоянием:

```text
data_freshness:
FRESH
RECENT
STALE
VERY_STALE
```

`STALE` не является sync status.

---

# 13. Background Synchronization

MVP не использует real-time blockchain monitoring.

Начальный ориентир:

> **15-minute synchronization interval**

Интервал является configuration value.

В будущем он может зависеть от:

- user activity;
- provider limits;
- wallet activity;
- subscription;
- data freshness requirements.

При успешном sync создаётся новый portfolio snapshot.

При неудачном sync новый snapshot не создаётся.

---

# 14. Snapshot Creation

Snapshot создаётся после успешного sync:

```text
Successful Wallet Sync
        ↓
Recalculate Portfolio
        ↓
Create PortfolioSnapshot
```

Snapshot является историческим источником состояния портфеля.

Если synchronization завершилась ошибкой, старый snapshot не перезаписывается.

---

# 15. Historical Data

MVP поддерживает:

- 24h;
- 7d;
- 30d;
- All time.

Обязательно сохраняются:

> **Portfolio snapshots**

Snapshot позволяет восстановить состояние портфеля в определённый момент времени.

Snapshot используется для:

- performance;
- historical charts;
- AI analysis;
- comparisons;
- scenario context.

---

# 16. Portfolio Performance

MVP не строит полноценный accounting engine.

Не рассчитываем:

- realized PnL;
- unrealized PnL;
- cost basis;
- tax lots;
- сложный transaction accounting.

Используем:

> **Portfolio Performance based on historical portfolio snapshots.**

### Calculation

Для выбранного периода используется текущий валидный snapshot и исторический snapshot, максимально близкий к началу периода.

```text
Current Snapshot
      ↓
Closest valid snapshot to T-period
      ↓
Performance
```

Формула:

```text
performance_pct =
(current_value - historical_value)
/
historical_value
× 100
```

Если historical snapshot отсутствует:

```text
UNAVAILABLE
```

Если snapshot содержит неполные данные:

```text
PARTIAL
```

### Periods

```text
24h
7d
30d
All time
```

`All time`:

```text
First valid snapshot
      ↓
Current valid snapshot
```

UI использует термин:

> **Portfolio Performance**

а не полноценный `PnL`.

---

# 17. Snapshot and Calculation Versioning

Каждый snapshot сохраняет:

```text
calculation_version
```

AI-related deterministic calculations могут иметь:

```text
calculation_id
calculation_version
```

Это позволяет воспроизводить и отслеживать результаты после изменения calculation logic.

---

# 18. Blockchain Data Architecture

Backend не зависит от конкретного blockchain provider.

Используется abstraction layer:

```text
BlockchainDataProvider
├── ZerionProvider
└── TatumProvider
```

Business logic работает только с normalized domain data.

```text
Providers
    ↓
Normalization
    ↓
Domain Model
    ↓
Business Logic
```

Provider является adapter.

Business logic никогда не знает структуру конкретного provider API.

---

# 19. Provider Registry and Resolver

Provider selection выполняется через:

```text
ProviderRegistry
      ↓
ProviderResolver
```

Resolver учитывает:

- blockchain;
- provider capabilities;
- primary provider;
- fallback provider;
- supported operations.

Например:

```text
Ethereum → Zerion
Bitcoin  → Tatum
Solana   → Zerion
```

В будущем:

```text
Bitcoin
 ├── Tatum
 └── Provider C
```

Business logic не содержит provider-specific branching.

---

# 20. Blockchain Provider — Zerion

Zerion используется как основной provider для сетей, которые хорошо покрываются его API.

MVP использует Zerion прежде всего для:

- Ethereum;
- BNB Chain;
- Solana;
- TRON;
- других поддерживаемых сетей в будущем.

Получаемые данные могут включать:

- balances;
- portfolio;
- transactions;
- DeFi positions;
- metadata.

Domain layer не зависит от структуры ответа Zerion.

---

# 21. Blockchain Provider — Tatum

Tatum используется как второй blockchain data provider.

Основная задача:

- Bitcoin;
- Litecoin;
- Dogecoin;
- XRP Ledger.

Tatum является implementation choice, а не частью domain architecture.

В будущем Tatum может быть заменён другим provider без изменения:

- PortfolioService;
- TransactionService;
- SnapshotService;
- REST API;
- frontend.

---

# 22. Market Data Provider

Основной market-data provider:

> **CoinGecko**

Используется для:

- current price;
- 24h change;
- historical prices;
- market cap;
- volume;
- asset metadata.

Архитектура:

```text
MarketDataProvider
        ↓
CoinGeckoProvider
        ↓
PriceService
        ↓
Portfolio Services
```

---

# 23. PriceService

Business logic никогда не обращается напрямую к CoinGecko.

Правильно:

```text
PortfolioService
      ↓
PriceService
      ↓
MarketDataProvider
      ↓
CoinGeckoProvider
```

PriceService отвечает за:

- получение цен;
- caching;
- freshness;
- normalization;
- source tracking;
- fallback logic в будущем.

---

# 24. Asset Model

Asset является отдельной domain entity.

Нельзя идентифицировать asset только через symbol.

Например:

```text
ETH
USDC
USDT
```

недостаточно для уникальной идентификации.

Asset:

```text
Asset
-----
id
chain_id
contract_address
symbol
name
decimals
asset_type

market_data_provider
market_data_id
```

Для native asset:

```text
ETH
contract_address = NULL
```

Для token:

```text
USDC
contract_address = 0x...
```

---

# 25. Asset → Market Data Mapping

Каждый asset, для которого возможна рыночная цена, имеет mapping:

```text
Blockchain Asset
      ↓
Market Data Mapping
      ↓
CoinGecko Asset ID
      ↓
Price
```

Например:

```text
ETH / Ethereum
→ CoinGecko
→ ethereum
```

Нельзя искать цену только по symbol.

Если надёжный mapping отсутствует:

```text
market_data_id = NULL
```

и цена считается неизвестной.

---

# 26. Price Data Model

Каждая используемая цена имеет metadata:

```text
asset
price
currency
timestamp
source
status
freshness
```

Например:

```json
{
  "asset": "ETH",
  "price": 4210.52,
  "currency": "USD",
  "timestamp": "...",
  "source": "coingecko",
  "status": "fresh",
  "freshness": "FRESH"
}
```

Цена не является вечным значением.

Она всегда имеет:

- timestamp;
- source;
- freshness.

---

# 27. Data Freshness

Начальные thresholds:

```text
FRESH
< 5 min

RECENT
5–15 min

STALE
15–60 min

VERY_STALE
> 60 min
```

Thresholds являются configuration values.

AI получает информацию о freshness и обязан учитывать её при формировании ответа.

---

# 28. WalletPosition

Position отражает текущее состояние конкретного asset в wallet.

```text
WalletPosition
--------------
wallet_id
asset_id
balance_raw
balance_normalized
updated_at
```

Текущая valuation не является единственным источником истины.

Стоимость рассчитывается через:

```text
balance × price
```

---

# 29. Portfolio Snapshot Model

```text
PortfolioSnapshot
-----------------
id
wallet_id
captured_at
total_value_usd
status
calculation_version
```

Отдельно:

```text
PortfolioSnapshotPosition
-------------------------
snapshot_id
asset_id
balance
price_usd
value_usd
allocation_pct
```

Snapshot должен позволять восстановить portfolio state на момент времени.

---

# 30. Portfolio Valuation

Основная формула:

```text
Portfolio Value =
Σ(balance × normalized price)
```

Но только для assets с валидной ценой.

### Цена доступна

```text
balance exists
+
valid price
→
include in valuation
```

### Цена отсутствует

```text
balance exists
+
price unavailable
→
exclude from valuation
```

Баланс при этом всё равно отображается пользователю.

---

# 31. Valuation Status

Portfolio valuation имеет состояние:

```text
COMPLETE
PARTIAL
UNAVAILABLE
```

Например:

```text
ETH     $8,200       ✓
BTC     $5,100       ✓
SOL     $2,300       ✓
TOKEN   unavailable  ⚠
```

В таком случае система не показывает результат как полностью достоверный.

Например:

> Portfolio value is partially calculated because the price of 1 asset is currently unavailable.

---

# 32. PARTIAL / STALE Data Policy

### COMPLETE

Все необходимые данные доступны.

AI и portfolio calculations работают нормально.

### PARTIAL

Dashboard работает, но:

- portfolio value является partial;
- performance является partial;
- AI получает `data_quality = PARTIAL`;
- AI не должен представлять неполные результаты как точные.

### STALE

Dashboard показывает данные с warning.

AI может работать, но должен учитывать freshness.

Например:

> Based on portfolio data last updated 42 minutes ago...

### UNAVAILABLE

Если необходимые данные отсутствуют:

- финансовый анализ не выполняется;
- AI не должен придумывать отсутствующие значения.

---

# 33. Spam / Dust Tokens

Blockchain wallet может содержать:

- spam tokens;
- dust;
- неизвестные assets;
- scam tokens.

Они не должны засорять основной dashboard.

Основной view:

```text
Assets

ETH       $8,200
BTC       $5,100
SOL       $2,300
USDC      $1,200

Hidden assets (17)
```

Пользователь может раскрыть скрытые assets отдельно.

---

# 34. Token Visibility

Asset visibility определяется backend logic.

Состояния:

```text
VISIBLE
HIDDEN_DUST
HIDDEN_SPAM
UNKNOWN
```

LLM не является системой определения spam tokens.

На MVP используются deterministic rules.

В будущем может появиться:

```text
TokenRiskService
```

---

# 35. Tokens Without Price

Если для asset нет надёжной рыночной цены:

> **не придумываем стоимость.**

Показываем:

```text
UNKNOWN TOKEN

Balance:
500,000

Price:
—
```

Принцип:

> **Количество показываем. Стоимость не показываем.**

Такой asset не должен некорректно влиять на portfolio valuation.

---

# 36. DeFi and NFT Scope

Advanced DeFi analytics не входит в MVP.

Сложные DeFi positions, LP positions и аналогичные сущности не включаются в portfolio valuation, если для них нет надёжной canonical representation.

NFT intelligence также не входит в MVP.

NFT не учитывается в portfolio valuation.

При необходимости UI может показывать:

> NFTs are not included in portfolio valuation.

---

# 37. Transaction Domain Model

Canonical transaction model:

```text
Transaction
-----------
id
wallet_id
chain_id

tx_hash
block_number
timestamp

status
type

from_address
to_address

asset_in
amount_in

asset_out
amount_out

fee_asset
fee_amount

protocol
counterparty

raw_reference

created_at
updated_at
```

### Transaction types

```text
TRANSFER
SWAP
STAKE
UNSTAKE
CLAIM
APPROVE
CONTRACT_INTERACTION
UNKNOWN
```

Raw provider responses не являются domain source of truth.

Хранятся только необходимые normalized данные и provider reference.

---

# 38. Transaction Classification

Transaction type определяется backend, а не LLM.

Pipeline:

```text
Provider
   ↓
TransactionNormalizer
   ↓
TransactionClassifier
   ↓
Canonical Transaction
   ↓
AI Explainer
```

Если backend не может надёжно определить тип:

```text
UNKNOWN
```

LLM не имеет права превращать `UNKNOWN` в подтверждённый `SWAP`, `BUY` или другой тип без backend evidence.

---

# 39. Transaction Explainer

AI переводит technical blockchain transaction в человеческое объяснение.

Backend предоставляет transaction facts.

AI занимается объяснением.

Пример:

> You swapped approximately 2.4 ETH for 8,200 USDC through Uniswap. The network fee was approximately $4.20.

AI не должен самостоятельно вычислять transaction amounts или fee.

---

# 40. AI Architecture

AI не получает всю database.

Pipeline:

```text
User Question
      ↓
AI Orchestrator
      ↓
Determine Intent / Tools
      ↓
Fetch Required Data
      ↓
Domain Calculations
      ↓
Build Minimal Context
      ↓
LLM
      ↓
Structured Response
      ↓
Frontend
```

---

# 41. AI Provider

Primary LLM provider:

> **OpenAI**

Используется abstraction:

```text
LLMProvider
     │
     └── OpenAIProvider
```

В будущем:

```text
OpenAIProvider
AnthropicProvider
GoogleProvider
LocalProvider
```

Конкретная модель задаётся configuration:

```text
LLM_PROVIDER=openai
LLM_MODEL=...
```

Модель не hardcoded в domain logic.

---

# 42. AI Model Strategy

MVP может использовать одну primary model.

В будущем:

```text
Simple question
      ↓
Fast / cheaper model

Complex analysis
      ↓
Stronger reasoning model
```

Выбор модели находится в AI infrastructure/configuration layer.

---

# 43. AI MVP Capabilities

MVP содержит:

1. Portfolio Analysis
2. Ask AI
3. Transaction Explainer
4. Scenario Simulator

---

# 44. Portfolio Analysis

Автоматический анализ:

- текущей стоимости;
- изменения;
- основных причин изменения;
- активов, повлиявших на результат;
- allocation;
- concentration.

Пример:

> Your portfolio is down 4.2% today.

> ETH contributed approximately 70% of the decline.

Все factual claims должны быть основаны на backend calculations.

---

# 45. Ask AI

Пользователь может задавать вопросы о своём portfolio.

Примеры:

> Why did I lose $800 today?

> What is my largest position?

> Which assets contribute most to my portfolio risk?

> How did my portfolio change this week?

AI отвечает только на основании доступных portfolio, market и historical data.

---

# 46. Scenario Simulator

Пользователь задаёт сценарий:

> What if ETH falls 20%?

Backend выполняет точный calculation:

```text
Current portfolio:
$24,850

ETH allocation:
52%

ETH scenario:
-20%

ETH impact:
-$2,585

Estimated portfolio:
$22,265
```

После этого AI объясняет результат.

### Critical rule

> **LLM не рассчитывает финансовый результат.**

Calculation выполняется deterministic backend logic.

LLM только интерпретирует результат.

---

# 47. Scenario Service

Scenario calculations находятся в отдельном domain service:

```text
ScenarioService
```

Принцип:

```text
User Scenario
      ↓
ScenarioService
      ↓
Deterministic Calculation
      ↓
Structured Result
      ↓
LLM
      ↓
Explanation
```

Поддерживаемые сценарии MVP:

- asset price up/down by percentage;
- portfolio impact;
- resulting estimated portfolio value.

---

# 48. AI Intent Model

MVP использует ограниченный набор intents:

```text
PORTFOLIO_SUMMARY
PORTFOLIO_PERFORMANCE
PORTFOLIO_ALLOCATION
TRANSACTION_EXPLANATION
SCENARIO_SIMULATION
GENERAL_PORTFOLIO_QUESTION
UNSUPPORTED
```

Intent используется как routing mechanism и не должен превращаться в большое количество hardcoded `if/else`.

---

# 49. AI Tools

AI Orchestrator использует domain tools:

```text
get_portfolio()
get_positions()
get_portfolio_performance()
get_transaction()
get_historical_portfolio()
get_asset_price()
simulate_scenario()
```

LLM выбирает необходимый tool.

Tool выполняется backend.

После получения результата LLM формирует explanation.

---

# 50. AI Calculations

AI никогда не является source of truth для числовых расчётов.

Неправильно:

```text
LLM
 ↓
calculate portfolio loss
```

Правильно:

```text
User
 ↓
AI Orchestrator
 ↓
PortfolioService / ScenarioService
 ↓
Deterministic calculation
 ↓
Structured facts
 ↓
LLM
 ↓
Explanation
```

Это относится к:

- portfolio value;
- allocation;
- performance;
- scenario simulation;
- percentage changes;
- asset contribution;
- transaction amounts;
- transaction fees.

---

# 51. AI Context

LLM получает:

```text
User question
+
Minimal relevant portfolio data
+
Market data
+
Required historical data
```

Не передаём:

- entire database;
- raw provider responses;
- private application data;
- лишние transactions;
- database internals.

Принцип:

> **Minimum necessary context.**

---

# 52. Structured AI Context

Предпочтительно передавать LLM специализированные DTO.

Например:

```json
{
  "portfolio": {
    "value_usd": 24850,
    "change_24h_pct": -4.2
  },
  "drivers": [
    {
      "asset": "ETH",
      "allocation_pct": 52,
      "contribution_usd": -3500
    }
  ],
  "data_quality": "COMPLETE"
}
```

Не передаём raw database entities.

---

# 53. AI Claims and Evidence

AI может:

- анализировать;
- интерпретировать;
- сравнивать;
- делать выводы;
- формировать analytical opinion.

Но факты должны основываться на backend data.

Принцип:

> **AI can reason freely, but must not present unsupported assumptions as verified facts.**

Factual claims желательно связывать с внутренним evidence:

```text
Claim:
ETH caused 70% of today's decline

Evidence:
portfolio_performance
period = 24h
calculation_id = ...
```

Evidence не обязательно показывать пользователю в MVP.

---

# 54. AI Response Contract

AI response является structured object, а не просто строкой.

Пример:

```json
{
  "answer": "...",
  "data_quality": "COMPLETE",
  "claims": [
    {
      "text": "...",
      "evidence": [
        {
          "type": "calculation",
          "id": "..."
        }
      ]
    }
  ],
  "references": [
    {
      "type": "asset",
      "id": "..."
    }
  ]
}
```

Это позволяет frontend в будущем:

- ссылаться на assets;
- ссылаться на transactions;
- показывать evidence;
- строить richer AI UI.

---

# 55. AI Streaming

AI responses используют SSE.

```text
POST /api/v1/ai/conversations/:id/messages
                  ↓
              SSE stream
                  ↓
       ┌──────────┼──────────┐
       ↓          ↓          ↓
tool_started  text_delta  completed
       ↓
tool_completed
```

REST остаётся основным API.

SSE используется для streaming AI responses.

---

# 56. AI Safety and Financial Advice

MaxAI Crypto является analytical intelligence product, а не trading advisor.

AI может:

```text
ANALYZE
EXPLAIN
COMPARE
SIMULATE
```

AI не должен выдавать персонализированные торговые команды:

```text
BUY
SELL
SHORT
LONG
USE 20x LEVERAGE
```

Например:

> Should I sell ETH?

AI отвечает в аналитическом формате:

> I can't determine whether you should sell, but I can show how ETH concentration affects your portfolio and simulate several scenarios.

---

# 57. Unsupported Questions

MVP AI работает с:

- user's portfolio;
- blockchain transactions;
- portfolio history;
- market data required for portfolio analysis.

Поддерживается:

> Why is my portfolio down?

Поддерживается:

> What happens if ETH drops 20%?

Не входит в MVP:

> What happened in the global Ethereum ecosystem today?

Это требует News Intelligence.

AI должен корректно сообщить, что такой тип информации пока не поддерживается.

---

# 58. External Data as Untrusted Input

Blockchain data и external metadata являются untrusted input.

Token names, transaction metadata, NFT metadata и другие внешние строки не должны рассматриваться как instructions для LLM.

AI context builder обязан отделять:

```text
System / Developer instructions
        ↓
Trusted backend facts
        ↓
Untrusted external data
```

External blockchain data никогда не получает приоритет над AI system policy.

---

# 59. Conversation History

Conversation history сохраняется.

Модель:

```text
Conversation
    ↓
ConversationMessage
```

Пользователь может вести последовательный диалог:

```text
User:
Why did my portfolio fall?

AI:
ETH caused most of the decline.

User:
What if ETH falls another 20%?

AI:
...
```

Сложная long-term AI memory не входит в MVP.

---

# 60. Conversation Context

Не отправляем всю историю разговора бесконечно.

MVP:

```text
Recent conversation messages
+
Current question
+
Required portfolio context
```

В будущем можно добавить conversation summarization.

---

# 61. AI Usage Limits

Free:

> **10 AI operations/day**

AI operation включает:

- AI Insight;
- Ask AI;
- Transaction Explainer;
- Scenario Simulation.

Каждая operation расходует одну usage unit.

Лимит контролируется backend.

Redis:

```text
ai:usage:{user_id}:{date}
```

Frontend не является механизмом защиты.

Внутренний AI cost budget также используется для защиты от abuse и неожиданных расходов.

---

# 62. Privacy

Критические правила:

Мы не получаем и не храним:

- private keys;
- seed phrases;
- signing credentials;
- данные, позволяющие управлять wallet.

Работаем с:

> **Public blockchain addresses**

Также:

> **User data is never used for model training.**

AI получает только минимально необходимый контекст.

Wallet address не передаётся LLM без необходимости.

---

# 63. Security Boundary

MaxAI Crypto — read-only analytics application.

MVP не выполняет:

- transaction signing;
- swaps;
- trading;
- bridge operations;
- wallet control.

```text
Wallet
  ↓
Public address
  ↓
Read-only data
  ↓
MaxAI
```

MaxAI не получает возможность распоряжаться средствами пользователя.

---

# 64. Monetization

Основная модель:

> **Freemium → Pro subscription**

## Free

- 1 wallet;
- базовый dashboard;
- historical portfolio data;
- ограниченный AI;
- 10 AI operations/day.

## Pro

В дальнейшем:

- multiple wallets;
- advanced analytics;
- increased AI usage;
- extended historical analytics;
- advanced scenarios;
- additional AI features.

Ориентир:

> **$10–20/month**

Финальная цена определяется после тестирования продукта и willingness-to-pay.

---

# 65. Subscription Architecture

Даже если Pro реализуется позже, backend должен иметь отдельную abstraction:

```text
Subscription
Plan
Entitlement
Usage
```

Feature access не должен быть распределён по проекту через множество:

```text
if user.IsPro
```

Вместо этого используется entitlement/plan layer.

---

# 66. Real-Time

Real-time не входит в MVP.

Не используем:

- WebSockets;
- permanent blockchain monitoring;
- streaming blockchain events;
- real-time portfolio recalculation.

В MVP:

```text
Cached data
+
Background synchronization
+
Manual refresh
```

Real-time может появиться после подтверждения product-market fit.

---

# 67. Redis

Redis используется как infrastructure layer.

Используем Redis для:

```text
1. Cache
2. Rate limiting
3. Distributed locks
4. Background job coordination
```

Примеры:

```text
price:{asset}
portfolio:{wallet}

ai:usage:{user_id}:{date}

sync:wallet:{wallet_id}
```

Redis не является source of truth.

---

# 68. Redis TTL

Начальные значения:

```text
Price cache:
30–60 seconds

Portfolio cache:
1–5 minutes

AI rate limit:
until next UTC day

Wallet sync lock:
5–15 minutes

Provider backoff:
30 sec → 1 min → 5 min
```

Все значения configuration values.

---

# 69. PostgreSQL

PostgreSQL является:

> **Primary source of truth**

Основные таблицы:

```text
users

wallets
chains
assets
wallet_positions

transactions

prices
portfolio_snapshots
portfolio_snapshot_positions

conversations
conversation_messages

ai_usage

subscriptions
```

Возможны дополнительные технические таблицы для jobs и sync state.

---

# 70. Database Principle

Не строим database как копию provider API.

Неправильно:

```text
Zerion response
      ↓
Store everything
```

Правильно:

```text
Provider
   ↓
Normalizer
   ↓
Canonical MaxAI domain model
   ↓
PostgreSQL
```

MaxAI хранит только данные, необходимые продукту.

Мы не строим собственный blockchain indexer.

---

# 71. Financial Precision

Финансовые значения не хранятся и не рассчитываются как обычные floating-point values.

PostgreSQL использует:

```text
NUMERIC
```

Go использует decimal arithmetic.

Это относится к:

- token balances;
- USD prices;
- portfolio values;
- percentages;
- fees;
- scenario calculations.

---

# 72. Job Architecture

Тяжёлые операции не выполняются внутри HTTP request.

Background jobs:

```text
InitialWalletSyncJob
WalletSyncJob
PortfolioSnapshotJob
PriceRefreshJob
```

Пример:

```text
POST /wallets
      ↓
Create wallet
      ↓
Enqueue InitialWalletSyncJob
      ↓
Return
```

Worker выполняет synchronization отдельно.

---

# 73. Job Idempotency

Все background jobs должны быть безопасны для retry.

Например:

```text
WalletSyncJob
 ↓
provider request
 ↓
worker crashes
 ↓
retry
```

Retry не должен создавать:

- duplicate transactions;
- duplicate snapshots;
- duplicate positions;
- duplicate AI usage charges.

Для этого используются:

- unique constraints;
- idempotency keys;
- deterministic job identifiers;
- transactional writes.

---

# 74. Provider Failure Handling

Если provider недоступен:

```text
ETH      $8,200       ✓
SOL      $2,400       ✓
BTC      unavailable  ⚠
```

Система не скрывает проблему.

Data quality:

```text
COMPLETE
PARTIAL
STALE
UNAVAILABLE
```

Frontend показывает пользователю состояние данных.

Provider-specific errors не просачиваются напрямую во frontend.

Вместо:

```text
Tatum 429 Too Many Requests
```

API возвращает domain-level состояние:

```text
PORTFOLIO_DATA_TEMPORARILY_UNAVAILABLE
```

---

# 75. Provider Rate Limits

Provider integrations должны учитывать:

- provider-specific quotas;
- concurrency;
- batching;
- retryable errors;
- non-retryable errors;
- exponential backoff;
- circuit breaker в будущем;
- quota monitoring.

Начальный backoff:

```text
30 sec
→ 1 min
→ 5 min
```

Все параметры configuration values.

---

# 76. API Architecture

Frontend взаимодействует с backend через:

> **REST API**

Base path:

```text
/api/v1
```

Примеры:

```text
POST   /api/v1/auth/...
POST   /api/v1/wallets
GET    /api/v1/wallets
GET    /api/v1/wallets/:id
GET    /api/v1/wallets/:id/portfolio
GET    /api/v1/wallets/:id/performance
GET    /api/v1/wallets/:id/transactions

POST   /api/v1/ai/conversations
POST   /api/v1/ai/conversations/:id/messages

POST   /api/v1/ai/scenarios
```

AI streaming выполняется через SSE.

---

# 77. API Contract Strategy

**OpenAPI является source of truth для REST API contracts.**

Contracts определяют:

- request DTO;
- response DTO;
- enums;
- pagination;
- errors;
- validation;
- nullable fields;
- timestamps.

API versioning:

```text
/api/v1
```

---

# 78. API Error Contract

Все API errors имеют единый формат:

```json
{
  "error": {
    "code": "PORTFOLIO_DATA_UNAVAILABLE",
    "message": "...",
    "details": {}
  }
}
```

Категории:

```text
VALIDATION_ERROR
AUTHENTICATION_ERROR
NOT_FOUND
PROVIDER_ERROR
DATA_UNAVAILABLE
RATE_LIMIT
INTERNAL_ERROR
```

Provider-specific details не являются frontend contract.

---

# 79. Pagination

Используется **cursor-based pagination**.

Пример:

```text
GET /api/v1/wallets/:id/transactions
    ?limit=50
    &cursor=...
```

Cursor pagination используется прежде всего для:

- transactions;
- conversation messages;
- conversations;
- больших списков assets при необходимости.

---

# 80. Frontend Architecture

Frontend:

```text
React
TypeScript
Vite
Tailwind CSS
```

State architecture:

```text
Zustand
    ↓
UI / Client State

TanStack Query
    ↓
API / Server State
```

### Zustand

Используется для:

- UI state;
- local interaction state;
- modal state;
- filters;
- temporary form state;
- client-only preferences.

### TanStack Query

Используется для:

- portfolio data;
- wallet data;
- transactions;
- performance;
- conversations;
- API cache;
- loading/error states;
- server synchronization.

Frontend не содержит финансовую business logic.

---

# 81. Frontend Data Flow

Правильно:

```text
React
 ↓
TanStack Query
 ↓
REST API
 ↓
Go Backend
 ↓
Domain Services
```

Неправильно:

```text
React
 ↓
calculate portfolio valuation
```

Финансовые calculations выполняются backend.

---

# 82. Onboarding UX

MVP onboarding:

```text
Landing
   ↓
Select blockchain
   ↓
Enter wallet address
   ↓
Analyze
   ↓
Initial synchronization
   ↓
Portfolio
```

На MVP сеть выбирается пользователем явно.

Automatic chain detection не является обязательной частью MVP.

---

# 83. Initial Sync UX

Во время initial sync frontend показывает progress state.

Например:

```text
Analyzing wallet...

Fetching balances...
Fetching transactions...
Normalizing assets...
Fetching market prices...
Calculating portfolio...
Preparing analysis...
```

UI не должен показывать пользователю ложный прогресс.

Стадии отображаются только если backend действительно сообщает соответствующее состояние.

---

# 84. AI Insight UX

AI Insight не генерируется автоматически после каждого background sync.

MVP:

```text
Portfolio ready
      ↓
Dashboard
      ↓
[Analyze]
      ↓
AI request
      ↓
Insight
```

Это позволяет контролировать LLM costs.

---

# 85. Product Analytics

После появления MVP измеряются:

```text
Landing visit
     ↓
Wallet entered
     ↓
Analysis started
     ↓
Portfolio loaded
     ↓
AI insight viewed
     ↓
First AI question
     ↓
Second AI question
     ↓
Scenario used
     ↓
Return visit
```

Главные product metrics:

- wallet analysis completion;
- time to first useful insight;
- AI operation rate;
- repeat usage;
- retention;
- conversion to Pro.

Analytics implementation должна использовать единый event schema.

---

# 86. Privacy and Data Retention

Необходимо определить lifecycle пользовательских данных:

- wallet deletion;
- account deletion;
- conversation deletion;
- transaction retention;
- snapshot retention;
- data export;
- GDPR deletion requirements.

Основной принцип:

> Пользовательские данные не используются для model training.

Удаление аккаунта должно приводить к удалению/анонимизации соответствующих пользовательских данных согласно установленной retention policy.

---

# 87. Observability

С MVP логируются:

- provider latency;
- provider errors;
- synchronization failures;
- AI request latency;
- AI token usage;
- AI errors;
- portfolio calculation errors;
- job failures;
- API errors.

Каждая background operation имеет identifier:

```text
sync_job_id
wallet_id
provider
started_at
finished_at
status
error
```

Нужны:

- structured logs;
- metrics;
- error tracking;
- tracing для критических flows.

Конкретный observability stack является infrastructure decision.

---

# 88. Testing Strategy

Минимальная стратегия:

```text
Unit tests
Integration tests
Provider adapter tests
Portfolio calculation tests
Scenario calculation tests
AI tool tests
API tests
E2E tests
```

Особенно важны deterministic test cases для:

- valuation;
- performance;
- allocation;
- contribution;
- scenario calculations;
- transaction classification.

Финансовые calculations должны иметь стабильные regression/golden cases.

---

# 89. Deployment

Production architecture должна поддерживать:

```text
Docker
CI/CD
Staging
Production
PostgreSQL migrations
Secrets management
Backups
Monitoring
```

Environment-specific configuration не должна попадать в source code.

---

# 90. Infrastructure Economics

Необходимо отслеживать unit economics:

```text
Active User
    ↓
Blockchain provider usage
+
Market data usage
+
LLM usage
+
Database
+
Redis
+
Compute
    ↓
Monthly COGS
```

Особенно контролируются:

- provider API costs;
- LLM token costs;
- AI operations/user;
- synchronization frequency.

Это необходимо для проверки viability модели `$10–20/month`.

---

# 91. What Is NOT in MVP

Сознательно исключаем:

- News Intelligence;
- real-time updates;
- transaction signing;
- trading;
- swaps;
- bridge;
- copy trading;
- NFT intelligence;
- advanced DeFi analytics;
- полноценный accounting engine;
- realized PnL;
- unrealized PnL;
- cost basis;
- tax lots;
- сложную AI memory;
- advanced security engine;
- automatic chain detection;
- personalized trading recommendations.

Это необходимо для защиты MVP от scope creep.

---

# 92. Core Backend Architecture

```text
                         Go API
                           │
             ┌─────────────┼─────────────┐
             │             │             │
             ▼             ▼             ▼
        Portfolio      Transaction    AI Orchestrator
         Domain          Domain             │
             │             │                │
             └──────┬──────┘                │
                    │                       │
              Domain Services              │
                    │                       │
        ┌───────────┼───────────┐           │
        │           │           │           ▼
        ▼           ▼           ▼      LLMProvider
   Blockchain   Market Data  Scenario       │
   Provider     Provider     Service        ▼
        │           │                  OpenAIProvider
   ┌────┴────┐      │
   ▼         ▼      ▼
Zerion     Tatum  CoinGecko
        │           │
        └─────┬─────┘
              ▼
       Normalized Domain Data
              │
       ┌──────┴───────┐
       ▼              ▼
 PostgreSQL          Redis
       │
       ▼
 Historical Snapshots
```

---

# 93. Backend Architecture Principles

### Principle 1 — Domain-first

Business logic не зависит от providers.

### Principle 2 — Provider abstraction

Каждый внешний provider является adapter.

### Principle 3 — Provider Registry

Provider selection выполняется через capabilities и resolver.

### Principle 4 — Deterministic calculations

Financial calculations выполняются backend.

### Principle 5 — AI as intelligence layer

LLM объясняет и интерпретирует данные.

### Principle 6 — PostgreSQL as source of truth

Redis не заменяет database.

### Principle 7 — Async heavy operations

Blockchain synchronization выполняется background workers.

### Principle 8 — Minimal context

AI получает только необходимые данные.

### Principle 9 — Read-only MVP

Нет операций с деньгами пользователя.

### Principle 10 — Untrusted external data

Blockchain/provider metadata никогда не рассматривается как trusted AI instruction.

### Principle 11 — Idempotent jobs

Retry не должен создавать duplicate domain state.

### Principle 12 — Versioned calculations

Критические calculations должны иметь version.

---

# 94. First Technical Milestone

Первый milestone:

```text
POST /wallets
      ↓
Wallet created
      ↓
Initial sync
      ↓
Provider Resolver
      ↓
Zerion / Tatum
      ↓
Normalized assets
      ↓
CoinGecko prices
      ↓
Portfolio valuation
      ↓
PostgreSQL
      ↓
Portfolio Snapshot
      ↓
GET /portfolio
      ↓
Frontend dashboard
```

После этого существует working portfolio core.

---

# 95. Second Technical Milestone

```text
Portfolio
   ↓
Snapshots
   ↓
Historical performance
   ↓
24h / 7d / 30d / All time
```

После этого появляется historical intelligence.

---

# 96. Third Technical Milestone

```text
Portfolio
   ↓
AI Orchestrator
   ↓
Tools
   ↓
OpenAI
   ↓
AI Insight
```

После этого появляется основная differentiator продукта.

---

# 97. Fourth Technical Milestone

Добавить:

```text
Transaction Explainer
Scenario Simulator
Conversation History
AI Limits
SSE Streaming
```

После этого основной MVP user journey завершён.

---

# 98. Development Order

## STEP 1 — Domain Model

Спроектировать:

```text
User
Wallet
Chain
Asset
WalletPosition
Transaction
Price
PortfolioSnapshot
PortfolioSnapshotPosition
Conversation
ConversationMessage
Subscription
```

## STEP 2 — PostgreSQL Schema

Создать database schema, constraints и indexes.

## STEP 3 — Provider Interfaces

```text
BlockchainDataProvider
MarketDataProvider
LLMProvider
```

## STEP 4 — Provider Registry / Resolver

Определить provider capabilities и routing.

## STEP 5 — Zerion Adapter

## STEP 6 — Tatum Adapter

## STEP 7 — CoinGecko Adapter

## STEP 8 — Normalization Layer

## STEP 9 — Asset / Price Mapping

## STEP 10 — PortfolioService

## STEP 11 — SnapshotService

## STEP 12 — TransactionService

## STEP 13 — ScenarioService

## STEP 14 — Redis + Background Workers

## STEP 15 — Sync State Machine

## STEP 16 — AI Orchestrator

## STEP 17 — AI Tools + Structured Context

## STEP 18 — REST API / OpenAPI

## STEP 19 — React Dashboard

## STEP 20 — AI UI + SSE

## STEP 21 — Landing Page

## STEP 22 — Analytics / Monitoring

## STEP 23 — Testing / Hardening

## STEP 24 — Deployment

---

# 99. Long-Term Architecture

Архитектура должна позволять без фундаментального переписывания добавить:

```text
Multiple wallets
       ↓
Aggregated portfolio
       ↓
Advanced analytics
       ↓
News intelligence
       ↓
DeFi intelligence
       ↓
NFT intelligence
       ↓
Advanced accounting
       ↓
Real-time monitoring
```

При этом core domain model остаётся стабильной.

---

# 100. Future Provider Strategy

Providers являются заменяемыми adapters.

### Blockchain

```text
BlockchainDataProvider
├── Zerion
├── Tatum
├── Provider C
└── Provider D
```

### Market data

```text
MarketDataProvider
├── CoinGecko
└── Future fallback
```

### AI

```text
LLMProvider
├── OpenAI
├── Anthropic
├── Google
└── Local
```

---

# 101. What MaxAI Crypto Is NOT

MaxAI Crypto не является:

- crypto wallet;
- exchange;
- trading terminal;
- trading bot;
- copy-trading platform;
- transaction signing application;
- blockchain explorer;
- generic chatbot.

Это:

> **AI financial intelligence layer for crypto portfolios.**

---

# 102. Core Differentiation

Обычный portfolio tracker:

```text
What do I own?
```

MaxAI Crypto:

```text
What do I own?
        +
What is happening?
        +
Why is it happening?
        +
What happens if...?
```

---

# 103. Final Product Definition

> **MaxAI Crypto is an AI-powered financial intelligence platform for crypto holders that transforms blockchain, market, and portfolio data into understandable explanations, analysis, and scenarios.**

---

# 104. Final Technology Decisions

| Area | Decision |
|---|---|
| Product | MaxAI Crypto |
| Positioning | AI Financial Intelligence for Crypto |
| Frontend | React + TypeScript + Vite + Tailwind CSS |
| Client state | Zustand |
| Server state | TanStack Query |
| Backend | Go |
| API | REST `/api/v1` |
| AI streaming | SSE |
| API contracts | OpenAPI |
| Primary DB | PostgreSQL |
| Cache / infrastructure | Redis |
| Pagination | Cursor |
| Blockchain provider #1 | Zerion |
| Blockchain provider #2 | Tatum |
| Market data | CoinGecko |
| AI provider | OpenAI |
| AI architecture | AI Orchestrator + domain tools |
| Provider routing | Registry + Resolver |
| Background jobs | Redis-backed workers |
| Wallet model | 1 user → N wallets |
| MVP UI | 1 wallet |
| Historical data | Portfolio snapshots |
| Performance | Snapshot-based |
| Accounting engine | No |
| Real-time | No |
| AI limit | 10 operations/day Free |
| Authentication | Guest + Google + Email |
| Auth mechanism | Access + Refresh Tokens |
| Wallet access | Read-only |
| Private keys | Never collected |
| Seed phrases | Never collected |
| Freemium | Yes |
| Pro | Future |
| Landing page | Yes |
| News Intelligence | Post-MVP |
| Trading | Excluded |
| Swaps | Excluded |
| Bridge | Excluded |
| Copy trading | Excluded |
| Advanced DeFi | Post-MVP |
| NFT Intelligence | Post-MVP |
| Advanced accounting | Post-MVP |
| Automatic chain detection | No for MVP |

---

# 105. Final Principles

### Principle 1

> **Do not build another crypto tracker.**

### Principle 2

> **Providers are replaceable. Domain logic is not.**

### Principle 3

> **Backend calculates. AI explains.**

### Principle 4

> **Never invent missing financial data.**

### Principle 5

> **AI gets minimum necessary context.**

### Principle 6

> **PostgreSQL is the source of truth.**

### Principle 7

> **Redis is infrastructure, not the database.**

### Principle 8

> **MVP is read-only.**

### Principle 9

> **No feature enters MVP unless it improves the core user journey.**

### Principle 10

> **External blockchain data is untrusted input.**

### Principle 11

> **Background jobs must be idempotent.**

### Principle 12

> **Critical calculations are deterministic and versioned.**

### Principle 13

> **AI must never present unsupported assumptions as verified facts.**

### Principle 14

> **Do not use floating-point arithmetic for financial values.**

### Principle 15

> **The frontend presents data; it does not own financial business logic.**

### Principle 16

> **MaxAI Crypto should answer three questions:**

```text
What is happening?

Why is it happening?

What happens if...?
```

---

# 106. Final Product Vision

MaxAI Crypto starts with one simple interaction:

```text
Select blockchain
      ↓
Enter wallet address
      ↓
Understand your portfolio
```

Но long-term product vision значительно шире:

```text
Blockchain Data
       +
Market Data
       +
Portfolio History
       +
AI Reasoning
       ↓
Financial Intelligence
```

Продукт должен постепенно стать intelligent financial layer над всем crypto portfolio пользователя.

Не просто:

> **"Here are your assets."**

А:

> **"Here is what is happening with your money, why it is happening, and what different scenarios could mean for you."**

---

# MaxAI Crypto

> **AI Financial Intelligence for Crypto**

> **Understand your crypto, not just track it.**
