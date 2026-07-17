import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import EnrollmentDetailView from '../src/features/course/EnrollmentDetailView.vue'
import { authStore } from '../src/stores/auth'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

function mockResponse(body: unknown, status = 200): Response {
  return { ok: status < 300, status, json: async () => body } as Response
}

function testI18n() {
  return createI18n({ legacy: false, locale: 'zh-CN', fallbackLocale: 'zh-CN', messages: { 'zh-CN': zhCN, 'ja-JP': jaJP, 'en-US': enUS } })
}

function testRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/', component: { template: '<div />' } }],
  })
}

describe('EnrollmentDetailView', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    authStore.clearSession()
    authStore.state.accessToken = 'test-token'
    authStore.state.role = 'OPERATOR'
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  function setupFetch(opts: { hasActiveAssignment?: boolean } = {}): void {
    globalThis.fetch = vi.fn().mockImplementation((url: string, method: any) => {
      const m = method?.method
      if (url.startsWith('/enrollments/5') && !url.includes('/assignments')) {
        if (m === 'PATCH') {
          return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 2, enrollmentType: 'REGULAR', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
        }
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, targetLevelId: 2, enrollmentType: 'REGULAR', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/enrollments/5/assignments')) {
        if (m === 'POST') {
          return Promise.resolve(mockResponse({ code: 0, data: { id: 2, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '', createdAt: '', updatedAt: '' } }, 201))
        }
        const items = opts.hasActiveAssignment
          ? [{ id: 1, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '', createdAt: '', updatedAt: '' }]
          : []
        return Promise.resolve(mockResponse({ code: 0, data: { items, page: 1, pageSize: 20, total: items.length } }))
      }
      if (url.startsWith('/assignments/') && url.endsWith('/end')) {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ENDED', startDate: '', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'Nihongo', code: 'n1', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [
          { id: 1, trackId: 1, name: 'N5', code: 'n5', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' },
          { id: 2, trackId: 1, name: 'N4', code: 'n4', sortOrder: 1, enabled: true, createdAt: '', updatedAt: '' },
        ], page: 1, pageSize: 20, total: 2 } }))
      }
      if (url.startsWith('/teachers')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 3, name: 'Sensei', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
  }

  it('renders enrollment detail with course selection and level change sections', async () => {
    setupFetch()
    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="enrollment-detail-view"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="course-selection-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="level-change-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="assignments-section"]').exists()).toBe(true)
  })

  it('course selection PATCH does not include currentLevelId', async () => {
    const calls: any[] = []
    setupFetch()
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url === '/enrollments/5' && opts?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url === '/enrollments/5') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, targetLevelId: 2, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/enrollments/5/assignments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, trackId: 1, name: 'L1', code: 'l1', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }, { id: 2, trackId: 1, name: 'L2', code: 'l2', sortOrder: 1, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 2 } }))
      }
      if (url.startsWith('/teachers')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="save-course-selection"]').trigger('click')
    await flushPromises()

    const patch = calls.find((c) => c.url === '/enrollments/5' && c.opts?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch.opts.body)
    expect(body.currentLevelId).toBeUndefined()
  })

  it('level change PATCH does not include domainId/trackId/targetLevelId', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url === '/enrollments/5' && opts?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 2, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url === '/enrollments/5') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/enrollments/5/assignments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, trackId: 1, name: 'L1', code: 'l1', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }, { id: 2, trackId: 1, name: 'L2', code: 'l2', sortOrder: 1, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 2 } }))
      }
      if (url.startsWith('/teachers')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    // Change current level to N4 (id=2) — directly update the component's
    // reactive ref since v-model.number on select doesn't work reliably in jsdom.
    const vm = wrapper.vm as any
    vm.selCurrentLevel = 2
    await wrapper.find('[data-testid="save-level-change"]').trigger('click')
    await flushPromises()

    const patch = calls.find((c) => c.url === '/enrollments/5' && c.opts?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch.opts.body)
    // currentLevelId should be 2 (N4). It may be string "2" or number 2.
    expect(Number(body.currentLevelId)).toBe(2)
    expect(body.domainId).toBeUndefined()
    expect(body.trackId).toBeUndefined()
    expect(body.targetLevelId).toBeUndefined()
  })

  it('same level change shows 42201 same-level error', async () => {
    setupFetch()
    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    // Current level is already 1; don't change it.
    await wrapper.find('[data-testid="save-level-change"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="level-change-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="level-change-error"]').text()).toContain(zhCN.enrollments.sameLevelRejected)
  })

  it('shows atomic replace hint when active assignment exists', async () => {
    setupFetch({ hasActiveAssignment: true })
    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="atomic-replace-hint"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="add-assignment-btn"]').text()).toContain(zhCN.enrollments.replaceAssignment)
  })

  it('shows no-teacher enrollment message when no assignments', async () => {
    setupFetch({ hasActiveAssignment: false })
    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="state-empty"]').text()).toContain(zhCN.enrollments.noTeacherEnrollment)
  })

  it('shows end assignment button for ACTIVE assignment', async () => {
    setupFetch({ hasActiveAssignment: true })
    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="end-assignment-btn"]').exists()).toBe(true)
  })

  it('end assignment shows confirmation dialog', async () => {
    setupFetch({ hasActiveAssignment: true })
    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="end-assignment-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(true)
  })

  it('creates assignment via form', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url === '/enrollments/5/assignments' && opts?.method === 'POST') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '', createdAt: '', updatedAt: '' } }, 201))
      }
      if (url === '/enrollments/5') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/enrollments/5/assignments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, trackId: 1, name: 'L', code: 'l', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/teachers')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 3, name: 'Sensei', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-assignment-btn"]').trigger('click')
    const vm = wrapper.vm as any
    vm.assignForm = { teacherId: 3, roleType: 'MAIN' }
    await wrapper.find('[data-testid="assignment-create-form"] form').trigger('submit.prevent')
    await flushPromises()

    const post = calls.find((c) => c.url === '/enrollments/5/assignments' && c.opts?.method === 'POST')
    expect(post).toBeDefined()
    const body = JSON.parse(post.opts.body)
    expect(body.teacherId).toBe(3)
    expect(body.roleType).toBe('MAIN')
  })

  it('ends assignment via confirmation dialog', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url === '/assignments/1/end' && opts?.method === 'POST') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ENDED', startDate: '', createdAt: '', updatedAt: '' } }))
      }
      if (url === '/enrollments/5') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/enrollments/5/assignments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [
          { id: 1, enrollmentId: 5, studentId: 10, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '', createdAt: '', updatedAt: '' },
        ], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, trackId: 1, name: 'L', code: 'l', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/teachers')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="end-assignment-btn"]').trigger('click')
    await wrapper.find('[data-testid="confirm-ok"]').trigger('click')
    await flushPromises()

    const post = calls.find((c) => c.url === '/assignments/1/end' && c.opts?.method === 'POST')
    expect(post).toBeDefined()
  })

  it('saves course selection via PATCH', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url === '/enrollments/5' && opts?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, targetLevelId: 2, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url === '/enrollments/5') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, targetLevelId: 2, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/enrollments/5/assignments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, trackId: 1, name: 'L1', code: 'l1', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }, { id: 2, trackId: 1, name: 'L2', code: 'l2', sortOrder: 1, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 2 } }))
      }
      if (url.startsWith('/teachers')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="save-course-selection"]').trigger('click')
    await flushPromises()

    const patch = calls.find((c) => c.url === '/enrollments/5' && c.opts?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch.opts.body)
    expect(body.domainId).toBe(1)
    expect(body.trackId).toBe(1)
  })

  it('shows error state on enrollment load failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 40401, message: 'NOT_FOUND', requestId: 'r1' }, 404),
    )

    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
  })

  it('shows assignment create error on POST failure', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      if (url === '/enrollments/5/assignments' && opts?.method === 'POST') {
        return Promise.resolve(mockResponse({ code: 42201, message: 'INVALID_STATE', requestId: 'r1' }, 422))
      }
      if (url === '/enrollments/5') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/enrollments/5/assignments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, trackId: 1, name: 'L', code: 'l', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/teachers')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 3, name: 'S', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-assignment-btn"]').trigger('click')
    const vm = wrapper.vm as any
    vm.assignForm = { teacherId: 3, roleType: 'MAIN' }
    await wrapper.find('[data-testid="assignment-create-form"] form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-testid="assignment-create-error"]').exists()).toBe(true)
  })

  it('cancels assignment create form', async () => {
    setupFetch()
    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-assignment-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="assignment-create-form"]').exists()).toBe(true)
    const cancelBtn = wrapper.find('[data-testid="assignment-create-form"] button[type="button"]')
    await cancelBtn.trigger('click')
    expect(wrapper.find('[data-testid="assignment-create-form"]').exists()).toBe(false)
  })

  it('cancels end assignment confirmation', async () => {
    setupFetch({ hasActiveAssignment: true })
    const wrapper = mount(EnrollmentDetailView, {
      props: { id: '5' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="end-assignment-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(true)
    await wrapper.find('[data-testid="confirm-cancel"]').trigger('click')
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(false)
  })
})
