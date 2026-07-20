import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import DashboardView from '../src/features/dashboard/DashboardView.vue'
import NotificationsView from '../src/features/notification/NotificationsView.vue'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'
import { authStore } from '../src/stores/auth'

function response(body: unknown, status = 200): Response {
  return { ok: status >= 200 && status < 300, status, json: async () => body } as Response
}

function i18n() {
  return createI18n({
    legacy: false,
    locale: 'zh-CN',
    fallbackLocale: 'zh-CN',
    messages: { 'zh-CN': zhCN, 'ja-JP': jaJP, 'en-US': enUS },
  })
}

describe('MVP operations views', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    authStore.clearSession()
    authStore.state.accessToken = 'test-token'
    authStore.state.role = 'OWNER'
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
    authStore.clearSession()
  })

  it('renders the five backend-owned dashboard facts and creates an Owner backup', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string, options?: RequestInit) => {
      if (url === '/dashboard') {
        return Promise.resolve(response({
          code: 0,
          data: {
            todayLessons: 1,
            pendingLessonConfirmations: 2,
            renewalNeededStudents: 3,
            teacherPayableAggregate: 4000,
            failedNotifications: 4,
          },
        }))
      }
      if (url === '/system/backups' && options?.method === 'POST') {
        return Promise.resolve(response({ code: 0, data: { file: 'zedu-test' } }, 201))
      }
      return Promise.reject(new Error(`unexpected request ${url}`))
    })

    const wrapper = mount(DashboardView, { global: { plugins: [i18n()] } })
    await flushPromises()
    expect(wrapper.find('[data-testid="dashboard-today-lessons"]').text()).toContain('1')
    expect(wrapper.find('[data-testid="dashboard-pending-confirmations"]').text()).toContain('2')
    expect(wrapper.find('[data-testid="dashboard-renewal-needed"]').text()).toContain('3')
    expect(wrapper.find('[data-testid="dashboard-teacher-payable"]').text()).toContain('4000')
    expect(wrapper.find('[data-testid="dashboard-failed-notifications"]').text()).toContain('4')

    await wrapper.find('[data-testid="dashboard-create-backup"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="dashboard-backup-created"]').text()).toContain('zedu-test')
  })

  it('does not expose the backup action to an Operator', async () => {
    authStore.state.role = 'OPERATOR'
    globalThis.fetch = vi.fn().mockResolvedValue(response({
      code: 0,
      data: { todayLessons: 0, pendingLessonConfirmations: 0, renewalNeededStudents: 0, teacherPayableAggregate: 0, failedNotifications: 0 },
    }))

    const wrapper = mount(DashboardView, { global: { plugins: [i18n()] } })
    await flushPromises()
    expect(wrapper.find('[data-testid="dashboard-create-backup"]').exists()).toBe(false)
    expect(wrapper.text()).not.toMatch(/restore|恢复|復元/i)
  })

  it('renders scheduled reminder status and prevents retry after the bounded limit', async () => {
    globalThis.fetch = vi.fn().mockImplementation((url: string) => {
      if (url === '/notifications/outbox') {
        return Promise.resolve(response({
          code: 0,
          data: {
            items: [
              { id: 1, lessonId: 10, eventType: 'LESSON_REMINDER', recipientEmail: 'masked@example.test', status: 'FAILED', attempts: 2 },
              { id: 2, lessonId: 11, eventType: 'LESSON_REMINDER', recipientEmail: 'masked@example.test', status: 'FAILED', attempts: 3 },
            ],
            page: 1,
            pageSize: 100,
            total: 2,
          },
        }))
      }
      return Promise.reject(new Error(`unexpected request ${url}`))
    })

    const wrapper = mount(NotificationsView, { global: { plugins: [i18n()] } })
    await flushPromises()
    expect(wrapper.find('[data-testid="notification-row-1"]').text()).toContain('课前提醒')
    expect(wrapper.findAll('[data-testid="notification-retry"]')).toHaveLength(1)
    expect(wrapper.find('[data-testid="notifications-manual-retry-hint"]').exists()).toBe(true)
  })
})
