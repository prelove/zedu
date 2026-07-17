import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import PaginationBar from '../src/components/PaginationBar.vue'
import LoadingState from '../src/components/LoadingState.vue'
import EmptyState from '../src/components/EmptyState.vue'
import ErrorState from '../src/components/ErrorState.vue'
import ConfirmDialog from '../src/components/ConfirmDialog.vue'
import { ApiError, ApiErrorCode, NetworkError } from '../src/api/http'
import { zhCN } from '../src/i18n/locales/zh-CN'
import { jaJP } from '../src/i18n/locales/ja-JP'
import { enUS } from '../src/i18n/locales/en-US'

function testI18n() {
  return createI18n({ legacy: false, locale: 'zh-CN', fallbackLocale: 'zh-CN', messages: { 'zh-CN': zhCN, 'ja-JP': jaJP, 'en-US': enUS } })
}

describe('PaginationBar', () => {
  it('renders page info and controls', () => {
    const wrapper = mount(PaginationBar, {
      props: { page: 2, pageSize: 20, total: 50 },
      global: { plugins: [testI18n()] },
    })
    expect(wrapper.find('[data-testid="pagination"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('50')
    expect(wrapper.text()).toContain('2 / 3')
  })

  it('disables prev on first page', () => {
    const wrapper = mount(PaginationBar, {
      props: { page: 1, pageSize: 20, total: 50 },
      global: { plugins: [testI18n()] },
    })
    expect(wrapper.find('[data-testid="pagination-prev"]').attributes('disabled')).toBeDefined()
  })

  it('disables next on last page', () => {
    const wrapper = mount(PaginationBar, {
      props: { page: 3, pageSize: 20, total: 50 },
      global: { plugins: [testI18n()] },
    })
    expect(wrapper.find('[data-testid="pagination-next"]').attributes('disabled')).toBeDefined()
  })

  it('emits page-change on next click', async () => {
    const wrapper = mount(PaginationBar, {
      props: { page: 1, pageSize: 20, total: 50 },
      global: { plugins: [testI18n()] },
    })
    await wrapper.find('[data-testid="pagination-next"]').trigger('click')
    expect(wrapper.emitted('page-change')).toBeTruthy()
    expect(wrapper.emitted('page-change')![0]).toEqual([2])
  })

  it('emits page-change on prev click', async () => {
    const wrapper = mount(PaginationBar, {
      props: { page: 2, pageSize: 20, total: 50 },
      global: { plugins: [testI18n()] },
    })
    await wrapper.find('[data-testid="pagination-prev"]').trigger('click')
    expect(wrapper.emitted('page-change')![0]).toEqual([1])
  })
})

describe('LoadingState', () => {
  it('renders loading message', () => {
    const wrapper = mount(LoadingState, { global: { plugins: [testI18n()] } })
    expect(wrapper.find('[data-testid="state-loading"]').exists()).toBe(true)
    expect(wrapper.text()).toContain(zhCN.common.loading)
  })
})

describe('EmptyState', () => {
  it('renders default no-data message', () => {
    const wrapper = mount(EmptyState, { global: { plugins: [testI18n()] } })
    expect(wrapper.find('[data-testid="state-empty"]').exists()).toBe(true)
    expect(wrapper.text()).toContain(zhCN.common.noData)
  })

  it('renders custom message', () => {
    const wrapper = mount(EmptyState, {
      props: { message: 'Custom empty' },
      global: { plugins: [testI18n()] },
    })
    expect(wrapper.text()).toContain('Custom empty')
  })
})

describe('ErrorState', () => {
  it('renders network error message', () => {
    const err = new NetworkError('fail')
    const wrapper = mount(ErrorState, {
      props: { error: err },
      global: { plugins: [testI18n()] },
    })
    expect(wrapper.find('[data-testid="state-error"]').exists()).toBe(true)
    expect(wrapper.text()).toContain(zhCN.errors.NETWORK_ERROR)
  })

  it('renders ApiError with stable key', () => {
    const err = new ApiError(ApiErrorCode.CONFLICT, 'CONFLICT', 'r1', 409)
    const wrapper = mount(ErrorState, {
      props: { error: err },
      global: { plugins: [testI18n()] },
    })
    expect(wrapper.text()).toContain(zhCN.apiErrors.CONFLICT)
  })

  it('emits retry on retry button click', async () => {
    const err = new NetworkError('fail')
    const wrapper = mount(ErrorState, {
      props: { error: err },
      global: { plugins: [testI18n()] },
    })
    await wrapper.find('[data-testid="state-error-retry"]').trigger('click')
    expect(wrapper.emitted('retry')).toBeTruthy()
  })
})

describe('ConfirmDialog', () => {
  it('renders title and message', () => {
    const wrapper = mount(ConfirmDialog, {
      props: { title: 'Confirm Title', message: 'Are you sure?' },
      global: { plugins: [testI18n()] },
    })
    expect(wrapper.find('[data-testid="confirm-dialog"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Confirm Title')
    expect(wrapper.text()).toContain('Are you sure?')
  })

  it('emits confirm on OK button', async () => {
    const wrapper = mount(ConfirmDialog, {
      props: { title: 'T', message: 'M' },
      global: { plugins: [testI18n()] },
    })
    await wrapper.find('[data-testid="confirm-ok"]').trigger('click')
    expect(wrapper.emitted('confirm')).toBeTruthy()
  })

  it('emits cancel on Cancel button', async () => {
    const wrapper = mount(ConfirmDialog, {
      props: { title: 'T', message: 'M' },
      global: { plugins: [testI18n()] },
    })
    await wrapper.find('[data-testid="confirm-cancel"]').trigger('click')
    expect(wrapper.emitted('cancel')).toBeTruthy()
  })
})
