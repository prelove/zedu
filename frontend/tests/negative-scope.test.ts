import { describe, it, expect } from 'vitest'
import { readFileSync, readdirSync } from 'node:fs'
import { join, resolve } from 'node:path'

describe('frontend negative scope: no forbidden pages or endpoints', () => {
  const forbiddenPatterns = [
    /\/attendance/,
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
        if (
          content.includes(`path: '${pattern.source}'`) ||
          content.includes(`fetch('${pattern.source}'`) ||
          content.includes(`httpRequest('${pattern.source}'`)
        ) {
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
        if (
          content.includes(`fetch('${pattern.source}'`) ||
          content.includes(`httpRequest('${pattern.source}'`)
        ) {
          violations.push(`${file}: requests ${pattern.source}`)
        }
      }
    }
    expect(violations, `Forbidden requests found: ${violations.join(', ')}`).toEqual([])
  })

  it('router defines only approved routes', () => {
    const routerContent = readFileSync(resolve(srcDir, 'router', 'index.ts'), 'utf-8')
    const pathMatches = routerContent.match(/path:\s*'[^']+'/g) ?? []
    const paths = pathMatches.map((match) => match.match(/'([^']+)'/)?.[1] ?? '')
    const approved = [
      '/login',
      '/',
      '/onboarding',
      '/students',
      '/students/:id',
      '/teachers',
      '/teachers/:id',
      '/courses',
      '/finance/payments',
      '/lessons',
      '/enrollments/:id',
    ]

    for (const path of approved) {
      expect(paths).toContain(path)
    }

    for (const path of paths) {
      for (const pattern of forbiddenPatterns) {
        expect(path, `route ${path} matches forbidden pattern ${pattern.source}`).not.toMatch(pattern)
      }
    }
  })
})
