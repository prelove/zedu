<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getEnrollment, listEnrollments, type Enrollment } from '../../api/course'
import { getStudent, listStudents, type Student } from '../../api/directory'
import { errorToI18nKey } from '../../api/error-mapping'
import {
  createPayment,
  downloadPaymentAttachment,
  getBaseCurrency,
  getPayment,
  listPaymentAttachments,
  listPaymentMethods,
  listPayments,
  listStudentLedger,
  uploadPaymentAttachment,
  voidPayment,
  type BaseCurrency,
  type PaymentAttachment,
  type PaymentDetail,
  type PaymentMethod,
  type PaymentSummary,
  type PaymentWrite,
  type StudentLedgerEntry,
} from '../../api/finance'
import ErrorState from '../../components/ErrorState.vue'
import LoadingState from '../../components/LoadingState.vue'
import { usePaginatedList } from '../../composables/usePaginatedList'
import { useDictionary } from '../../composables/useDictionary'
import { authStore } from '../../stores/auth'
import FinancePaymentForm from './components/FinancePaymentForm.vue'
import FinancePaymentsTable from './components/FinancePaymentsTable.vue'
import PaymentDetailPanel from './components/PaymentDetailPanel.vue'
import PaymentVoidDialog from './components/PaymentVoidDialog.vue'
import StudentLedgerPanel from './components/StudentLedgerPanel.vue'

const { t } = useI18n()
const { domains, tracks, load: loadDictionary } = useDictionary()

const loading = ref(true)
const loadError = ref<unknown>(null)

const baseCurrency = ref<BaseCurrency | null>(null)
const paymentMethods = ref<PaymentMethod[]>([])
const students = ref<Student[]>([])
const formEnrollments = ref<Enrollment[]>([])

const formSubmitting = ref(false)
const formErrorKey = ref<string | null>(null)
const formSuccessKey = ref<string | null>(null)
const formResetVersion = ref(0)

const payments = usePaginatedList<PaymentSummary>(20)
const paymentsError = ref<unknown>(null)
const paymentFilter = reactive({ paymentNo: '', status: '' })
const selectedPaymentId = ref<number | null>(null)
const paymentDetail = ref<PaymentDetail | null>(null)
const detailError = ref<unknown>(null)

const attachments = usePaginatedList<PaymentAttachment>(20)
const attachmentsError = ref<unknown>(null)
const uploadPending = ref(false)
const uploadErrorKey = ref<string | null>(null)
const uploadSuccessKey = ref<string | null>(null)
const downloadPendingId = ref<number | null>(null)

const ledger = usePaginatedList<StudentLedgerEntry>(20)
const ledgerError = ref<unknown>(null)

const voidDialogOpen = ref(false)
const voidPending = ref(false)
const voidErrorKey = ref<string | null>(null)

const studentNameById = reactive<Record<number, string>>({})
const enrollmentLabelById = reactive<Record<number, string>>({})

function mapError(error: unknown): string {
  return errorToI18nKey(error) ?? 'errors.UNKNOWN'
}

function studentName(studentId: number): string {
  return studentNameById[studentId] ?? `#${studentId}`
}

function enrollmentLabel(enrollmentId: number): string {
  return enrollmentLabelById[enrollmentId] ?? `#${enrollmentId}`
}

function buildEnrollmentLabel(item: Enrollment): string {
  const domainName = domains.value.find((domain) => domain.id === item.domainId)?.name
  const trackName = tracks.value.find((track) => track.id === item.trackId)?.name
  return [domainName, trackName].filter(Boolean).join(' / ') || `#${item.id}`
}

async function ensureReferences(items: Array<Pick<PaymentSummary, 'studentId' | 'enrollmentId'>>): Promise<void> {
  const missingStudentIds = [...new Set(items.map((item) => item.studentId))].filter((id) => !studentNameById[id])
  const missingEnrollmentIds = [...new Set(items.map((item) => item.enrollmentId))].filter((id) => !enrollmentLabelById[id])

  await Promise.all([
    ...missingStudentIds.map(async (studentId) => {
      try {
        const student = await authStore.authedRequest((token) => getStudent(token, studentId))
        studentNameById[studentId] = student.name
      } catch {
        studentNameById[studentId] = `#${studentId}`
      }
    }),
    ...missingEnrollmentIds.map(async (enrollmentId) => {
      try {
        const item = await authStore.authedRequest((token) => getEnrollment(token, enrollmentId))
        enrollmentLabelById[enrollmentId] = buildEnrollmentLabel(item)
      } catch {
        enrollmentLabelById[enrollmentId] = `#${enrollmentId}`
      }
    }),
  ])
}

async function loadStudentsList(): Promise<void> {
  const data = await authStore.authedRequest((token) =>
    listStudents(token, { page: 1, pageSize: 100, status: 'ACTIVE' }),
  )
  students.value = data.items
  for (const student of data.items) {
    studentNameById[student.id] = student.name
  }
}

