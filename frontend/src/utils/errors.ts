import type { Locale } from '../i18n/config'
import { zhCN } from '../i18n/locales/zh-CN'
import { jaJP } from '../i18n/locales/ja-JP'
import { enUS } from '../i18n/locales/en-US'

/**
 * Stable internal error states that map to localized user-facing messages.
 * These are NOT HTTP status codes — they are application-level error categories.
 */
export type ApiErrorState = 'NETWORK_ERROR' | 'SERVER_ERROR' | 'TIMEOUT' | 'UNKNOWN'

const localeMessages: Record<Locale, Record<ApiErrorState, string>> = {
  'zh-CN': zhCN.errors,
  'ja-JP': jaJP.errors,
  'en-US': enUS.errors,
}

/**
 * Map a stable internal error state to a localized user-facing message.
 * Never exposes raw exception text, stack traces, or internal details.
 */
export function mapApiError(state: ApiErrorState, locale: Locale): string {
  const messages = localeMessages[locale]
  return messages[state] ?? messages.UNKNOWN
}

/**
 * Classify a caught error into a stable ApiErrorState.
 * This is the single point where raw exceptions are converted to stable states.
 */
export function classifyError(error: unknown): ApiErrorState {
  if (error instanceof TypeError && error.message.includes('fetch')) {
    return 'NETWORK_ERROR'
  }
  if (error instanceof Error) {
    if (error.name === 'TimeoutError' || error.message.toLowerCase().includes('timeout')) {
      return 'TIMEOUT'
    }
    if (error.message.includes('500') || error.message.includes('502') || error.message.includes('503')) {
      return 'SERVER_ERROR'
    }
  }
  return 'UNKNOWN'
}
