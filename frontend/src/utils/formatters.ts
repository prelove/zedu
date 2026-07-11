import { TIMEZONE, type Locale } from '../i18n/config'

/**
 * Format a date for display using the specified locale and Asia/Tokyo timezone.
 * Does not depend on the Windows system language or timezone.
 */
export function formatDate(date: Date, locale: Locale): string {
  return new Intl.DateTimeFormat(locale, {
    timeZone: TIMEZONE,
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(date)
}

/**
 * Format a date with time for display using the specified locale and Asia/Tokyo timezone.
 */
export function formatDateTime(date: Date, locale: Locale): string {
  return new Intl.DateTimeFormat(locale, {
    timeZone: TIMEZONE,
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

/**
 * Format an integer JPY amount for display.
 * JPY has no fractional units, so the input must be an integer.
 * Uses explicit locale and currency; does not depend on system locale.
 */
export function formatJPY(amountYen: number, locale: Locale): string {
  if (!Number.isInteger(amountYen)) {
    throw new TypeError('JPY amount must be an integer; float values are prohibited.')
  }
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency: 'JPY',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amountYen)
}
