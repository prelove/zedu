import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import StudentEditForm from '../src/features/directory/components/StudentEditForm.vue'
import ParentsSection from '../src/features/directory/components/ParentsSection.vue'
import EnrollmentsSection from '../src/features/directory/components/EnrollmentsSection.vue'
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

const sampleStudent = {
  id: 1, name: 'Taro', nameLocal: '太郎', email: 't@x.com', phone: '123',
  nationality: 'JP', timezone: 'Asia/Tokyo', status: 'ACTIVE' as const,
  sourceChannel: 'web', note: 'note', createdAt: '', updatedAt: '',
}

const sampleDomains = [{ id: 1, name: 'JP', code: 'jp', type: 'L' as const, sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }]
const sampleTracks = [{ id: 1, domainId: 1, name: 'T', code: 't', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' }]
const sampleLevels = [
  { id: 1, trackId: 1, name: 'L1', code: 'l1', sortOrder: 0, enabled: true, createdAt: '', updatedAt: '' },
  { id: 2, trackId: 1, name: 'L2', code: 'l2', sortOrder: 1, enabled: true, createdAt: '', updatedAt: '' },
]

describe('StudentEditForm', () => {
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

  it('renders edit form with student data', () => {
    const wrapper = mount(StudentEditForm, {
      props: { student: sampleStudent },
      global: { plugins: [testI18n(), testRouter()] },
    })
    expect(wrapper.find('[data-testid="student-edit-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="edit-student-name"]').element.value).toBe('Taro')
  })

  it('shows no-changes when submitting without changes', async () => {
    const wrapper = mount(StudentEditForm, {
      props: { student: sampleStudent },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="student-edit-section"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="student-no-changes"]').exists()).toBe(true)
  })

  it('saves student edit via PATCH and emits saved', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 0, data: { ...sampleStudent, name: 'Updated' } }))
    const wrapper = mount(StudentEditForm, {
      props: { student: sampleStudent },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="edit-student-name"]').setValue('Updated')
    await wrapper.find('[data-testid="student-edit-section"] form').trigger('submit.prevent')
    await flushPromises()
    const calls = (globalThis.fetch as any).mock.calls
    const patch = calls.find((c: any[]) => c[0] === '/students/1' && c[1]?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch[1].body)
    expect(body.name).toBe('Updated')
    expect(wrapper.emitted('saved')).toBeTruthy()
  })

  it('shows save error on PATCH failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
    const wrapper = mount(StudentEditForm, {
      props: { student: sampleStudent },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="edit-student-name"]').setValue('Updated')
    await wrapper.find('[data-testid="student-edit-section"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="student-save-error"]').exists()).toBe(true)
  })

  it('shows email bypass warning on CONFLICT error', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
    const wrapper = mount(StudentEditForm, {
      props: { student: sampleStudent },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="edit-student-name"]').setValue('Updated')
    await wrapper.find('[data-testid="student-edit-section"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="student-save-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="student-email-no-bypass"]').exists()).toBe(true)
  })
})

