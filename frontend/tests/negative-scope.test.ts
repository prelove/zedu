import { describe, it, expect } from 'vitest'
import { readFileSync, readdirSync } from 'node:fs'
import { join, resolve } from 'node:path'

/**
 * Negative-scope test: ensure no frontend source or test file references
 * forbidden capabilities — lesson, attendance, payment, notification, backup,
 * report, payout, settlement — or student/teacher/parent login pages.
 *
 * M2-KIMI-02 allows students, teachers, parents, enrollments, assignments,
 * course-domains, tracks, levels, and capability-tags paths.
 */
describe('M2-KIMI-02 negative scope: no forbidden pages or endpoints', () => {
  // Forbidden patterns: capabilities NOT in M2-KIMI-02 scope.
  const forbiddenPatterns = [
    /\/lesson/,
    /\/attendance/,
    /\/payment/,
    /\/notification/,
    /\/backup/,
    /\/report/,
    /\/payout/,
    /\/settlement/,
  ]

  const srcDir = resolve(__dirname, '..', 'src')
  const testsDir = resolve(__dirname)

  function collectTsVueFiles(dir: string, files: string[] = []): string[] {
    const entries = readdirSync(dir, { withFileTypes: true })
    for (const entry of entries) {
      const fullPath = join(dir, entry.name)
      if (entry.isDirectory() && entry.name !== 'node_modules' && entry.name !== 'dist') {
        collectTsVueFiles(fullPath, files)
      } else if (entry.isFile() && (entry.name.endsWith('.ts') || entry.name.endsWith('.vue'))) {
        files.push(fullPath)
      }
    }
    return files
  }

  it('no src/ file registers a forbidden route or API path', () => {
    const files = collectTsVueFiles(srcDir)
    const violations: string[] = []
    for (const file of files) {
      const content = readFileSync(file, 'utf-8')
      for (const pattern of forbiddenPatterns) {
        // Check for route registration patterns like path: '/lesson' or
        // API call patterns like fetch('/lesson') or httpRequest('/lesson')
        if (content.includes(`path: '${pattern.source}'`) ||
            content.includes(`fetch('${pattern.source}'`) ||
            content.includes(`httpRequest('${pattern.source}'`)) {
          violations.push(`${file}: matches ${pattern.source}`)
        }
      }
    }
    expect(violations, `Forbidden paths found: ${violations.join(', ')}`).toEqual([])
  })

  it('no tests/ file makes a request to a forbidden endpoint', () => {
    const files = collectTsVueFiles(testsDir)
    const violations: string[] = []
    for (const file of files) {
      if (file.endsWith('negative-scope.test.ts')) continue
      const content = readFileSync(file, 'utf-8')
      for (const pattern of forbiddenPatterns) {
        if (content.includes(`fetch('${pattern.source}'`) ||
            content.includes(`httpRequest('${pattern.source}'`)) {
          violations.push(`${file}: requests ${pattern.source}`)
        }
      }
    }
    expect(violations, `Forbidden requests found: ${violations.join(', ')}`).toEqual([])
  })

  it('router defines only approved routes', () => {
    const routerContent = readFileSync(resolve(srcDir, 'router', 'index.ts'), 'utf-8')
    const pathMatches = routerContent.match(/path:\s*'[^']+'/g) ?? []
    const paths = pathMatches.map((m) => m.match(/'([^']+)'/)?.[1] ?? '')
    // Approved routes for M2-KIMI-01 + M2-KIMI-02.
    const approved = [
      '/login',
      '/',
      '/onboarding',
      '/students',
      '/students/:id',
      '/teachers',
      '/teachers/:id',
      '/courses',
      '/enrollments/:id',
    ]
    for (const p of approved) {
      expect(paths).toContain(p)
    }
    // No forbidden routes (lesson, attendance, payment, etc.).
    for (const p of paths) {
      for (const pattern of forbiddenPatterns) {
        expect(p, `route ${p} matches forbidden pattern ${pattern.source}`).not.toMatch(pattern)
      }
    }
  })
})
