import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

function nonEmptyLineCount(relativePath: string): number {
  const absolutePath = resolve(import.meta.dirname, '..', relativePath)
  return readFileSync(absolutePath, 'utf8')
    .split(/\r?\n/)
    .filter((line) => line.trim().length > 0)
    .length
}

describe('frontend component line limits', () => {
  const componentLimit = 200
  const viewLimit = 350
  const testLimit = 600

  it.each([
    'src/features/directory/components/ParentsSection.vue',
    'src/features/directory/components/EnrollmentsSection.vue',
    'src/features/directory/components/CapabilitiesSection.vue',
    'src/features/directory/components/AvailabilitySection.vue',
    'src/features/course/components/AssignmentsSection.vue',
    'src/features/course/components/DictionaryTab.vue',
  ])('%s stays within the shared component limit', (relativePath) => {
    expect(nonEmptyLineCount(relativePath)).toBeLessThanOrEqual(componentLimit)
  })

  it.each([
    'src/features/course/CourseDictionaryView.vue',
    'src/features/course/EnrollmentDetailView.vue',
  ])('%s stays within the view limit', (relativePath) => {
    expect(nonEmptyLineCount(relativePath)).toBeLessThanOrEqual(viewLimit)
  })

  it.each([
    'tests/student-subcomponents.test.ts',
    'tests/teacher-subcomponents.test.ts',
    'tests/course-subcomponents.test.ts',
    'tests/course-dict-view.test.ts',
    'tests/enrollment-detail-view.test.ts',
  ])('%s stays within the test file limit', (relativePath) => {
    expect(nonEmptyLineCount(relativePath)).toBeLessThanOrEqual(testLimit)
  })
})
