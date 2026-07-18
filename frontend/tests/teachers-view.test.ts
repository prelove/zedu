import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import TeachersListView from '../src/features/directory/TeachersListView.vue'
import TeacherDetailView from '../src/features/directory/TeacherDetailView.vue'
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
      { path: '/teachers', name: 'teachers', component: { template: '<div />' } },
      { path: '/teachers/:id', name: 'teacher-detail', component: { template: '<div />' } },
    ],
  })
}

describe('TeachersListView', () => {
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

  it('renders teacher list with formatted JPY rate', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 0, data: { items: [
        { id: 1, name: 'Sensei', email: 's@b.com', phone: '', defaultRate: 3000, status: 'ACTIVE', createdAt: '', updatedAt: '' },
      ], page: 1, pageSize: 20, total: 1 } }),
    )

    const wrapper = mount(TeachersListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="teachers-table"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="teacher-row"]')).toHaveLength(1)
    // JPY 3000 should be formatted (contains the number with grouping).
    expect(wrapper.find('[data-testid="teacher-row"]').text()).toContain('3,000')
  })

  it('shows empty state when no teachers', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }),
    )

    const wrapper = mount(TeachersListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
  })

  it('opens create form and submits teacher', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (opts?.method === 'POST') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, name: 'New', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' } }, 201))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(TeachersListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="teachers-create-btn"]').trigger('click')
    await wrapper.find('[data-testid="teacher-form-name"]').setValue('New Teacher')
    await wrapper.find('[data-testid="teacher-form-rate"]').setValue('5000')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const post = calls.find((c) => c.opts?.method === 'POST')
    expect(post).toBeDefined()
    const body = JSON.parse(post.opts.body)
    expect(body.name).toBe('New Teacher')
    expect(body.defaultRate).toBe(5000)
  })

  it('shows error state on API failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 500, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500),
    )

    const wrapper = mount(TeachersListView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
  })

  it('navigates to teacher detail on view button click', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 0, data: { items: [
        { id: 42, name: 'Sensei', email: '', phone: '', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' },
      ], page: 1, pageSize: 20, total: 1 } }),
    )

    const router = testRouter()
    const pushSpy = vi.spyOn(router, 'push')
    const wrapper = mount(TeachersListView, { global: { plugins: [testI18n(), router] } })
    await flushPromises()

    await wrapper.find('[data-testid="teacher-view-btn"]').trigger('click')
    expect(pushSpy).toHaveBeenCalledWith({ name: 'teacher-detail', params: { id: 42 } })
  })
})

