/**
 * M2-KIMI-01 browser-equivalent smoke test.
 *
 * This script exercises the real backend through the Vite dev proxy,
 * verifying the same scenarios a browser user would trigger:
 *   1. Login success → access token in response
 *   2. /auth/me with token → user profile
 *   3. Unauthenticated /auth/me → 40101
 *   4. Owner /onboarding/initialize → success
 *   5. Owner repeat /onboarding/initialize → reused=true
 *   6. Operator /onboarding/initialize → 40301
 *   7. /auth/refresh with cookie → new token
 *   8. /auth/logout → session revoked
 *   9. Login failure → 40102
 *
 * This script performs state-changing initialization/reset requests. It is
 * intentionally disabled unless pointed at a disposable environment.
 *
 * Run (PowerShell):
 *   $env:ZEDU_SMOKE_BASE_URL='http://localhost:5173'
 *   $env:ZEDU_SMOKE_OWNER_USERNAME='smoke-owner'
 *   $env:ZEDU_SMOKE_OWNER_PASSWORD='<disposable-password>'
 *   $env:ZEDU_SMOKE_OPERATOR_USERNAME='smoke-operator'
 *   $env:ZEDU_SMOKE_OPERATOR_PASSWORD='<disposable-password>'
 *   $env:ZEDU_SMOKE_ALLOW_MUTATION='1'
 *   node tests/smoke.mjs
 *
 * Never point it at a shared or production database.
 */

const PROXY = process.env.ZEDU_SMOKE_BASE_URL
const OWNER_USERNAME = process.env.ZEDU_SMOKE_OWNER_USERNAME
const OWNER_PASSWORD = process.env.ZEDU_SMOKE_OWNER_PASSWORD
const OPERATOR_USERNAME = process.env.ZEDU_SMOKE_OPERATOR_USERNAME
const OPERATOR_PASSWORD = process.env.ZEDU_SMOKE_OPERATOR_PASSWORD

if (
  !PROXY ||
  !OWNER_USERNAME ||
  !OWNER_PASSWORD ||
  !OPERATOR_USERNAME ||
  !OPERATOR_PASSWORD ||
  process.env.ZEDU_SMOKE_ALLOW_MUTATION !== '1'
) {
  console.error('Refusing to run mutable smoke test without disposable-environment variables and ZEDU_SMOKE_ALLOW_MUTATION=1.')
  process.exit(2)
}

function assert(condition, message) {
  if (!condition) {
    console.error(`FAIL: ${message}`)
    process.exit(1)
  }
  console.warn(`PASS: ${message}`)
}

/** Manual cookie jar — Node fetch doesn't auto-manage cookies like a browser. */
let cookieJar = ''

async function jsonFetch(url, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...options.headers }
  if (cookieJar) {
    headers['Cookie'] = cookieJar
  }
  const res = await fetch(url, {
    method: options.method ?? 'GET',
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined,
    redirect: 'manual',
  })
  // Capture Set-Cookie.
  const setCookie = res.headers.get('set-cookie')
  if (setCookie) {
    // Extract the cookie name=value part (before the first ;).
    const cookiePart = setCookie.split(';')[0]
    if (cookiePart) {
      // Merge with existing cookies.
      const existing = cookieJar ? cookieJar.split('; ').filter(c => c) : []
      const cookieName = cookiePart.split('=')[0]
      const filtered = existing.filter(c => !c.startsWith(`${cookieName}=`))
      filtered.push(cookiePart)
      cookieJar = filtered.join('; ')
    }
  }
  const text = await res.text()
  let body
  try { body = JSON.parse(text) } catch { body = { raw: text } }
  return { status: res.status, body, headers: res.headers }
}

