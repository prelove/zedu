import { describe, it, expect } from 'vitest'
import { isSupportedLocale, SUPPORTED_LOCALES, DEFAULT_LOCALE } from '../src/i18n/config'

describe('isSupportedLocale', () => {
  it('returns true for zh-CN', () => {
    expect(isSupportedLocale('zh-CN')).toBe(true)
  })

  it('returns true for ja-JP', () => {
    expect(isSupportedLocale('ja-JP')).toBe(true)
  })

  it('returns true for en-US', () => {
    expect(isSupportedLocale('en-US')).toBe(true)
  })

  it('returns false for unsupported locale', () => {
    expect(isSupportedLocale('ko-KR')).toBe(false)
  })

  it('returns false for empty string', () => {
    expect(isSupportedLocale('')).toBe(false)
  })

  it('SUPPORTED_LOCALES has exactly 3 entries', () => {
    expect(SUPPORTED_LOCALES).toHaveLength(3)
  })

  it('DEFAULT_LOCALE is zh-CN', () => {
    expect(DEFAULT_LOCALE).toBe('zh-CN')
  })
})
