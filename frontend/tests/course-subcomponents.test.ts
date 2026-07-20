import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import CourseSelectionSection from '../src/features/course/components/CourseSelectionSection.vue'
import LevelChangeSection from '../src/features/course/components/LevelChangeSection.vue'
import AssignmentsSection from '../src/features/course/components/AssignmentsSection.vue'
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

const sampleEnrollment = {
  id: 5, studentId: 10, domainId: 1, trackId: 1, currentLevelId: 1, targetLevelId: 2,
  enrollmentType: 'R' as const, status: 'ACTIVE' as const, createdAt: '', updatedAt: '',
}

const sampleTeachers = [
  { id: 3, name: 'Sensei', defaultRate: 3000, status: 'ACTIVE' as const, createdAt: '', updatedAt: '' },
]

describe('CourseSelectionSection', () => {
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

  it('renders course selection section with form fields', () => {
    const wrapper = mount(CourseSelectionSection, {
      props: { enrollment: sampleEnrollment, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    expect(wrapper.find('[data-testid="course-selection-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="sel-domain"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="sel-track"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="sel-target-level"]').exists()).toBe(true)
  })

  it('shows dict error and disables save button', () => {
    const wrapper = mount(CourseSelectionSection, {
      props: { enrollment: sampleEnrollment, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: 'errors.NETWORK_ERROR' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    expect(wrapper.find('[data-testid="course-dict-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="save-course-selection"]').attributes('disabled')).toBeDefined()
  })

  it('emits retryDict when retry button clicked', async () => {
    const wrapper = mount(CourseSelectionSection, {
      props: { enrollment: sampleEnrollment, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: 'errors.NETWORK_ERROR' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="course-dict-retry"]').trigger('click')
    expect(wrapper.emitted('retryDict')).toBeTruthy()
  })

  it('saves course selection via PATCH and emits saved', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 0, data: { ...sampleEnrollment, domainId: 1, trackId: 1, targetLevelId: 2 } }))
    const wrapper = mount(CourseSelectionSection, {
      props: { enrollment: sampleEnrollment, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="save-course-selection"]').trigger('click')
    await flushPromises()
    const calls = (globalThis.fetch as any).mock.calls
    const patch = calls.find((c: any[]) => c[0] === '/enrollments/5' && c[1]?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch[1].body)
    expect(body.domainId).toBe(1)
    expect(body.trackId).toBe(1)
    expect(body.targetLevelId).toBe(2)
    expect(body.currentLevelId).toBeUndefined()
    expect(wrapper.emitted('saved')).toBeTruthy()
  })

  it('shows save error on PATCH failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
    const wrapper = mount(CourseSelectionSection, {
      props: { enrollment: sampleEnrollment, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="save-course-selection"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="course-selection-error"]').exists()).toBe(true)
  })

  it('resets track and target level when domain changes', async () => {
    const wrapper = mount(CourseSelectionSection, {
      props: { enrollment: sampleEnrollment, domains: sampleDomains, tracks: sampleTracks, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    const trackSelect = wrapper.find('[data-testid="sel-track"]')
    await trackSelect.setValue(1)
    await wrapper.find('[data-testid="sel-domain"]').trigger('change')
    // After domain change, track should reset to 0
    expect((trackSelect.element as HTMLSelectElement).value).toBe('0')
  })
})

describe('LevelChangeSection', () => {
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

  it('renders level change section', () => {
    const wrapper = mount(LevelChangeSection, {
      props: { enrollment: sampleEnrollment, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    expect(wrapper.find('[data-testid="level-change-section"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="sel-current-level"]').exists()).toBe(true)
  })

  it('shows dict error and disables save button', () => {
    const wrapper = mount(LevelChangeSection, {
      props: { enrollment: sampleEnrollment, levels: sampleLevels, dictError: 'errors.NETWORK_ERROR' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    expect(wrapper.find('[data-testid="level-dict-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="save-level-change"]').attributes('disabled')).toBeDefined()
  })

  it('emits retryDict when retry button clicked', async () => {
    const wrapper = mount(LevelChangeSection, {
      props: { enrollment: sampleEnrollment, levels: sampleLevels, dictError: 'errors.NETWORK_ERROR' },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="level-dict-retry"]').trigger('click')
    expect(wrapper.emitted('retryDict')).toBeTruthy()
  })

  it('rejects same level change client-side', async () => {
    const wrapper = mount(LevelChangeSection, {
      props: { enrollment: sampleEnrollment, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    // currentLevelId is 1, selCurrentLevel is 1 — same level
    await wrapper.find('[data-testid="save-level-change"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="level-change-error"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="level-change-error"]').text()).toContain(zhCN.enrollments.sameLevelRejected)
  })

  it('saves level change via PATCH and emits saved', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 0, data: { ...sampleEnrollment, currentLevelId: 2 } }))
    const wrapper = mount(LevelChangeSection, {
      props: { enrollment: sampleEnrollment, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="sel-current-level"]').setValue(2)
    await wrapper.find('[data-testid="save-level-change"]').trigger('click')
    await flushPromises()
    const calls = (globalThis.fetch as any).mock.calls
    const patch = calls.find((c: any[]) => c[0] === '/enrollments/5' && c[1]?.method === 'PATCH')
    expect(patch).toBeDefined()
    const body = JSON.parse(patch[1].body)
    expect(Number(body.currentLevelId)).toBe(2)
    expect(body.domainId).toBeUndefined()
    expect(body.trackId).toBeUndefined()
    expect(body.targetLevelId).toBeUndefined()
    expect(wrapper.emitted('saved')).toBeTruthy()
  })

  it('rejects the same level after a successful transition without a second PATCH', async () => {
    // The backend deliberately keeps enrollment.currentLevelId as its initial
    // snapshot and records later changes as level events. The UI must retain
    // the effective level for this open page rather than trust that snapshot.
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 0, data: sampleEnrollment }))
    const wrapper = mount(LevelChangeSection, {
      props: { enrollment: sampleEnrollment, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })

    await wrapper.find('[data-testid="sel-current-level"]').setValue(2)
    await wrapper.find('[data-testid="save-level-change"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="save-level-change"]').trigger('click')
    await flushPromises()

    const patches = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls
      .filter(([url, options]) => url === '/enrollments/5' && options?.method === 'PATCH')
    expect(patches).toHaveLength(1)
    expect(wrapper.find('[data-testid="level-change-error"]').text())
      .toContain(zhCN.enrollments.sameLevelRejected)
  })

  it('shows save error on PATCH failure', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
    const wrapper = mount(LevelChangeSection, {
      props: { enrollment: sampleEnrollment, levels: sampleLevels, dictError: null },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await wrapper.find('[data-testid="sel-current-level"]').setValue(2)
    await wrapper.find('[data-testid="save-level-change"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="level-change-error"]').exists()).toBe(true)
  })
})

describe('AssignmentsSection', () => {
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

  function setupFetch(opts: { assignments?: any[]; error?: boolean; postError?: boolean; endDuplicate?: boolean } = {}): void {
    const assigns = opts.assignments ?? []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts2?: any) => {
      if (url.includes('/assignments/') && url.includes('/end')) {
        if (opts.endDuplicate) return Promise.resolve(mockResponse({ code: 42201, message: 'INVALID_STATE', requestId: 'r1' }, 422))
        return Promise.resolve(mockResponse({ code: 0, data: { id: 1, enrollmentId: 5, teacherId: 3, roleType: 'MAIN', status: 'ENDED', startDate: '', createdAt: '', updatedAt: '' } }))
      }
      if (url.includes('/assignments') && opts2?.method === 'POST') {
        if (opts.postError) return Promise.resolve(mockResponse({ code: 40901, message: 'CONFLICT', requestId: 'r1' }, 409))
        return Promise.resolve(mockResponse({ code: 0, data: { id: 2, enrollmentId: 5, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '', createdAt: '', updatedAt: '' } }))
      }
      if (url.includes('/assignments')) {
        if (opts.error) return Promise.resolve(mockResponse({ code: 50001, message: 'INTERNAL_ERROR', requestId: 'r1' }, 500))
        return Promise.resolve(mockResponse({ code: 0, data: { items: assigns, page: 1, pageSize: 20, total: assigns.length } }))
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })
  }

  it('shows empty state when no assignments', async () => {
    setupFetch({ assignments: [] })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
  })

  it('shows error state on load failure', async () => {
    setupFetch({ error: true })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
  })

  it('renders assignments table with rows', async () => {
    setupFetch({
      assignments: [
        { id: 1, enrollmentId: 5, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '2024-01-01', endDate: null, createdAt: '', updatedAt: '' },
      ],
    })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="assignments-table"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="assignment-row"]')).toHaveLength(1)
    expect(wrapper.find('[data-testid="end-assignment-btn"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="atomic-replace-hint"]').exists()).toBe(true)
  })

  it('shows ended assignment without end button', async () => {
    setupFetch({
      assignments: [
        { id: 1, enrollmentId: 5, teacherId: 3, roleType: 'MAIN', status: 'ENDED', startDate: '2024-01-01', endDate: '2024-02-01', createdAt: '', updatedAt: '' },
      ],
    })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="end-assignment-btn"]').exists()).toBe(false)
  })

  it('opens create form and creates assignment', async () => {
    setupFetch({ assignments: [] })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-assignment-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="assignment-create-form"]').exists()).toBe(true)
    await wrapper.find('[data-testid="assign-form-teacher"]').setValue(3)
    await wrapper.find('[data-testid="assignment-create-form"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="assignment-create-form"]').exists()).toBe(false)
  })

  it('shows create error on POST failure', async () => {
    setupFetch({ assignments: [], postError: true })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-assignment-btn"]').trigger('click')
    await wrapper.find('[data-testid="assign-form-teacher"]').setValue(3)
    await wrapper.find('[data-testid="assignment-create-form"] form').trigger('submit.prevent')
    await flushPromises()
    expect(wrapper.find('[data-testid="assignment-create-error"]').exists()).toBe(true)
  })

  it('cancels create form', async () => {
    setupFetch({ assignments: [] })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="add-assignment-btn"]').trigger('click')
    await wrapper.find('[data-testid="assignment-create-form"] button[type="button"]').trigger('click')
    expect(wrapper.find('[data-testid="assignment-create-form"]').exists()).toBe(false)
  })

  it('ends assignment via confirmation dialog', async () => {
    setupFetch({
      assignments: [
        { id: 1, enrollmentId: 5, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '2024-01-01', endDate: null, createdAt: '', updatedAt: '' },
      ],
    })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="end-assignment-btn"]').trigger('click')
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(true)
    await wrapper.find('[data-testid="confirm-ok"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(false)
  })

  it('cancels end assignment confirmation', async () => {
    setupFetch({
      assignments: [
        { id: 1, enrollmentId: 5, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '2024-01-01', endDate: null, createdAt: '', updatedAt: '' },
      ],
    })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="end-assignment-btn"]').trigger('click')
    await wrapper.find('[data-testid="confirm-cancel"]').trigger('click')
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(false)
  })

  it('shows end duplicate error on 42201', async () => {
    setupFetch({
      assignments: [
        { id: 1, enrollmentId: 5, teacherId: 3, roleType: 'MAIN', status: 'ACTIVE', startDate: '2024-01-01', endDate: null, createdAt: '', updatedAt: '' },
      ],
      endDuplicate: true,
    })
    const wrapper = mount(AssignmentsSection, {
      props: { enrollmentId: 5, teachers: sampleTeachers },
      global: { plugins: [testI18n(), testRouter()] },
    })
    await flushPromises()
    await wrapper.find('[data-testid="end-assignment-btn"]').trigger('click')
    await wrapper.find('[data-testid="confirm-ok"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="assignment-end-duplicate"]').exists()).toBe(true)
  })
})