async function main() {
  // 1. Login success.
  const loginRes = await jsonFetch(`${PROXY}/auth/login`, {
    method: 'POST',
    body: { username: OWNER_USERNAME, password: OWNER_PASSWORD },
  })
  assert(loginRes.status === 200, `login returns 200 (got ${loginRes.status})`)
  assert(loginRes.body.code === 0, 'login returns code 0')
  assert(typeof loginRes.body.data?.accessToken === 'string', 'login returns accessToken')
  assert(loginRes.body.data?.role === 'OWNER', 'login returns role OWNER')
  const token = loginRes.body.data.accessToken
  console.warn(`  token length: ${token.length}`)
  console.warn(`  cookie jar: ${cookieJar ? '(set)' : '(empty)'}`)

  // 2. /auth/me with token.
  const meRes = await jsonFetch(`${PROXY}/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  assert(meRes.status === 200, `me returns 200 (got ${meRes.status})`)
  assert(meRes.body.data?.username === OWNER_USERNAME, 'me returns correct username')

  // 3. Unauthenticated /auth/me → 40101.
  const savedCookie = cookieJar
  cookieJar = '' // clear cookies for unauthenticated test
  const meUnauthRes = await jsonFetch(`${PROXY}/auth/me`)
  assert(meUnauthRes.status === 401, `unauthenticated me returns 401 (got ${meUnauthRes.status})`)
  assert(meUnauthRes.body.code === 40101, `unauthenticated me returns 40101 (got ${meUnauthRes.body.code})`)
  cookieJar = savedCookie // restore cookies

  // 4. Owner /onboarding/initialize (japanese).
  const initRes = await jsonFetch(`${PROXY}/onboarding/initialize`, {
    method: 'POST',
    body: { template: 'japanese' },
    headers: { Authorization: `Bearer ${token}` },
  })
  assert(initRes.status === 200, `initialize returns 200 (got ${initRes.status})`)
  assert(initRes.body.code === 0, 'initialize returns code 0')
  assert(initRes.body.data?.template === 'japanese', 'initialize returns template japanese')
  console.warn(`  reused: ${initRes.body.data?.reused}`)

  // 5. Owner repeat initialize → reused=true.
  const initRepeatRes = await jsonFetch(`${PROXY}/onboarding/initialize`, {
    method: 'POST',
    body: { template: 'japanese' },
    headers: { Authorization: `Bearer ${token}` },
  })
  assert(initRepeatRes.body.code === 0, 'repeat initialize returns code 0')
  assert(initRepeatRes.body.data?.reused === true, 'repeat initialize returns reused=true')

  // 6. Operator /onboarding/initialize → 40301.
  // Login as operator (separate cookie context).
  const savedOwnerCookie = cookieJar
  cookieJar = ''
  const opLoginRes = await jsonFetch(`${PROXY}/auth/login`, {
    method: 'POST',
    body: { username: OPERATOR_USERNAME, password: OPERATOR_PASSWORD },
  })
  const opToken = opLoginRes.body.data?.accessToken
  const opInitRes = await jsonFetch(`${PROXY}/onboarding/initialize`, {
    method: 'POST',
    body: { template: 'japanese' },
    headers: { Authorization: `Bearer ${opToken}` },
  })
  assert(opInitRes.status === 403, `operator initialize returns 403 (got ${opInitRes.status})`)
  assert(opInitRes.body.code === 40301, `operator initialize returns 40301 (got ${opInitRes.body.code})`)
  // Logout operator to clean up.
  await jsonFetch(`${PROXY}/auth/logout`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${opToken}` },
  })
  // Restore owner cookies.
  cookieJar = savedOwnerCookie

  // 7. Login failure → 40102.
  const failRes = await jsonFetch(`${PROXY}/auth/login`, {
    method: 'POST',
    body: { username: OWNER_USERNAME, password: 'wrong-password' },
  })
  assert(failRes.status === 401, `failed login returns 401 (got ${failRes.status})`)
  assert(failRes.body.code === 40102, `failed login returns 40102 (got ${failRes.body.code})`)
  assert(failRes.body.message === 'LOGIN_FAILED', 'failed login returns stable key LOGIN_FAILED')

  // 8. /auth/refresh (uses cookie from login).
  const refreshRes = await jsonFetch(`${PROXY}/auth/refresh`, {
    method: 'POST',
  })
  assert(refreshRes.status === 200, `refresh returns 200 (got ${refreshRes.status})`)
  assert(refreshRes.body.code === 0, 'refresh returns code 0')
  assert(typeof refreshRes.body.data?.accessToken === 'string', 'refresh returns new accessToken')
  const refreshedToken = refreshRes.body.data.accessToken
  assert(refreshedToken !== token, 'refresh returns a different token')

  // 9. /auth/logout with refreshed token.
  const logoutRes = await jsonFetch(`${PROXY}/auth/logout`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${refreshedToken}` },
  })
  assert(logoutRes.status === 200, `logout returns 200 (got ${logoutRes.status})`)
  assert(logoutRes.body.code === 0, 'logout returns code 0')

  // 10. After logout, old refresh cookie should be rejected.
  const postLogoutRefresh = await jsonFetch(`${PROXY}/auth/refresh`, {
    method: 'POST',
  })
  assert(postLogoutRefresh.status === 401, `post-logout refresh returns 401 (got ${postLogoutRefresh.status})`)
  assert(postLogoutRefresh.body.code === 40101, `post-logout refresh returns 40101 (got ${postLogoutRefresh.body.code})`)

  // 11. Owner reset (no business data → should succeed).
  // Re-login since we logged out.
  cookieJar = ''
  const reLoginRes = await jsonFetch(`${PROXY}/auth/login`, {
    method: 'POST',
    body: { username: OWNER_USERNAME, password: OWNER_PASSWORD },
  })
  const reToken = reLoginRes.body.data?.accessToken
  const resetRes = await jsonFetch(`${PROXY}/onboarding/reset`, {
    method: 'POST',
    body: { template: 'blank' },
    headers: { Authorization: `Bearer ${reToken}` },
  })
  assert(resetRes.status === 200, `reset returns 200 (got ${resetRes.status})`)
  assert(resetRes.body.code === 0, 'reset returns code 0')
  assert(resetRes.body.data?.template === 'blank', 'reset returns template blank')

  console.warn('\n=== ALL SMOKE TESTS PASSED ===')
}

main().catch((err) => {
  console.error('SMOKE TEST ERROR:', err)
  process.exit(1)
})