describe('ParentsSection', () => {
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

  function setupFetch(opts: { parents?: any[]; error?: boolean; postError?: boolean; patchError?: boolean } = {}): void {
    const parents = opts.parents ?? [
      { id: 1, studentId: 1, name: 'Parent', email: 'p@x.com', phone: '456', relationship: 'father', isPrimary: true, createdAt: '', updatedAt: '' },
    ]
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts2?: any) => {
      if (opts2?.method === 'PATCH' && url.includes('/parents/')) {
        if (opts.patchError) return Promise.resolve(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
        return Promise.resolve(mockResponse({ code: 0, data: { ...parents[0], name: 'Updated Parent' } }))
      }
      if (opts2?.method === 'POST' && url.includes('/parents')) {
        if (opts.postError) return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, studentId: 1, name: 'New Parent', email: 'n@x.com', phone: '789', relationship: 'mother', isPrimary: false, createdAt: '', updatedAt: '' } }))
      }
      if (url.includes('/parents')) {
        if (opts.error) return Promise.resolve(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
        return Promise.resolve(mockResponse({ code: 0, data: { items: parents, page: 1, pageSize: 20, total: parents.length } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
  }

  it('renders parents table with rows', async () => {
    setupFetch()
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="parents-table"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="parent-row"]')).toHaveLength(1)
  })

  it('shows empty state when no parents', async () => {
    setupFetch({ parents: [] })
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
  })

  it('shows error state on load failure', async () => {
    setupFetch({ error: true })
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
  })

  it('opens create form and creates parent', async () => {
    setupFetch()
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-parent-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="parent-create-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="parent-form-name"]').setValue('New Parent')
    await wrapper.find('[data-testid="parent-create-form"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="parent-create-form"]').exists()).toBe(false)
  })

  it('shows create error on POST failure', async () => {
    setupFetch({ postError: true })
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-parent-btn"]').trigger('click')
    await wrapper.find('[data-testid="parent-form-name"]').setValue('New Parent')
    await wrapper.find('[data-testid="parent-create-form"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="parent-create-error"]').exists()).toBe(true)
  })

  it('cancels create form', async () => {
    setupFetch()
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-parent-btn"]').trigger('click')
    await wrapper.find('[data-testid="parent-create-cancel"]').trigger('click')
    expect(wrapper.find('[data-testid="parent-create-form"]').exists()).toBe(false)
  })

  it('opens edit form and saves changes', async () => {
    setupFetch()
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="parent-edit-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="parent-edit-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="parent-edit-name"]').setValue('Updated Parent')
    await wrapper.find('[data-testid="parent-edit-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="parent-edit-form"]').exists()).toBe(false)
  })

  it('shows no-changes when edit has no changes', async () => {
    setupFetch()
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="parent-edit-btn"]').trigger('click')
    await wrapper.find('[data-testid="parent-edit-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="parent-edit-no-changes"]').exists()).toBe(true)
  })

  it('cancels edit form', async () => {
    setupFetch()
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="parent-edit-btn"]').trigger('click')
    await wrapper.find('[data-testid="parent-edit-cancel"]').trigger('click')
    expect(wrapper.find('[data-testid="parent-edit-form"]').exists()).toBe(false)
  })

  it('shows edit error on PATCH failure', async () => {
    setupFetch({ patchError: true })
    const wrapper = mount(ParentsSection, {
      props: { studentId: 1 },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="parent-edit-btn"]').trigger('click')
    await wrapper.find('[data-testid="parent-edit-name"]').setValue('Updated')
    await wrapper.find('[data-testid="parent-edit-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="parent-edit-error"]').exists()).toBe(true)
  })
})

