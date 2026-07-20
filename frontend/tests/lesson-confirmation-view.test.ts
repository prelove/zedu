import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import LessonsView from '../src/features/lesson/LessonsView.vue'
import { authStore } from '../src/stores/auth'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

function response(data: unknown): Response {
  return { ok: true, status: 200, json: async () => ({ code: 0, data }) } as Response
}

function i18n() {
  return createI18n({ legacy: false, locale: 'zh-CN', messages: { 'zh-CN': zhCN, 'ja-JP': jaJP, 'en-US': enUS } })
}

describe('LessonsView confirmation', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    authStore.clearSession()
    authStore.state.accessToken = 'test-token'
    authStore.state.role = 'OPERATOR'
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('uses the selected attendance outcome and explicitly entered confirmation facts', async () => {
    const calls: Array<{ url: string; options?: RequestInit }> = []
    globalThis.fetch = vi.fn().mockImplementation((url: string, options?: RequestInit) => {
      calls.push({ url, options })
      if (url.startsWith('/lessons?')) {
        return Promise.resolve(response({ items: [{ id: 8, lessonNo: 'L-8', enrollmentId: 1, assignmentId: 2, teacherId: 3, studentId: 4, scheduledStartAt: '2026-08-01T10:00:00Z', scheduledEndAt: '2026-08-01T11:00:00Z', durationMin: 60, timezone: 'Asia/Tokyo', meetingType: 'OFFLINE', status: 'SCHEDULED' }], page: 1, pageSize: 20, total: 1 }))
      }
      if (url === '/system/attendance-outcomes') {
        return Promise.resolve(response([{ code: 'ATTENDED', name: 'Attended', suggestedLessonDeducted: '1', suggestedChargeRatio: '1', suggestedTeacherPayRatio: '1' }, { code: 'STUDENT_LEAVE', name: 'Student leave', suggestedLessonDeducted: '0.5', suggestedChargeRatio: '0.5', suggestedTeacherPayRatio: '1' }]))
      }
      if (url === '/lessons/8/confirm') return Promise.resolve(response({ lessonId: 8 }))
      return Promise.resolve(response({ items: [], page: 1, pageSize: 20, total: 0 }))
    })
    const wrapper = mount(LessonsView, { global: { plugins: [i18n()] } })
    await flushPromises()
    await wrapper.get('[data-testid="lesson-confirm-open"]').trigger('click')
    await flushPromises()
    expect(wrapper.get('[data-testid="lesson-confirm-dialog"]').exists()).toBe(true)
    await wrapper.get('[data-testid="lesson-confirm-outcome"]').setValue('STUDENT_LEAVE')
    await wrapper.get('[data-testid="lesson-confirm-deducted"]').setValue('0.5')
    expect((wrapper.get('[data-testid="lesson-confirm-deducted"]').element as HTMLInputElement).checkValidity()).toBe(true)
    await wrapper.get('[data-testid="lesson-confirm-charge"]').setValue('1200')
    await wrapper.get('[data-testid="lesson-confirm-teacher-pay"]').setValue('800')
    await wrapper.get('[data-testid="lesson-confirm-submit"]').trigger('submit')
    await flushPromises()
    const confirmation = calls.find((call) => call.url === '/lessons/8/confirm')
    expect(confirmation).toBeTruthy()
    expect(JSON.parse(String(confirmation?.options?.body))).toMatchObject({ outcomeType: 'STUDENT_LEAVE', lessonDeducted: '0.5', chargeAmount: 1200, teacherPayAmount: 800, actualDurationMin: 60 })
  })
})
