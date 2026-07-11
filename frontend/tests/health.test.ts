import { describe, it, expect, vi, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import HealthStatus from '../src/components/HealthStatus.vue'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

function createTestI18n(locale = 'zh-CN') {
  return createI18n({
    legacy: false,
    locale,
    fallbackLocale: 'zh-CN',
    messages: {
      'zh-CN': zhCN,
      'ja-JP': jaJP,
      'en-US': enUS,
    },
  })
}

describe('HealthStatus component', () => {
  const originalFetch = globalThis.fetch

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('shows loading state initially', () => {
    globalThis.fetch = vi.fn(() => new Promise(() => {})) // never resolves
    const i18n = createTestI18n('zh-CN')
    const wrapper = mount(HealthStatus, {
      global: { plugins: [i18n] },
    })
    expect(wrapper.text()).toContain(zhCN.health.loading)
  })

  it('shows healthy state when API returns ok', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ status: 'ok' }),
    } as Response)
    const i18n = createTestI18n('zh-CN')
    const wrapper = mount(HealthStatus, {
      global: { plugins: [i18n] },
    })
    await flushPromises()
    expect(globalThis.fetch).toHaveBeenCalledWith('/healthz')
    expect(wrapper.text()).toContain(zhCN.health.healthy)
  })

  it('shows unavailable state when API is unreachable', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'))
    const i18n = createTestI18n('zh-CN')
    const wrapper = mount(HealthStatus, {
      global: { plugins: [i18n] },
    })
    await flushPromises()
    expect(wrapper.text()).toContain(zhCN.health.unavailable)
  })

  it('shows unavailable state when API returns non-200', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 503,
      json: async () => ({}),
    } as Response)
    const i18n = createTestI18n('zh-CN')
    const wrapper = mount(HealthStatus, {
      global: { plugins: [i18n] },
    })
    await flushPromises()
    expect(wrapper.text()).toContain(zhCN.health.unavailable)
  })

  it('displays localized text in ja-JP', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ status: 'ok' }),
    } as Response)
    const i18n = createTestI18n('ja-JP')
    const wrapper = mount(HealthStatus, {
      global: { plugins: [i18n] },
    })
    await flushPromises()
    expect(wrapper.text()).toContain(jaJP.health.healthy)
  })

  it('displays localized text in en-US', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ status: 'ok' }),
    } as Response)
    const i18n = createTestI18n('en-US')
    const wrapper = mount(HealthStatus, {
      global: { plugins: [i18n] },
    })
    await flushPromises()
    expect(wrapper.text()).toContain(enUS.health.healthy)
  })

  it('does not expose raw exception stack traces to user', async () => {
    const errorMsg = 'NetworkError: raw stack trace details at line 42'
    globalThis.fetch = vi.fn().mockRejectedValue(new Error(errorMsg))
    const i18n = createTestI18n('zh-CN')
    const wrapper = mount(HealthStatus, {
      global: { plugins: [i18n] },
    })
    await flushPromises()
    expect(wrapper.text()).not.toContain('stack trace')
    expect(wrapper.text()).not.toContain('line 42')
    expect(wrapper.text()).not.toContain('NetworkError')
  })
})
