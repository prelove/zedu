import { describe, it, expect } from 'vitest'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

/**
 * Recursively collect all key paths from a locale object.
 */
function collectKeys(obj: Record<string, any>, prefix = ''): string[] {
  const keys: string[] = []
  for (const key of Object.keys(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key
    if (typeof obj[key] === 'object' && obj[key] !== null && !Array.isArray(obj[key])) {
      keys.push(...collectKeys(obj[key], fullKey))
    } else {
      keys.push(fullKey)
    }
  }
  return keys
}

describe('M2-KIMI-02 i18n key parity', () => {
  const zhKeys = collectKeys(zhCN).sort()
  const jaKeys = collectKeys(jaJP).sort()
  const enKeys = collectKeys(enUS).sort()

  it('zh-CN and ja-JP have identical key sets', () => {
    expect(jaKeys).toEqual(zhKeys)
  })

  it('zh-CN and en-US have identical key sets', () => {
    expect(enKeys).toEqual(zhKeys)
  })

  it('all locale values are non-empty strings', () => {
    function checkNonEmpty(obj: Record<string, any>, path: string): void {
      for (const key of Object.keys(obj)) {
        const v = obj[key]
        const p = `${path}.${key}`
        if (typeof v === 'object' && v !== null) {
          checkNonEmpty(v, p)
        } else {
          expect(typeof v, `${p} should be string`).toBe('string')
          expect((v as string).length, `${p} should be non-empty`).toBeGreaterThan(0)
        }
      }
    }
    checkNonEmpty(zhCN, 'zh-CN')
    checkNonEmpty(jaJP, 'ja-JP')
    checkNonEmpty(enUS, 'en-US')
  })

  it('students key group exists with required keys', () => {
    expect(zhCN.students).toBeDefined()
    expect(zhCN.students.title).toBeDefined()
    expect(zhCN.students.create).toBeDefined()
    expect(zhCN.students.emailOptional).toBeDefined()
    expect(zhCN.students.emailDuplicateBlocked).toBeDefined()
    expect(zhCN.students.noBypass).toBeDefined()
    expect(zhCN.students.parents).toBeDefined()
    expect(zhCN.students.enrollments).toBeDefined()
  })

  it('teachers key group exists with required keys', () => {
    expect(zhCN.teachers).toBeDefined()
    expect(zhCN.teachers.capabilityDuplicate).toBeDefined()
    expect(zhCN.teachers.capabilityEnd).toBeDefined()
    expect(zhCN.teachers.availabilityInvalidTime).toBeDefined()
    expect(zhCN.teachers.availabilityInvalidWeekday).toBeDefined()
  })

  it('courses key group exists with required keys', () => {
    expect(zhCN.courses).toBeDefined()
    expect(zhCN.courses.referencedCannotDisable).toBeDefined()
    expect(zhCN.courses.domains).toBeDefined()
    expect(zhCN.courses.tracks).toBeDefined()
    expect(zhCN.courses.levels).toBeDefined()
    expect(zhCN.courses.tags).toBeDefined()
  })

  it('enrollments key group exists with required keys', () => {
    expect(zhCN.enrollments).toBeDefined()
    expect(zhCN.enrollments.saveCourseSelection).toBeDefined()
    expect(zhCN.enrollments.saveLevelChange).toBeDefined()
    expect(zhCN.enrollments.courseSelectionHint).toBeDefined()
    expect(zhCN.enrollments.levelChangeHint).toBeDefined()
    expect(zhCN.enrollments.sameLevelRejected).toBeDefined()
    expect(zhCN.enrollments.atomicReplaceHint).toBeDefined()
    expect(zhCN.enrollments.endAssignmentDuplicate).toBeDefined()
  })

  it('nav key group exists with required keys', () => {
    expect(zhCN.nav).toBeDefined()
    expect(zhCN.nav.students).toBeDefined()
    expect(zhCN.nav.teachers).toBeDefined()
    expect(zhCN.nav.courses).toBeDefined()
  })
})
