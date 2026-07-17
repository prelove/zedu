import { httpRequest } from './http'

/** POST /auth/login response data: { accessToken, role }. */
export interface LoginData {
  accessToken: string
  role: 'OWNER' | 'OPERATOR'
}

/** GET /auth/me response data: { id, username, role, displayName }. */
export interface MeData {
  id: number
  username: string
  role: 'OWNER' | 'OPERATOR'
  displayName: string
}

/** POST /auth/refresh response data: { accessToken, role }. */
export interface RefreshData {
  accessToken: string
  role: 'OWNER' | 'OPERATOR'
}

/**
 * POST /auth/login — public. Sets refresh cookie via Set-Cookie header
 * (handled by the browser; we never read the cookie from JS).
 */
export function login(username: string, password: string): Promise<LoginData> {
  return httpRequest<LoginData>('/auth/login', {
    method: 'POST',
    body: { username, password },
    skipAuthRetry: true,
  }).then((res) => res.data)
}

/**
 * POST /auth/refresh — uses refresh cookie (credentials: include).
 * Never send a body or access token.
 */
export function refresh(): Promise<RefreshData> {
  return httpRequest<RefreshData>('/auth/refresh', {
    method: 'POST',
    skipAuthRetry: true,
  }).then((res) => res.data)
}

/**
 * POST /auth/logout — authenticated. Revokes the refresh session and
 * clears the refresh cookie server-side.
 */
export function logout(token: string): Promise<void> {
  return httpRequest<void>('/auth/logout', {
    method: 'POST',
    token,
    skipAuthRetry: true,
  }).then(() => undefined)
}

/**
 * GET /auth/me — authenticated. Returns the current account profile.
 */
export function me(token: string): Promise<MeData> {
  return httpRequest<MeData>('/auth/me', {
    method: 'GET',
    token,
    skipAuthRetry: true,
  }).then((res) => res.data)
}
