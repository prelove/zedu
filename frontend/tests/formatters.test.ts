import { describe, it, expect } from 'vitest'
import { formatDate, formatJPY, formatDateTime } from '../src/utils/formatters'

describe('formatDate', () => {
  // Fixed UTC timestamp: 2026-07-11T03:00:00.000Z
  const testDate = new Date('2026-07-11T03:00:00.000Z')

  it('formats date in zh-CN with Asia/Tokyo timezone', () => {
    const result = formatDate(testDate, 'zh-CN')
    // 2026-07-11T03:00:00Z = 2026-07-11 12:00 JST
    expect(result).toContain('2026')
    expect(result).toContain('7')
    expect(result).toContain('11')
  })

  it('formats date in ja-JP with Asia/Tokyo timezone', () => {
    const result = formatDate(testDate, 'ja-JP')
    expect(result).toContain('2026')
    expect(result).toContain('7')
    expect(result).toContain('11')
  })

  it('formats date in en-US with Asia/Tokyo timezone', () => {
    const result = formatDate(testDate, 'en-US')
    // en-US formats month as a word: "July 11, 2026"
    expect(result).toContain('2026')
    expect(result).toContain('July')
    expect(result).toContain('11')
  })

  it('does not depend on Windows system locale', () => {
    // The formatter should always use Asia/Tokyo regardless of system locale
    const result = formatDate(testDate, 'en-US')
    // Should produce a consistent result (not dependent on system language)
    expect(result).toMatch(/2026/)
  })
})

describe('formatDateTime', () => {
  const testDate = new Date('2026-07-11T03:00:00.000Z')

  it('includes time component in ja-JP', () => {
    const result = formatDateTime(testDate, 'ja-JP')
    // 12:00 JST
    expect(result).toContain('12')
  })
})

describe('formatJPY', () => {
  it('formats integer amount in zh-CN', () => {
    const result = formatJPY(1000, 'zh-CN')
    expect(result).toContain('1,000')
    expect(result).toContain('¥')
  })

  it('formats integer amount in ja-JP', () => {
    const result = formatJPY(1000, 'ja-JP')
    expect(result).toContain('1,000')
    // ja-JP locale uses fullwidth yen sign ￥ (U+FFE5)
    expect(result).toMatch(/[¥￥]/)
  })

  it('formats integer amount in en-US', () => {
    const result = formatJPY(1000, 'en-US')
    expect(result).toContain('1,000')
    expect(result).toContain('¥')
  })

  it('does not use float for currency (accepts integer cents or integer yen)', () => {
    // JPY has no fractional units; formatJPY should accept integer yen
    const result = formatJPY(12345, 'ja-JP')
    expect(result).toContain('12,345')
  })

  it('handles zero correctly', () => {
    const result = formatJPY(0, 'zh-CN')
    expect(result).toContain('0')
    expect(result).toContain('¥')
  })

  it('handles large amounts without float precision loss', () => {
    const result = formatJPY(100000000, 'ja-JP')
    expect(result).toContain('100,000,000')
  })

  it('throws TypeError for float values (prohibited by coding standard)', () => {
    expect(() => formatJPY(1000.5, 'zh-CN')).toThrow(TypeError)
    expect(() => formatJPY(1000.5, 'zh-CN')).toThrow(/integer/)
  })

  it('throws TypeError for NaN', () => {
    expect(() => formatJPY(NaN, 'zh-CN')).toThrow(TypeError)
  })
})
