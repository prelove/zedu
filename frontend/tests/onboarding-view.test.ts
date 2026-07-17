import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createRouter, createMemoryHistory } from 'vue-router'
import OnboardingView from '../src/features/onboarding/OnboardingView.vue'
import { authStore } from '../src/stores/auth'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

function createTestI18n(locale = 'zh-CN') {
  return createI18n({
    legacy: false,
    locale,
    fallbackLocale: 'zh-CN',
    messages: { 'zh-CN': zhCN, 'ja-JP': jaJP, 'en-US': enUS },
  })
}

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: { template: '<div>home</div>' } },
      { path: '/onboarding', name: 'onboarding', component: OnboardingView },
    ],
  })
}

function mockResponse(body: unknown, status = 200): Response {
  return { ok: status < 300, status, json: async () => body } as Response
}

function successEnvelope(data: unknown) {
  return { code: 0, data }
}

function errorEnvelope(code: number, message: string) {
  return { code, message, requestId: 'rid-test' }
}

describe('OnboardingView', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    authStore.clearSession()
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  function mountAsOwner(locale = 'zh-CN') {
    authStore.state.accessToken = 'tok-owner'
    authStore.state.role = 'OWNER'
    authStore.state.user = { id: 1, username: 'owner', role: 'OWNER', displayName: 'Owner' }

    const i18n = createTestI18n(locale)
    const router = createTestRouter()
    router.push('/onboarding')
    return mount(OnboardingView, { global: { plugins: [i18n, router] } })
  }

  function mountAsOperator() {
    authStore.state.accessToken = 'tok-op'
    authStore.state.role = 'OPERATOR'
    authStore.state.user = { id: 2, username: 'op', role: 'OPERATOR', displayName: 'Operator' }

    const i18n = createTestI18n('zh-CN')
    const router = createTestRouter()
    router.push('/onboarding')
    return mount(OnboardingView, { global: { plugins: [i18n, router] } })
  }

  it('Owner sees template selection and initialize button', () => {
    const wrapper = mountAsOwner()
    expect(wrapper.find('[data-testid="onboarding-template-input"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="onboarding-initialize"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="onboarding-reset"]').exists()).toBe(true)
  })

  it('Operator sees owner-only notice, no initialize button, no API call', () => {
    let fetchCalled = false
    globalThis.fetch = vi.fn(() => {
      fetchCalled = true
      return Promise.resolve(mockResponse({}))
    })

    const wrapper = mountAsOperator()
    expect(wrapper.find('[data-testid="onboarding-owner-only"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="onboarding-initialize"]').exists()).toBe(false)
    expect(fetchCalled).toBe(false)
  })

  it('initialize sends correct template body and shows success result', async () => {
    let capturedBody: unknown = null
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      if (url === '/onboarding/initialize') {
        capturedBody = JSON.parse(opts?.body as string)
        return Promise.resolve(mockResponse(successEnvelope({ template: 'japanese', reused: false })))
      }
      return Promise.reject(new Error(`unexpected: ${url}`))
    })

    const wrapper = mountAsOwner()

    // Select the japanese template.
    const radios = wrapper.findAll('[data-testid="onboarding-template-input"]')
    await radios[0].setValue(true) // japanese is first
    await wrapper.find('[data-testid="onboarding-initialize"]').trigger('click')
    await flushPromises()

    expect(capturedBody).toEqual({ template: 'japanese' })
    const result = wrapper.find('[data-testid="onboarding-result"]')
    expect(result.exists()).toBe(true)
    expect(result.text()).toContain(zhCN.onboarding.templateJapanese)
  })

  it('reused=true shows "existing template" result, not "initialized"', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse(successEnvelope({ template: 'japanese', reused: true })),
    )

    const wrapper = mountAsOwner()
    const radios = wrapper.findAll('[data-testid="onboarding-template-input"]')
    await radios[0].setValue(true)
    await wrapper.find('[data-testid="onboarding-initialize"]').trigger('click')
    await flushPromises()

    const result = wrapper.find('[data-testid="onboarding-result"]')
    expect(result.text()).toContain(zhCN.onboarding.resultReused.split('{template}')[0])
  })

  it('42201/INVALID_TEMPLATE shows localized error', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse(errorEnvelope(42201, 'INVALID_TEMPLATE'), 422),
    )

    const wrapper = mountAsOwner()
    const radios = wrapper.findAll('[data-testid="onboarding-template-input"]')
    await radios[0].setValue(true)
    await wrapper.find('[data-testid="onboarding-initialize"]').trigger('click')
    await flushPromises()

    const errorEl = wrapper.find('[data-testid="onboarding-error"]')
    expect(errorEl.exists()).toBe(true)
    expect(errorEl.text()).toContain(zhCN.apiErrors.INVALID_TEMPLATE)
    expect(errorEl.text()).not.toContain('INVALID_TEMPLATE')
  })

  it('42201/RESET_NOT_ALLOWED shows localized error on reset', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse(errorEnvelope(42201, 'RESET_NOT_ALLOWED'), 422),
    )

    const wrapper = mountAsOwner()
    const radios = wrapper.findAll('[data-testid="onboarding-template-input"]')
    await radios[0].setValue(true)
    // Open reset confirmation.
    await wrapper.find('[data-testid="onboarding-reset"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="onboarding-reset-confirm"]').exists()).toBe(true)
    // Confirm reset.
    await wrapper.find('[data-testid="onboarding-reset-confirm-ok"]').trigger('click')
    await flushPromises()

    const errorEl = wrapper.find('[data-testid="onboarding-error"]')
    expect(errorEl.exists()).toBe(true)
    expect(errorEl.text()).toContain(zhCN.apiErrors.RESET_NOT_ALLOWED)
  })

  it('40301/FORBIDDEN shows localized forbidden error', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse(errorEnvelope(40301, 'FORBIDDEN'), 403),
    )

    const wrapper = mountAsOwner()
    const radios = wrapper.findAll('[data-testid="onboarding-template-input"]')
    await radios[0].setValue(true)
    await wrapper.find('[data-testid="onboarding-initialize"]').trigger('click')
    await flushPromises()

    const errorEl = wrapper.find('[data-testid="onboarding-error"]')
    expect(errorEl.text()).toContain(zhCN.apiErrors.FORBIDDEN)
  })

  it('reset requires confirmation — cancel does not call API', async () => {
    let fetchCalled = false
    globalThis.fetch = vi.fn(() => {
      fetchCalled = true
      return Promise.resolve(mockResponse({}))
    })

    const wrapper = mountAsOwner()
    const radios = wrapper.findAll('[data-testid="onboarding-template-input"]')
    await radios[0].setValue(true)
    await wrapper.find('[data-testid="onboarding-reset"]').trigger('click')
    await flushPromises()

    // Cancel the reset.
    await wrapper.find('[data-testid="onboarding-reset-confirm-cancel"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="onboarding-reset-confirm"]').exists()).toBe(false)
    expect(fetchCalled).toBe(false)
  })

  it('reset sends correct template body after confirmation', async () => {
    let capturedBody: unknown = null
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      if (url === '/onboarding/reset') {
        capturedBody = JSON.parse(opts?.body as string)
        return Promise.resolve(mockResponse(successEnvelope({ template: 'k12', reused: false })))
      }
      return Promise.reject(new Error(`unexpected: ${url}`))
    })

    const wrapper = mountAsOwner()
    const radios = wrapper.findAll('[data-testid="onboarding-template-input"]')
    await radios[1].setValue(true) // k12 is second
    await wrapper.find('[data-testid="onboarding-reset"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-testid="onboarding-reset-confirm-ok"]').trigger('click')
    await flushPromises()

    expect(capturedBody).toEqual({ template: 'k12' })
    const result = wrapper.find('[data-testid="onboarding-result"]')
    expect(result.text()).toContain(zhCN.onboarding.templateK12)
  })

  it('three template options are rendered (japanese, k12, blank)', () => {
    const wrapper = mountAsOwner()
    const radios = wrapper.findAll('[data-testid="onboarding-template-input"]')
    expect(radios).toHaveLength(3)
  })

  it('localizes template labels in ja-JP', () => {
    const wrapper = mountAsOwner('ja-JP')
    expect(wrapper.text()).toContain(jaJP.onboarding.templateJapanese)
    expect(wrapper.text()).toContain(jaJP.onboarding.templateK12)
    expect(wrapper.text()).toContain(jaJP.onboarding.templateBlank)
  })

  it('localizes template labels in en-US', () => {
    const wrapper = mountAsOwner('en-US')
    expect(wrapper.text()).toContain(enUS.onboarding.templateJapanese)
    expect(wrapper.text()).toContain(enUS.onboarding.templateK12)
    expect(wrapper.text()).toContain(enUS.onboarding.templateBlank)
  })
})
