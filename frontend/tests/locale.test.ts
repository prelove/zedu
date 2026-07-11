import { describe, it, expect, beforeEach } from 'vitest'
import { createI18n } from 'vue-i18n'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'
import { SUPPORTED_LOCALES, DEFAULT_LOCALE, type Locale } from '../src/i18n/config'

describe('i18n config', () => {
  it('supports exactly three locales: zh-CN, ja-JP, en-US', () => {
    expect(SUPPORTED_LOCALES).toEqual(['zh-CN', 'ja-JP', 'en-US'])
  })

  it('default locale is zh-CN', () => {
    expect(DEFAULT_LOCALE).toBe('zh-CN')
  })
})

describe('i18n instance', () => {
  let i18n: ReturnType<typeof createI18n>

  beforeEach(() => {
    i18n = createI18n({
      legacy: false,
      locale: DEFAULT_LOCALE,
      fallbackLocale: DEFAULT_LOCALE,
      messages: {
        'zh-CN': zhCN,
        'ja-JP': jaJP,
        'en-US': enUS,
      },
    })
  })

  it('translates a key in zh-CN', () => {
    const t = i18n.global.t
    expect(t('app.name')).toBe(zhCN.app.name)
  })

  it('switches to ja-JP and translates', () => {
    i18n.global.locale.value = 'ja-JP' as Locale
    const t = i18n.global.t
    expect(t('app.name')).toBe(jaJP.app.name)
  })

  it('switches to en-US and translates', () => {
    i18n.global.locale.value = 'en-US' as Locale
    const t = i18n.global.t
    expect(t('app.name')).toBe(enUS.app.name)
  })

  it('falls back to default locale when key is missing in current locale', () => {
    // Create a minimal i18n with a missing key in ja-JP
    const partialJa = JSON.parse(JSON.stringify(jaJP)) as typeof jaJP
    delete (partialJa as Record<string, unknown>).health
    const partialI18n = createI18n({
      legacy: false,
      locale: 'ja-JP',
      fallbackLocale: 'zh-CN',
      messages: {
        'zh-CN': zhCN,
        'ja-JP': partialJa,
        'en-US': enUS,
      },
    })
    const t = partialI18n.global.t
    // health.healthy exists in zh-CN but not in partial ja-JP
    expect(t('health.healthy')).toBe(zhCN.health.healthy)
  })
})