async function loadFormEnrollments(studentId: number | null): Promise<void> {
  formEnrollments.value = []
  if (!studentId) {
    return
  }
  const data = await authStore.authedRequest((token) =>
    listEnrollments(token, studentId, { page: 1, pageSize: 100 }),
  )
  formEnrollments.value = data.items.filter((item) => item.status === 'ACTIVE')
  for (const item of formEnrollments.value) {
    enrollmentLabelById[item.id] = buildEnrollmentLabel(item)
  }
}

async function loadPaymentsPage(page = payments.page.value): Promise<void> {
  payments.loading.value = true
  paymentsError.value = null
  try {
    const data = await authStore.authedRequest((token) =>
      listPayments(token, { page, pageSize: payments.pageSize.value, ...paymentFilter }),
    )
    payments.setData(data)
    await ensureReferences(data.items)
  } catch (error) {
    paymentsError.value = error
  } finally {
    payments.loading.value = false
  }
}

function applyPaymentFilter(): void {
  void loadPaymentsPage(1)
}

async function loadAttachmentsPage(page = attachments.page.value): Promise<void> {
  if (!selectedPaymentId.value) {
    return
  }
  attachments.loading.value = true
  attachmentsError.value = null
  try {
    const data = await authStore.authedRequest((token) =>
      listPaymentAttachments(token, selectedPaymentId.value!, { page, pageSize: attachments.pageSize.value }),
    )
    attachments.setData(data)
  } catch (error) {
    attachmentsError.value = error
  } finally {
    attachments.loading.value = false
  }
}

async function loadLedgerPage(page = ledger.page.value): Promise<void> {
  if (!paymentDetail.value) {
    return
  }
  ledger.loading.value = true
  ledgerError.value = null
  try {
    const data = await authStore.authedRequest((token) =>
      listStudentLedger(token, paymentDetail.value!.studentId, { page, pageSize: ledger.pageSize.value }),
    )
    ledger.setData(data)
  } catch (error) {
    ledgerError.value = error
  } finally {
    ledger.loading.value = false
  }
}

async function selectPayment(paymentId: number): Promise<void> {
  selectedPaymentId.value = paymentId
  detailError.value = null
  paymentDetail.value = null
  try {
    const detail = await authStore.authedRequest((token) => getPayment(token, paymentId))
    paymentDetail.value = detail
    await ensureReferences([detail])
    await Promise.all([loadAttachmentsPage(1), loadLedgerPage(1)])
  } catch (error) {
    detailError.value = error
  }
}

async function createNewPayment(payload: PaymentWrite): Promise<void> {
  formSubmitting.value = true
  formErrorKey.value = null
  formSuccessKey.value = null
  try {
    const created = await authStore.authedRequest((token) => createPayment(token, payload))
    formSuccessKey.value = 'financePayments.createSuccess'
    formResetVersion.value += 1
    await loadPaymentsPage(1)
    await selectPayment(created.id)
  } catch (error) {
    formErrorKey.value = mapError(error)
  } finally {
    formSubmitting.value = false
  }
}

async function confirmVoid(reason: string): Promise<void> {
  if (!selectedPaymentId.value) {
    return
  }
  voidPending.value = true
  voidErrorKey.value = null
  try {
    await authStore.authedRequest((token) => voidPayment(token, selectedPaymentId.value!, { reason }))
    voidDialogOpen.value = false
    await loadPaymentsPage(payments.page.value)
    await selectPayment(selectedPaymentId.value)
  } catch (error) {
    voidErrorKey.value = mapError(error)
  } finally {
    voidPending.value = false
  }
}

async function uploadAttachment(file: File): Promise<void> {
  if (!selectedPaymentId.value) {
    return
  }
  uploadPending.value = true
  uploadErrorKey.value = null
  uploadSuccessKey.value = null
  try {
    await authStore.authedRequest((token) => uploadPaymentAttachment(token, selectedPaymentId.value!, file))
    uploadSuccessKey.value = 'financePayments.attachmentUploadSuccess'
    await loadAttachmentsPage(1)
  } catch (error) {
    uploadErrorKey.value = mapError(error)
  } finally {
    uploadPending.value = false
  }
}

async function downloadAttachment(item: PaymentAttachment): Promise<void> {
  if (!selectedPaymentId.value) {
    return
  }
  downloadPendingId.value = item.id
  try {
    const payload = await authStore.authedRequest((token) =>
      downloadPaymentAttachment(token, selectedPaymentId.value!, item.id),
    )
    const url = URL.createObjectURL(payload.blob)
    const link = document.createElement('a')
    link.href = url
    link.download = payload.fileName
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } finally {
    downloadPendingId.value = null
  }
}

