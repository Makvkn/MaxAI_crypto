import type { RequestOptions } from './client'
import type {
  AIStreamEvent,
  AIUsage,
  AuthSession,
  Conversation,
  ConversationListParams,
  ConversationMessage,
  CreateConversationRequest,
  CreateWalletRequest,
  CursorPage,
  CursorParams,
  EmailCredentials,
  GoogleAuthRequest,
  PerformancePeriod,
  Portfolio,
  PortfolioPerformance,
  ScenarioRequest,
  ScenarioResult,
  SendMessageRequest,
  Transaction,
  TransactionListParams,
  UpgradeAccountRequest,
  User,
  Wallet,
} from './types'

/**
 * The MaxAI API surface, as consumed by the frontend.
 *
 * Both adapters implement this one interface:
 *   - `createHttpApi`  -> real REST calls against `/api/v1`
 *   - `createMockApi`  -> in-memory backend simulation for development
 *
 * Because the contract is identical, switching `VITE_API_MODE` changes no
 * feature code. When the OpenAPI client is generated, the generated functions
 * are wired into `createHttpApi` and this interface stays put.
 */
export interface MaxAIApi {
  auth: AuthApi
  wallets: WalletsApi
  portfolio: PortfolioApi
  performance: PerformanceApi
  transactions: TransactionsApi
  conversations: ConversationsApi
  ai: AiUsageApi
  scenarios: ScenariosApi
}

export interface AuthApi {
  /** Creates an anonymous account. Guests are real users with real data. */
  createGuestSession(options?: RequestOptions): Promise<AuthSession>
  registerWithEmail(
    credentials: EmailCredentials,
    options?: RequestOptions,
  ): Promise<AuthSession>
  loginWithEmail(
    credentials: EmailCredentials,
    options?: RequestOptions,
  ): Promise<AuthSession>
  loginWithGoogle(
    request: GoogleAuthRequest,
    options?: RequestOptions,
  ): Promise<AuthSession>
  /** Attaches credentials to the current anonymous account; `user.id` is kept. */
  upgradeAccount(
    request: UpgradeAccountRequest,
    options?: RequestOptions,
  ): Promise<AuthSession>
  /** Resolves the current user from the stored access token. */
  getCurrentUser(options?: RequestOptions): Promise<User>
  /**
   * Bootstraps auth on app start: refreshes an expired access token when
   * needed, then validates the session. Protected queries should wait for this.
   */
  initializeSession(options?: RequestOptions): Promise<User>
  logout(options?: RequestOptions): Promise<void>
}

export interface WalletsApi {
  list(params?: CursorParams, options?: RequestOptions): Promise<CursorPage<Wallet>>
  get(walletId: string, options?: RequestOptions): Promise<Wallet>
  create(
    request: CreateWalletRequest,
    options?: RequestOptions,
  ): Promise<Wallet>
}

export interface PortfolioApi {
  get(walletId: string, options?: RequestOptions): Promise<Portfolio>
}

export interface PerformanceApi {
  get(
    walletId: string,
    period: PerformancePeriod,
    options?: RequestOptions,
  ): Promise<PortfolioPerformance>
}

export interface TransactionsApi {
  list(
    walletId: string,
    params?: TransactionListParams,
    options?: RequestOptions,
  ): Promise<CursorPage<Transaction>>
  get(
    walletId: string,
    transactionId: string,
    options?: RequestOptions,
  ): Promise<Transaction>
}

export interface ConversationsApi {
  list(
    params?: ConversationListParams,
    options?: RequestOptions,
  ): Promise<CursorPage<Conversation>>
  create(
    request: CreateConversationRequest,
    options?: RequestOptions,
  ): Promise<Conversation>
  listMessages(
    conversationId: string,
    params?: CursorParams,
    options?: RequestOptions,
  ): Promise<CursorPage<ConversationMessage>>
  /**
   * Sends a message and streams the reply.
   *
   * Returns an async iterable of domain events rather than a string, so tool
   * activity and the final structured response are both observable. Aborting
   * `options.signal` cancels the stream.
   */
  streamMessage(
    conversationId: string,
    request: SendMessageRequest,
    options?: { signal?: AbortSignal },
  ): AsyncIterable<AIStreamEvent>
}

export interface AiUsageApi {
  /** Daily AI operation budget. Enforcement is server-side. */
  getUsage(options?: RequestOptions): Promise<AIUsage>
}

export interface ScenariosApi {
  simulate(
    request: ScenarioRequest,
    options?: RequestOptions,
  ): Promise<ScenarioResult>
}
