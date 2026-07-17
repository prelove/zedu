import { describe, it, expect, vi, afterEach } from 'vitest'
import { httpRequest, ApiError, ApiErrorCode, NetworkError } from '../src/api/http'

function mockResponse(body: unknown, status = 200): Response {
  return { ok: status < 300, status, json: async () => body } as Response
}

describe('httpRequest', () => {
  const originalFetch = globalThis.fetch

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('parses success envelope { code: 0, data }', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 0, data: { hello: 'world' } }))

    const res = await httpRequest<{ hello: string }>('/test')
    expect(res.data.hello).toBe('world')
  })

  it('throws ApiError on non-zero code with stable key and requestId', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ code: 40102, message: 'LOGIN_FAILED', requestId: 'rid-1' }, 401))

    try {
      await httpRequest('/test')
      expect.fail('should throw')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      const apiErr = err as ApiError
      expect(apiErr.code).toBe(40102)
      expect(apiErr.stableKey).toBe('LOGIN_FAILED')
      expect(apiErr.requestId).toBe('rid-1')
      expect(apiErr.httpStatus).toBe(401)
    }
  })

  it('throws NetworkError on fetch failure', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'))

    try {
      await httpRequest('/test')
      expect.fail('should throw')
    } catch (err) {
      expect(err).toBeInstanceOf(NetworkError)
    }
  })

  it('sends Authorization Bearer header when token is provided', async () => {
    let capturedHeaders: Record<string, string> | undefined
    globalThis.fetch = vi.fn().mockImplementation((_url: string, opts?: RequestInit) => {
      capturedHeaders = opts?.headers as Record<string, string>
      return Promise.resolve(mockResponse({ code: 0, data: null }))
    })

    await httpRequest('/test', { token: 'tok-123' })
    expect(capturedHeaders?.['Authorization']).toBe('Bearer tok-123')
  })

  it('sends Content-Type JSON when body is provided', async () => {
    let capturedHeaders: Record<string, string> | undefined
    globalThis.fetch = vi.fn().mockImplementation((_url: string, opts?: RequestInit) => {
      capturedHeaders = opts?.headers as Record<string, string>
      return Promise.resolve(mockResponse({ code: 0, data: null }))
    })

    await httpRequest('/test', { method: 'POST', body: { foo: 'bar' } })
    expect(capturedHeaders?.['Content-Type']).toBe('application/json; charset=utf-8')
  })

  it('sends credentials: include for refresh cookie', async () => {
    let capturedCredentials: RequestCredentials | undefined
    globalThis.fetch = vi.fn().mockImplementation((_url: string, opts?: RequestInit) => {
      capturedCredentials = opts?.credentials
      return Promise.resolve(mockResponse({ code: 0, data: null }))
    })

    await httpRequest('/test')
    expect(capturedCredentials).toBe('include')
  })

  it('throws ApiError INTERNAL_ERROR on non-JSON 500 response', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 502,
      json: async () => {
        throw new Error('not JSON')
      },
    } as Response)

    try {
      await httpRequest('/test')
      expect.fail('should throw')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      expect((err as ApiError).code).toBe(ApiErrorCode.INTERNAL_ERROR)
    }
  })

  it('throws ApiError on malformed envelope (missing code)', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(mockResponse({ foo: 'bar' }))

    try {
      await httpRequest('/test')
      expect.fail('should throw')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      expect((err as ApiError).code).toBe(ApiErrorCode.INTERNAL_ERROR)
    }
  })

  it('serializes body as JSON string', async () => {
    let capturedBody: string | undefined
    globalThis.fetch = vi.fn().mockImplementation((_url: string, opts?: RequestInit) => {
      capturedBody = opts?.body as string
      return Promise.resolve(mockResponse({ code: 0, data: null }))
    })

    await httpRequest('/test', { method: 'POST', body: { template: 'japanese' } })
    expect(capturedBody).toBe(JSON.stringify({ template: 'japanese' }))
  })
})
