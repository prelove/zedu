import { describe, expect, it } from 'vitest'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

describe('finance config i18n', () => {
  it('contains required finance config keys in all locales', () => {
    const requiredKeys = [
      'title',
      'baseCurrencyTitle',
      'baseCurrencyDescription',
      'baseCurrencyLabel',
      'baseCurrencyLocked',
      'baseCurrencySave',
      'paymentMethodsTitle',
      'paymentMethodsDescription',
      'paymentMethodsEmpty',
      'methodCode',
      'methodName',
      'methodSortOrder',
      'methodCreate',
      'methodUpdate',
    ]

    for (const locale of [zhCN, jaJP, enUS]) {
      expect(locale.nav).toHaveProperty('financeConfig')
      expect(locale).toHaveProperty('financeConfig')
      for (const key of requiredKeys) {
        expect(locale.financeConfig).toHaveProperty(key)
        expect((locale.financeConfig as Record<string, string>)[key]).not.toBe('')
      }
    }
  })
})
