import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import CapabilitiesSection from '../src/features/directory/components/CapabilitiesSection.vue'
import AvailabilitySection from '../src/features/directory/components/AvailabilitySection.vue'
import TeacherEditForm from '../src/features/directory/components/TeacherEditForm.vue'
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

const sampleDomains = [{ id: 1, name: 'JP', code: 'jp', type: 'L' as const, sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }]
const sampleTracks = [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }]
const sampleLevels = [
  { id: 1, trackId: 1, name: 'L1', code: 'l1', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' },
  { id: 2, trackId: 1, name: 'L2', code: 'l2', sortOrder: 1, enabled: true, createdAt: '', updatedAt: '' },
]

describe('CapabilitiesSection', () => {
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

  function setupFetch(opts: { capabilities?: any[]; error?: boolean } = {}): void {
    const caps = opts.capabilities ?? [
      { id: 1, teacherId: 1, domainId: 1, trackId: 1, levelId: 1, status: 'ACTIVE', verified: false, createdAt: '', updatedAt: '' },
    ]
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts2?: any) => {
      if (url.includes('/capabilities') && opts2?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { ...caps[0], status: 'ENDED' } }))
      }
      if (url.includes('/capabilities') && opts2?.method === 'POST') {
        if (opts.error) return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, teacherId: 1, domainId: 1, trackId: 1, levelId: 2, status: 'ACTIVE', verified: false, createdAt: '', updatedAt: '' } }))
      }
      if (url.includes('/capabilities')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: caps, page: 1, pageSize: 20, total: caps.length } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
  }

  it('renders capabilities table with rows', async () => {
    setupFetch()
    const wrapper = mount(CapabilitiesSection, {
      props: { teacherId: 1, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="capabilities-table"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="capability-row"]')).toHaveLength(1)
  })

  it('shows empty state when no capabilities', async () => {
    setupFetch({ capabilities: [] })
    const wrapper = mount(CapabilitiesSection, {
      props: { teacherId: 1, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
  })

  it('shows error state on load failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
    const wrapper = mount(CapabilitiesSection, {
      props: { teacherId: 1, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
  })

  it('shows dict error and disables add button', async () => {
    setupFetch()
    const wrapper = mount(CapabilitiesSection, {
      props: { teacherId: 1, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: 'errors.NETWORK_ERROR' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="add-capability-btn"]').attributes('disabled')).toBeDefined()
  })

  it('opens create form and creates capability', async () => {
    setupFetch()
    const wrapper = mount(CapabilitiesSection, {
      props: { teacherId: 1, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-capability-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="capability-create-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="cap-form-domain"]').setValue(1)
    await wrapper.find('[data-testid="cap-form-track"]').setValue(1)
    await wrapper.find('[data-testid="cap-form-level"]').setValue(2)
    await wrapper.find('[data-testid="capability-create-form"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="capability-create-form"]').exists()).toBe(false)
  })

  it('shows duplicate hint on 409 conflict', async () => {
    setupFetch({ error: true })
    const wrapper = mount(CapabilitiesSection, {
      props: { teacherId: 1, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
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

  it('cancels create form', async () => {
    setupFetch()
    const wrapper = mount(CapabilitiesSection, {
      props: { teacherId: 1, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-capability-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="capability-create-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="capability-create-form"] button[type="button"]').trigger('click')
    expect(wrapper.find('[data-testid="capability-create-form"]').exists()).toBe(false)
  })

  it('ends capability via confirmation dialog', async () => {
    setupFetch()
    const wrapper = mount(CapabilitiesSection, {
      props: { teacherId: 1, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="capability-end-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(true)
    await wrapper.find('[data-testid="confirm-ok"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(false)
  })

  it('cancels end capability confirmation', async () => {
    setupFetch()
    const wrapper = mount(CapabilitiesSection, {
      props: { teacherId: 1, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="capability-end-btn"]').trigger('click')
    await wrapper.find('[data-testid="confirm-cancel"]').trigger('click')
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(false)
  })
})

describe('AvailabilitySection', () => {
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

  function setupFetch(opts: { availability?: any[]; error?: boolean; postError?: boolean } = {}): void {
    const avails = opts.availability ?? [
      { id: 1, teacherId: 1, weekday: 1, startTime: '09:00', endTime: '10:00', createdAt: '', updatedAt: '' },
    ]
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts2?: any) => {
      if (url.includes('/availability') && opts2?.method === 'POST') {
        if (opts.postError) return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, teacherId: 1, weekday: 2, startTime: '10:00', endTime: '11:00', createdAt: '', updatedAt: '' } }))
      }
      if (url.includes('/availability') && opts2?.method === 'PATCH') {
        return Promise.resolve(mockResponse({ code: 0, data: { ...avails[0], weekday: 3 } }))
      }
      if (url.includes('/availability')) {
        if (opts.error) return Promise.resolve(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
        return Promise.resolve(mockResponse({ code: 0, data: { items: avails, page: 1, pageSize: 20, total: avails.length } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
  }

  it('renders availability table with rows', async () => {
    setupFetch()
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="availability-table"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="availability-row"]')).toHaveLength(1)
  })

  it('shows empty state when no availability', async () => {
    setupFetch({ availability: [] })
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
  })

  it('shows error state on load failure', async () => {
    setupFetch({ error: true })
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
    await wrapper.find('[data-testid="state-error-retry"]').trigger('click')
  })

  it('opens create form and creates availability', async () => {
    setupFetch()
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-availability-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="availability-create-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="avail-form-weekday"]').setValue(2)
    await wrapper.find('[data-testid="avail-form-start"]').setValue('10:00')
    await wrapper.find('[data-testid="avail-form-end"]').setValue('11:00')
    await wrapper.find('[data-testid="availability-create-form"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="availability-create-form"]').exists()).toBe(false)
  })

  it('shows client error for invalid time (start >= end)', async () => {
    setupFetch()
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-availability-btn"]').trigger('click')
    await wrapper.find('[data-testid="avail-form-start"]').setValue('11:00')
    await wrapper.find('[data-testid="avail-form-end"]').setValue('10:00')
    await wrapper.find('[data-testid="availability-create-form"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="availability-client-error"]').exists()).toBe(true)
  })

  it('shows create error on POST failure', async () => {
    setupFetch({ postError: true })
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-availability-btn"]').trigger('click')
    await wrapper.find('[data-testid="avail-form-weekday"]').setValue(2)
    await wrapper.find('[data-testid="avail-form-start"]').setValue('10:00')
    await wrapper.find('[data-testid="avail-form-end"]').setValue('11:00')
    await wrapper.find('[data-testid="availability-create-form"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="availability-create-error"]').exists()).toBe(true)
  })

  it('cancels create form', async () => {
    setupFetch()
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-availability-btn"]').trigger('click')
    await wrapper.find('[data-testid="availability-create-form"] button[type="button"]').trigger('click')
    expect(wrapper.find('[data-testid="availability-create-form"]').exists()).toBe(false)
  })

  it('opens edit form and saves changes', async () => {
    setupFetch()
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="availability-edit-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="availability-edit-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="avail-edit-weekday"]').setValue(3)
    await wrapper.find('[data-testid="availability-edit-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="availability-edit-form"]').exists()).toBe(false)
  })

  it('shows no-changes message when edit has no changes', async () => {
    setupFetch()
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="availability-edit-btn"]').trigger('click')
    await wrapper.find('[data-testid="availability-edit-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="avail-edit-no-changes"]').exists()).toBe(true)
  })

  it('shows client error in edit form for invalid time', async () => {
    setupFetch()
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="availability-edit-btn"]').trigger('click')
    await wrapper.find('[data-testid="avail-edit-start-time"]').setValue('10:00')
    await wrapper.find('[data-testid="avail-edit-end-time"]').setValue('09:00')
    await wrapper.find('[data-testid="availability-edit-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="availability-client-error"]').exists()).toBe(true)
  })

  it('cancels edit form', async () => {
    setupFetch()
    const wrapper = mount(AvailabilitySection, {
      props: { teacherId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="availability-edit-btn"]').trigger('click')
    await wrapper.find('[data-testid="avail-edit-cancel"]').trigger('click')
    expect(wrapper.find('[data-testid="availability-edit-form"]').exists()).toBe(false)
  })
})

describe('TeacherEditForm', () => {
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

  const sampleTeacher = {
    id: 1, name: 'Sensei', nameLocal: '先生', email: 's@x.com', phone: '123',
    defaultRate: 3000, status: 'ACTIVE' as const, bio: 'bio', note: 'note',
    createdAt: '', updatedAt: '',
  }

  it('renders edit form with teacher data', async () => {
    const wrapper = mount(TeacherEditForm, {
      props: { teacher: sampleTeacher },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="teacher-edit-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="edit-teacher-name"]').element.value).toBe('Sensei')
  })

  it('shows no-changes when submitting without changes', async () => {
    const wrapper = mount(TeacherEditForm, {
      props: { teacher: sampleTeacher },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="teacher-edit-section"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="teacher-no-changes"]').exists()).toBe(true)
  })

  it('saves teacher edit via PATCH', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 0, data: { ...sampleTeacher, name: 'Updated' } }))
    const wrapper = mount(TeacherEditForm, {
      props: { teacher: sampleTeacher },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="edit-teacher-name"]').setValue('Updated')
    await wrapper.find('[data-testid="teacher-edit-section"] form').trigger('submit.prevent')
    await flushPromises()
    const calls = (globalThis.fetch as any).mock.calls
    const patch = calls.find((c: any[]) => c[0] === '/teachers/1' && c[1]?.method === 'PATCH')
    expect(patch).toBeDefined()
  })

  it('shows save error on PATCH failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
    const wrapper = mount(TeacherEditForm, {
      props: { teacher: sampleTeacher },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="edit-teacher-name"]').setValue('Updated')
    await wrapper.find('[data-testid="teacher-edit-section"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="teacher-save-error"]').exists()).toBe(true)
  })
})
