<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import EmptyState from '../../../components/EmptyState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import LoadingState from '../../../components/LoadingState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import type { PaymentSummary } from '../../../api/finance'

defineProps<{
  payments: PaymentSummary[]
  selectedPaymentId: number | null
  page: number
  pageSize: number
  total: number
  loading: boolean
  error: unknown
  studentName: (studentId: number) => string
  enrollmentLabel: (enrollmentId: number) => string
}>()

const emit = defineEmits<{
  (e: 'retry'): void
  (e: 'select', paymentId: number): void
  (e: 'page-change', page: number): void
}>()

const { t } = useI18n()
</script>

<template>
  <section class="payments-table-card">
    <header class="section-header">
      <div>
        <h2>{{ t('financePayments.listTitle') }}</h2>
        <p>{{ t('financePayments.listDescription') }}</p>
      </div>
    </header>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="emit('retry')"
    />
    <template v-else>
      <EmptyState
        v-if="payments.length === 0"
        :message="t('financePayments.emptyPayments')"
      />

      <div
        v-else
        class="payments-table-wrap"
      >
        <table
          class="payments-table"
          data-testid="payments-table"
        >
          <thead>
            <tr>
              <th>{{ t('financePayments.paymentNo') }}</th>
              <th>{{ t('financePayments.student') }}</th>
              <th>{{ t('financePayments.enrollment') }}</th>
              <th>{{ t('financePayments.amountBase') }}</th>
              <th>{{ t('financePayments.lessonsAdded') }}</th>
              <th>{{ t('common.status') }}</th>
              <th>{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="payment in payments"
              :key="payment.id"
              :class="{ selected: payment.id === selectedPaymentId }"
              data-testid="payments-row"
            >
              <td>{{ payment.paymentNo }}</td>
              <td>{{ studentName(payment.studentId) }}</td>
              <td>{{ enrollmentLabel(payment.enrollmentId) }}</td>
              <td>{{ payment.amountBase }}</td>
              <td>{{ payment.lessonsAdded }}</td>
              <td>{{ payment.status }}</td>
              <td>
                <button
                  type="button"
                  :data-testid="`payments-select-${payment.id}`"
                  @click="emit('select', payment.id)"
                >
                  {{ t('financePayments.viewDetail') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <PaginationBar
          :page="page"
          :page-size="pageSize"
          :total="total"
          @page-change="(nextPage) => emit('page-change', nextPage)"
        />
      </div>
    </template>
  </section>
</template>

<style scoped>
.payments-table-card {
  border: 1px solid #d8dee4;
  border-radius: 0.5rem;
  padding: 1rem;
  background: #fff;
}

.section-header h2,
.section-header p {
  margin: 0;
}

.section-header p {
  color: #57606a;
  margin-top: 0.35rem;
}

.payments-table-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.payments-table {
  width: 100%;
  border-collapse: collapse;
}

.payments-table th,
.payments-table td {
  padding: 0.75rem;
  border-bottom: 1px solid #e5e7eb;
  text-align: left;
}

.payments-table tr.selected {
  background: #eff6ff;
}

button {
  padding: 0.35rem 0.7rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
  background: #fff;
  cursor: pointer;
}
</style>
