import { describe, expect, it } from 'vitest'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

describe('finance payments i18n', () => {
  it('contains required finance payments keys in all locales', () => {
    const requiredKeys = [
      'title',
      'createTitle',
      'createDescription',
      'listTitle',
      'detailTitle',
      'ledgerTitle',
      'paymentNo',
      'student',
      'enrollment',
      'paymentMethod',
      'originalAmount',
      'originalCurrency',
      'fxRateToBase',
      'amountBase',
      'lessonsAdded',
      'paidAt',
      'submitPayment',
      'attachmentsTitle',
      'uploadAttachment',
      'downloadAttachment',
      'voidTitle',
      'voidReason',
    ]

    for (const locale of [zhCN, jaJP, enUS]) {
      const nav = locale.nav as Record<string, string>
      const financePayments = (locale as Record<string, any>).financePayments as Record<string, string>
      expect(nav).toHaveProperty('financePayments')
      for (const key of requiredKeys) {
        expect(financePayments).toHaveProperty(key)
        expect(financePayments[key]).not.toBe('')
      }
    }
  })
})
