export type Locale = 'zh-CN' | 'ja-JP' | 'en-US'

export const SUPPORTED_LOCALES: readonly Locale[] = ['zh-CN', 'ja-JP', 'en-US'] as const

export const DEFAULT_LOCALE: Locale = 'zh-CN'

export const TIMEZONE = 'Asia/Tokyo'

export function isSupportedLocale(value: string): value is Locale {
  return (SUPPORTED_LOCALES as readonly string[]).includes(value)
}
