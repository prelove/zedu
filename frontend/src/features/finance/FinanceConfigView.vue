<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import EmptyState from '../../components/EmptyState.vue'
import ErrorState from '../../components/ErrorState.vue'
import LoadingState from '../../components/LoadingState.vue'
import {
  createPaymentMethod,
  getBaseCurrency,
  listPaymentMethods,
  updateBaseCurrency,
  updatePaymentMethod,
  type BaseCurrency,
  type PaymentMethod,
} from '../../api/finance'
import { errorToI18nKey } from '../../api/error-mapping'
import { ApiError, NetworkError } from '../../api/http'
import { authStore } from '../../stores/auth'

type SupportedCurrency = BaseCurrency['currency']

interface MethodFormState {
  code: string
  name: string
  sortOrder: string
  enabled: boolean
}

const { t } = useI18n()

const loading = ref(true)
const loadError = ref<unknown>(null)

const baseCurrency = ref<BaseCurrency | null>(null)
const baseCurrencyDraft = ref<SupportedCurrency>('JPY')
const baseCurrencySaving = ref(false)
const baseCurrencyError = ref<string | null>(null)
const baseCurrencySuccess = ref<string | null>(null)

const paymentMethods = ref<PaymentMethod[]>([])
const methodSaving = ref(false)
const methodError = ref<string | null>(null)
const methodSuccess = ref<string | null>(null)
const editingCode = ref<string | null>(null)
const methodForm = reactive<MethodFormState>({
  code: '',
  name: '',
  sortOrder: '0',
  enabled: true,
})

const currencyOptions: SupportedCurrency[] = ['JPY', 'CNY', 'USD']

const isEditingMethod = computed(() => editingCode.value !== null)
const isBaseCurrencyLocked = computed(() => baseCurrency.value?.locked ?? false)
const canSaveBaseCurrency = computed(() => {
  if (baseCurrency.value === null) {
    return false
  }
  return (
    !baseCurrencySaving.value &&
    !isBaseCurrencyLocked.value &&
    baseCurrencyDraft.value !== baseCurrency.value.currency
  )
})

function toErrorKey(error: unknown): string {
  if (error instanceof NetworkError) {
    return 'errors.NETWORK_ERROR'
  }
  if (error instanceof ApiError) {
    return errorToI18nKey(error) ?? 'errors.UNKNOWN'
  }
  return 'errors.UNKNOWN'
}

function resetMethodForm(): void {
  editingCode.value = null
  methodForm.code = ''
  methodForm.name = ''
  methodForm.sortOrder = '0'
  methodForm.enabled = true
}

function sortPaymentMethods(items: PaymentMethod[]): PaymentMethod[] {
  return [...items].sort((left, right) => {
    if (left.sortOrder !== right.sortOrder) {
      return left.sortOrder - right.sortOrder
    }
    return left.code.localeCompare(right.code)
  })
}

async function loadFinanceConfig(): Promise<void> {
  loading.value = true
  loadError.value = null

  try {
    const [currency, methods] = await Promise.all([
      authStore.authedRequest((token) => getBaseCurrency(token)),
      authStore.authedRequest((token) => listPaymentMethods(token)),
    ])
    baseCurrency.value = currency
    baseCurrencyDraft.value = currency.currency
    paymentMethods.value = sortPaymentMethods(methods)
  } catch (error) {
    loadError.value = error
  } finally {
    loading.value = false
  }
}

async function refreshPaymentMethods(): Promise<void> {
  const items = await authStore.authedRequest((token) => listPaymentMethods(token))
  paymentMethods.value = sortPaymentMethods(items)
}

async function saveBaseCurrency(): Promise<void> {
  if (!canSaveBaseCurrency.value) {
    return
  }

  baseCurrencySaving.value = true
  baseCurrencyError.value = null
  baseCurrencySuccess.value = null

  try {
    const updated = await authStore.authedRequest((token) =>
      updateBaseCurrency(token, baseCurrencyDraft.value),
    )
    baseCurrency.value = updated
    baseCurrencyDraft.value = updated.currency
    baseCurrencySuccess.value = 'financeConfig.baseCurrencySaved'
  } catch (error) {
    baseCurrencyError.value = toErrorKey(error)
  } finally {
    baseCurrencySaving.value = false
  }
}

function startEditMethod(item: PaymentMethod): void {
  editingCode.value = item.code
  methodForm.code = item.code
  methodForm.name = item.name
  methodForm.sortOrder = String(item.sortOrder)
  methodForm.enabled = item.enabled
  methodError.value = null
  methodSuccess.value = null
}

