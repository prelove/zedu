import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createRouter, createMemoryHistory, type Router } from 'vue-router'
import { authStore } from '../src/stores/auth'

/**
 * Create a test router with the same guard logic as the real router.
 * We use createMemoryHistory for test environments.
 */
function createTestRouter(): Router {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/login',
        name: 'login',
        component: { template: '<div data-testid="login-view">login</div>' },
        meta: { public: true },
      },
      {
        path: '/',
        name: 'home',
        component: { template: '<div data-testid="home-view">home</div>' },
        meta: { requiresAuth: true },
      },
      {
        path: '/onboarding',
        name: 'onboarding',
        component: { template: '<div data-testid="onboarding-view">onboarding</div>' },
        meta: { requiresAuth: true, requiresOwner: true },
      },
    ],
  })

  router.beforeEach((to) => {
    const isAuthenticated = authStore.isAuthenticated.value
    if (to.meta.requiresAuth && !isAuthenticated) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
    if (to.meta.public && isAuthenticated) {
      return { name: 'home' }
    }
  if (to.meta.requiresOwner && !authStore.isOwner.value) {
      return { name: 'home', query: { denied: 'owner' } }
    }
    return true
  })

  return router
}

describe('router guard', () => {
  beforeEach(() => {
    authStore.clearSession()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('unauthenticated access to / redirects to /login with redirect query', async () => {
    const router = createTestRouter()
    await router.push('/')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('login')
    expect(router.currentRoute.value.query.redirect).toBe('/')
  })

  it('unauthenticated access to /onboarding redirects to /login with redirect query', async () => {
    const router = createTestRouter()
    await router.push('/onboarding')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('login')
    expect(router.currentRoute.value.query.redirect).toBe('/onboarding')
  })

  it('authenticated user accessing /login redirects to /', async () => {
    authStore.state.accessToken = 'tok-123'
    authStore.state.role = 'OPERATOR'

    const router = createTestRouter()
    await router.push('/login')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('home')
  })

  it('authenticated Owner can access /', async () => {
    authStore.state.accessToken = 'tok-123'
    authStore.state.role = 'OWNER'

    const router = createTestRouter()
    await router.push('/')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('home')
  })

  it('authenticated Owner can access /onboarding', async () => {
    authStore.state.accessToken = 'tok-123'
    authStore.state.role = 'OWNER'

    const router = createTestRouter()
    await router.push('/onboarding')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('onboarding')
  })

  it('authenticated Operator is redirected from /onboarding to /', async () => {
    authStore.state.accessToken = 'tok-123'
    authStore.state.role = 'OPERATOR'

    const router = createTestRouter()
    await router.push('/onboarding')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('home')
    expect(router.currentRoute.value.query.denied).toBe('owner')
  })

  it('login redirect preserves the original target path', async () => {
    const router = createTestRouter()
    await router.push('/onboarding')
    await router.isReady()
    // The redirect query should contain /onboarding so we can return after login.
    expect(router.currentRoute.value.query.redirect).toBe('/onboarding')
  })
})
