<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { cancelLesson, confirmLesson, createLesson, listAttendanceOutcomes, listLessons, type AttendanceOutcome, type Lesson, type LessonWrite } from '../../api/lesson'
import { errorToI18nKey } from '../../api/error-mapping'
import ErrorState from '../../components/ErrorState.vue'
import LoadingState from '../../components/LoadingState.vue'
import PaginationBar from '../../components/PaginationBar.vue'
import { authStore } from '../../stores/auth'

const { t } = useI18n()
const loading = ref(true)
const error = ref<unknown>(null)
const submitting = ref(false)
const actionError = ref<string | null>(null)
const lessons = ref<Lesson[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const filters = reactive({ status: '' })
const confirmationTarget = ref<Lesson | null>(null)
const attendanceOutcomes = ref<AttendanceOutcome[]>([])
const confirmationForm = reactive({ outcomeType: '', lessonDeducted: '0', chargeAmount: 0, teacherPayAmount: 0, actualDurationMin: 0, note: '' })
const confirmationOutcome = computed(() => attendanceOutcomes.value.find((outcome) => outcome.code === confirmationForm.outcomeType))
const form = reactive<LessonWrite>({
  enrollmentId: 0,
  assignmentId: 0,
  startAt: '',
  durationMin: 60,
  timezone: 'Asia/Tokyo',
  meetingType: 'OFFLINE',
  meetingLink: '',
  lessonTopic: '',
  note: '',
})

function errorKey(value: unknown): string {
  return errorToI18nKey(value) ?? 'errors.UNKNOWN'
}

async function load(targetPage = page.value): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) => listLessons(token, { page: targetPage, pageSize, status: filters.status || undefined }))
    lessons.value = data.items
    total.value = data.total
    page.value = data.page
  } catch (caught) {
    error.value = caught
  } finally {
    loading.value = false
  }
}

async function submit(): Promise<void> {
  submitting.value = true
  actionError.value = null
  try {
    await authStore.authedRequest((token) => createLesson(token, { ...form }))
    await load(1)
  } catch (caught) {
    actionError.value = errorKey(caught)
  } finally {
    submitting.value = false
  }
}

async function cancel(item: Lesson): Promise<void> {
  const reason = window.prompt(t('lessons.cancelReason'))
  if (!reason?.trim()) return
  actionError.value = null
  try {
    await authStore.authedRequest((token) => cancelLesson(token, item.id, reason))
    await load(page.value)
  } catch (caught) {
    actionError.value = errorKey(caught)
  }
}

async function openConfirmation(item: Lesson): Promise<void> {
  actionError.value = null
  try {
    attendanceOutcomes.value = await authStore.authedRequest((token) => listAttendanceOutcomes(token))
    const first = attendanceOutcomes.value[0]
    if (!first) {
      actionError.value = 'errors.UNKNOWN'
      return
    }
    confirmationTarget.value = item
    confirmationForm.outcomeType = first.code
    confirmationForm.lessonDeducted = first.suggestedLessonDeducted || '0'
    confirmationForm.chargeAmount = 0
    confirmationForm.teacherPayAmount = 0
    confirmationForm.actualDurationMin = item.durationMin
    confirmationForm.note = ''
  } catch (caught) { actionError.value = errorKey(caught) }
}

function applyOutcomeSuggestion(): void {
  if (confirmationOutcome.value) confirmationForm.lessonDeducted = confirmationOutcome.value.suggestedLessonDeducted || '0'
}

async function submitConfirmation(): Promise<void> {
  if (!confirmationTarget.value) return
  submitting.value = true
  actionError.value = null
  try {
    await authStore.authedRequest((token) => confirmLesson(token, confirmationTarget.value!.id, { ...confirmationForm }))
    confirmationTarget.value = null
    await load(page.value)
  } catch (caught) { actionError.value = errorKey(caught) } finally { submitting.value = false }
}

onMounted(() => { void load() })
</script>

