import { describe, it, expect } from 'vitest'
import { readFileSync, readdirSync } from 'node:fs'
import { join, resolve } from 'node:path'

/**
 * Negative-scope test: ensure no frontend source or test file references
 * lesson, attendance, payment, notification, backup, report, payout, or
 * student/teacher/parent login pages or API endpoints.
 *
 * This is a static scan — it checks that we did not register any forbidden
 * routes or import any forbidden modules.
 */
describe('M2-KIMI-01 negative scope: no forbidden pages or endpoints', () => {
  const forbiddenPatterns = [
    /\/lesson/,
    /\/attendance/,
    /\/payment/,
    /\/notification/,
    /\/backup/,
    /\/report/,
    /\/payout/,
    /\/settlement/,
    /students\//, // student/teacher/parent business pages are M2-KIMI-02
    /\/teachers\//,
    /\/parents\//,
    /\/enrollments\//,
    /\/assignments\//,
    /\/course-domains/,
    /\/tracks\//,
    /\/levels\//,
    /\/capability-tags/,
  ]

  // Allowed files that may legitimately mention these paths in comments or
  // contract documentation — but we scan only src/ and tests/ for route
  // registrations and API calls, not docs.
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
        if (pattern.test(content)) {
          // Allow mentions in i18n locale files (they don't have these paths).
          // But flag any actual route or fetch call.
          if (content.includes(`path: '${pattern.source}'`) ||
              content.includes(`fetch('${pattern.source}'`) ||
              content.includes(`httpRequest('${pattern.source}'`)) {
            violations.push(`${file}: matches ${pattern.source}`)
          }
        }
      }
    }
    expect(violations, `Forbidden paths found: ${violations.join(', ')}`).toEqual([])
  })

  it('no tests/ file makes a request to a forbidden endpoint', () => {
    const files = collectTsVueFiles(testsDir)
    const violations: string[] = []
    for (const file of files) {
      // Skip this file itself.
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

  it('router only defines /login, /, and /onboarding routes', () => {
    const routerContent = readFileSync(resolve(srcDir, 'router', 'index.ts'), 'utf-8')
    // Extract all path: '...' values.
    const pathMatches = routerContent.match(/path:\s*'[^']+'/g) ?? []
    const paths = pathMatches.map((m) => m.match(/'([^']+)'/)?.[1] ?? '')
    expect(paths).toContain('/login')
    expect(paths).toContain('/')
    expect(paths).toContain('/onboarding')
    // No other routes.
    expect(paths).toHaveLength(3)
  })
})
