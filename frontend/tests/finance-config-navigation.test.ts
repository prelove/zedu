import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'
import HomeView from '../src/features/auth/HomeView.vue'
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

function createTestRouter(): Router {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', name: 'login', component: { template: '<div />' }, meta: { public: true } },
      { path: '/', name: 'home', component: HomeView, meta: { requiresAuth: true } },
      { path: '/students', name: 'students', component: { template: '<div />' }, meta: { requiresAuth: true } },
      { path: '/teachers', name: 'teachers', component: { template: '<div />' }, meta: { requiresAuth: true } },
      { path: '/courses', name: 'courses', component: { template: '<div />' }, meta: { requiresAuth: true } },
      { path: '/finance/payments', name: 'finance-payments', component: { template: '<div />' }, meta: { requiresAuth: true } },
      { path: '/onboarding', name: 'onboarding', component: { template: '<div />' }, meta: { requiresAuth: true, requiresOwner: true } },
      { path: '/finance/config', name: 'finance-config', component: { template: '<div />' }, meta: { requiresAuth: true, requiresOwner: true } },
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

describe('finance config navigation', () => {
  beforeEach(() => {
    authStore.clearSession()
  })

  it('redirects Operator away from /finance/config', async () => {
    authStore.state.accessToken = 'tok-op'
    authStore.state.role = 'OPERATOR'

    const router = createTestRouter()
    await router.push('/finance/config')
    await router.isReady()

    expect(router.currentRoute.value.name).toBe('home')
    expect(router.currentRoute.value.query.denied).toBe('owner')
  })

  it('shows finance config nav for Owner only', async () => {
    authStore.state.accessToken = 'tok-owner'
    authStore.state.role = 'OWNER'
    authStore.state.user = { id: 1, username: 'owner', role: 'OWNER', displayName: 'Owner' }

    const router = createTestRouter()
    await router.push('/')
    await router.isReady()

    const wrapper = mount(HomeView, { global: { plugins: [testI18n(), router] } })
    expect(wrapper.find('[data-testid="nav-finance-payments"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="nav-finance-config"]').exists()).toBe(true)
  })

  it('does not show finance config nav for Operator', async () => {
    authStore.state.accessToken = 'tok-op'
    authStore.state.role = 'OPERATOR'
    authStore.state.user = { id: 2, username: 'op', role: 'OPERATOR', displayName: 'Operator' }

    const router = createTestRouter()
    await router.push('/')
    await router.isReady()

    const wrapper = mount(HomeView, { global: { plugins: [testI18n(), router] } })
    expect(wrapper.find('[data-testid="nav-finance-payments"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="nav-finance-config"]').exists()).toBe(false)
  })
})
