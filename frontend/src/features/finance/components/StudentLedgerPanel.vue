<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import EmptyState from '../../../components/EmptyState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import LoadingState from '../../../components/LoadingState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import type { StudentLedgerEntry } from '../../../api/finance'
import { formatTimestamp } from '../payment-utils'

defineProps<{
  studentName: string
  ledger: StudentLedgerEntry[]
  page: number
  pageSize: number
  total: number
  loading: boolean
  error: unknown
}>()

const emit = defineEmits<{
  (e: 'retry'): void
  (e: 'page-change', page: number): void
}>()

const { t } = useI18n()
</script>

<template>
  <section class="ledger-card">
    <header class="section-header">
      <div>
        <h2>{{ t('financePayments.ledgerTitle') }}</h2>
        <p>{{ t('financePayments.ledgerDescription', { student: studentName || '-' }) }}</p>
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
        v-if="ledger.length === 0"
        :message="t('financePayments.emptyLedger')"
      />
      <div
        v-else
        class="ledger-table-wrap"
      >
        <table
          class="ledger-table"
          data-testid="student-ledger-table"
        >
          <thead>
            <tr>
              <th>{{ t('financePayments.ledgerType') }}</th>
              <th>{{ t('financePayments.enrollment') }}</th>
              <th>{{ t('financePayments.amountDelta') }}</th>
              <th>{{ t('financePayments.lessonDelta') }}</th>
              <th>{{ t('financePayments.balanceAfter') }}</th>
              <th>{{ t('financePayments.lessonBalanceAfter') }}</th>
              <th>{{ t('common.note') }}</th>
              <th>{{ t('financePayments.createdAt') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="entry in ledger"
              :key="entry.id"
              data-testid="student-ledger-row"
            >
              <td>{{ entry.bizType }}</td>
              <td>#{{ entry.enrollmentId }}</td>
              <td>{{ entry.amountDelta }}</td>
              <td>{{ entry.lessonDelta }}</td>
              <td>{{ entry.balanceAfter }}</td>
              <td>{{ entry.lessonBalanceAfter }}</td>
              <td>{{ entry.note || '-' }}</td>
              <td>{{ formatTimestamp(entry.createdAt) }}</td>
            </tr>
          </tbody>
        </table>

        <PaginationBar
          :page="page"
          :page-size="pageSize"
          :total="total"
          @page-change="(page) => emit('page-change', page)"
        />
      </div>
    </template>
  </section>
</template>

<style scoped>
.ledger-card {
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

.ledger-table-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.ledger-table {
  width: 100%;
  border-collapse: collapse;
}

.ledger-table th,
.ledger-table td {
  padding: 0.75rem;
  border-bottom: 1px solid #e5e7eb;
  text-align: left;
}
</style>
