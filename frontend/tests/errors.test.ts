import { describe, it, expect } from 'vitest'
import { mapApiError, type ApiErrorState } from '../src/utils/errors'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

describe('mapApiError', () => {
  it('maps network error to localized message', () => {
    const result = mapApiError('NETWORK_ERROR', 'zh-CN')
    expect(result).toBe(zhCN.errors.NETWORK_ERROR)
  })

  it('maps network error in ja-JP', () => {
    const result = mapApiError('NETWORK_ERROR', 'ja-JP')
    expect(result).toBe(jaJP.errors.NETWORK_ERROR)
  })

  it('maps network error in en-US', () => {
    const result = mapApiError('NETWORK_ERROR', 'en-US')
    expect(result).toBe(enUS.errors.NETWORK_ERROR)
  })

  it('maps server error to localized message', () => {
    const result = mapApiError('SERVER_ERROR', 'zh-CN')
    expect(result).toBe(zhCN.errors.SERVER_ERROR)
  })

  it('maps unknown error to fallback message', () => {
    const result = mapApiError('UNKNOWN_ERROR' as ApiErrorState, 'zh-CN')
    expect(result).toBe(zhCN.errors.UNKNOWN)
  })

  it('never returns raw exception text', () => {
    const result = mapApiError('NETWORK_ERROR', 'zh-CN')
    expect(result).not.toContain('Error')
    expect(result).not.toContain('stack')
    expect(result).not.toContain('at line')
  })

  it('all error codes exist in all three locales', () => {
    const errorCodes: ApiErrorState[] = ['NETWORK_ERROR', 'SERVER_ERROR', 'TIMEOUT', 'UNKNOWN']
    for (const code of errorCodes) {
      expect(zhCN.errors).toHaveProperty(code)
      expect(jaJP.errors).toHaveProperty(code)
      expect(enUS.errors).toHaveProperty(code)
    }
  })
})
