import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import CourseDictionaryView from '../src/features/course/CourseDictionaryView.vue'
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

describe('CourseDictionaryView', () => {
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
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      if (opts?.method === 'POST') {
        if (url.startsWith('/course-domains')) {
          return Promise.resolve(mockResponse({ code: 0, data: { id: 3, name: 'New', code: 'new', type: 'LANGUAGE', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' } }, 201))
        }
        return Promise.resolve(mockResponse({ code: 0, data: { id: 3, name: 'New', code: 'new', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' } }, 201))
      }
      if (opts?.method === 'PATCH') {
        if (url.startsWith('/course-domains')) {
          return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'JP', code: 'jp', type: 'LANGUAGE', sortOrder: 0, enabled: false, createdAt: '', updatedAt: '' } }))
        }
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'X', code: 'x', sortOrder: 0, enabled: false, createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [
          { id: 1, name: 'JP', code: 'jp', type: 'LANGUAGE', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' },
        ], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [
          { id: 1, domainId: 1, name: 'Nihongo', code: 'n1', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' },
        ], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [
          { id: 1, trackId: 1, name: 'N5', code: 'n5', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' },
        ], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/capability-tags')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [
          { id: 1, domainId: 1, name: 'Tag1', code: 't1', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' },
        ], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
  }

  it('renders domains tab by default', async () => {
    setupFetch()
    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="course-dict-view"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="course-table"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="course-row"]')).toHaveLength(1)
  })

  it('switches to tracks tab and shows tracks', async () => {
    setupFetch()
    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="tab-tracks"]').trigger('click')
    await flushPromises()

    expect(wrapper.findAll('[data-testid="course-row"]')).toHaveLength(1)
  })

  it('switches to levels tab and shows levels', async () => {
    setupFetch()
    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="tab-levels"]').trigger('click')
    await flushPromises()

    expect(wrapper.findAll('[data-testid="course-row"]')).toHaveLength(1)
  })

  it('switches to tags tab and shows tags', async () => {
    setupFetch()
    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="tab-tags"]').trigger('click')
    await flushPromises()

    expect(wrapper.findAll('[data-testid="course-row"]')).toHaveLength(1)
  })

  it('creates a new domain via form', async () => {
    const calls: any[] = []
    setupFetch()
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (opts?.method === 'POST' && url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 3, name: 'New', code: 'new', type: 'LANGUAGE', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' } }, 201))
      }
      if (opts?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'JP', code: 'jp', type: 'LANGUAGE', sortOrder: 0, enabled: false, createdAt: '', updatedAt: '' } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'LANGUAGE', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="course-create-btn"]').trigger('click')
    await wrapper.find('[data-testid="d-form-name"]').setValue('New Domain')
    await wrapper.find('[data-testid="d-form-code"]').setValue('new')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const post = calls.find((c) => c.opts?.method === 'POST')
    expect(post).toBeDefined()
    const body = JSON.parse(post.opts.body)
    expect(body.name).toBe('New Domain')
    expect(body.code).toBe('new')
  })

  it('shows 42201 error when disabling referenced item', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      if (opts?.method === 'PATCH' && url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 42201, message: 'INVALID_STATE', requestId: 'r1' }, 422))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'LANGUAGE', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="course-toggle-1"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="course-form-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="course-referenced-hint"]').exists()).toBe(true)
  })

  it('toggles domain enabled status', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (opts?.method === 'PATCH' && url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, name: 'JP', code: 'jp', type: 'L', sortOrder: 0, enabled: false, createdAt: '', updatedAt: '' } }))
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
      if (url.startsWith('/capability-tags')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'Tag', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="course-toggle-1"]').trigger('click')
    await flushPromises()

    const patch = calls.find((c) => c.opts?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch.opts.body)
    expect(body.enabled).toBe(false)
  })

  it('creates a track via tracks tab form', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (opts?.method === 'POST' && url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, domainId: 1, name: 'NewTrack', code: 'nt', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' } }, 201))
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
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="tab-tracks"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="course-create-btn"]').trigger('click')
    await wrapper.find('[data-testid="t-form-name"]').setValue('NewTrack')
    await wrapper.find('[data-testid="t-form-code"]').setValue('nt')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const post = calls.find((c) => c.opts?.method === 'POST' && c.url.startsWith('/tracks'))
    expect(post).toBeDefined()
    const body = JSON.parse(post.opts.body)
    expect(body.name).toBe('NewTrack')
    expect(body.domainId).toBe(1)
  })

  it('creates a level via levels tab form', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (opts?.method === 'POST' && url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, trackId: 1, name: 'NewLevel', code: 'nl', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' } }, 201))
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
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="tab-levels"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="course-create-btn"]').trigger('click')
    await wrapper.find('[data-testid="l-form-name"]').setValue('NewLevel')
    await wrapper.find('[data-testid="l-form-code"]').setValue('nl')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const post = calls.find((c) => c.opts?.method === 'POST' && c.url.startsWith('/levels'))
    expect(post).toBeDefined()
    const body = JSON.parse(post.opts.body)
    expect(body.name).toBe('NewLevel')
    expect(body.trackId).toBe(1)
  })

  it('creates a tag via tags tab form', async () => {
    const calls: any[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: any) => {
      calls.push({ url, opts })
      if (opts?.method === 'POST' && url.startsWith('/capability-tags')) {
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, domainId: 1, name: 'NewTag', code: 'nt', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' } }, 201))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, name: 'JP', code: 'jp', type: 'L', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/capability-tags')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [{ id: 1, domainId: 1, name: 'Tag', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }], page: 1, pageSize: 20, total: 1 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="tab-tags"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="course-create-btn"]').trigger('click')
    await wrapper.find('[data-testid="g-form-name"]').setValue('NewTag')
    await wrapper.find('[data-testid="g-form-code"]').setValue('nt')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    const post = calls.find((c) => c.opts?.method === 'POST' && c.url.startsWith('/capability-tags'))
    expect(post).toBeDefined()
    const body = JSON.parse(post.opts.body)
    expect(body.name).toBe('NewTag')
    expect(body.domainId).toBe(1)
  })

  it('cancels create form', async () => {
    setupFetch()
    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    await wrapper.find('[data-testid="course-create-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="course-create-form"]').exists()).toBe(true)
    const cancelBtn = wrapper.find('[data-testid="course-create-form"] button[type="button"]')
    await cancelBtn.trigger('click')
    expect(wrapper.find('[data-testid="course-create-form"]').exists()).toBe(false)
  })

  it('shows empty state when no domains', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.startsWith('/course-domains') || url.startsWith('/tracks') || url.startsWith('/levels') || url.startsWith('/capability-tags')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
  })

  it('shows error state on API failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse({ code: 500, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500),
    )

    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
  })

  it('requests and renders the next page for the active tab', async () => {
    const calls: string[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calls.push(url)
      if (url.startsWith('/course-domains?page=2&pageSize=1')) {
        return Promise.resolve(mockResponse({
          code: 0,
          data: {
            items: [{ id: 2, name: 'EN', code: 'en', type: 'LANGUAGE', sortOrder: 1, enabled: true, createdAt: '', updatedAt: '' }],
            page: 2,
            pageSize: 1,
            total: 2,
          },
        }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({
          code: 0,
          data: {
            items: [{ id: 1, name: 'JP', code: 'jp', type: 'LANGUAGE', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }],
            page: 1,
            pageSize: 1,
            total: 2,
          },
        }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      if (url.startsWith('/capability-tags')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    const wrapper = mount(CourseDictionaryView, { global: { plugins: [testI18n(), testRouter()] } })
    await flushPromises()

    expect(wrapper.text()).toContain('JP')
    await wrapper.find('[data-testid="pagination-next"]').trigger('click')
    await flushPromises()

    expect(calls).toContain('/course-domains?page=2&pageSize=1')
    expect(wrapper.text()).toContain('EN')
    expect(wrapper.text()).not.toContain('JP')
  })
})
