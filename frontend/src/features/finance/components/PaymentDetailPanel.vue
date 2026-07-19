<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import EmptyState from '../../../components/EmptyState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import LoadingState from '../../../components/LoadingState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import type { PaymentAttachment, PaymentDetail } from '../../../api/finance'
import { formatFileSize, formatTimestamp } from '../payment-utils'

const props = defineProps<{
  payment: PaymentDetail | null
  studentName: string
  enrollmentLabel: string
  attachments: PaymentAttachment[]
  attachmentsPage: number
  attachmentsPageSize: number
  attachmentsTotal: number
  attachmentsLoading: boolean
  attachmentsError: unknown
  uploadPending: boolean
  uploadErrorKey: string | null
  uploadSuccessKey: string | null
  downloadPendingId: number | null
}>()

const emit = defineEmits<{
  (e: 'retry-attachments'): void
  (e: 'attachment-page-change', page: number): void
  (e: 'upload-file', file: File): void
  (e: 'download', attachment: PaymentAttachment): void
  (e: 'request-void'): void
}>()

const { t } = useI18n()
const selectedFile = ref<File | null>(null)
const canUpload = computed(() => props.payment?.status === 'CONFIRMED')

watch(
  () => props.uploadSuccessKey,
  (value) => {
    if (value) {
      selectedFile.value = null
    }
  },
)

function onFileChange(event: Event): void {
  const input = event.target as HTMLInputElement
  selectedFile.value = input.files?.[0] ?? null
}

function uploadSelectedFile(): void {
  if (selectedFile.value) {
    emit('upload-file', selectedFile.value)
  }
}
</script>

<template>
  <section class="payment-detail-card">
    <header class="section-header">
      <div>
        <h2>{{ t('financePayments.detailTitle') }}</h2>
        <p>{{ t('financePayments.detailDescription') }}</p>
      </div>
      <button
        v-if="payment?.status === 'CONFIRMED'"
        type="button"
        data-testid="payment-void-button"
        @click="emit('request-void')"
      >
        {{ t('financePayments.voidAction') }}
      </button>
    </header>

    <EmptyState
      v-if="!payment"
      :message="t('financePayments.selectPaymentHint')"
    />

    <template v-else>
      <dl class="detail-grid">
        <div>
          <dt>{{ t('financePayments.paymentNo') }}</dt>
          <dd data-testid="payment-detail-no">
            {{ payment.paymentNo }}
          </dd>
        </div>
        <div>
          <dt>{{ t('financePayments.student') }}</dt>
          <dd>{{ studentName }}</dd>
        </div>
        <div>
          <dt>{{ t('financePayments.enrollment') }}</dt>
          <dd>{{ enrollmentLabel }}</dd>
        </div>
        <div>
          <dt>{{ t('common.status') }}</dt>
          <dd>{{ payment.status }}</dd>
        </div>
        <div>
          <dt>{{ t('financePayments.originalAmount') }}</dt>
          <dd>{{ payment.originalAmount }} {{ payment.originalCurrency }}</dd>
        </div>
        <div>
          <dt>{{ t('financePayments.amountBase') }}</dt>
          <dd>{{ payment.amountBase }}</dd>
        </div>
        <div>
          <dt>{{ t('financePayments.fxRateToBase') }}</dt>
          <dd>{{ payment.fxRateToBase }}</dd>
        </div>
        <div>
          <dt>{{ t('financePayments.lessonsAdded') }}</dt>
          <dd>{{ payment.lessonsAdded }}</dd>
        </div>
        <div>
          <dt>{{ t('financePayments.paymentMethod') }}</dt>
          <dd>{{ payment.paymentMethodName }} ({{ payment.paymentMethodCode }})</dd>
        </div>
        <div>
          <dt>{{ t('financePayments.paidAt') }}</dt>
          <dd>{{ formatTimestamp(payment.paidAt) }}</dd>
        </div>
        <div class="detail-note">
          <dt>{{ t('common.note') }}</dt>
          <dd>
            {{ payment.note || '-' }}
          </dd>
        </div>
      </dl>

      <section class="attachments-section">
        <header class="section-header">
          <div>
            <h3>{{ t('financePayments.attachmentsTitle') }}</h3>
            <p>{{ t('financePayments.attachmentsDescription') }}</p>
          </div>
        </header>

        <div class="upload-row">
          <input
            type="file"
            accept=".jpg,.jpeg,.png,.webp,.pdf"
            data-testid="payment-attachment-input"
            :disabled="!canUpload || uploadPending"
            @change="onFileChange"
          >
          <button
            type="button"
            data-testid="payment-attachment-upload"
            :disabled="!canUpload || uploadPending || !selectedFile"
            @click="uploadSelectedFile"
          >
            {{ uploadPending ? t('common.saving') : t('financePayments.uploadAttachment') }}
          </button>
        </div>

        <p
          v-if="uploadSuccessKey"
          class="status-success"
          role="status"
          aria-live="polite"
          data-testid="payment-attachment-success"
        >
          {{ t(uploadSuccessKey) }}
        </p>

        <p
          v-if="uploadErrorKey"
          class="status-error"
          role="alert"
          aria-live="assertive"
          data-testid="payment-attachment-error"
        >
          {{ t(uploadErrorKey) }}
        </p>

        <LoadingState v-if="attachmentsLoading" />
        <ErrorState
          v-else-if="attachmentsError"
          :error="attachmentsError"
          @retry="emit('retry-attachments')"
        />
        <template v-else>
          <EmptyState
            v-if="attachments.length === 0"
            :message="t('financePayments.attachmentsEmpty')"
          />
          <div
            v-else
            class="attachments-list"
          >
            <ul>
              <li
                v-for="attachment in attachments"
                :key="attachment.id"
                data-testid="payment-attachment-row"
              >
                <div>
                  <strong>{{ attachment.fileName }}</strong>
                  <span>{{ formatFileSize(attachment.fileSize) }} · {{ formatTimestamp(attachment.uploadedAt) }}</span>
                </div>
                <button
                  type="button"
                  :data-testid="`payment-attachment-download-${attachment.id}`"
                  :disabled="downloadPendingId === attachment.id"
                  @click="emit('download', attachment)"
                >
                  {{ downloadPendingId === attachment.id ? t('common.loading') : t('financePayments.downloadAttachment') }}
                </button>
              </li>
            </ul>

            <PaginationBar
              :page="attachmentsPage"
              :page-size="attachmentsPageSize"
              :total="attachmentsTotal"
              @page-change="(page) => emit('attachment-page-change', page)"
            />
          </div>
        </template>
      </section>
    </template>
  </section>
</template>

<style scoped>
.payment-detail-card {
  border: 1px solid #d8dee4;
  border-radius: 0.5rem;
  padding: 1rem;
  background: #fff;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: start;
}

.section-header h2,
.section-header h3,
.section-header p {
  margin: 0;
}

.section-header p {
  color: #57606a;
  margin-top: 0.35rem;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 0.75rem;
  margin: 0;
}

.detail-grid dt {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.detail-grid dd {
  margin: 0;
}

.detail-note {
  grid-column: 1 / -1;
}

.attachments-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.upload-row {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.attachments-list ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.attachments-list li {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: center;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 0.75rem;
}

.attachments-list li div {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

button {
  padding: 0.45rem 0.8rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
  background: #fff;
  cursor: pointer;
}

.status-success {
  color: #166534;
}

.status-error {
  color: #dc2626;
}
</style>
