import { describe, it, expect } from 'vitest'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

/**
 * Recursively collect all leaf key paths from a nested locale object.
 * e.g. { app: { title: 'X' } } => ['app.title']
 */
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

describe('i18n key parity', () => {
  it('zh-CN and ja-JP have identical key sets', () => {
    const zhKeys = collectKeys(zhCN)
    const jaKeys = collectKeys(jaJP)
    expect(jaKeys).toEqual(zhKeys)
  })

  it('zh-CN and en-US have identical key sets', () => {
    const zhKeys = collectKeys(zhCN)
    const enKeys = collectKeys(enUS)
    expect(enKeys).toEqual(zhKeys)
  })

  it('all three locales have at least the required top-level groups', () => {
    const required = ['app', 'health', 'errors', 'common']
    for (const locale of [zhCN, jaJP, enUS]) {
      for (const group of required) {
        expect(locale).toHaveProperty(group)
      }
    }
  })

  it('locale values contain no empty strings', () => {
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

  it('supports CJK and emoji characters in values', () => {
    // zh-CN app name contains Chinese characters (CJK Unified Ideographs)
    expect(zhCN.app.name).toMatch(/[\u4e00-\u9fff]/u)
    // ja-JP app name contains Japanese characters
    expect(jaJP.app.name).toMatch(/[\u3040-\u30ff\u4e00-\u9fff]/u)
    // At least one locale value should contain emoji
    const allValues = JSON.stringify(zhCN) + JSON.stringify(jaJP) + JSON.stringify(enUS)
    expect(allValues).toMatch(/[\u{1F300}-\u{1F9FF}]/u)
  })
})