<template>
  <main
    class="lessons-view"
    data-testid="lessons-view"
  >
    <h1>{{ t('lessons.title') }}</h1>
    <form
      class="lesson-form"
      @submit.prevent="submit"
    >
      <label>{{ t('lessons.enrollmentId') }}<input
        v-model.number="form.enrollmentId"
        type="number"
        min="1"
        required
      ></label>
      <label>{{ t('lessons.assignmentId') }}<input
        v-model.number="form.assignmentId"
        type="number"
        min="1"
        required
      ></label>
      <label>{{ t('lessons.startAt') }}<input
        v-model="form.startAt"
        type="datetime-local"
        step="1"
        required
      ></label>
      <label>{{ t('lessons.duration') }}<input
        v-model.number="form.durationMin"
        type="number"
        min="10"
        max="480"
        required
      ></label>
      <label>{{ t('lessons.timezone') }}<input
        v-model="form.timezone"
        required
      ></label>
      <label>{{ t('lessons.meetingType') }}<select v-model="form.meetingType"><option value="OFFLINE">{{ t('lessons.offline') }}</option><option value="WECHAT">WeChat</option><option value="ONLINE">Online</option></select></label>
      <label>{{ t('lessons.meetingLink') }}<input
        v-model="form.meetingLink"
        type="url"
      ></label>
      <label>{{ t('lessons.topic') }}<input v-model="form.lessonTopic"></label>
      <label>{{ t('common.note') }}<input v-model="form.note"></label>
      <button
        type="submit"
        :disabled="submitting"
      >
        {{ submitting ? t('common.creating') : t('lessons.create') }}
      </button>
    </form>
    <p
      v-if="actionError"
      role="alert"
      aria-live="assertive"
    >
      {{ t(actionError) }}
    </p>
    <label>{{ t('lessons.status') }}<select
      v-model="filters.status"
      @change="load(1)"
    ><option value="">{{ t('common.all') }}</option><option value="SCHEDULED">{{ t('lessons.scheduled') }}</option><option value="CANCELLED">{{ t('lessons.cancelled') }}</option><option value="COMPLETED">{{ t('lessons.completed') }}</option></select></label>
    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="load()"
    />
    <template v-else>
      <p v-if="lessons.length === 0">
        {{ t('lessons.empty') }}
      </p>
      <table v-else>
        <thead><tr><th>{{ t('lessons.startAt') }}</th><th>{{ t('lessons.topic') }}</th><th>{{ t('lessons.status') }}</th><th>{{ t('common.actions') }}</th></tr></thead>
        <tbody>
          <tr
            v-for="item in lessons"
            :key="item.id"
          >
            <td>{{ item.scheduledStartAt }}</td><td>{{ item.lessonTopic || '-' }}</td><td>{{ item.status }}</td><td>
              <button
                v-if="item.status === 'SCHEDULED'"
                type="button"
                @click="cancel(item)"
              >
                {{ t('lessons.cancel') }}
              </button>
              <button
                v-if="item.status === 'SCHEDULED'"
                type="button"
                data-testid="lesson-confirm-open"
                @click="openConfirmation(item)"
              >
                {{ t('lessons.confirm') }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <PaginationBar
        :page="page"
        :page-size="pageSize"
        :total="total"
        @page-change="load"
      />
    </template>
    <section
      v-if="confirmationTarget"
      class="confirmation-dialog"
      data-testid="lesson-confirm-dialog"
      role="dialog"
      aria-modal="true"
      :aria-label="t('lessons.confirmTitle')"
    >
      <form @submit.prevent="submitConfirmation">
        <h2>{{ t('lessons.confirmTitle') }}</h2>
        <label>{{ t('lessons.outcome') }}<select
          v-model="confirmationForm.outcomeType"
          data-testid="lesson-confirm-outcome"
          required
          @change="applyOutcomeSuggestion"
        ><option
          v-for="outcome in attendanceOutcomes"
          :key="outcome.code"
          :value="outcome.code"
        >{{ outcome.name }} ({{ outcome.code }})</option></select></label>
        <label>{{ t('lessons.actualDuration') }}<input
          v-model.number="confirmationForm.actualDurationMin"
          data-testid="lesson-confirm-duration"
          type="number"
          min="0"
          :max="confirmationTarget.durationMin * 2"
          required
        ></label>
        <label>{{ t('lessons.lessonDeducted') }}<input
          v-model="confirmationForm.lessonDeducted"
          data-testid="lesson-confirm-deducted"
          inputmode="decimal"
          pattern="\d+(\.\d{1,3})?"
          required
        ></label>
        <label>{{ t('lessons.chargeAmount') }}<input
          v-model.number="confirmationForm.chargeAmount"
          data-testid="lesson-confirm-charge"
          type="number"
          min="0"
          step="1"
          required
        ></label>
        <label>{{ t('lessons.teacherPayAmount') }}<input
          v-model.number="confirmationForm.teacherPayAmount"
          data-testid="lesson-confirm-teacher-pay"
          type="number"
          min="0"
          step="1"
          required
        ></label>
        <label>{{ t('common.note') }}<input v-model="confirmationForm.note"></label>
        <div class="confirmation-actions">
          <button
            type="button"
            @click="confirmationTarget = null"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            data-testid="lesson-confirm-submit"
            :disabled="submitting"
          >
            {{ submitting ? t('common.saving') : t('lessons.confirm') }}
          </button>
        </div>
      </form>
    </section>
  </main>
</template>

<style scoped>
.lessons-view { max-width: 1100px; margin: 0 auto; padding: 1rem; }
.lesson-form { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: .75rem; margin-bottom: 1rem; }
label { display: grid; gap: .25rem; } input, select, button { padding: .4rem; }
table { width: 100%; border-collapse: collapse; margin-top: 1rem; } th, td { border: 1px solid #ddd; padding: .5rem; text-align: left; }
.confirmation-dialog { position: fixed; inset: 0; z-index: 10; display: grid; place-items: center; padding: 1rem; background: rgb(0 0 0 / 45%); }
.confirmation-dialog form { display: grid; gap: .75rem; width: min(100%, 460px); padding: 1rem; border-radius: .4rem; background: #fff; }
.confirmation-actions { display: flex; justify-content: flex-end; gap: .5rem; }
</style>
