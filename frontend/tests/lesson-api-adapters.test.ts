import { afterEach, describe, expect, it, vi } from 'vitest'
import { cancelLesson, createLesson, listLessons, updateLesson } from '../src/api/lesson'

function response(data: unknown): Response {
  return { ok: true, status: 200, json: async () => ({ code: 0, data }) } as Response
}

describe('lesson API adapter', () => {
  const originalFetch = globalThis.fetch

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('uses the frozen list filters without an /api prefix', async () => {
    let calledURL = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => { calledURL = url; return Promise.resolve(response({ items: [], page: 1, pageSize: 20, total: 0 })) })
    await listLessons('access', { studentId: 7, status: 'SCHEDULED', page: 2, pageSize: 10 })
    expect(calledURL).toBe('/lessons?studentId=7&status=SCHEDULED&page=2&pageSize=10')
  })

  it('creates and updates with the distinct M4a payload shapes', async () => {
    const calls: Array<{ url: string; method: string; body: unknown }> = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, options: RequestInit) => {
      calls.push({ url, method: options.method ?? 'GET', body: JSON.parse(String(options.body)) })
      return Promise.resolve(response({ id: 1 }))
    })
    await createLesson('access', { enrollmentId: 3, assignmentId: 4, startAt: '2026-08-01T19:00:00', durationMin: 60, timezone: 'Asia/Tokyo', meetingType: 'OFFLINE' })
    await updateLesson('access', 1, { startAt: '2026-08-02T19:00:00', durationMin: 90, timezone: 'Asia/Tokyo', meetingType: 'OFFLINE' })
    expect(calls[0]).toMatchObject({ url: '/lessons', method: 'POST', body: { enrollmentId: 3, assignmentId: 4 } })
    expect(calls[1]).toMatchObject({ url: '/lessons/1', method: 'PATCH' })
    expect(calls[1].body).not.toHaveProperty('enrollmentId')
    expect(calls[1].body).not.toHaveProperty('assignmentId')
  })

  it('cancels by POST with a required reason', async () => {
    let called: { url: string; method: string; body: unknown } | undefined
    globalThis.fetch = vi.fn().mockImplementation((url: string, options: RequestInit) => { called = { url, method: options.method ?? 'GET', body: JSON.parse(String(options.body)) }; return Promise.resolve(response({ id: 1 })) })
    await cancelLesson('access', 1, 'student absence')
    expect(called).toEqual({ url: '/lessons/1/cancel', method: 'POST', body: { reason: 'student absence' } })
  })
})
