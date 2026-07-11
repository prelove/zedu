import { describe, it, expect } from 'vitest'
import { i18n, setLocale, getLocale } from '../src/i18n/index'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

describe('i18n module', () => {
  it('getLocale returns default locale initially', () => {
    expect(getLocale()).toBe('zh-CN')
  })

  it('setLocale changes the current locale', () => {
    setLocale('ja-JP')
    expect(getLocale()).toBe('ja-JP')
    expect(i18n.global.t('app.name')).toBe(jaJP.app.name)
  })

  it('setLocale to en-US works', () => {
    setLocale('en-US')
    expect(getLocale()).toBe('en-US')
    expect(i18n.global.t('app.name')).toBe(enUS.app.name)
  })

  it('setLocale back to zh-CN works', () => {
    setLocale('zh-CN')
    expect(getLocale()).toBe('zh-CN')
    expect(i18n.global.t('app.name')).toBe(zhCN.app.name)
  })
})
