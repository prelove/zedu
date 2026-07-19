import { describe, expect, it, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
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

describe('HomeView', () => {
  beforeEach(() => {
    authStore.clearSession()
    authStore.state.accessToken = 'test-token'
    authStore.state.role = 'OPERATOR'
  })

  it('shows a localized permission notice after the Owner-only route guard redirects', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: HomeView },
        { path: '/students', component: { template: '<div />' } },
        { path: '/teachers', component: { template: '<div />' } },
        { path: '/courses', component: { template: '<div />' } },
        { path: '/finance/payments', component: { template: '<div />' } },
        { path: '/onboarding', component: { template: '<div />' } },
      ],
    })
    await router.push('/?denied=owner')
    await router.isReady()
    const wrapper = mount(HomeView, { global: { plugins: [testI18n(), router] } })

    const notice = wrapper.find('[data-testid="home-owner-required"]')
    expect(notice.exists()).toBe(true)
    expect(notice.text()).toContain(zhCN.routeGuard.ownerRequired)
  })
})