function cancelEditMethod(): void {
  resetMethodForm()
  methodError.value = null
}

async function submitMethod(): Promise<void> {
  if (methodSaving.value) {
    return
  }

  const trimmedCode = methodForm.code.trim().toUpperCase()
  const trimmedName = methodForm.name.trim()
  const parsedSortOrder = Number.parseInt(methodForm.sortOrder, 10)
  if (
    trimmedName.length === 0 ||
    (editingCode.value === null && trimmedCode.length === 0) ||
    Number.isNaN(parsedSortOrder) ||
    parsedSortOrder < 0
  ) {
    methodError.value = 'apiErrors.INVALID_STATE'
    return
  }

  methodSaving.value = true
  methodError.value = null
  methodSuccess.value = null

  try {
    if (editingCode.value === null) {
      await authStore.authedRequest((token) =>
        createPaymentMethod(token, {
          code: trimmedCode,
          name: trimmedName,
          sortOrder: parsedSortOrder,
          enabled: methodForm.enabled,
        }),
      )
      methodSuccess.value = 'financeConfig.methodCreateSuccess'
    } else {
      await authStore.authedRequest((token) =>
        updatePaymentMethod(token, editingCode.value!, {
          name: trimmedName,
          sortOrder: parsedSortOrder,
          enabled: methodForm.enabled,
        }),
      )
      methodSuccess.value = 'financeConfig.methodUpdateSuccess'
    }

    await refreshPaymentMethods()
    resetMethodForm()
  } catch (error) {
    methodError.value = toErrorKey(error)
  } finally {
    methodSaving.value = false
  }
}

onMounted(() => {
  void loadFinanceConfig()
})
</script>

