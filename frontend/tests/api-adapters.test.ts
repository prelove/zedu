import { describe, it, expect, vi, afterEach } from 'vitest'
import {
  listStudents, createStudent, getStudent, updateStudent,
  listParents, createParent, updateParent,
  listTeachers, createTeacher,
  listCapabilities, createCapability,
  listAvailability,
} from '../src/api/directory'
import {
  listDomains, createDomain, updateDomain,
  listTracks,
  listLevels,
  listTags,
  listEnrollments, createEnrollment, getEnrollment, updateEnrollment,
  listAssignments, createAssignment, endAssignment,
} from '../src/api/course'

function mockResponse(body: unknown, status = 200): Response {
  return { ok: status < 300, status, json: async () => body } as Response
}

describe('directory API adapter', () => {
  const originalFetch = globalThis.fetch

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('listStudents calls /students with pagination query and Bearer token', async () => {
    let calledUrl = ''
    let calledHeaders: Record<string, string> = {}
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calledUrl = url
      calledHeaders = opts.headers
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listStudents('tok-123', { page: 2, pageSize: 50, search: 'alice' })
    expect(calledUrl).toBe('/students?page=2&pageSize=50&search=alice')
    expect(calledHeaders['Authorization']).toBe('Bearer tok-123')
  })

  it('createStudent POSTs to /students with body', async () => {
    let calledUrl = ''
    let calledMethod = ''
    let calledBody: any
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calledUrl = url
      calledMethod = opts.method
      calledBody = JSON.parse(opts.body)
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Test', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await createStudent('tok', { name: 'Test', email: '' })
    expect(calledUrl).toBe('/students')
    expect(calledMethod).toBe('POST')
    expect(calledBody.name).toBe('Test')
  })

  it('getStudent calls /students/{id} with GET', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 42, name: 'X', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await getStudent('tok', 42)
    expect(calledUrl).toBe('/students/42')
  })

  it('updateStudent PATCHes /students/{id}', async () => {
    let calledUrl = ''
    let calledMethod = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calledUrl = url
      calledMethod = opts.method
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Updated', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await updateStudent('tok', 1, { name: 'Updated' })
    expect(calledUrl).toBe('/students/1')
    expect(calledMethod).toBe('PATCH')
  })

  it('listParents calls /students/{id}/parents', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listParents('tok', 5)
    expect(calledUrl).toBe('/students/5/parents')
  })

  it('createParent POSTs to /students/{id}/parents', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string, _opts: any) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, studentId: 5, name: 'P', isPrimary: false, createdAt: '', updatedAt: '' } }))
    })

    await createParent('tok', 5, { name: 'P' })
    expect(calledUrl).toBe('/students/5/parents')
  })

  it('updateParent PATCHes /students/{id}/parents/{parentId}', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 3, studentId: 5, name: 'P', isPrimary: false, createdAt: '', updatedAt: '' } }))
    })

    await updateParent('tok', 5, 3, { name: 'Updated' })
    expect(calledUrl).toBe('/students/5/parents/3')
  })

  it('listTeachers calls /teachers with search and status', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listTeachers('tok', { status: 'ACTIVE', search: 'bob' })
    expect(calledUrl).toBe('/teachers?search=bob&status=ACTIVE')
  })

  it('createTeacher POSTs to /teachers', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'T', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await createTeacher('tok', { name: 'T' })
    expect(calledUrl).toBe('/teachers')
  })

  it('listCapabilities calls /teachers/{id}/capabilities', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listCapabilities('tok', 7)
    expect(calledUrl).toBe('/teachers/7/capabilities')
  })

  it('createCapability POSTs to /teachers/{id}/capabilities', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, teacherId: 7, domainId: 1, trackId: 1, levelId: 1, status: 'ACTIVE', verified: false, createdAt: '', updatedAt: '' } }))
    })

    await createCapability('tok', 7, { domainId: 1, trackId: 1, levelId: 1 })
    expect(calledUrl).toBe('/teachers/7/capabilities')
  })

  it('listAvailability calls /teachers/{id}/availability', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listAvailability('tok', 7)
    expect(calledUrl).toBe('/teachers/7/availability')
  })

  it('no path has /api prefix', async () => {
    const urls: string[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      urls.push(url)
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listStudents('tok')
    await listTeachers('tok')
    await listParents('tok', 1)
    await listCapabilities('tok', 1)
    await listAvailability('tok', 1)

    for (const u of urls) {
      expect(u.startsWith('/api')).toBe(false)
    }
  })
})

