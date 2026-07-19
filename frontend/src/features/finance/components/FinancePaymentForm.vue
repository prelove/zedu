<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Enrollment } from '../../../api/course'
import type { Student } from '../../../api/directory'
import type { BaseCurrency, PaymentMethod, PaymentWrite } from '../../../api/finance'
import { fromDatetimeLocalValue, generatePaymentNo, toDatetimeLocalValue } from '../payment-utils'

const props = defineProps<{
  baseCurrency: BaseCurrency
  paymentMethods: PaymentMethod[]
  students: Student[]
  enrollments: Enrollment[]
  submitting: boolean
  errorKey: string | null
  successKey: string | null
  resetVersion: number
}>()

const emit = defineEmits<{
  (e: 'submit', payload: PaymentWrite): void
  (e: 'student-change', studentId: number | null): void
}>()

const { t } = useI18n()

interface FormState {
  paymentNo: string
  studentId: string
  enrollmentId: string
  originalAmount: string
  originalCurrency: BaseCurrency['currency']
  fxRateToBase: string
  lessonsAdded: string
  paymentMethodCode: string
  paidAt: string
  note: string
}

const form = reactive<FormState>(createEmptyForm())
const validationError = computed(() => currentValidationError.value)
const localError = computed(() => props.errorKey ?? validationError.value)
const currentValidationError = computed({
  get: () => validationErrorState,
  set: (value: string) => {
    validationErrorState = value
  },
})
let validationErrorState = ''

function createEmptyForm(): FormState {
  return {
    paymentNo: generatePaymentNo(),
    studentId: '',
    enrollmentId: '',
    originalAmount: '',
    originalCurrency: props.baseCurrency.currency,
    fxRateToBase: '1',
    lessonsAdded: '1',
    paymentMethodCode: props.paymentMethods[0]?.code ?? '',
    paidAt: toDatetimeLocalValue(),
    note: '',
  }
}

function resetForm(): void {
  Object.assign(form, createEmptyForm())
  currentValidationError.value = ''
  emit('student-change', null)
}

watch(
  () => props.resetVersion,
  () => resetForm(),
)

watch(
  () => props.paymentMethods,
  (methods) => {
    if (methods.length === 0) {
      form.paymentMethodCode = ''
      return
    }
    if (!methods.some((item) => item.code === form.paymentMethodCode)) {
      form.paymentMethodCode = methods[0].code
    }
  },
  { immediate: true },
)

watch(
  () => props.baseCurrency.currency,
  (currency) => {
    if (form.originalCurrency === currency) {
      form.fxRateToBase = '1'
    }
  },
)

watch(
  () => form.originalCurrency,
  (currency) => {
    if (currency === props.baseCurrency.currency) {
      form.fxRateToBase = '1'
    } else if (form.fxRateToBase === '1') {
      form.fxRateToBase = ''
    }
  },
)

watch(
  () => form.studentId,
  (studentId) => {
    form.enrollmentId = ''
    emit('student-change', studentId ? Number(studentId) : null)
  },
)

watch(
  () => props.enrollments,
  (items) => {
    if (items.length === 0) {
      form.enrollmentId = ''
    } else if (!items.some((item) => String(item.id) === form.enrollmentId)) {
      form.enrollmentId = String(items[0].id)
    }
  },
)

function submit(): void {
  currentValidationError.value = ''

  const studentId = Number(form.studentId)
  const enrollmentId = Number(form.enrollmentId)
  const lessonsAdded = Number.parseInt(form.lessonsAdded, 10)
  if (
    form.paymentNo.trim() === '' ||
    !Number.isInteger(studentId) ||
    studentId <= 0 ||
    !Number.isInteger(enrollmentId) ||
    enrollmentId <= 0 ||
    form.originalAmount.trim() === '' ||
    form.fxRateToBase.trim() === '' ||
    !Number.isInteger(lessonsAdded) ||
    lessonsAdded <= 0 ||
    form.paymentMethodCode.trim() === '' ||
    form.paidAt.trim() === ''
  ) {
    currentValidationError.value = 'apiErrors.INVALID_STATE'
    return
  }

  emit('submit', {
    paymentNo: form.paymentNo,
    studentId,
    enrollmentId,
    originalAmount: form.originalAmount.trim(),
    originalCurrency: form.originalCurrency,
    fxRateToBase: form.fxRateToBase.trim(),
    lessonsAdded,
    paymentMethodCode: form.paymentMethodCode,
    paidAt: fromDatetimeLocalValue(form.paidAt),
    note: form.note.trim(),
  })
}
</script>

