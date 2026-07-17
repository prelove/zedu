import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createRouter, createMemoryHistory } from 'vue-router'
import LoginView from '../src/features/auth/LoginView.vue'
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
      { path: '/login', name: 'login', component: LoginView },
      { path: '/', name: 'home', component: { template: '<div>home</div>' } },
      { path: '/onboarding', name: 'onboarding', component: { template: '<div>onboarding</div>' } },
    ],
  })
}

function mockSuccessEnvelope(data: unknown) {
  return { code: 0, data }
}

function mockErrorEnvelope(code: number, message: string) {
  return { code, message, requestId: 'rid-test' }
}

function mockResponse(body: unknown, status = 200): Response {
  return { ok: status < 300, status, json: async () => body } as Response
}

describe('LoginView', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    // Clear auth store state before each test.
    vi.resetModules()
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('renders form with username and password fields and labels', () => {
    const i18n = createTestI18n('zh-CN')
    const router = createTestRouter()
    router.push('/login')
    const wrapper = mount(LoginView, { global: { plugins: [i18n, router] } })

    expect(wrapper.find('[data-testid="login-username"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="login-password"]').exists()).toBe(true)
    // Labels are associated via for/id.
    expect(wrapper.find('label[for="login-username"]').exists()).toBe(true)
    expect(wrapper.find('label[for="login-password"]').exists()).toBe(true)
  })

  it('submit button is disabled when fields are empty', () => {
    const i18n = createTestI18n('zh-CN')
    const router = createTestRouter()
    router.push('/login')
    const wrapper = mount(LoginView, { global: { plugins: [i18n, router] } })

    const button = wrapper.find('[data-testid="login-submit"]')
    expect(button.attributes('disabled')).toBeDefined()
  })

  it('successful login redirects to redirect query path', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (url === '/auth/login') {
        return Promise.resolve(mockResponse(mockSuccessEnvelope({ accessToken: 'tok', role: 'OWNER' })))
      }
      if (url === '/auth/me') {
        return Promise.resolve(mockResponse(mockSuccessEnvelope({ id: 1, username: 'admin', role: 'OWNER', displayName: 'Admin' })))
      }
      return Promise.reject(new Error(`unexpected: ${url}`))
    })

    const i18n = createTestI18n('zh-CN')
    const router = createTestRouter()
    await router.push('/login?redirect=/onboarding')
    const wrapper = mount(LoginView, { global: { plugins: [i18n, router] } })

    await wrapper.find('[data-testid="login-username"]').setValue('admin')
    await wrapper.find('[data-testid="login-password"]').setValue('pass')
    await wrapper.find('[data-testid="login-form"]').trigger('submit.prevent')
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/onboarding')
  })

  it('login failed (40102) shows localized error, not raw response', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse(mockErrorEnvelope(40102, 'LOGIN_FAILED'), 401),
    )

    const i18n = createTestI18n('zh-CN')
    const router = createTestRouter()
    await router.push('/login')
    const wrapper = mount(LoginView, { global: { plugins: [i18n, router] } })

    await wrapper.find('[data-testid="login-username"]').setValue('bad')
    await wrapper.find('[data-testid="login-password"]').setValue('creds')
    await wrapper.find('[data-testid="login-form"]').trigger('submit.prevent')
    await flushPromises()

    const errorEl = wrapper.find('[data-testid="login-error"]')
    expect(errorEl.exists()).toBe(true)
    expect(errorEl.text()).toContain(zhCN.apiErrors.LOGIN_FAILED)
    // Must NOT contain raw error key or requestId.
    expect(errorEl.text()).not.toContain('LOGIN_FAILED')
    expect(errorEl.text()).not.toContain('rid-test')
  })

  it('account locked (40103) shows localized locked message in all three locales', async () => {
    for (const [locale, msgs] of [['zh-CN', zhCN], ['ja-JP', jaJP], ['en-US', enUS]] as const) {
      globalThis.fetch = vi.fn().mockResolvedValue(
        mockResponse(mockErrorEnvelope(40103, 'ACCOUNT_LOCKED'), 401),
      )

      const i18n = createTestI18n(locale)
      const router = createTestRouter()
      await router.push('/login')
      const wrapper = mount(LoginView, { global: { plugins: [i18n, router] } })

      await wrapper.find('[data-testid="login-username"]').setValue('locked')
      await wrapper.find('[data-testid="login-password"]').setValue('user')
      await wrapper.find('[data-testid="login-form"]').trigger('submit.prevent')
      await flushPromises()

      const errorEl = wrapper.find('[data-testid="login-error"]')
      expect(errorEl.text()).toContain(msgs.apiErrors.ACCOUNT_LOCKED)
    }
  })

  it('network error shows localized network error, not raw exception', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch: ECONNREFUSED at 127.0.0.1:8080'))

    const i18n = createTestI18n('zh-CN')
    const router = createTestRouter()
    await router.push('/login')
    const wrapper = mount(LoginView, { global: { plugins: [i18n, router] } })

    await wrapper.find('[data-testid="login-username"]').setValue('admin')
    await wrapper.find('[data-testid="login-password"]').setValue('pass')
    await wrapper.find('[data-testid="login-form"]').trigger('submit.prevent')
    await flushPromises()

    const errorEl = wrapper.find('[data-testid="login-error"]')
    expect(errorEl.exists()).toBe(true)
    expect(errorEl.text()).toContain(zhCN.errors.NETWORK_ERROR)
    // Must not leak raw error details.
    expect(errorEl.text()).not.toContain('ECONNREFUSED')
    expect(errorEl.text()).not.toContain('127.0.0.1')
  })

  it('500 error shows localized server error', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(
      mockResponse(mockErrorEnvelope(50001, 'INTERNAL_ERROR'), 500),
    )

    const i18n = createTestI18n('en-US')
    const router = createTestRouter()
    await router.push('/login')
    const wrapper = mount(LoginView, { global: { plugins: [i18n, router] } })

    await wrapper.find('[data-testid="login-username"]').setValue('admin')
    await wrapper.find('[data-testid="login-password"]').setValue('pass')
    await wrapper.find('[data-testid="login-form"]').trigger('submit.prevent')
    await flushPromises()

    const errorEl = wrapper.find('[data-testid="login-error"]')
    expect(errorEl.text()).toContain(enUS.apiErrors.INTERNAL_ERROR)
  })

  it('submit button is disabled and shows submitting text during submission', async () => {
    // Never-resolving fetch to keep submitting state active.
    globalThis.fetch = vi.fn(() => new Promise(() => {}))

    const i18n = createTestI18n('zh-CN')
    const router = createTestRouter()
    await router.push('/login')
    const wrapper = mount(LoginView, { global: { plugins: [i18n, router] } })

    await wrapper.find('[data-testid="login-username"]').setValue('admin')
    await wrapper.find('[data-testid="login-password"]').setValue('pass')
    await wrapper.find('[data-testid="login-form"]').trigger('submit.prevent')
    await flushPromises()

    const button = wrapper.find('[data-testid="login-submit"]')
    expect(button.attributes('disabled')).toBeDefined()
    expect(button.text()).toContain(zhCN.auth.submitting)
  })

  it('does not render password value in the DOM beyond the input value attribute', async () => {
    const i18n = createTestI18n('zh-CN')
    const router = createTestRouter()
    await router.push('/login')
    const wrapper = mount(LoginView, { global: { plugins: [i18n, router] } })

    const passwordInput = wrapper.find('[data-testid="login-password"]')
    await passwordInput.setValue('secret123')
    // The password is in the input's value (as expected for a password field),
    // but must not appear elsewhere in the rendered text.
    expect(wrapper.text()).not.toContain('secret123')
  })
})
