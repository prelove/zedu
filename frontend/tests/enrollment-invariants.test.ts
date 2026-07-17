import { describe, it, expect, vi, afterEach } from 'vitest'
import { updateEnrollment, createAssignment, endAssignment } from '../src/api/course'

function mockResponse(body: unknown, status = 200): Response {
  return { ok: status < 300, status, json: async () => body } as Response
}

describe('enrollment invariant: course selection and level change are separate PATCHes', () => {
  const originalFetch = globalThis.fetch

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('course selection PATCH sends domainId/trackId/targetLevelId but NOT currentLevelId', async () => {
    let calledBody: any
    globalThis.fetch = vi.fn().mockImplementation((_url: string, opts: any) => {
      calledBody = JSON.parse(opts.body)
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, studentId: 1, domainId: 1, trackId: 1, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await updateEnrollment('tok', 1, { domainId: 2, trackId: 3, targetLevelId: 5 })
    expect(calledBody.domainId).toBe(2)
    expect(calledBody.trackId).toBe(3)
    expect(calledBody.targetLevelId).toBe(5)
    expect(calledBody.currentLevelId).toBeUndefined()
  })

  it('level change PATCH sends currentLevelId but NOT domainId/trackId/targetLevelId', async () => {
    let calledBody: any
    globalThis.fetch = vi.fn().mockImplementation((_url: string, opts: any) => {
      calledBody = JSON.parse(opts.body)
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, studentId: 1, domainId: 1, trackId: 1, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await updateEnrollment('tok', 1, { currentLevelId: 7 })
    expect(calledBody.currentLevelId).toBe(7)
    expect(calledBody.domainId).toBeUndefined()
    expect(calledBody.trackId).toBeUndefined()
    expect(calledBody.targetLevelId).toBeUndefined()
  })

  it('assignment create POSTs to /enrollments/{id}/assignments (atomic replace)', async () => {
    let calledUrl = ''
    let calledMethod = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calledUrl = url
      calledMethod = opts.method
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '', createdAt: '', updatedAt: '' } }))
    })

    await createAssignment('tok', 5, { teacherId: 3 })
    expect(calledUrl).toBe('/enrollments/5/assignments')
    expect(calledMethod).toBe('POST')
  })

  it('end assignment POSTs to /assignments/{id}/end (not DELETE)', async () => {
    let calledUrl = ''
    let calledMethod = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calledUrl = url
      calledMethod = opts.method
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ENDED', startDate: '', createdAt: '', updatedAt: '' } }))
    })

    await endAssignment('tok', 1)
    expect(calledUrl).toBe('/assignments/1/end')
    expect(calledMethod).toBe('POST')
    expect(calledMethod).not.toBe('DELETE')
  })

  it('no financial fields in enrollment write body', async () => {
    let calledBody: any
    globalThis.fetch = vi.fn().mockImplementation((_url: string, opts: any) => {
      calledBody = JSON.parse(opts.body)
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, studentId: 1, domainId: 1, trackId: 1, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await updateEnrollment('tok', 1, { domainId: 2 })
    expect(calledBody).not.toHaveProperty('rateAmount')
    expect(calledBody).not.toHaveProperty('balance')
    expect(calledBody).not.toHaveProperty('chargeAmount')
    expect(calledBody).not.toHaveProperty('lessonHours')
    expect(calledBody).not.toHaveProperty('paymentAmount')
  })

  it('no financial fields in assignment write body', async () => {
    let calledBody: any
    globalThis.fetch = vi.fn().mockImplementation((_url: string, opts: any) => {
      calledBody = JSON.parse(opts.body)
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '', createdAt: '', updatedAt: '' } }))
    })

    await createAssignment('tok', 5, { teacherId: 3, roleType: 'MAIN' })
    expect(calledBody).not.toHaveProperty('rateAmount')
    expect(calledBody).not.toHaveProperty('balance')
    expect(calledBody).not.toHaveProperty('chargeAmount')
    expect(calledBody).not.toHaveProperty('lessonHours')
  })
})