describe('EnrollmentsSection', () => {
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

  function setupFetch(opts: { enrollments?: any[]; error?: boolean; postError?: boolean } = {}): void {
    const enrollments = opts.enrollments ?? [
      { id: 5, studentId: 1, domainId: 1, trackId: 1, currentLevelId: 1, targetLevelId: 2, enrollmentType: 'R', status: 'ACTIVE', createdAt: '', updatedAt: '' },
    ]
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts2?: any) => {
      if (opts2?.method === 'POST' && url.includes('/enrollments')) {
        if (opts.postError) return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
        return Promise.resolve(mockResponse({ code: 0, data: { id: 6, studentId: 1, domainId: 1, trackId: 1, currentLevelId: 0, targetLevelId: 1, enrollmentType: 'T', status: 'ACTIVE', createdAt: '', updatedAt: '' } }))
      }
      if (url.includes('/enrollments') && (!opts2?.method || opts2.method === 'GET')) {
        if (opts.error) return Promise.resolve(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
        return Promise.resolve(mockResponse({ code: 0, data: { items: enrollments, page: 1, pageSize: 20, total: enrollments.length } }))
      }
      if (url.startsWith('/course-domains')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: sampleDomains, page: 1, pageSize: 100, total: 1 } }))
      }
      if (url.startsWith('/tracks')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: sampleTracks, page: 1, pageSize: 100, total: 1 } }))
      }
      if (url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: sampleLevels, page: 1, pageSize: 100, total: 2 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
  }

  it('renders enrollments table with rows', async () => {
    setupFetch()
    const wrapper = mount(EnrollmentsSection, {
      props: { studentId: 1, studentStatus: 'ACTIVE', domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="enrollments-table"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="enrollment-row"]')).toHaveLength(1)
    expect(wrapper.find('[data-testid="view-enrollment-btn"]').exists()).toBe(true)
  })

  it('shows empty state when no enrollments', async () => {
    setupFetch({ enrollments: [] })
    const wrapper = mount(EnrollmentsSection, {
      props: { studentId: 1, studentStatus: 'ACTIVE', domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
  })

  it('shows error state on load failure', async () => {
    setupFetch({ error: true })
    const wrapper = mount(EnrollmentsSection, {
      props: { studentId: 1, studentStatus: 'ACTIVE', domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
  })

  it('shows dict error and disables add button when dictionary fails', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.startsWith('/course-domains') || url.startsWith('/tracks') || url.startsWith('/levels')) {
        return Promise.resolve(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
      }
      if (url.includes('/enrollments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
    const wrapper = mount(EnrollmentsSection, {
      props: { studentId: 1, studentStatus: 'ACTIVE' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="enrollment-dict-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="add-enrollment-btn"]').attributes('disabled')).toBeDefined()
  })

  it('retries dictionary load when retry button clicked', async () => {
    let dictFailed = true
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.startsWith('/course-domains') || url.startsWith('/tracks') || url.startsWith('/levels')) {
        if (dictFailed) return Promise.resolve(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 100, total: 0 } }))
      }
      if (url.includes('/enrollments')) {
        return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
    const wrapper = mount(EnrollmentsSection, {
      props: { studentId: 1, studentStatus: 'ACTIVE' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="enrollment-dict-error"]').exists()).toBe(true)
    dictFailed = false
    await wrapper.find('[data-testid="enrollment-dict-retry"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="enrollment-dict-error"]').exists()).toBe(false)
  })

  it('opens create form and creates enrollment', async () => {
    setupFetch({ enrollments: [] })
    const wrapper = mount(EnrollmentsSection, {
      props: { studentId: 1, studentStatus: 'ACTIVE', domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-enrollment-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="enrollment-create-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="enrollment-form-domain"]').setValue(1)
    await wrapper.find('[data-testid="enrollment-form-track"]').setValue(1)
    await wrapper.find('[data-testid="enrollment-form-target-level"]').setValue(2)
    await wrapper.find('[data-testid="enrollment-form-type"]').setValue('T')
    await wrapper.find('[data-testid="enrollment-create-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="enrollment-create-form"]').exists()).toBe(false)
  })

  it('shows create error on POST failure', async () => {
    setupFetch({ enrollments: [], postError: true })
    const wrapper = mount(EnrollmentsSection, {
      props: { studentId: 1, studentStatus: 'ACTIVE', domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-enrollment-btn"]').trigger('click')
    await wrapper.find('[data-testid="enrollment-form-domain"]').setValue(1)
    await wrapper.find('[data-testid="enrollment-form-track"]').setValue(1)
    await wrapper.find('[data-testid="enrollment-form-target-level"]').setValue(2)
    await wrapper.find('[data-testid="enrollment-form-type"]').setValue('T')
    await wrapper.find('[data-testid="enrollment-create-form"]').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="enrollment-create-error"]').exists()).toBe(true)
  })

  it('cancels create form', async () => {
    setupFetch({ enrollments: [] })
    const wrapper = mount(EnrollmentsSection, {
      props: { studentId: 1, studentStatus: 'ACTIVE', domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-enrollment-btn"]').trigger('click')
    await wrapper.find('[data-testid="enrollment-create-cancel"]').trigger('click')
    expect(wrapper.find('[data-testid="enrollment-create-form"]').exists()).toBe(false)
  })

  it('emits view when view button clicked', async () => {
    setupFetch()
    const wrapper = mount(EnrollmentsSection, {
      props: { studentId: 1, studentStatus: 'ACTIVE', domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="view-enrollment-btn"]').trigger('click')
    expect(wrapper.emitted('viewEnrollment')).toBeTruthy()
  })
})