describe('TeacherDetailView', () => {
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

  function setupFetch(): void {
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.startsWith('/teachers/1')) {
        if (url.includes('/capabilities')) {
          return Promise.resolve(mockResponse({ code: 0, data: { items: [
            { id: 1, teacherId: 1, domainId: 1, trackId: 1, levelId: 1, status: 'ACTIVE', verified: false, createdAt: '', updatedAt: '' },
          ], page: 1, pageSize: 20, total: 1 } }))
        }
        if (url.includes('/availability')) {
          return Promise.resolve(mockResponse({ code: 0, data: { items: [
            { id: 1, teacherId: 1, weekday: 1, startTime: '09:00', endTime: '10:00', createdAt: '', updatedAt: '' },
          ], page: 1, pageSize: 20, total: 1 } }))
        }
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Sensei', defaultRate: 3000, status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'LANGUAGE', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'Nihongo', code: 'n1', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, trackId: 1, name: 'N5', code: 'n5', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
  }

  it('renders teacher detail with capabilities and availability', async () => {
    setupFetch()
    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="teacher-detail-view"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="teacher-edit-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="capabilities-table"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="availability-table"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="capability-row"]')).toHaveLength(1)
    expect(wrapper.findAll('[data-testid="availability-row"]')).toHaveLength(1)
  })

  it('shows capability end button only for ACTIVE capabilities', async () => {
    setupFetch()
    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    expect(wrapper.find('[data-testid="capability-end-btn"]').exists()).toBe(true)
  })

  it('client-side validates availability time (start < end)', async () => {
    setupFetch()
    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-availability-btn"]').trigger('click')
    await wrapper.find('[data-testid="avail-form-start"]').setValue('10:00')
    await wrapper.find('[data-testid="avail-form-end"]').setValue('09:00')
    await wrapper.find('[data-testid="availability-create-form"] form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-testid="availability-client-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="availability-client-error"]').text()).toContain(zhCN.teachers.availabilityInvalidTime)
  })

  it('shows capability duplicate hint on 40901', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      if (url.startsWith('/teachers/1')) {
        if (url.includes('/capabilities')) {
          if (opts?.method === 'POST') {
            return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
          }
          return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
        }
        if (url.includes('/availability')) {
          return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
        }
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'S', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'D', code: 'd', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, trackId: 1, name: 'L', code: 'l', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-capability-btn"]').trigger('click')
    await wrapper.find('[data-testid="cap-form-domain"]').setValue(1)
    await wrapper.find('[data-testid="cap-form-track"]').setValue(1)
    await wrapper.find('[data-testid="cap-form-level"]').setValue(1)
    await wrapper.find('[data-testid="capability-create-form"] form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-testid="capability-create-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="capability-duplicate-hint"]').exists()).toBe(true)
  })

  it('saves teacher edit via PATCH', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url === '/teachers/1' && opts?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Updated', defaultRate: 5000, status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/teachers/1')) {
        if (url.includes('/capabilities') || url.includes('/availability')) {
          return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
        }
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'Sensei', defaultRate: 3000, status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/course-domains') || url.startsWith('/tracks') || url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="edit-teacher-name"]').setValue('Updated Name')
    await wrapper.find('[data-testid="teacher-edit-section"] form').trigger('submit.prevent')
    await flushPromises()

    const patch = calls.find((c) => c.url === '/teachers/1' && c.opts?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch.opts.body)
    expect(body.name).toBe('Updated Name')
  })

  it('creates availability with valid time', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url.includes('/availability') && opts?.method === 'POST') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, teacherId: 1, weekday: 2, startTime: '10:00', endTime: '11:00', createdAt: '', updatedAt: '' } }, 201))
      }
      if (url.startsWith('/teachers/1')) {
        if (url.includes('/capabilities')) {
          return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
        }
        if (url.includes('/availability')) {
          return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
        }
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'S', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/course-domains') || url.startsWith('/tracks') || url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-availability-btn"]').trigger('click')
    await wrapper.find('[data-testid="avail-form-weekday"]').setValue(2)
    await wrapper.find('[data-testid="avail-form-start"]').setValue('10:00')
    await wrapper.find('[data-testid="avail-form-end"]').setValue('11:00')
    await wrapper.find('[data-testid="availability-create-form"] form').trigger('submit.prevent')
    await flushPromises()

    const post = calls.find((c) => c.url.includes('/availability') && c.opts?.method === 'POST')
    expect(post).toBeDefined()
    const body = JSON.parse(post.opts.body)
    expect(body.weekday).toBe(2)
    expect(body.startTime).toBe('10:00')
    expect(body.endTime).toBe('11:00')
  })

  it('ends capability via confirmation dialog', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (url.includes('/capabilities/') && opts?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, teacherId: 1, domainId: 1, trackId: 1, levelId: 1, status: 'ENDED', verified: false, createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/teachers/1')) {
        if (url.includes('/capabilities')) {
          return Promise.resolve(mockResponse({ code: 0, data: { items: [
            { id: 1, teacherId: 1, domainId: 1, trackId: 1, levelId: 1, status: 'ACTIVE', verified: false, createdAt: '', updatedAt: '' },
          ], page: 1, pageSize: 20, total: 1 } }))
        }
        if (url.includes('/availability')) {
          return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
        }
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'S', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'D', code: 'd', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, trackId: 1, name: 'L', code: 'l', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="capability-end-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(true)
    await wrapper.find('[data-testid="confirm-ok"]').trigger('click')
    await flushPromises()

    const patch = calls.find((c) => c.url.includes('/capabilities/') && c.opts?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch.opts.body)
    expect(body.status).toBe('ENDED')
  })

  it('validates availability weekday (1-7)', async () => {
    setupFetch()
    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-availability-btn"]').trigger('click')
    const vm = wrapper.vm as any
    // Set invalid weekday (0).
    vm.availForm = { weekday: 0, startTime: '09:00', endTime: '10:00' }
    await wrapper.find('[data-testid="availability-create-form"] form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-testid="availability-client-error"]').exists()).toBe(true)
  })

  it('shows teacher save error on PATCH failure', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      if (url === '/teachers/1' && opts?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
      }
      if (url.startsWith('/teachers/1')) {
        if (url.includes('/capabilities') || url.includes('/availability')) {
          return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
        }
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'S', defaultRate: 0, status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/course-domains') || url.startsWith('/tracks') || url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="edit-teacher-name"]').setValue('Updated')
    await wrapper.find('[data-testid="teacher-edit-section"] form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-testid="teacher-save-error"]').exists()).toBe(true)
  })

  it('cancels capability create form', async () => {
    setupFetch()
    const wrapper = mount(TeacherDetailView, {
      props: { id: '1' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()

    await wrapper.find('[data-testid="add-capability-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="capability-create-form"]').exists()).toBe(true)
    // Click cancel button inside the form.
    const cancelBtn = wrapper.find('[data-testid="capability-create-form"] button[type="button"]')
    await cancelBtn.trigger('click')
    expect(wrapper.find('[data-testid="capability-create-form"]').exists()).toBe(false)
  })
})
