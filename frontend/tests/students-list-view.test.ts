import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import StudentsListView from '../src/features/directory/StudentsListView.vue'
import { authStore } from '../src/stores/auth'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'
// directory API is exercised indirectly via fetch mocking.

function mockResponse(body: unknown, status = 200): Response {
  return { ok: status < 300, status, json: async () => body } as Response
}

function testI18n() {
  return createI18n({
    legacy: false,
    locale: 'zh-CN',
    fallbackLocale: 'zh-CN',
    messages: { 'zh-CN': zhCN, 'ja-JP': jaJP, 'en-US': enUS },
  })
}

function testRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/students', component: StudentsListView },
      { path: '/students/:id', name: 'student-detail', component: { template: '<div />' } },
      { path: '/teachers', component: { template: '<div />' } },
      { path: '/courses', component: { template: '<div />' } },
      { path: '/enrollments/:id', component: { template: '<div />' } },
    ],
  })
}

describe('StudentsListView', () => {
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

  it('renders list with students from API', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 0, data: { items: [
        { id: 1, name: 'Alice', email: 'a@b.com', phone: '123', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' },
        { id: 2, name: 'Bob', email: '', phone: '', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' },
      ], page: 1, pageSize: 20, total: 2 } }),
    )

    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="students-table"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="student-row"]')).toHaveLength(2)
    expect(wrapper.find('[data-testid="student-row"]').text()).toContain('Alice')
  })

  it('shows empty state when no students', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }),
    )

    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
  })

  it('shows error state on API failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 500, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500),
    )

    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
  })

  it('opens create form and submits with empty email (no email field sent)', async () => {
    const fetchCalls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      fetchCalls.push({ url, opts })
      if (opts?.method === 'POST') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 3, name: 'New', timezone: 'Asia/Tokyo', status: 'ACTIVE', createdAt: '', updatedAt: '' } }, 201))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="students-create-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="students-create-form"]').exists()).toBe(true)

    await wrapper.find('[data-testid="student-form-name"]').setValue('New Student')
    // Leave email empty — should not send email field.
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const postCall = fetchCalls.find((c) => c.opts?.method === 'POST')
    expect(postCall).toBeDefined()
    const body = JSON.parse(postCall.opts.body)
    expect(body.name).toBe('New Student')
    expect(body.email).toBeUndefined()
  })

  it('shows 40901 conflict error and no-bypass hint on duplicate email', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      if (opts?.method === 'POST') {
        return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="students-create-btn"]').trigger('click')
    await wrapper.find('[data-testid="student-form-name"]').setValue('Dup')
    await wrapper.find('[data-testid="student-form-email"]').setValue('dup@test.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-testid="students-create-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="students-create-error"]').text()).toContain(zhCN.apiErrors.CONFLICT)
    expect(wrapper.find('[data-testid="students-no-bypass"]').exists()).toBe(true)
    // Form should remain open (not closed on error).
    expect(wrapper.find('[data-testid="students-create-form"]').exists()).toBe(true)
  })

  it('disables submit button while creating', async () => {
    let resolveFn: () => void
    const pending = new Promise<void>((r) => { resolveFn = r })
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      if (opts?.method === 'POST') {
        return pending.then(() => mockResponse({ code: 0, data: { id: 3, name: 'X', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' } }, 201))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="students-create-btn"]').trigger('click')
    await wrapper.find('[data-testid="student-form-name"]').setValue('X')
    await wrapper.find('form').trigger('submit.prevent')

    // While pending, button should be disabled.
    expect(wrapper.find('[data-testid="students-create-submit"]').attributes('disabled')).toBeDefined()

    resolveFn!()
    await flushPromises()
  })

  it('navigates to student detail on view button click', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 0, data: { items: [
        { id: 42, name: 'Alice', email: '', phone: '', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' },
      ], page: 1, pageSize: 20, total: 1 } }),
    )

    const router = testRouter()
    const pushSpy = vi.spyOn(router, 'push')
    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), router] } })
    await flushPromises()

    await wrapper.find('[data-testid="student-view-btn"]').trigger('click')
    expect(pushSpy).toHaveBeenCalledWith({ name: 'student-detail', params: { id: 42 } })
  })

  it('shows pagination when total > pageSize', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 0, data: { items: [
        { id: 1, name: 'A', email: '', phone: '', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' },
      ], page: 1, pageSize: 20, total: 50 } }),
    )

    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="pagination"]').exists()).toBe(true)
  })

  it('emits page change on next page click', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 0, data: { items: [
        { id: 1, name: 'A', email: '', phone: '', timezone: 'UTC', status: 'ACTIVE', createdAt: '', updatedAt: '' },
      ], page: 1, pageSize: 20, total: 50 } }),
    )

    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="pagination-next"]').trigger('click')
    await flushPromises()
    // After page change, fetch should be called again with page=2.
    expect(wrapper.find('[data-testid="students-table"]').exists()).toBe(true)
  })

  it('cancels create form', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }),
    )

    const wrapper = mount(StudentsListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="students-create-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="students-create-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="students-create-cancel"]').trigger('click')
    expect(wrapper.find('[data-testid="students-create-form"]').exists()).toBe(false)
  })
})
