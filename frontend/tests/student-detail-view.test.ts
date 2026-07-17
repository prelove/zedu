import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import StudentDetailView from '../src/features/directory/StudentDetailView.vue'
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
    routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/students', name: 'students', component: { template: '<div />' } },
      { path: '/students/:id', name: 'student-detail', component: StudentDetailView },
      { path: '/enrollments/:id', name: 'enrollment-detail', component: { template: '<div />' } },
    ],
  })
}

describe('StudentDetailView', () => {
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

  function setupFetch(opts: { hasParents?: boolean; hasEnrollments?: boolean } = {}): void {
    globalThis.fetch = vi.fn().mockImplementation((url: string, methodObj: any) => {
      const m = methodObj?.method
      // Student detail GET/PATCH
      if (url === '/students/1' && m === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Updated', email: 'a@b.com', phone: '123', timezone: 'Asia/Tokyo', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url === '/students/1') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Alice', email: 'a@b.com', phone: '123', timezone: 'Asia/Tokyo', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      // Parents
      if (url === '/students/1/parents' && m === 'POST') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, studentId: 1, name: 'Parent2', isPrimary: false, createdAt: '', updatedAt: '' } }, 201))
      }
      if (url.startsWith('/students/1/parents')) {
        const items = opts.hasParents
          ? [{ id: 1, studentId: 1, name: 'Dad', email: 'd@b.com', phone: '555', isPrimary: true, createdAt: '', updatedAt: '' }]
          : []
        return Promise.resolve(mockResponse({ code: 0, data: { items, page: 1, pageSize: 20, total: items.length } }))
      }
      // Enrollments
      if (url.startsWith('/students/1/enrollments')) {
        const items = opts.hasEnrollments
          ? [{ id: 10, studentId: 1, domainId: 1, trackId: 1, enrollmentType: 'REGULAR', status: 'ACTIVE', createdAt: '', updatedAt: '' }]
          : []
        return Promise.resolve(mockResponse({ code: 0, data: { items, page: 1, pageSize: 20, total: items.length } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
  }

  it('renders student detail with edit form, parents section, and enrollments section', async () => {
    setupFetch({ hasParents: true, hasEnrollments: true })
    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="student-detail-view"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="student-edit-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="student-parents-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="student-enrollments-section"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="parent-row"]')).toHaveLength(1)
    expect(wrapper.findAll('[data-testid="enrollment-row"]')).toHaveLength(1)
  })

  it('shows empty states when no parents and no enrollments', async () => {
    setupFetch({ hasParents: false, hasEnrollments: false })
    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    // Should have empty states for both parents and enrollments.
    const emptyStates = wrapper.findAll('[data-testid="state-empty"]')
    expect(emptyStates.length).toBeGreaterThanOrEqual(2)
  })

  it('saves student edit via PATCH', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url === '/students/1' && opts?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Updated', email: 'a@b.com', phone: '123', timezone: 'Asia/Tokyo', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url === '/students/1') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Alice', email: 'a@b.com', phone: '123', timezone: 'Asia/Tokyo', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/students/1/parents') || url.startsWith('/students/1/enrollments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    // Change the name.
    await wrapper.find('[data-testid="edit-student-name"]').setValue('Updated Name')
    await wrapper.find('[data-testid="student-edit-section"] form').trigger('submit.prevent')
    await flushPromises()

    const patch = calls.find((c) => c.url === '/students/1' && c.opts?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch.opts.body)
    expect(body.name).toBe('Updated Name')
  })

  it('creates a new parent', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url === '/students/1/parents' && opts?.method === 'POST') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, studentId: 1, name: 'Mom', isPrimary: false, createdAt: '', updatedAt: '' } }, 201))
      }
      if (url === '/students/1') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Alice', email: '', phone: '', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/students/1/parents')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/students/1/enrollments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-parent-btn"]').trigger('click')
    await wrapper.find('[data-testid="parent-form-name"]').setValue('Mom')
    await wrapper.find('[data-testid="parent-create-form"] form').trigger('submit.prevent')
    await flushPromises()

    const post = calls.find((c) => c.url === '/students/1/parents' && c.opts?.method === 'POST')
    expect(post).toBeDefined()
    const body = JSON.parse(post.opts.body)
    expect(body.name).toBe('Mom')
  })

  it('navigates to enrollment detail on view button click', async () => {
    setupFetch({ hasParents: false, hasEnrollments: true })
    const router = testRouter()
    const pushSpy = vi.spyOn(router, 'push')
    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), router] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="view-enrollment-btn"]').trigger('click')
    expect(pushSpy).toHaveBeenCalledWith({ name: 'enrollment-detail', params: { id: 10 } })
  })

  it('shows error state on API failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 500, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500),
    )

    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
  })

  it('navigates back to students list', async () => {
    setupFetch()
    const router = testRouter()
    const pushSpy = vi.spyOn(router, 'push')
    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), router] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="back-to-students"]').trigger('click')
    expect(pushSpy).toHaveBeenCalledWith({ name: 'students' })
  })

  it('shows save error on PATCH failure', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      if (url === '/students/1' && opts?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
      }
      if (url === '/students/1') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Alice', email: 'a@b.com', phone: '', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/students/1/parents') || url.startsWith('/students/1/enrollments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="edit-student-name"]').setValue('Updated')
    await wrapper.find('[data-testid="student-edit-section"] form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-testid="student-save-error"]').exists()).toBe(true)
  })

  it('shows parent create error on POST failure', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      if (url === '/students/1/parents' && opts?.method === 'POST') {
        return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
      }
      if (url === '/students/1') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Alice', email: '', phone: '', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/students/1/parents') || url.startsWith('/students/1/enrollments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-parent-btn"]').trigger('click')
    await wrapper.find('[data-testid="parent-form-name"]').setValue('Dup')
    await wrapper.find('[data-testid="parent-create-form"] form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-testid="parent-create-error"]').exists()).toBe(true)
  })

  it('cancels parent create form', async () => {
    setupFetch()
    const wrapper = mount(StudentDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-parent-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="parent-create-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="parent-create-cancel"]').trigger('click')
    expect(wrapper.find('[data-testid="parent-create-form"]').exists()).toBe(false)
  })
})
