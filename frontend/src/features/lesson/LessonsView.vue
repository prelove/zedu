<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { cancelLesson, confirmLesson, listAttendanceOutcomes, listLessons, type Lesson, type LessonWrite } from '../../api/lesson'
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

async function confirm(item: Lesson): Promise<void> {
  const outcomes = await authStore.authedRequest((token) => listAttendanceOutcomes(token))
  const outcomeType = window.prompt(`Attendance outcome (${outcomes.map((outcome) => outcome.code).join(', ')})`, outcomes[0]?.code ?? '')
  if (!outcomeType) return
  try {
    await authStore.authedRequest((token) => confirmLesson(token, item.id, { outcomeType, lessonDeducted: '1', chargeAmount: 0, teacherPayAmount: 0, actualDurationMin: item.durationMin }))
    await load(page.value)
  } catch (caught) { actionError.value = errorKey(caught) }
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
                @click="confirm(item)"
              >
                Confirm
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
  </main>
</template>

<style scoped>
.lessons-view { max-width: 1100px; margin: 0 auto; padding: 1rem; }
.lesson-form { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: .75rem; margin-bottom: 1rem; }
label { display: grid; gap: .25rem; } input, select, button { padding: .4rem; }
table { width: 100%; border-collapse: collapse; margin-top: 1rem; } th, td { border: 1px solid #ddd; padding: .5rem; text-align: left; }
</style>