describe('course API adapter', () => {
  const originalFetch = globalThis.fetch

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('listDomains calls /course-domains', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listDomains('tok')
    expect(calledUrl).toBe('/course-domains')
  })

  it('listTracks calls /tracks with domainId filter', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listTracks('tok', { domainId: 3 })
    expect(calledUrl).toBe('/tracks?domainId=3')
  })

  it('listLevels calls /levels with trackId filter', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listLevels('tok', { trackId: 5 })
    expect(calledUrl).toBe('/levels?trackId=5')
  })

  it('listTags calls /capability-tags with domainId filter', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listTags('tok', { domainId: 2 })
    expect(calledUrl).toBe('/capability-tags?domainId=2')
  })

  it('createDomain POSTs to /course-domains', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'D', code: 'd1', type: 'LANGUAGE', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' } }))
    })

    await createDomain('tok', { name: 'D', code: 'd1' })
    expect(calledUrl).toBe('/course-domains')
  })

  it('updateDomain PATCHes /course-domains/{id}', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'D', code: 'd1', type: 'LANGUAGE', sortOrder: 0, enabled: false, createdAt: '', updatedAt: '' } }))
    })

    await updateDomain('tok', 1, { enabled: false })
    expect(calledUrl).toBe('/course-domains/1')
  })

  it('listEnrollments calls /students/{id}/enrollments', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listEnrollments('tok', 10)
    expect(calledUrl).toBe('/students/10/enrollments')
  })

  it('createEnrollment POSTs to /students/{id}/enrollments', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, studentId: 10, domainId: 1, trackId: 1, enrollmentType: 'REGULAR', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await createEnrollment('tok', 10, { domainId: 1, trackId: 1 })
    expect(calledUrl).toBe('/students/10/enrollments')
  })

  it('getEnrollment calls /enrollments/{id}', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, enrollmentType: 'REGULAR', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await getEnrollment('tok', 5)
    expect(calledUrl).toBe('/enrollments/5')
  })

  it('updateEnrollment PATCHes /enrollments/{id}', async () => {
    let calledUrl = ''
    let calledBody: any
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calledUrl = url
      calledBody = JSON.parse(opts.body)
      return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, enrollmentType: 'REGULAR', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
    })

    await updateEnrollment('tok', 5, { domainId: 2 })
    expect(calledUrl).toBe('/enrollments/5')
    expect(calledBody.domainId).toBe(2)
    // Must NOT include currentLevelId in course selection PATCH.
    expect(calledBody.currentLevelId).toBeUndefined()
  })

  it('listAssignments calls /enrollments/{id}/assignments', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listAssignments('tok', 5)
    expect(calledUrl).toBe('/enrollments/5/assignments')
  })

  it('createAssignment POSTs to /enrollments/{id}/assignments', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '', createdAt: '', updatedAt: '' } }))
    })

    await createAssignment('tok', 5, { teacherId: 3, roleType: 'MAIN' })
    expect(calledUrl).toBe('/enrollments/5/assignments')
  })

  it('endAssignment POSTs to /assignments/{id}/end', async () => {
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
  })

  it('no course path has /api prefix', async () => {
    const urls: string[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      urls.push(url)
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listDomains('tok')
    await listTracks('tok')
    await listLevels('tok')
    await listTags('tok')
    await listEnrollments('tok', 1)
    await listAssignments('tok', 1)

    for (const u of urls) {
      expect(u.startsWith('/api')).toBe(false)
    }
  })

  it('no forbidden API paths are used', async () => {
    const urls: string[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      urls.push(url)
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listDomains('tok')
    await listTracks('tok')
    await listEnrollments('tok', 1)
    await listAssignments('tok', 1)
    await endAssignment('tok', 1)

    const forbidden = ['/lesson', '/attendance', '/payment', '/notification', '/backup', '/report', '/payout', '/settlement']
    for (const u of urls) {
      for (const f of forbidden) {
        expect(u).not.toContain(f)
      }
    }
  })
})
