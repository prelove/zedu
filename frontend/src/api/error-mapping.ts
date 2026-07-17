import { ApiError, NetworkError } from './http'
import { ApiErrorCode } from './http'

/**
 * Map a caught error to a stable i18n key under `apiErrors.*`.
 * This is the single point where backend stable keys / error codes are
 * converted to i18n keys. The caller then uses `t(key)` to render the
 * localized message.
 *
 * Returns null if the error is not recognized (caller should fall back
 * to a generic message).
 */
export function errorToI18nKey(err: unknown): string | null {
  if (err instanceof NetworkError) {
    return 'apiErrors.AUTH_REQUIRED' // network failure during auth → treat as auth required
  }
  if (!(err instanceof ApiError)) {
    return 'apiErrors.INTERNAL_ERROR'
  }
  // Map by stable key first (the backend `message` field is a stable key).
  const keyMap: Record<string, string> = {
    AUTH_REQUIRED: 'apiErrors.AUTH_REQUIRED',
    LOGIN_FAILED: 'apiErrors.LOGIN_FAILED',
    ACCOUNT_LOCKED: 'apiErrors.ACCOUNT_LOCKED',
    FORBIDDEN: 'apiErrors.FORBIDDEN',
    NOT_FOUND: 'apiErrors.NOT_FOUND',
    CONFLICT: 'apiErrors.CONFLICT',
    INVALID_STATE: 'apiErrors.INVALID_STATE',
    INVALID_TEMPLATE: 'apiErrors.INVALID_TEMPLATE',
    RESET_NOT_ALLOWED: 'apiErrors.RESET_NOT_ALLOWED',
    DATABASE_ERROR: 'apiErrors.DATABASE_ERROR',
    INTERNAL_ERROR: 'apiErrors.INTERNAL_ERROR',
  }
  if (err.stableKey in keyMap) {
    return keyMap[err.stableKey]
  }
  // Fall back to code-based mapping.
  switch (err.code) {
    case ApiErrorCode.AUTH_REQUIRED:
      return 'apiErrors.AUTH_REQUIRED'
    case ApiErrorCode.LOGIN_FAILED:
      return 'apiErrors.LOGIN_FAILED'
    case ApiErrorCode.ACCOUNT_LOCKED:
      return 'apiErrors.ACCOUNT_LOCKED'
    case ApiErrorCode.FORBIDDEN:
      return 'apiErrors.FORBIDDEN'
    case ApiErrorCode.NOT_FOUND:
      return 'apiErrors.NOT_FOUND'
    case ApiErrorCode.CONFLICT:
      return 'apiErrors.CONFLICT'
    case ApiErrorCode.INVALID_STATE:
      return 'apiErrors.INVALID_STATE'
    case ApiErrorCode.INTERNAL_ERROR:
      return 'apiErrors.INTERNAL_ERROR'
    case ApiErrorCode.DATABASE_ERROR:
      return 'apiErrors.DATABASE_ERROR'
    default:
      return 'apiErrors.INTERNAL_ERROR'
  }
}
