import { describe, it, expect, vi, afterEach } from 'vitest'
import {
  createPayment,
  createPaymentMethod,
  downloadPaymentAttachment,
  getPayment,
  getBaseCurrency,
  listPaymentAttachments,
  listPayments,
  listPaymentMethods,
  listStudentLedger,
  uploadPaymentAttachment,
  updateBaseCurrency,
  updatePaymentMethod,
  voidPayment,
} from '../src/api/finance'

function mockResponse(body: unknown, status = 200): Response {
  return { ok: status < 300, status, json: async () => body } as Response
}

describe('finance API adapter', () => {
  const originalFetch = globalThis.fetch

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('getBaseCurrency calls /system/base-currency with Bearer token', async () => {
    let calledUrl = ''
    let calledHeaders: Record<string, string> = {}
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: RequestInit) => {
      calledUrl = url
      calledHeaders = opts.headers as Record<string, string>
      return Promise.resolve(mockResponse({ code: 0, data: { currency: 'JPY', locked: false } }))
    })

    await getBaseCurrency('tok-owner')

    expect(calledUrl).toBe('/system/base-currency')
    expect(calledHeaders.Authorization).toBe('Bearer tok-owner')
  })

  it('updateBaseCurrency PUTs currency body to /system/base-currency', async () => {
    let calledUrl = ''
    let calledMethod = ''
    let calledBody: unknown = null
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: RequestInit) => {
      calledUrl = url
      calledMethod = opts.method ?? 'GET'
      calledBody = JSON.parse(String(opts.body))
      return Promise.resolve(mockResponse({ code: 0, data: { currency: 'USD', locked: false } }))
    })

    await updateBaseCurrency('tok-owner', 'USD')

    expect(calledUrl).toBe('/system/base-currency')
    expect(calledMethod).toBe('PUT')
    expect(calledBody).toEqual({ currency: 'USD' })
  })

  it('listPaymentMethods calls /system/payment-methods', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: [] }))
    })

    await listPaymentMethods('tok-owner')

    expect(calledUrl).toBe('/system/payment-methods')
  })

  it('createPaymentMethod POSTs the new method payload', async () => {
    let calledUrl = ''
    let calledMethod = ''
    let calledBody: unknown = null
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: RequestInit) => {
      calledUrl = url
      calledMethod = opts.method ?? 'GET'
      calledBody = JSON.parse(String(opts.body))
      return Promise.resolve(mockResponse({ code: 0, data: { code: 'CARD', name: 'Card', sortOrder: 50, enabled: true } }, 201))
    })

    await createPaymentMethod('tok-owner', {
      code: 'CARD',
      name: 'Card',
      sortOrder: 50,
      enabled: true,
    })

    expect(calledUrl).toBe('/system/payment-methods')
    expect(calledMethod).toBe('POST')
    expect(calledBody).toEqual({
      code: 'CARD',
      name: 'Card',
      sortOrder: 50,
      enabled: true,
    })
  })

  it('updatePaymentMethod PATCHes /system/payment-methods/{code} with editable fields only', async () => {
    let calledUrl = ''
    let calledMethod = ''
    let calledBody: unknown = null
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: RequestInit) => {
      calledUrl = url
      calledMethod = opts.method ?? 'GET'
      calledBody = JSON.parse(String(opts.body))
      return Promise.resolve(mockResponse({ code: 0, data: { code: 'CASH', name: 'Cash desk', sortOrder: 10, enabled: false } }))
    })

    await updatePaymentMethod('tok-owner', 'CASH', {
      name: 'Cash desk',
      sortOrder: 10,
      enabled: false,
    })

    expect(calledUrl).toBe('/system/payment-methods/CASH')
    expect(calledMethod).toBe('PATCH')
    expect(calledBody).toEqual({
      name: 'Cash desk',
      sortOrder: 10,
      enabled: false,
    })
  })

  it('createPayment POSTs string monetary fields and paymentNo to /finance/payments', async () => {
    let calledUrl = ''
    let calledMethod = ''
    let calledBody: unknown = null
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: RequestInit) => {
      calledUrl = url
      calledMethod = opts.method ?? 'GET'
      calledBody = JSON.parse(String(opts.body))
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, paymentNo: 'uuid-1', studentId: 2, enrollmentId: 3, amountBase: 1000, lessonsAdded: 5, status: 'CONFIRMED' } }, 201))
    })

    await createPayment('tok-owner', {
      paymentNo: '11111111-1111-4111-8111-111111111111',
      studentId: 2,
      enrollmentId: 3,
      originalAmount: '500.00',
      originalCurrency: 'CNY',
      fxRateToBase: '2',
      lessonsAdded: 5,
      paymentMethodCode: 'CASH',
      paidAt: '2026-07-19T00:00:00.000Z',
      note: 'memo',
    })

    expect(calledUrl).toBe('/finance/payments')
    expect(calledMethod).toBe('POST')
    expect(calledBody).toEqual({
      paymentNo: '11111111-1111-4111-8111-111111111111',
      studentId: 2,
      enrollmentId: 3,
      originalAmount: '500.00',
      originalCurrency: 'CNY',
      fxRateToBase: '2',
      lessonsAdded: 5,
      paymentMethodCode: 'CASH',
      paidAt: '2026-07-19T00:00:00.000Z',
      note: 'memo',
    })
  })

  it('listPayments adds pagination and frozen M3 filters', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 2, pageSize: 10, total: 0 } }))
    })

    await listPayments('tok-owner', { page: 2, pageSize: 10, paymentNo: 'abc', status: 'CONFIRMED' })

    expect(calledUrl).toBe('/finance/payments?page=2&pageSize=10&paymentNo=abc&status=CONFIRMED')
  })

  it('getPayment and voidPayment call the detail and void endpoints', async () => {
    const calls: Array<{ url: string; method: string }> = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      calls.push({ url, method: opts?.method ?? 'GET' })
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, paymentNo: 'uuid-1', studentId: 2, enrollmentId: 3, amountBase: 1000, lessonsAdded: 5, status: 'VOIDED', originalAmount: '500', originalCurrency: 'JPY', fxRateToBase: '1', paymentMethodCode: 'CASH', paidAt: '2026-07-19T00:00:00Z', note: '' } }))
    })

    await getPayment('tok-owner', 1)
    await voidPayment('tok-owner', 1, { reason: 'wrong entry' })

    expect(calls).toEqual([
      { url: '/finance/payments/1', method: 'GET' },
      { url: '/finance/payments/1/void', method: 'POST' },
    ])
  })

  it('listStudentLedger calls /finance/ledger/student/{studentId}', async () => {
    let calledUrl = ''
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calledUrl = url
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listStudentLedger('tok-owner', 9, { page: 1, pageSize: 20 })

    expect(calledUrl).toBe('/finance/ledger/student/9?page=1&pageSize=20')
  })

  it('uploadPaymentAttachment sends multipart without forcing JSON content-type', async () => {
    let calledUrl = ''
    let contentType = ''
    let isFormData = false
    globalThis.fetch = vi.fn().mockImplementation((url: string, opts: RequestInit) => {
      calledUrl = url
      contentType = new Headers(opts.headers).get('Content-Type') ?? ''
      isFormData = opts.body instanceof FormData
      return Promise.resolve(mockResponse({ code: 0, data: { id: 1, paymentId: 1, fileName: 'proof.png', fileType: 'image/png', fileSize: 12, uploadedBy: 1, uploadedAt: '2026-07-19T00:00:00Z' } }, 201))
    })

    await uploadPaymentAttachment('tok-owner', 1, new File(['png'], 'proof.png', { type: 'image/png' }))

    expect(calledUrl).toBe('/finance/payments/1/attachments')
    expect(contentType).toBe('')
    expect(isFormData).toBe(true)
  })

  it('listPaymentAttachments and downloadPaymentAttachment call attachment endpoints', async () => {
    const calls: string[] = []
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      calls.push(url)
      if (url.endsWith('/content')) {
        return Promise.resolve(
          new Response(new Blob(['proof']), {
            status: 200,
            headers: {
              'Content-Type': 'application/pdf',
              'Content-Disposition': 'attachment; filename=\"proof.pdf\"',
            },
          }),
        )
      }
      return Promise.resolve(mockResponse({ code: 0, data: { items: [], page: 1, pageSize: 20, total: 0 } }))
    })

    await listPaymentAttachments('tok-owner', 1, { page: 1, pageSize: 20 })
    const downloaded = await downloadPaymentAttachment('tok-owner', 1, 2)

    expect(calls).toEqual([
      '/finance/payments/1/attachments?page=1&pageSize=20',
      '/finance/payments/1/attachments/2/content',
    ])
    expect(downloaded.fileName).toBe('proof.pdf')
    expect(downloaded.fileType).toBe('application/pdf')
  })
})
