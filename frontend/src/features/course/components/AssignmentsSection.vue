<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { createAssignment, endAssignment, listAssignments, type Assignment, type AssignmentWrite } from '../../../api/course'
import { type Teacher } from '../../../api/directory'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import LoadingState from '../../../components/LoadingState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import ConfirmDialog from '../../../components/ConfirmDialog.vue'
import { usePaginatedList } from '../../../composables/usePaginatedList'
import AssignmentCreateForm from './AssignmentCreateForm.vue'
import AssignmentsTable from './AssignmentsTable.vue'
import TeacherLoadNotice from './TeacherLoadNotice.vue'

const props = withDefaults(defineProps<{
  enrollmentId: number
  teachers: Teacher[]
  canManageAssignments?: boolean
  teacherLoadError?: string | null
}>(), {
  canManageAssignments: true,
  teacherLoadError: null,
})

const emit = defineEmits<{ retryTeachers: [] }>()

const { t } = useI18n()
const { items: assignments, page, pageSize, total, loading, error, setData } = usePaginatedList<Assignment>(20)

const showForm = ref(false)
const saving = ref(false)
const formError = ref<string | null>(null)
const assignToEnd = ref<Assignment | null>(null)
const endError = ref<string | null>(null)
const endDuplicate = ref(false)

const hasActiveAssignment = computed(() => assignments.value.some((assignment) => assignment.status === 'ACTIVE'))
const assignmentsEnabled = computed(() => props.canManageAssignments)

function toErrorKey(err: unknown): string {
  if (err instanceof NetworkError) return 'errors.NETWORK_ERROR'
  if (err instanceof ApiError) return errorToI18nKey(err) ?? 'errors.UNKNOWN'
  return 'errors.UNKNOWN'
}

async function loadAssignments(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) => listAssignments(token, props.enrollmentId, { page: page.value, pageSize: pageSize.value }))
    setData(data)
  } catch (err) {
    error.value = toErrorKey(err)
  } finally {
    loading.value = false
  }
}

async function handleCreate(body: AssignmentWrite): Promise<void> {
  saving.value = true
  formError.value = null
  try {
    await authStore.authedRequest((token) => createAssignment(token, props.enrollmentId, body))
    showForm.value = false
    await loadAssignments()
  } catch (err) {
    formError.value = toErrorKey(err)
  } finally {
    saving.value = false
  }
}

async function handleEndAssignment(): Promise<void> {
  if (!assignToEnd.value) return
  endError.value = null
  endDuplicate.value = false
  try {
    await authStore.authedRequest((token) => endAssignment(token, assignToEnd.value!.id))
    assignToEnd.value = null
    await loadAssignments()
  } catch (err) {
    if (err instanceof ApiError && err.code === 42201) endDuplicate.value = true
    else endError.value = toErrorKey(err)
  }
}

const teacherName = (teacherId: number) => props.teachers.find((teacher) => teacher.id === teacherId)?.name ?? String(teacherId)
const statusLabel = (status: string) => ({ ACTIVE: t('enrollments.statusActive'), ENDED: t('students.statusEnded') }[status] ?? status)
const roleLabel = (role: string) => ({ MAIN: t('enrollments.roleTypeMain'), ASSISTANT: t('enrollments.roleTypeAssistant'), OBSERVER: t('enrollments.roleTypeObserver') }[role] ?? role)
const openForm = () => { formError.value = null; showForm.value = true }

onMounted(() => void loadAssignments())
watch(page, () => void loadAssignments())
watch(assignmentsEnabled, (enabled) => {
  if (!enabled) showForm.value = false
})
</script>

<template>
  <section
    class="assignments-section"
    data-testid="assignments-section"
  >
    <h2>{{ t('enrollments.assignments') }}</h2>

    <TeacherLoadNotice
      v-if="teacherLoadError"
      :error-key="teacherLoadError"
      @retry="emit('retryTeachers')"
    />

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadAssignments"
    />
    <AssignmentsTable
      v-else
      :assignments="assignments"
      :page="page"
      :page-size="pageSize"
      :total="total"
      :teacher-name="teacherName"
      :status-label="statusLabel"
      :role-label="roleLabel"
      @end="(assignment) => (assignToEnd = assignment)"
      @page-change="(nextPage) => (page = nextPage)"
    />

    <button
      type="button"
      :disabled="!assignmentsEnabled"
      data-testid="add-assignment-btn"
      @click="openForm"
    >
      {{ hasActiveAssignment ? t('enrollments.replaceAssignment') : t('enrollments.addAssignment') }}
    </button>
    <p
      v-if="hasActiveAssignment"
      class="hint"
      data-testid="atomic-replace-hint"
    >
      {{ t('enrollments.atomicReplaceHint') }}
    </p>
    <AssignmentCreateForm
      v-if="showForm"
      :teachers="teachers"
      :saving="saving"
      :form-error="formError"
      @submit="handleCreate"
      @cancel="showForm = false"
    />

    <ConfirmDialog
      v-if="assignToEnd"
      :title="t('enrollments.endAssignment')"
      :message="t('enrollments.endAssignmentConfirm')"
      @confirm="handleEndAssignment"
      @cancel="assignToEnd = null"
    />
    <p
      v-if="endDuplicate"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="assignment-end-duplicate"
    >
      {{ t('enrollments.endAssignmentDuplicate') }}
    </p>
    <p
      v-if="endError"
      class="form-error"
      role="alert"
      aria-live="assertive"
    >
      {{ t(endError) }}
    </p>
  </section>
</template>

<style scoped>
.assignments-section {
  margin-top: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-error {
  margin: 0;
  color: #dc3545;
  font-size: 0.875rem;
}

button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}

button:disabled {
  color: #6c757d;
  cursor: not-allowed;
}
</style>
