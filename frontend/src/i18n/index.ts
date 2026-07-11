import { createI18n } from 'vue-i18n'
import { DEFAULT_LOCALE, type Locale } from './config'
import { zhCN } from './locales/zh-CN'
import { jaJP } from './locales/ja-JP'
import { enUS } from './locales/en-US'

export const i18n = createI18n({
  legacy: false,
  locale: DEFAULT_LOCALE,
  fallbackLocale: DEFAULT_LOCALE,
  messages: {
    'zh-CN': zhCN,
    'ja-JP': jaJP,
    'en-US': enUS,
  },
})

export function setLocale(locale: Locale): void {
  i18n.global.locale.value = locale
}

export function getLocale(): Locale {
  return i18n.global.locale.value as Locale
}

export { type Locale } from './config'
