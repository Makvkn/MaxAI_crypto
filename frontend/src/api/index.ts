import { env } from '@/app/config/env'
import {
  apiCreateConversation as apiCreateConversationHttp,
  apiCreateGuestSession as apiCreateGuestSessionHttp,
  apiCreateWallet as apiCreateWalletHttp,
  apiGetAiUsage as apiGetAiUsageHttp,
  apiGetConversationMessages as apiGetConversationMessagesHttp,
  apiGetConversations as apiGetConversationsHttp,
  apiGetCurrentUser as apiGetCurrentUserHttp,
  apiGetPerformance as apiGetPerformanceHttp,
  apiGetPortfolio as apiGetPortfolioHttp,
  apiGetTransaction as apiGetTransactionHttp,
  apiGetTransactions as apiGetTransactionsHttp,
  apiGetWallet as apiGetWalletHttp,
  apiGetWallets as apiGetWalletsHttp,
  apiInitializeSession as apiInitializeSessionHttp,
  apiLoginWithEmail as apiLoginWithEmailHttp,
  apiLoginWithGoogle as apiLoginWithGoogleHttp,
  apiLogout as apiLogoutHttp,
  apiRegisterWithEmail as apiRegisterWithEmailHttp,
  apiSimulateScenario as apiSimulateScenarioHttp,
  apiStreamConversationMessage as apiStreamConversationMessageHttp,
  apiUpgradeAccount as apiUpgradeAccountHttp,
} from './endpoints'
import { installRefreshHandler } from './endpoints/auth'
import { createMockApi } from './mock'
import { tokenStore } from './tokenStore'

installRefreshHandler()

let mockApi = env.apiMode === 'mock' ? createMockApi() : null

function getMockApi() {
  if (!mockApi) mockApi = createMockApi()
  return mockApi
}

export function resetMockApi(): void {
  mockApi = createMockApi()
}

export const apiCreateGuestSession = (...args: Parameters<typeof apiCreateGuestSessionHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().auth.createGuestSession(...args)
    : apiCreateGuestSessionHttp(...args)

export const apiRegisterWithEmail = (...args: Parameters<typeof apiRegisterWithEmailHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().auth.registerWithEmail(...args)
    : apiRegisterWithEmailHttp(...args)

export const apiLoginWithEmail = (...args: Parameters<typeof apiLoginWithEmailHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().auth.loginWithEmail(...args)
    : apiLoginWithEmailHttp(...args)

export const apiLoginWithGoogle = (...args: Parameters<typeof apiLoginWithGoogleHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().auth.loginWithGoogle(...args)
    : apiLoginWithGoogleHttp(...args)

export const apiUpgradeAccount = (...args: Parameters<typeof apiUpgradeAccountHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().auth.upgradeAccount(...args)
    : apiUpgradeAccountHttp(...args)

export const apiGetCurrentUser = (...args: Parameters<typeof apiGetCurrentUserHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().auth.getCurrentUser(...args)
    : apiGetCurrentUserHttp(...args)

export const apiInitializeSession = (...args: Parameters<typeof apiInitializeSessionHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().auth.initializeSession(...args)
    : apiInitializeSessionHttp(...args)

export const apiLogout = (...args: Parameters<typeof apiLogoutHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().auth.logout(...args)
    : apiLogoutHttp(...args)

export const apiGetWallets = (...args: Parameters<typeof apiGetWalletsHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().wallets.list(...args)
    : apiGetWalletsHttp(...args)

export const apiGetWallet = (...args: Parameters<typeof apiGetWalletHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().wallets.get(args[0].walletId, args[1])
    : apiGetWalletHttp(...args)

export const apiCreateWallet = (...args: Parameters<typeof apiCreateWalletHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().wallets.create(...args)
    : apiCreateWalletHttp(...args)

export const apiGetPortfolio = (...args: Parameters<typeof apiGetPortfolioHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().portfolio.get(args[0].walletId, args[1])
    : apiGetPortfolioHttp(...args)

export const apiGetPerformance = (...args: Parameters<typeof apiGetPerformanceHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().performance.get(args[0].walletId, args[1], args[2])
    : apiGetPerformanceHttp(...args)

export const apiGetTransactions = (...args: Parameters<typeof apiGetTransactionsHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().transactions.list(args[0].walletId, args[1], args[2])
    : apiGetTransactionsHttp(...args)

export const apiGetTransaction = (...args: Parameters<typeof apiGetTransactionHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().transactions.get(args[0].walletId, args[0].transactionId, args[1])
    : apiGetTransactionHttp(...args)

export const apiGetConversations = (...args: Parameters<typeof apiGetConversationsHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().conversations.list(...args)
    : apiGetConversationsHttp(...args)

export const apiCreateConversation = (...args: Parameters<typeof apiCreateConversationHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().conversations.create(...args)
    : apiCreateConversationHttp(...args)

export const apiGetConversationMessages = (
  ...args: Parameters<typeof apiGetConversationMessagesHttp>
) =>
  env.apiMode === 'mock'
    ? getMockApi().conversations.listMessages(args[0].conversationId, args[1], args[2])
    : apiGetConversationMessagesHttp(...args)

export const apiStreamConversationMessage = (
  ...args: Parameters<typeof apiStreamConversationMessageHttp>
) =>
  env.apiMode === 'mock'
    ? getMockApi().conversations.streamMessage(args[0].conversationId, args[1], args[2])
    : apiStreamConversationMessageHttp(...args)

export const apiGetAiUsage = (...args: Parameters<typeof apiGetAiUsageHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().ai.getUsage(...args)
    : apiGetAiUsageHttp(...args)

export const apiSimulateScenario = (...args: Parameters<typeof apiSimulateScenarioHttp>) =>
  env.apiMode === 'mock'
    ? getMockApi().scenarios.simulate(...args)
    : apiSimulateScenarioHttp(...args)

export { tokenStore }
export { ApiError } from './errors'
export * from './types'
