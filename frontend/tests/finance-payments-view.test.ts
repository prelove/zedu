import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import FinancePaymentsView from '../src/features/finance/FinancePaymentsView.vue'
import { authStore } from '../src/stores/auth'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

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
      { path: '/finance/payments', name: 'finance-payments', component: FinancePaymentsView },
    ],
  })
}

function successEnvelope(data: unknown) {
  return { code: 0, data }
}

function mockJsonResponse(data: unknown, status = 200): Response {
  return { ok: status < 300, status, json: async () => data, headers: new Headers() } as Response
}

describe('FinancePaymentsView', () => {
  const originalFetch = globalThis.fetch
  const originalCreateObjectUrl = URL.createObjectURL
  const originalRevokeObjectUrl = URL.revokeObjectURL
  const originalRandomUUID = globalThis.crypto.randomUUID

  beforeEach(() => {
    authStore.clearSession()
    authStore.state.accessToken = 'tok-op'
    authStore.state.role = 'OPERATOR'
    authStore.state.user = { id: 1, username: 'op', role: 'OPERATOR', displayName: 'Operator' }
    vi.spyOn(globalThis.crypto, 'randomUUID').mockReturnValue('33333333-3333-4333-8333-333333333333')
    URL.createObjectURL = vi.fn(() => 'blob:proof')
    URL.revokeObjectURL = vi.fn()
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    URL.createObjectURL = originalCreateObjectUrl
    URL.revokeObjectURL = originalRevokeObjectUrl
    globalThis.crypto.randomUUID = originalRandomUUID
    vi.restoreAllMocks()
  })

  function mountView() {
    const router = testRouter()
    router.push('/finance/payments')
    return mount(FinancePaymentsView, { global: { plugins: [testI18n(), router] } })
  }

  it('creates a payment, then loads detail, ledger, and attachments for the selected payment', async () => {
    let createdBody: unknown = null
    let paymentsListCount = 0

    globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      const method = opts?.method ?? 'GET'

      if (url === '/system/base-currency') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ currency: 'JPY', locked: false })))
      }
      if (url === '/system/payment-methods') {
        return Promise.resolve(mockJsonResponse(successEnvelope([{ code: 'CASH', name: 'Cash', sortOrder: 10, enabled: true }])))
      }
      if (url === '/students?page=1&pageSize=100&status=ACTIVE') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 7, name: 'Alice', timezone: 'Asia/Tokyo', status: 'ACTIVE', createdAt: '2026-07-19T00:00:00Z', updatedAt: '2026-07-19T00:00:00Z' }], page: 1, pageSize: 100, total: 1 })))
      }
      if (url === '/finance/payments?page=1&pageSize=20') {
        paymentsListCount += 1
        const items = paymentsListCount > 1
          ? [{ id: 9, paymentNo: '33333333-3333-4333-8333-333333333333', studentId: 7, enrollmentId: 8, amountBase: 1000, lessonsAdded: 4, status: 'CONFIRMED' }]
          : [{ id: 1, paymentNo: 'old-payment', studentId: 7, enrollmentId: 8, amountBase: 500, lessonsAdded: 2, status: 'CONFIRMED' }]
        return Promise.resolve(mockJsonResponse(successEnvelope({ items, page: 1, pageSize: 20, total: items.length })))
      }
      if (url === '/students/7/enrollments?page=1&pageSize=100') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 8, studentId: 7, domainId: 1, trackId: 2, enrollmentType: 'R', status: 'ACTIVE', createdAt: '2026-07-19T00:00:00Z', updatedAt: '2026-07-19T00:00:00Z' }], page: 1, pageSize: 100, total: 1 })))
      }
      if (url === '/finance/payments' && method === 'POST') {
        createdBody = JSON.parse(String(opts?.body))
        return Promise.resolve(mockJsonResponse(successEnvelope({ id: 9, paymentNo: '33333333-3333-4333-8333-333333333333', studentId: 7, enrollmentId: 8, amountBase: 1000, lessonsAdded: 4, status: 'CONFIRMED' }), 201))
      }
      if (url === '/finance/payments/1' || url === '/finance/payments/9') {
        const id = url.endsWith('/9') ? 9 : 1
        return Promise.resolve(mockJsonResponse(successEnvelope({ id, paymentNo: id === 9 ? '33333333-3333-4333-8333-333333333333' : 'old-payment', studentId: 7, enrollmentId: 8, amountBase: id === 9 ? 1000 : 500, lessonsAdded: id === 9 ? 4 : 2, status: 'CONFIRMED', originalAmount: id === 9 ? '1000' : '500', originalCurrency: 'JPY', fxRateToBase: '1', paymentMethodCode: 'CASH', paidAt: '2026-07-19T00:00:00Z', note: id === 9 ? 'new note' : '' })))
      }
      if (url === '/finance/payments/1/attachments?page=1&pageSize=20' || url === '/finance/payments/9/attachments?page=1&pageSize=20') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [], page: 1, pageSize: 20, total: 0 })))
      }
      if (url === '/finance/ledger/student/7?page=1&pageSize=20') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 1, enrollmentId: 8, bizType: 'RECHARGE', amountDelta: 1000, lessonDelta: 4, balanceAfter: 1500, lessonBalanceAfter: 6, note: 'new note', createdAt: '2026-07-19T00:00:00Z' }], page: 1, pageSize: 20, total: 1 })))
      }
      if (url === '/students/7') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ id: 7, name: 'Alice', timezone: 'Asia/Tokyo', status: 'ACTIVE', createdAt: '2026-07-19T00:00:00Z', updatedAt: '2026-07-19T00:00:00Z' })))
      }
      if (url === '/enrollments/8') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ id: 8, studentId: 7, domainId: 1, trackId: 2, enrollmentType: 'R', status: 'ACTIVE', createdAt: '2026-07-19T00:00:00Z', updatedAt: '2026-07-19T00:00:00Z' })))
      }
      if (url === '/course-domains?page=1&pageSize=100') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 1, code: 'JP', name: 'Japanese', type: 'LANGUAGE', sortOrder: 1, enabled: true }], page: 1, pageSize: 100, total: 1 })))
      }
      if (url === '/course-tracks?page=1&pageSize=100') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 2, domainId: 1, code: 'REG', name: 'Regular', sortOrder: 1, enabled: true }], page: 1, pageSize: 100, total: 1 })))
      }
      if (url === '/course-levels?page=1&pageSize=100' || url === '/course-tags?page=1&pageSize=100') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [], page: 1, pageSize: 100, total: 0 })))
      }
      return Promise.reject(new Error(`unexpected ${method} ${url}`))
    })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-testid="payment-form-student"]').setValue('7')
    await flushPromises()
    await wrapper.find('[data-testid="payment-form-enrollment"]').setValue('8')
    await wrapper.find('[data-testid="payment-form-amount"]').setValue('1000')
    await wrapper.find('[data-testid="payment-form-lessons"]').setValue('4')
    await wrapper.find('[data-testid="payment-form-note"]').setValue('new note')
    await wrapper.find('[data-testid="payment-form-submit"]').trigger('click')
    await flushPromises()

    expect(createdBody).toEqual({
      paymentNo: '33333333-3333-4333-8333-333333333333',
      studentId: 7,
      enrollmentId: 8,
      originalAmount: '1000',
      originalCurrency: 'JPY',
      fxRateToBase: '1',
      lessonsAdded: 4,
      paymentMethodCode: 'CASH',
      paidAt: expect.stringMatching(/^20/),
      note: 'new note',
    })
    expect(wrapper.find('[data-testid="payment-detail-no"]').text()).toBe('33333333-3333-4333-8333-333333333333')
    expect(wrapper.findAll('[data-testid="student-ledger-row"]')).toHaveLength(1)
    expect(wrapper.find('[data-testid="payment-form-success"]').text()).toContain('充值已创建')
  })

  it('uploads, downloads, and voids attachments for a confirmed payment', async () => {
    const anchorClick = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
    let voidBody: unknown = null

    globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      const method = opts?.method ?? 'GET'

      if (url === '/system/base-currency') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ currency: 'JPY', locked: false })))
      }
      if (url === '/system/payment-methods') {
        return Promise.resolve(mockJsonResponse(successEnvelope([{ code: 'CASH', name: 'Cash', sortOrder: 10, enabled: true }])))
      }
      if (url === '/students?page=1&pageSize=100&status=ACTIVE') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 7, name: 'Alice', timezone: 'Asia/Tokyo', status: 'ACTIVE', createdAt: '2026-07-19T00:00:00Z', updatedAt: '2026-07-19T00:00:00Z' }], page: 1, pageSize: 100, total: 1 })))
      }
      if (url === '/finance/payments?page=1&pageSize=20') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 1, paymentNo: 'pay-1', studentId: 7, enrollmentId: 8, amountBase: 500, lessonsAdded: 2, status: 'CONFIRMED' }], page: 1, pageSize: 20, total: 1 })))
      }
      if (url === '/finance/payments/1') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ id: 1, paymentNo: 'pay-1', studentId: 7, enrollmentId: 8, amountBase: 500, lessonsAdded: 2, status: 'CONFIRMED', originalAmount: '500', originalCurrency: 'JPY', fxRateToBase: '1', paymentMethodCode: 'CASH', paidAt: '2026-07-19T00:00:00Z', note: '' })))
      }
      if (url === '/finance/payments/1/attachments?page=1&pageSize=20') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 4, paymentId: 1, fileName: 'proof.pdf', fileType: 'application/pdf', fileSize: 128, uploadedBy: 1, uploadedAt: '2026-07-19T00:00:00Z' }], page: 1, pageSize: 20, total: 1 })))
      }
      if (url === '/finance/ledger/student/7?page=1&pageSize=20') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [], page: 1, pageSize: 20, total: 0 })))
      }
      if (url === '/students/7') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ id: 7, name: 'Alice', timezone: 'Asia/Tokyo', status: 'ACTIVE', createdAt: '2026-07-19T00:00:00Z', updatedAt: '2026-07-19T00:00:00Z' })))
      }
      if (url === '/enrollments/8') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ id: 8, studentId: 7, domainId: 1, trackId: 2, enrollmentType: 'R', status: 'ACTIVE', createdAt: '2026-07-19T00:00:00Z', updatedAt: '2026-07-19T00:00:00Z' })))
      }
      if (url === '/course-domains?page=1&pageSize=100') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 1, code: 'JP', name: 'Japanese', type: 'LANGUAGE', sortOrder: 1, enabled: true }], page: 1, pageSize: 100, total: 1 })))
      }
      if (url === '/course-tracks?page=1&pageSize=100') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [{ id: 2, domainId: 1, code: 'REG', name: 'Regular', sortOrder: 1, enabled: true }], page: 1, pageSize: 100, total: 1 })))
      }
      if (url === '/course-levels?page=1&pageSize=100' || url === '/course-tags?page=1&pageSize=100') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ items: [], page: 1, pageSize: 100, total: 0 })))
      }
      if (url === '/finance/payments/1/attachments' && method === 'POST') {
        return Promise.resolve(mockJsonResponse(successEnvelope({ id: 5, paymentId: 1, fileName: 'upload.png', fileType: 'image/png', fileSize: 32, uploadedBy: 1, uploadedAt: '2026-07-19T00:00:00Z' }), 201))
      }
      if (url === '/finance/payments/1/attachments/4/content') {
        return Promise.resolve(new Response(new Blob(['proof']), { status: 200, headers: { 'Content-Type': 'application/pdf', 'Content-Disposition': 'attachment; filename=\"proof.pdf\"' } }))
      }
      if (url === '/finance/payments/1/void' && method === 'POST') {
        voidBody = JSON.parse(String(opts?.body))
        return Promise.resolve(mockJsonResponse(successEnvelope({ id: 1, paymentNo: 'pay-1', studentId: 7, enrollmentId: 8, amountBase: 500, lessonsAdded: 2, status: 'VOIDED', originalAmount: '500', originalCurrency: 'JPY', fxRateToBase: '1', paymentMethodCode: 'CASH', paidAt: '2026-07-19T00:00:00Z', note: '' })))
      }
      return Promise.reject(new Error(`unexpected ${method} ${url}`))
    })

    const wrapper = mountView()
    await flushPromises()

    const fileInput = wrapper.find('[data-testid="payment-attachment-input"]')
    const inputEl = fileInput.element as HTMLInputElement
    Object.defineProperty(inputEl, 'files', {
      configurable: true,
      value: [new File(['png'], 'upload.png', { type: 'image/png' })],
    })
    await fileInput.trigger('change')
    await wrapper.find('[data-testid="payment-attachment-upload"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="payment-attachment-success"]').text()).toContain('付款凭证已上传')

    await wrapper.find('[data-testid="payment-attachment-download-4"]').trigger('click')
    await flushPromises()
    expect(anchorClick).toHaveBeenCalled()
    expect(URL.createObjectURL).toHaveBeenCalled()

    await wrapper.find('[data-testid="payment-void-button"]').trigger('click')
    await wrapper.find('[data-testid="payment-void-reason"]').setValue('录入错误')
    await wrapper.find('[data-testid="payment-void-confirm"]').trigger('click')
    await flushPromises()

    expect(voidBody).toEqual({ reason: '录入错误' })
  })
})
