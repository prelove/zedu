import { reactive, computed } from 'vue'
import * as authApi from '../api/auth'
import { ApiError, ApiErrorCode, NetworkError } from '../api/http'
import type { MeData } from '../api/auth'

/**
 * Auth store — the single source of truth for the in-memory session.
 *
 * INVARIANTS (frozen by M2-KIMI-01 contract):
 * 1. The access token lives ONLY in memory (this reactive state). It is never
 *    written to localStorage, sessionStorage, cookies, or logs.
 * 2. The refresh cookie is handled entirely by the browser (credentials: include).
 *    This store never reads or writes the cookie from JS.
 * 3. On a 40101 from a protected request, refresh exactly ONCE and replay ONCE.
 *    If refresh fails or the replay still returns 40101, clear the session.
 * 4. Concurrent 40101s share a single in-flight refresh promise — no refresh storms.
 * 5. login and refresh requests themselves are NEVER retried (no recursion).
 */

interface AuthState {
  accessToken: string | null
  role: 'OWNER' | 'OPERATOR' | null
  user: MeData | null
}

const state = reactive<AuthState>({
  accessToken: null,
  role: null,
  user: null,
})

/** In-flight refresh promise — concurrent 40101s share this. */
let refreshPromise: Promise<string> | null = null

export const authStore = {
  state,

  isAuthenticated: computed(() => state.accessToken !== null),
  isOwner: computed(() => state.role === 'OWNER'),
  currentUser: computed(() => state.user),

  /**
   * Log in with username/password. On success, stores the access token and role
   * in memory and fetches the user profile. On failure, throws ApiError.
   */
  async login(username: string, password: string): Promise<void> {
    const data = await authApi.login(username, password)
    state.accessToken = data.accessToken
    state.role = data.role
    // Fetch the user profile for display; failure here does NOT discard the
    // session — the token is valid, we just don't have a display name yet.
    try {
      state.user = await authApi.me(data.accessToken)
    } catch {
      state.user = null
    }
  },

  /**
   * Restore the user profile from /auth/me using the existing in-memory token.
   * Called at app initialization. If there is no token, does nothing.
   * If the token is expired (40101), attempts one refresh; if that also fails,
   * clears the session.
   */
  async restore(): Promise<void> {
    if (!state.accessToken) {
      return
    }
    try {
      state.user = await authApi.me(state.accessToken)
    } catch (err) {
      if (err instanceof ApiError && err.code === ApiErrorCode.AUTH_REQUIRED) {
        // Try one refresh; if it works, retry me once.
        try {
          const newToken = await ensureFreshToken()
          state.user = await authApi.me(newToken)
        } catch {
          clearSession()
        }
      } else {
        // Non-auth error (network, 500) — keep the token; user can retry.
        state.user = null
      }
    }
  },

  /**
   * Log out. Calls the backend to revoke the refresh session, then clears
   * the in-memory session regardless. If the logout request fails with a
   * network/500 error, the caller should keep the user on the page and show
   * an error (the session is still valid). If it fails with 40101, the session
   * is already invalid — clear it.
   *
   * Returns true if the session was cleared (success or already-expired),
   * false if the logout request failed and the session is still valid.
   */
  async logout(): Promise<boolean> {
    if (!state.accessToken) {
      clearSession()
      return true
    }
    try {
      await authApi.logout(state.accessToken)
      clearSession()
      return true
    } catch (err) {
      if (err instanceof ApiError && err.code === ApiErrorCode.AUTH_REQUIRED) {
        // Session already invalid — clear and report success.
        clearSession()
        return true
      }
      // Network or server error — session may still be valid. Keep it.
      return false
    }
  },

  /**
   * Authenticated request wrapper. On 40101, refreshes once and replays once.
   * If refresh fails or the replay still 40101s, clears the session and rethrows.
   *
   * The login and refresh endpoints call the raw API functions directly
   * (not this wrapper) to avoid recursion.
   */
  async authedRequest<T>(
    fn: (token: string) => Promise<T>,
  ): Promise<T> {
    if (!state.accessToken) {
      throw new ApiError(ApiErrorCode.AUTH_REQUIRED, 'AUTH_REQUIRED', 'unknown', 401)
    }
    try {
      return await fn(state.accessToken)
    } catch (err) {
      if (!(err instanceof ApiError) || err.code !== ApiErrorCode.AUTH_REQUIRED) {
        throw err
      }
      // 40101 — refresh once, replay once.
      const newToken = await ensureFreshToken()
      try {
        return await fn(newToken)
      } catch (replayErr) {
        // If the replay also returns 40101, the session is invalid — clear it.
        if (replayErr instanceof ApiError && replayErr.code === ApiErrorCode.AUTH_REQUIRED) {
          clearSession()
        }
        throw replayErr
      }
    }
  },

  /** Clear the in-memory session (used by router guard on hard auth failure). */
  clearSession,

  /** Get the current access token (for router guard checks). */
  getToken(): string | null {
    return state.accessToken
  },
}

function clearSession(): void {
  state.accessToken = null
  state.role = null
  state.user = null
  refreshPromise = null
}

/**
 * Ensure a fresh access token. If a refresh is already in flight, share it.
 * On success, updates the in-memory token/role and returns the new token.
 * On failure, clears the session and rethrows.
 *
 * This function is NOT recursive — it calls authApi.refresh directly,
 * never authedRequest.
 */
async function ensureFreshToken(): Promise<string> {
  if (refreshPromise) {
    return refreshPromise
  }
  refreshPromise = (async () => {
    try {
      const data = await authApi.refresh()
      state.accessToken = data.accessToken
      state.role = data.role
      return data.accessToken
    } catch (err) {
      clearSession()
      if (err instanceof ApiError) {
        throw err
      }
      if (err instanceof NetworkError) {
        throw err
      }
      // Unknown error — wrap as internal.
      throw new ApiError(ApiErrorCode.INTERNAL_ERROR, 'INTERNAL_ERROR', 'unknown', 500)
    } finally {
      // Clear the in-flight flag so a future 40101 can attempt another refresh.
      // This is set AFTER the await resolves/rejects, so concurrent callers
      // already sharing this promise are unaffected.
      refreshPromise = null
    }
  })()
  return refreshPromise
}