<template>
  <section class="payment-form-card">
    <header class="section-header">
      <div>
        <h2>{{ t('financePayments.createTitle') }}</h2>
        <p>{{ t('financePayments.createDescription') }}</p>
      </div>
    </header>

    <div class="form-grid">
      <div class="field-group">
        <label for="payment-form-number">{{ t('financePayments.paymentNo') }}</label>
        <input
          id="payment-form-number"
          v-model="form.paymentNo"
          data-testid="payment-form-no"
          readonly
          class="field-input"
        >
      </div>

      <div class="field-group">
        <label for="payment-form-student">{{ t('financePayments.student') }}</label>
        <select
          id="payment-form-student"
          v-model="form.studentId"
          data-testid="payment-form-student"
          class="field-input"
          :disabled="submitting"
        >
          <option value="">
            {{ t('financePayments.selectStudent') }}
          </option>
          <option
            v-for="student in students"
            :key="student.id"
            :value="student.id"
          >
            {{ student.name }}
          </option>
        </select>
      </div>

      <div class="field-group">
        <label for="payment-form-enrollment">{{ t('financePayments.enrollment') }}</label>
        <select
          id="payment-form-enrollment"
          v-model="form.enrollmentId"
          data-testid="payment-form-enrollment"
          class="field-input"
          :disabled="submitting || enrollments.length === 0"
        >
          <option value="">
            {{ t('financePayments.selectEnrollment') }}
          </option>
          <option
            v-for="enrollment in enrollments"
            :key="enrollment.id"
            :value="enrollment.id"
          >
            #{{ enrollment.id }}
          </option>
        </select>
      </div>

      <div class="field-group">
        <label for="payment-form-method">{{ t('financePayments.paymentMethod') }}</label>
        <select
          id="payment-form-method"
          v-model="form.paymentMethodCode"
          data-testid="payment-form-method"
          class="field-input"
          :disabled="submitting"
        >
          <option
            v-for="method in paymentMethods"
            :key="method.code"
            :value="method.code"
          >
            {{ method.name }}
          </option>
        </select>
      </div>

      <div class="field-group">
        <label for="payment-form-amount">{{ t('financePayments.originalAmount') }}</label>
        <input
          id="payment-form-amount"
          v-model="form.originalAmount"
          data-testid="payment-form-amount"
          class="field-input"
          inputmode="decimal"
          :disabled="submitting"
        >
      </div>

      <div class="field-group">
        <label for="payment-form-currency">{{ t('financePayments.originalCurrency') }}</label>
        <select
          id="payment-form-currency"
          v-model="form.originalCurrency"
          data-testid="payment-form-currency"
          class="field-input"
          :disabled="submitting"
        >
          <option value="JPY">
            JPY
          </option>
          <option value="CNY">
            CNY
          </option>
          <option value="USD">
            USD
          </option>
        </select>
      </div>

      <div class="field-group">
        <label for="payment-form-rate">{{ t('financePayments.fxRateToBase') }}</label>
        <input
          id="payment-form-rate"
          v-model="form.fxRateToBase"
          data-testid="payment-form-rate"
          class="field-input"
          inputmode="decimal"
          :disabled="submitting || form.originalCurrency === baseCurrency.currency"
        >
      </div>

      <div class="field-group">
        <label for="payment-form-lessons">{{ t('financePayments.lessonsAdded') }}</label>
        <input
          id="payment-form-lessons"
          v-model="form.lessonsAdded"
          data-testid="payment-form-lessons"
          class="field-input"
          type="number"
          min="1"
          :disabled="submitting"
        >
      </div>

      <div class="field-group">
        <label for="payment-form-paid-at">{{ t('financePayments.paidAt') }}</label>
        <input
          id="payment-form-paid-at"
          v-model="form.paidAt"
          data-testid="payment-form-paid-at"
          class="field-input"
          type="datetime-local"
          :disabled="submitting"
        >
      </div>
    </div>

    <div class="field-group field-group-wide">
      <label for="payment-form-note">{{ t('common.note') }}</label>
      <textarea
        id="payment-form-note"
        v-model="form.note"
        data-testid="payment-form-note"
        class="field-input textarea-input"
        :disabled="submitting"
      />
    </div>

    <p class="helper-text">
      {{ t('financePayments.attachmentAfterCreate') }}
    </p>

    <p
      v-if="successKey"
      class="status-message status-success"
      role="status"
      aria-live="polite"
      data-testid="payment-form-success"
    >
      {{ t(successKey) }}
    </p>

    <p
      v-if="localError"
      class="status-message status-error"
      role="alert"
      aria-live="assertive"
      data-testid="payment-form-error"
    >
      {{ t(localError) }}
    </p>

    <div class="form-actions">
      <button
        type="button"
        class="primary-button"
        :disabled="submitting"
        data-testid="payment-form-submit"
        @click="submit"
      >
        {{ submitting ? t('common.saving') : t('financePayments.submitPayment') }}
      </button>
      <button
        type="button"
        :disabled="submitting"
        data-testid="payment-form-reset"
        @click="resetForm"
      >
        {{ t('common.cancel') }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.payment-form-card,
.field-group {
  display: flex;
  flex-direction: column;
}

.payment-form-card {
  gap: 0.75rem;
  border: 1px solid #d8dee4;
  border-radius: 0.5rem;
  padding: 1rem;
  background: #fff;
}

.section-header h2,
.section-header p {
  margin: 0;
}

.section-header p,
.helper-text {
  color: #57606a;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.75rem;
}

.field-group {
  gap: 0.35rem;
}

.field-group-wide {
  gap: 0.35rem;
}

.field-input {
  padding: 0.45rem 0.6rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
}

.textarea-input {
  min-height: 5rem;
}

.form-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.primary-button,
button {
  padding: 0.45rem 0.8rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
  background: #fff;
  cursor: pointer;
}

button:disabled {
  cursor: not-allowed;
  color: #94a3b8;
}

.status-success {
  color: #166534;
}

.status-error {
  color: #dc2626;
}
</style>
