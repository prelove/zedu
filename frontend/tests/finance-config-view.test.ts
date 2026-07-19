import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import FinanceConfigView from '../src/features/finance/FinanceConfigView.vue'
import { authStore } from '../src/stores/auth'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

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
      { path: '/', name: 'home', component: { template: '<div>home</div>' } },
      { path: '/finance/config', name: 'finance-config', component: FinanceConfigView },
    ],
  })
}

function successEnvelope(data: unknown) {
  return { code: 0, data }
}

function errorEnvelope(code: number, message: string) {
  return { code, message, requestId: 'rid-test' }
}

describe('FinanceConfigView', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    authStore.clearSession()
    authStore.state.accessToken = 'tok-owner'
    authStore.state.role = 'OWNER'
    authStore.state.user = { id: 1, username: 'owner', role: 'OWNER', displayName: 'Owner' }
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  function mountView() {
    const router = testRouter()
    router.push('/finance/config')
    return mount(FinanceConfigView, { global: { plugins: [testI18n(), router] } })
  }

  it('loads base currency and payment methods, including disabled methods', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (url === '/system/base-currency') {
        return Promise.resolve(mockResponse(successEnvelope({ currency: 'JPY', locked: false })))
      }
      if (url === '/system/payment-methods') {
        return Promise.resolve(
          mockResponse(
            successEnvelope([
              { code: 'CASH', name: 'Cash', sortOrder: 10, enabled: true },
              { code: 'OTHER', name: 'Other', sortOrder: 99, enabled: false },
            ]),
          ),
        )
      }
      return Promise.reject(new Error(`unexpected ${url}`))
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-testid="finance-config-view"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="finance-base-currency-select"]').element).toHaveProperty('value', 'JPY')
    expect(wrapper.findAll('[data-testid="finance-payment-method-row"]')).toHaveLength(2)
    expect(wrapper.text()).toContain('OTHER')
    expect(wrapper.text()).toContain(zhCN.common.disabled)
  })

  it('disables base currency editing when locked', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (url === '/system/base-currency') {
        return Promise.resolve(mockResponse(successEnvelope({ currency: 'JPY', locked: true })))
      }
      if (url === '/system/payment-methods') {
        return Promise.resolve(mockResponse(successEnvelope([])))
      }
      return Promise.reject(new Error(`unexpected ${url}`))
    })

    const wrapper = mountView()
    await flushPromises()

    const select = wrapper.find('[data-testid="finance-base-currency-select"]')
    const saveButton = wrapper.find('[data-testid="finance-base-currency-save"]')
    expect(select.attributes('disabled')).toBeDefined()
    expect(saveButton.attributes('disabled')).toBeDefined()
    expect(wrapper.find('[data-testid="finance-base-currency-locked"]').exists()).toBe(true)
  })

  it('PUTs the selected base currency and prevents double submit while pending', async () => {
    let putCalls = 0
    let resolvePut: (() => void) | null = null
    let capturedBody: unknown = null

    globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      if (url === '/system/base-currency' && (!opts || !opts.method || opts.method === 'GET')) {
        return Promise.resolve(mockResponse(successEnvelope({ currency: 'JPY', locked: false })))
      }
      if (url === '/system/payment-methods') {
        return Promise.resolve(mockResponse(successEnvelope([])))
      }
      if (url === '/system/base-currency' && opts?.method === 'PUT') {
        putCalls += 1
        capturedBody = JSON.parse(String(opts.body))
        return new Promise((resolve) => {
          resolvePut = () => resolve(mockResponse(successEnvelope({ currency: 'USD', locked: false })))
        })
      }
      return Promise.reject(new Error(`unexpected ${url}`))
    })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-testid="finance-base-currency-select"]').setValue('USD')
    await wrapper.find('[data-testid="finance-base-currency-save"]').trigger('click')
    await wrapper.find('[data-testid="finance-base-currency-save"]').trigger('click')

    expect(putCalls).toBe(1)
    expect(capturedBody).toEqual({ currency: 'USD' })

    resolvePut?.()
    await flushPromises()

    expect(wrapper.find('[data-testid="finance-base-currency-success"]').text()).toContain(
      zhCN.financeConfig.baseCurrencySaved,
    )
  })

  it('POSTs a new payment method', async () => {
    let capturedBody: unknown = null
    let getMethodsCount = 0

    globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      if (url === '/system/base-currency') {
        return Promise.resolve(mockResponse(successEnvelope({ currency: 'JPY', locked: false })))
      }
      if (url === '/system/payment-methods' && (!opts || !opts.method || opts.method === 'GET')) {
        getMethodsCount += 1
        const rows =
          getMethodsCount > 1
            ? [{ code: 'CARD', name: 'Card', sortOrder: 50, enabled: true }]
            : []
        return Promise.resolve(mockResponse(successEnvelope(rows)))
      }
      if (url === '/system/payment-methods' && opts?.method === 'POST') {
        capturedBody = JSON.parse(String(opts.body))
        return Promise.resolve(
          mockResponse(successEnvelope({ code: 'CARD', name: 'Card', sortOrder: 50, enabled: true }), 201),
        )
      }
      return Promise.reject(new Error(`unexpected ${url}`))
    })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-testid="finance-method-create-code"]').setValue('CARD')
    await wrapper.find('[data-testid="finance-method-create-name"]').setValue('Card')
    await wrapper.find('[data-testid="finance-method-create-sort-order"]').setValue('50')
    await wrapper.find('[data-testid="finance-method-create-submit"]').trigger('click')
    await flushPromises()

    expect(capturedBody).toEqual({
      code: 'CARD',
      name: 'Card',
      sortOrder: 50,
      enabled: true,
    })
    expect(wrapper.text()).toContain('CARD')
  })

  it('shows localized create error without raw stable key', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      if (url === '/system/base-currency') {
        return Promise.resolve(mockResponse(successEnvelope({ currency: 'JPY', locked: false })))
      }
      if (url === '/system/payment-methods' && (!opts || !opts.method || opts.method === 'GET')) {
        return Promise.resolve(mockResponse(successEnvelope([])))
      }
      if (url === '/system/payment-methods' && opts?.method === 'POST') {
        return Promise.resolve(mockResponse(errorEnvelope(40901, 'CONFLICT'), 409))
      }
      return Promise.reject(new Error(`unexpected ${url}`))
    })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-testid="finance-method-create-code"]').setValue('CARD')
    await wrapper.find('[data-testid="finance-method-create-name"]').setValue('Card')
    await wrapper.find('[data-testid="finance-method-create-submit"]').trigger('click')
    await flushPromises()

    const errorEl = wrapper.find('[data-testid="finance-method-create-error"]')
    expect(errorEl.exists()).toBe(true)
    expect(errorEl.text()).toContain(zhCN.apiErrors.CONFLICT)
    expect(errorEl.text()).not.toContain('CONFLICT')
  })
})
