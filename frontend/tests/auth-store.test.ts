import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { authStore } from '../src/stores/auth'
import { ApiError, ApiErrorCode } from '../src/api/http'

/**
 * Helper: create a mock Response with a JSON body.
 */
function mockResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  } as Response
}

function successEnvelope(data: unknown) {
  return { code: 0, data }
}

function errorEnvelope(code: number, message: string, requestId = 'rid-test') {
  return { code, message, requestId }
}

describe('authStore', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    authStore.clearSession()
    refreshCallCount = 0
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  // Track refresh calls across tests.
  let refreshCallCount = 0

  describe('login', () => {
    it('stores access token in memory only (not localStorage/sessionStorage)', async () => {
      globalThis.fetch = vi.fn().mockImplementation((url: string) => {
        if (url === '/auth/login') {
          return Promise.resolve(mockResponse(successEnvelope({ accessToken: 'tok-123', role: 'OWNER' })))
        }
        if (url === '/auth/me') {
          return Promise.resolve(mockResponse(successEnvelope({ id: 1, username: 'admin', role: 'OWNER', displayName: 'Admin' })))
        }
        return Promise.reject(new Error(`unexpected url: ${url}`))
      })

      await authStore.login('admin', 'pass')

      expect(authStore.getToken()).toBe('tok-123')
      expect(authStore.state.role).toBe('OWNER')
      expect(authStore.isAuthenticated.value).toBe(true)
      // Token must NOT be in any storage.
      expect(localStorage.getItem('accessToken')).toBeNull()
      expect(sessionStorage.getItem('accessToken')).toBeNull()
    })

    it('login failed (40102) throws ApiError and does not store token', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(
        mockResponse(errorEnvelope(40102, 'LOGIN_FAILED'), 401),
      )

      await expect(authStore.login('bad', 'creds')).rejects.toThrow()
      expect(authStore.getToken()).toBeNull()
      expect(authStore.isAuthenticated.value).toBe(false)
    })

    it('account locked (40103) throws ApiError with code 40103', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(
        mockResponse(errorEnvelope(40103, 'ACCOUNT_LOCKED'), 401),
      )

      try {
        await authStore.login('locked', 'user')
        expect.fail('should have thrown')
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError)
        expect((err as ApiError).code).toBe(40103)
      }
    })

    it('network error throws and does not store token', async () => {
      globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'))

      await expect(authStore.login('admin', 'pass')).rejects.toThrow()
      expect(authStore.getToken()).toBeNull()
    })

    it('500 error throws ApiError with INTERNAL_ERROR or DATABASE_ERROR', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(
        mockResponse(errorEnvelope(50001, 'INTERNAL_ERROR'), 500),
      )

      try {
        await authStore.login('admin', 'pass')
        expect.fail('should have thrown')
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError)
        expect((err as ApiError).code).toBe(50001)
      }
    })
  })

  describe('refresh-once retry on 40101', () => {
    it('refreshes once and replays the request on 40101', async () => {
      let protectedCallCount = 0
      globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
        const method = opts?.method ?? 'GET'
        if (url === '/auth/refresh') {
          refreshCallCount++
          return Promise.resolve(mockResponse(successEnvelope({ accessToken: 'tok-fresh', role: 'OWNER' })))
        }
        if (url === '/onboarding/initialize' && method === 'POST') {
          protectedCallCount++
          // First call uses the stale token → 40101.
          // Second call (after refresh) uses the fresh token → success.
          const authHeader = opts?.headers?.['Authorization'] as string | undefined
          if (authHeader === 'Bearer tok-stale') {
            return Promise.resolve(mockResponse(errorEnvelope(40101, 'AUTH_REQUIRED'), 401))
          }
          return Promise.resolve(mockResponse(successEnvelope({ template: 'japanese', reused: false })))
        }
        return Promise.reject(new Error(`unexpected: ${url} ${method}`))
      })

      // Set up a stale session.
      authStore.state.accessToken = 'tok-stale'
      authStore.state.role = 'OWNER'

      // Use authedRequest to make a protected call.
      const result = await authStore.authedRequest(async (token) => {
        const res = await fetch('/onboarding/initialize', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
          body: JSON.stringify({ template: 'japanese' }),
          credentials: 'include',
        })
        const body = await res.json()
        if (body.code !== 0) {
          throw new ApiError(body.code, body.message, body.requestId, res.status)
        }
        return body.data as { template: string; reused: boolean }
      })

      expect(protectedCallCount).toBe(2) // initial + replay
      expect(refreshCallCount).toBe(1)
      expect(result.template).toBe('japanese')
      expect(authStore.getToken()).toBe('tok-fresh')
    })

    it('does NOT retry login request on 40101 (no recursion)', async () => {
      let loginCallCount = 0
      globalThis.fetch = vi.fn().mockImplementation((url: string) => {
        if (url === '/auth/login') {
          loginCallCount++
          return Promise.resolve(mockResponse(errorEnvelope(40101, 'AUTH_REQUIRED'), 401))
        }
        return Promise.reject(new Error(`unexpected: ${url}`))
      })

      try {
        await authStore.login('admin', 'pass')
        expect.fail('should have thrown')
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError)
        expect((err as ApiError).code).toBe(40101)
      }
      expect(loginCallCount).toBe(1) // no retry
    })

    it('does NOT retry refresh request on 40101 (no recursion)', async () => {
      let refreshCount = 0
      globalThis.fetch = vi.fn().mockImplementation((url: string) => {
        if (url === '/auth/refresh') {
          refreshCount++
          return Promise.resolve(mockResponse(errorEnvelope(40101, 'AUTH_REQUIRED'), 401))
        }
        if (url === '/onboarding/initialize') {
          return Promise.resolve(mockResponse(errorEnvelope(40101, 'AUTH_REQUIRED'), 401))
        }
        return Promise.reject(new Error(`unexpected: ${url}`))
      })

      // Set up a stale session and trigger a protected request that 40101s.
      authStore.state.accessToken = 'tok-stale'
      authStore.state.role = 'OWNER'

      await expect(
        authStore.authedRequest(async (token) => {
          const res = await fetch('/onboarding/initialize', {
            method: 'POST',
            headers: { Authorization: `Bearer ${token}` },
            credentials: 'include',
          })
          const body = await res.json()
          if (body.code !== 0) {
            throw new ApiError(body.code, body.message, body.requestId, res.status)
          }
          return body.data
        }),
      ).rejects.toThrow()

      // Refresh was called once, and the 40101 from refresh was NOT retried.
      expect(refreshCount).toBe(1)
      // Session should be cleared.
      expect(authStore.getToken()).toBeNull()
    })

    it('repeated 40101 after refresh clears session and does not loop', async () => {
      let protectedCount = 0
      globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
        const method = opts?.method ?? 'GET'
        if (url === '/auth/refresh') {
          return Promise.resolve(mockResponse(successEnvelope({ accessToken: 'tok-fresh', role: 'OWNER' })))
        }
        if (url === '/onboarding/initialize' && method === 'POST') {
          protectedCount++
          // Always returns 40101, even after refresh.
          return Promise.resolve(mockResponse(errorEnvelope(40101, 'AUTH_REQUIRED'), 401))
        }
        return Promise.reject(new Error(`unexpected: ${url}`))
      })

      authStore.state.accessToken = 'tok-stale'
      authStore.state.role = 'OWNER'

      await expect(
        authStore.authedRequest(async (token) => {
          const res = await fetch('/onboarding/initialize', {
            method: 'POST',
            headers: { Authorization: `Bearer ${token}` },
            credentials: 'include',
          })
          const body = await res.json()
          if (body.code !== 0) {
            throw new ApiError(body.code, body.message, body.requestId, res.status)
          }
          return body.data
        }),
      ).rejects.toThrow()

      // Only 2 calls: initial + one replay. No loop.
      expect(protectedCount).toBe(2)
      expect(authStore.getToken()).toBeNull()
    })

    it('concurrent 40101s share a single refresh (no storm)', async () => {
      let refreshCount = 0
      globalThis.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
        const method = opts?.method ?? 'GET'
        if (url === '/auth/refresh') {
          refreshCount++
          return Promise.resolve(mockResponse(successEnvelope({ accessToken: 'tok-fresh', role: 'OWNER' })))
        }
        if (url === '/onboarding/initialize' && method === 'POST') {
          const authHeader = opts?.headers?.['Authorization'] as string | undefined
          if (authHeader === 'Bearer tok-stale') {
            return Promise.resolve(mockResponse(errorEnvelope(40101, 'AUTH_REQUIRED'), 401))
          }
          return Promise.resolve(mockResponse(successEnvelope({ template: 'japanese', reused: false })))
        }
        return Promise.reject(new Error(`unexpected: ${url}`))
      })

      authStore.state.accessToken = 'tok-stale'
      authStore.state.role = 'OWNER'

      // Fire 3 concurrent protected requests.
      const makeRequest = () =>
        authStore.authedRequest(async (token) => {
          const res = await fetch('/onboarding/initialize', {
            method: 'POST',
            headers: { Authorization: `Bearer ${token}` },
            credentials: 'include',
          })
          const body = await res.json()
          if (body.code !== 0) {
            throw new ApiError(body.code, body.message, body.requestId, res.status)
          }
          return body.data
        })

      const results = await Promise.all([makeRequest(), makeRequest(), makeRequest()])

      // All 3 should succeed.
      for (const r of results) {
        expect(r.template).toBe('japanese')
      }
      // Only 1 refresh call for all 3 concurrent 40101s.
      expect(refreshCount).toBe(1)
    })
  })

  describe('logout', () => {
    it('clears session on successful logout', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(mockResponse(successEnvelope(null), 200))

      authStore.state.accessToken = 'tok-123'
      authStore.state.role = 'OWNER'

      const cleared = await authStore.logout()
      expect(cleared).toBe(true)
      expect(authStore.getToken()).toBeNull()
      expect(authStore.isAuthenticated.value).toBe(false)
    })

    it('logout failure (network) keeps session and returns false', async () => {
      globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'))

      authStore.state.accessToken = 'tok-123'
      authStore.state.role = 'OWNER'

      const cleared = await authStore.logout()
      expect(cleared).toBe(false)
      expect(authStore.getToken()).toBe('tok-123')
    })

    it('logout 40101 clears session (already expired)', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(
        mockResponse(errorEnvelope(40101, 'AUTH_REQUIRED'), 401),
      )

      authStore.state.accessToken = 'tok-expired'
      authStore.state.role = 'OWNER'

      const cleared = await authStore.logout()
      expect(cleared).toBe(true)
      expect(authStore.getToken()).toBeNull()
    })
  })

  describe('restore', () => {
    it('does nothing when there is no token', async () => {
      let fetchCalled = false
      globalThis.fetch = vi.fn(() => {
        fetchCalled = true
        return Promise.resolve(mockResponse({}))
      })

      await authStore.restore()
      expect(fetchCalled).toBe(false)
    })

    it('fetches /auth/me when token exists', async () => {
      globalThis.fetch = vi.fn().mockImplementation((url: string) => {
        if (url === '/auth/me') {
          return Promise.resolve(mockResponse(successEnvelope({ id: 1, username: 'admin', role: 'OWNER', displayName: 'Admin' })))
        }
        return Promise.reject(new Error(`unexpected: ${url}`))
      })

      authStore.state.accessToken = 'tok-123'
      authStore.state.role = 'OWNER'

      await authStore.restore()
      expect(authStore.currentUser.value?.username).toBe('admin')
    })
  })

  describe('authedRequest edge cases', () => {
    it('throws AUTH_REQUIRED when no token is set', async () => {
      authStore.clearSession()
      try {
        await authStore.authedRequest(async () => 'should not reach')
        expect.fail('should throw')
      } catch (err) {
        expect(err).toBeInstanceOf(ApiError)
        expect((err as ApiError).code).toBe(ApiErrorCode.AUTH_REQUIRED)
      }
    })

    it('rethrows non-ApiError errors without retry', async () => {
      authStore.state.accessToken = 'tok-123'
      authStore.state.role = 'OWNER'
      globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('network'))

      try {
        await authStore.authedRequest(async () => {
          // Simulate a non-ApiError throw.
          throw new TypeError('custom error')
        })
        expect.fail('should throw')
      } catch (err) {
        expect(err).toBeInstanceOf(TypeError)
      }
    })
  })
})