async function loadPage(): Promise<void> {
  loading.value = true
  loadError.value = null
  try {
    await Promise.all([
      loadDictionary(),
      loadStudentsList(),
      authStore.authedRequest((token) => getBaseCurrency(token)).then((value) => {
        baseCurrency.value = value
      }),
      authStore.authedRequest((token) => listPaymentMethods(token)).then((items) => {
        paymentMethods.value = items
      }),
      loadPaymentsPage(1),
    ])
    if (payments.items.value[0]) {
      await selectPayment(payments.items.value[0].id)
    }
  } catch (error) {
    loadError.value = error
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadPage()
})
</script>

<template>
  <div
    class="finance-payments-view"
    data-testid="finance-payments-view"
  >
    <h1>{{ t('financePayments.title') }}</h1>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="loadError"
      :error="loadError"
      @retry="loadPage"
    />
    <template v-else-if="baseCurrency">
      <div class="finance-layout">
        <FinancePaymentForm
          :base-currency="baseCurrency"
          :payment-methods="paymentMethods"
          :students="students"
          :enrollments="formEnrollments"
          :submitting="formSubmitting"
          :error-key="formErrorKey"
          :success-key="formSuccessKey"
          :reset-version="formResetVersion"
          @student-change="loadFormEnrollments"
          @submit="createNewPayment"
        />

        <form
          class="payment-filter"
          data-testid="payment-filter"
          @submit.prevent="applyPaymentFilter"
        >
          <label>
            {{ t('financePayments.paymentNo') }}
            <input
              v-model.trim="paymentFilter.paymentNo"
              type="search"
            >
          </label>
          <label>
            {{ t('common.status') }}
            <select v-model="paymentFilter.status">
              <option value="">{{ t('common.all') }}</option>
              <option value="CONFIRMED">{{ t('enrollments.statusActive') }}</option>
              <option value="VOIDED">{{ t('enrollments.statusCancelled') }}</option>
            </select>
          </label>
          <button type="submit">
            {{ t('common.search') }}
          </button>
        </form>

        <FinancePaymentsTable
          :payments="payments.items.value"
          :selected-payment-id="selectedPaymentId"
          :page="payments.page.value"
          :page-size="payments.pageSize.value"
          :total="payments.total.value"
          :loading="payments.loading.value"
          :error="paymentsError"
          :student-name="studentName"
          :enrollment-label="enrollmentLabel"
          @retry="loadPaymentsPage"
          @page-change="loadPaymentsPage"
          @select="selectPayment"
        />

        <ErrorState
          v-if="detailError"
          :error="detailError"
          @retry="selectedPaymentId ? selectPayment(selectedPaymentId) : undefined"
        />

        <PaymentDetailPanel
          v-else
          :payment="paymentDetail"
          :student-name="paymentDetail ? studentName(paymentDetail.studentId) : '-'"
          :enrollment-label="paymentDetail ? enrollmentLabel(paymentDetail.enrollmentId) : '-'"
          :attachments="attachments.items.value"
          :attachments-page="attachments.page.value"
          :attachments-page-size="attachments.pageSize.value"
          :attachments-total="attachments.total.value"
          :attachments-loading="attachments.loading.value"
          :attachments-error="attachmentsError"
          :upload-pending="uploadPending"
          :upload-error-key="uploadErrorKey"
          :upload-success-key="uploadSuccessKey"
          :download-pending-id="downloadPendingId"
          @attachment-page-change="loadAttachmentsPage"
          @download="downloadAttachment"
          @request-void="voidDialogOpen = true; voidErrorKey = null"
          @retry-attachments="loadAttachmentsPage"
          @upload-file="uploadAttachment"
        />

        <StudentLedgerPanel
          :student-name="paymentDetail ? studentName(paymentDetail.studentId) : '-'"
          :ledger="ledger.items.value"
          :page="ledger.page.value"
          :page-size="ledger.pageSize.value"
          :total="ledger.total.value"
          :loading="ledger.loading.value"
          :error="ledgerError"
          @page-change="loadLedgerPage"
          @retry="loadLedgerPage"
        />
      </div>
    </template>

    <PaymentVoidDialog
      :open="voidDialogOpen"
      :pending="voidPending"
      :error-key="voidErrorKey"
      @cancel="voidDialogOpen = false"
      @confirm="confirmVoid"
    />
  </div>
</template>

<style scoped>
.finance-payments-view {
  max-width: 1100px;
  margin: 0 auto;
  padding: 1rem;
}

.finance-layout {
  display: grid;
  gap: 1rem;
}

.payment-filter {
  display: flex;
  gap: 0.75rem;
  align-items: end;
  flex-wrap: wrap;
}

.payment-filter label {
  display: grid;
  gap: 0.25rem;
}

h1 {
  margin-bottom: 1rem;
}
</style>
