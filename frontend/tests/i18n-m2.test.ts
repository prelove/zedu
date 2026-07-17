import { describe, it, expect } from 'vitest'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'
import { errorToI18nKey } from '../src/api/error-mapping'
import { ApiError, ApiErrorCode } from '../src/api/http'

function collectKeys(obj: Record<string, unknown>, prefix = ''): string[] {
  const keys: string[] = []
  for (const key of Object.keys(obj)) {
    const path = prefix ? `${prefix}.${key}` : key
    const value = obj[key]
    if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
      keys.push(...collectKeys(value as Record<string, unknown>, path))
    } else {
      keys.push(path)
    }
  }
  return keys.sort()
}

describe('i18n key parity for M2-KIMI-01', () => {
  it('zh-CN and ja-JP have identical key sets (including new auth/onboarding/apiErrors)', () => {
    const zhKeys = collectKeys(zhCN)
    const jaKeys = collectKeys(jaJP)
    expect(jaKeys).toEqual(zhKeys)
  })

  it('zh-CN and en-US have identical key sets (including new auth/onboarding/apiErrors)', () => {
    const zhKeys = collectKeys(zhCN)
    const enKeys = collectKeys(enUS)
    expect(enKeys).toEqual(zhKeys)
  })

  it('all new keys have non-empty values in all three locales', () => {
    function checkEmpty(obj: Record<string, unknown>, prefix = ''): void {
      for (const key of Object.keys(obj)) {
        const path = prefix ? `${prefix}.${key}` : key
        const value = obj[key]
        if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
          checkEmpty(value as Record<string, unknown>, path)
        } else {
          expect(value, `key "${path}" has empty string value`).not.toBe('')
        }
      }
    }
    for (const locale of [zhCN, jaJP, enUS]) {
      checkEmpty(locale)
    }
  })

  it('apiErrors group exists in all locales with all required keys', () => {
    const requiredApiErrorKeys = [
      'AUTH_REQUIRED',
      'LOGIN_FAILED',
      'ACCOUNT_LOCKED',
      'FORBIDDEN',
      'NOT_FOUND',
      'CONFLICT',
      'INVALID_STATE',
      'INVALID_TEMPLATE',
      'RESET_NOT_ALLOWED',
      'DATABASE_ERROR',
      'INTERNAL_ERROR',
    ]
    for (const locale of [zhCN, jaJP, enUS]) {
      for (const key of requiredApiErrorKeys) {
        expect(locale.apiErrors).toHaveProperty(key)
        expect((locale.apiErrors as Record<string, string>)[key].length).toBeGreaterThan(0)
      }
    }
  })

  it('auth group has login/session keys in all locales', () => {
    const requiredAuthKeys = [
      'loginTitle',
      'usernameLabel',
      'passwordLabel',
      'submit',
      'submitting',
      'logout',
      'logoutFailed',
      'sessionExpired',
    ]
    for (const locale of [zhCN, jaJP, enUS]) {
      for (const key of requiredAuthKeys) {
        expect(locale.auth).toHaveProperty(key)
      }
    }
  })

  it('onboarding group has template/initialize/reset keys in all locales', () => {
    const requiredOnboardingKeys = [
      'title',
      'templateLabel',
      'templateJapanese',
      'templateK12',
      'templateBlank',
      'initialize',
      'reset',
      'resetConfirmTitle',
      'resetConfirmMessage',
      'resultInitialized',
      'resultReused',
      'resultReset',
    ]
    for (const locale of [zhCN, jaJP, enUS]) {
      for (const key of requiredOnboardingKeys) {
        expect(locale.onboarding).toHaveProperty(key)
      }
    }
  })

  it('routeGuard group exists in all locales', () => {
    for (const locale of [zhCN, jaJP, enUS]) {
      expect(locale).toHaveProperty('routeGuard')
      expect(locale.routeGuard).toHaveProperty('loginRequired')
      expect(locale.routeGuard).toHaveProperty('ownerRequired')
    }
  })
})

describe('errorToI18nKey mapping', () => {
  it('maps all frozen error codes to non-empty i18n keys', () => {
    const codes = [
      { code: ApiErrorCode.AUTH_REQUIRED, key: 'AUTH_REQUIRED' },
      { code: ApiErrorCode.LOGIN_FAILED, key: 'LOGIN_FAILED' },
      { code: ApiErrorCode.ACCOUNT_LOCKED, key: 'ACCOUNT_LOCKED' },
      { code: ApiErrorCode.FORBIDDEN, key: 'FORBIDDEN' },
      { code: ApiErrorCode.NOT_FOUND, key: 'NOT_FOUND' },
      { code: ApiErrorCode.CONFLICT, key: 'CONFLICT' },
      { code: ApiErrorCode.INVALID_STATE, key: 'INVALID_STATE' },
      { code: ApiErrorCode.INTERNAL_ERROR, key: 'INTERNAL_ERROR' },
      { code: ApiErrorCode.DATABASE_ERROR, key: 'DATABASE_ERROR' },
    ]
    for (const { code, key } of codes) {
      const err = new ApiError(code, key, 'rid', 400)
      const i18nKey = errorToI18nKey(err)
      expect(i18nKey).not.toBeNull()
      // The mapped key must exist in all locales.
      for (const locale of [zhCN, jaJP, enUS]) {
        const parts = i18nKey!.split('.')
        let obj: Record<string, unknown> = locale as unknown as Record<string, unknown>
        for (const part of parts) {
          expect(obj).toHaveProperty(part)
          obj = obj[part] as Record<string, unknown>
        }
        expect(typeof obj).toBe('string')
        expect((obj as unknown as string).length).toBeGreaterThan(0)
      }
    }
  })

  it('maps INVALID_TEMPLATE and RESET_NOT_ALLOWED stable keys', () => {
    const templateErr = new ApiError(42201, 'INVALID_TEMPLATE', 'rid', 422)
    expect(errorToI18nKey(templateErr)).toBe('apiErrors.INVALID_TEMPLATE')

    const resetErr = new ApiError(42201, 'RESET_NOT_ALLOWED', 'rid', 422)
    expect(errorToI18nKey(resetErr)).toBe('apiErrors.RESET_NOT_ALLOWED')
  })
})