<template>
  <div
    class="finance-config-view"
    data-testid="finance-config-view"
  >
    <h1>{{ t('financeConfig.title') }}</h1>

    <LoadingState v-if="loading" />

    <ErrorState
      v-else-if="loadError"
      :error="loadError"
      @retry="loadFinanceConfig"
    />

    <template v-else>
      <section class="finance-section">
        <header class="section-header">
          <div>
            <h2>{{ t('financeConfig.baseCurrencyTitle') }}</h2>
            <p>{{ t('financeConfig.baseCurrencyDescription') }}</p>
          </div>
        </header>

        <label
          class="field-label"
          for="finance-base-currency-select"
        >
          {{ t('financeConfig.baseCurrencyLabel') }}
        </label>
        <select
          id="finance-base-currency-select"
          v-model="baseCurrencyDraft"
          class="field-input"
          :disabled="isBaseCurrencyLocked || baseCurrencySaving"
          data-testid="finance-base-currency-select"
          @change="baseCurrencyError = null; baseCurrencySuccess = null"
        >
          <option
            v-for="currency in currencyOptions"
            :key="currency"
            :value="currency"
          >
            {{ currency }}
          </option>
        </select>

        <p
          v-if="isBaseCurrencyLocked"
          class="status-message status-warning"
          role="status"
          aria-live="polite"
          data-testid="finance-base-currency-locked"
        >
          {{ t('financeConfig.baseCurrencyLocked') }}
        </p>

        <p
          v-if="baseCurrencySuccess"
          class="status-message status-success"
          role="status"
          aria-live="polite"
          data-testid="finance-base-currency-success"
        >
          {{ t(baseCurrencySuccess) }}
        </p>

        <p
          v-if="baseCurrencyError"
          class="status-message status-error"
          role="alert"
          aria-live="assertive"
          data-testid="finance-base-currency-error"
        >
          {{ t(baseCurrencyError) }}
        </p>

        <button
          type="button"
          class="primary-button"
          :disabled="!canSaveBaseCurrency"
          data-testid="finance-base-currency-save"
          @click="saveBaseCurrency"
        >
          {{ baseCurrencySaving ? t('common.saving') : t('financeConfig.baseCurrencySave') }}
        </button>
      </section>

      <section class="finance-section">
        <header class="section-header">
          <div>
            <h2>{{ t('financeConfig.paymentMethodsTitle') }}</h2>
            <p>{{ t('financeConfig.paymentMethodsDescription') }}</p>
          </div>
        </header>

        <div class="method-form-card">
          <h3>{{ t(isEditingMethod ? 'financeConfig.methodUpdate' : 'financeConfig.methodCreate') }}</h3>

          <div class="form-grid">
            <div class="field-group">
              <label
                class="field-label"
                for="finance-method-code"
              >
                {{ t('financeConfig.methodCode') }}
              </label>
              <input
                id="finance-method-code"
                v-model="methodForm.code"
                class="field-input"
                :disabled="isEditingMethod || methodSaving"
                data-testid="finance-method-create-code"
              >
            </div>

            <div class="field-group">
              <label
                class="field-label"
                for="finance-method-name"
              >
                {{ t('financeConfig.methodName') }}
              </label>
              <input
                id="finance-method-name"
                v-model="methodForm.name"
                class="field-input"
                :disabled="methodSaving"
                data-testid="finance-method-create-name"
              >
            </div>

            <div class="field-group">
              <label
                class="field-label"
                for="finance-method-sort-order"
              >
                {{ t('financeConfig.methodSortOrder') }}
              </label>
              <input
                id="finance-method-sort-order"
                v-model="methodForm.sortOrder"
                class="field-input"
                type="number"
                min="0"
                :disabled="methodSaving"
                data-testid="finance-method-create-sort-order"
              >
            </div>

            <label class="checkbox-label">
              <input
                v-model="methodForm.enabled"
                type="checkbox"
                :disabled="methodSaving"
                data-testid="finance-method-create-enabled"
              >
              {{ t('common.enabled') }}
            </label>
          </div>

          <p
            v-if="methodSuccess"
            class="status-message status-success"
            role="status"
            aria-live="polite"
          >
            {{ t(methodSuccess) }}
          </p>

          <p
            v-if="methodError"
            class="status-message status-error"
            role="alert"
            aria-live="assertive"
            data-testid="finance-method-create-error"
          >
            {{ t(methodError) }}
          </p>

          <div class="form-actions">
            <button
              type="button"
              class="primary-button"
              :disabled="methodSaving"
              data-testid="finance-method-create-submit"
              @click="submitMethod"
            >
              {{ methodSaving ? t('common.saving') : t(isEditingMethod ? 'financeConfig.methodUpdate' : 'financeConfig.methodCreate') }}
            </button>
            <button
              v-if="isEditingMethod"
              type="button"
              :disabled="methodSaving"
              @click="cancelEditMethod"
            >
              {{ t('common.cancel') }}
            </button>
          </div>
        </div>

        <EmptyState
          v-if="paymentMethods.length === 0"
          :message="t('financeConfig.paymentMethodsEmpty')"
        />

        <table
          v-else
          class="methods-table"
        >
          <thead>
            <tr>
              <th>{{ t('financeConfig.methodCode') }}</th>
              <th>{{ t('financeConfig.methodName') }}</th>
              <th>{{ t('financeConfig.methodSortOrder') }}</th>
              <th>{{ t('common.status') }}</th>
              <th>{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in paymentMethods"
              :key="item.code"
              data-testid="finance-payment-method-row"
            >
              <td>{{ item.code }}</td>
              <td>{{ item.name }}</td>
              <td>{{ item.sortOrder }}</td>
              <td>{{ item.enabled ? t('common.enabled') : t('common.disabled') }}</td>
              <td>
                <button
                  type="button"
                  @click="startEditMethod(item)"
                >
                  {{ t('common.edit') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </section>
    </template>
  </div>
</template>

<style scoped>
.finance-config-view {
  max-width: 960px;
  margin: 0 auto;
  padding: 1rem;
}

.finance-section {
  border: 1px solid #d8dee4;
  border-radius: 0.5rem;
  padding: 1rem;
  margin-top: 1rem;
  background: #fff;
}

.section-header h2,
.method-form-card h3 {
  margin: 0 0 0.5rem 0;
}

.section-header p {
  margin: 0 0 1rem 0;
  color: #57606a;
}

.method-form-card {
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  padding: 1rem;
  margin-bottom: 1rem;
  background: #f8fafc;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.75rem;
  align-items: end;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.field-label {
  font-weight: 600;
}

.field-input {
  padding: 0.45rem 0.6rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
}

.checkbox-label {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  min-height: 2.4rem;
}

.form-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1rem;
}

.primary-button,
.methods-table button {
  padding: 0.45rem 0.8rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
  background: #fff;
  cursor: pointer;
}

.primary-button:disabled,
.methods-table button:disabled {
  cursor: not-allowed;
  color: #94a3b8;
}

.status-message {
  margin: 0.75rem 0;
}

.status-warning {
  color: #b45309;
}

.status-success {
  color: #166534;
}

.status-error {
  color: #dc2626;
}

.methods-table {
  width: 100%;
  border-collapse: collapse;
}

.methods-table th,
.methods-table td {
  padding: 0.75rem;
  border-bottom: 1px solid #e5e7eb;
  text-align: left;
}
</style>
