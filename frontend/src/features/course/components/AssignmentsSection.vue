<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { listAssignments, createAssignment, endAssignment, type Assignment, type AssignmentWrite } from '../../../api/course'
import { type Teacher } from '../../../api/directory'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import LoadingState from '../../../components/LoadingState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import ConfirmDialog from '../../../components/ConfirmDialog.vue'
import { usePaginatedList } from '../../../composables/usePaginatedList'

const props = defineProps<{ enrollmentId: number; teachers: Teacher[] }>()
const { t } = useI18n()
const { items: assignments, page, pageSize, total, loading, error, setData } = usePaginatedList<Assignment>(20)

const showForm = ref(false)
const form = ref<AssignmentWrite>({ teacherId: 0, roleType: 'MAIN' })
const saving = ref(false)
const formError = ref<string | null>(null)
const assignToEnd = ref<Assignment | null>(null)
const endError = ref<string | null>(null)
const endDuplicate = ref(false)

const hasActiveAssignment = computed(() => assignments.value.some((a) => a.status === 'ACTIVE'))
const activeTeachers = computed(() => props.teachers.filter((tc) => tc.status === 'ACTIVE'))

async function loadAssignments() {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) => listAssignments(token, props.enrollmentId, { page: page.value, pageSize: pageSize.value }))
    setData(data)
  } catch (err) {
    if (err instanceof NetworkError) error.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) error.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else error.value = 'errors.UNKNOWN'
  } finally {
    loading.value = false
  }
}

onMounted(loadAssignments)
watch(page, loadAssignments)

function openForm() {
  showForm.value = true
  formError.value = null
  form.value = { teacherId: 0, roleType: 'MAIN' }
}

async function handleCreate() {
  saving.value = true
  formError.value = null
  try {
    await authStore.authedRequest((token) => createAssignment(token, props.enrollmentId, form.value))
    showForm.value = false
    await loadAssignments()
  } catch (err) {
    if (err instanceof NetworkError) formError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) formError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else formError.value = 'errors.UNKNOWN'
  } finally {
    saving.value = false
  }
}

async function handleEndAssignment() {
  if (!assignToEnd.value) return
  endError.value = null
  endDuplicate.value = false
  try {
    await authStore.authedRequest((token) => endAssignment(token, assignToEnd.value.id))
    assignToEnd.value = null
    await loadAssignments()
  } catch (err) {
    if (err instanceof ApiError && err.code === 42201) endDuplicate.value = true
    else if (err instanceof NetworkError) endError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) endError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else endError.value = 'errors.UNKNOWN'
  }
}

function teacherName(id: number): string { return props.teachers.find((tc) => tc.id === id)?.name ?? String(id) }
function statusLabel(status: string): string {
  const map: Record<string, string> = { ACTIVE: t('enrollments.statusActive'), ENDED: t('students.statusEnded') }
  return map[status] ?? status
}
function roleLabel(role: string): string {
  const map: Record<string, string> = { MAIN: t('enrollments.roleTypeMain'), ASSISTANT: t('enrollments.roleTypeAssistant'), OBSERVER: t('enrollments.roleTypeObserver') }
  return map[role] ?? role
}
</script>

<template>
  <section
    class="assignments-section"
    data-testid="assignments-section"
  >
    <h2>{{ t('enrollments.assignments') }}</h2>
    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadAssignments"
    />
    <EmptyState
      v-else-if="assignments.length === 0"
      :message="t('enrollments.noTeacherEnrollment')"
    />
    <table
      v-else
      data-testid="assignments-table"
    >
      <thead>
        <tr>
          <th>{{ t('enrollments.assignmentTeacher') }}</th>
          <th>{{ t('enrollments.assignmentRoleType') }}</th>
          <th>{{ t('enrollments.assignmentStatus') }}</th>
          <th>{{ t('enrollments.assignmentStartDate') }}</th>
          <th>{{ t('enrollments.assignmentEndDate') }}</th>
          <th>{{ t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="a in assignments"
          :key="a.id"
          data-testid="assignment-row"
        >
          <td>{{ teacherName(a.teacherId) }}</td>
          <td>{{ roleLabel(a.roleType) }}</td>
          <td>{{ statusLabel(a.status) }}</td>
          <td>{{ a.startDate }}</td>
          <td>{{ a.endDate ?? '—' }}</td>
          <td>
            <button
              v-if="a.status === 'ACTIVE'"
              type="button"
              data-testid="end-assignment-btn"
              @click="assignToEnd = a"
            >
              {{ t('enrollments.endAssignment') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
    <PaginationBar
      v-if="!loading && !error && total > 0"
      :page="page"
      :page-size="pageSize"
      :total="total"
      data-testid="assignments-pagination"
      @page-change="(p) => (page = p)"
    />
    <button
      type="button"
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
    <div
      v-if="showForm"
      class="create-form"
      data-testid="assignment-create-form"
    >
      <form @submit.prevent="handleCreate">
        <div class="form-field">
          <label for="assign-form-teacher">{{ t('enrollments.assignmentTeacher') }} *</label>
          <select
            id="assign-form-teacher"
            v-model.number="form.teacherId"
            required
            data-testid="assign-form-teacher"
          >
            <option :value="0">
              —
            </option>
            <option
              v-for="tc in activeTeachers"
              :key="tc.id"
              :value="tc.id"
            >
              {{ tc.name }}
            </option>
          </select>
        </div>
        <div class="form-field">
          <label for="assign-form-role">{{ t('enrollments.assignmentRoleType') }}</label>
          <select
            id="assign-form-role"
            v-model="form.roleType"
            data-testid="assign-form-role"
          >
            <option value="MAIN">
              {{ t('enrollments.roleTypeMain') }}
            </option>
            <option value="ASSISTANT">
              {{ t('enrollments.roleTypeAssistant') }}
            </option>
            <option value="OBSERVER">
              {{ t('enrollments.roleTypeObserver') }}
            </option>
          </select>
        </div>
        <p
          v-if="formError"
          class="form-error"
          role="alert"
          aria-live="assertive"
          data-testid="assignment-create-error"
        >
          {{ t(formError) }}
        </p>
        <div class="form-actions">
          <button
            type="button"
            :disabled="saving"
            @click="showForm = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="saving || !form.teacherId"
            data-testid="assignment-create-submit"
          >
            {{ saving ? t('common.creating') : t('common.create') }}
          </button>
        </div>
      </form>
    </div>
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
.assignments-section { margin-top: 1.5rem; }
.assignments-list { list-style: none; padding: 0; margin: 0 0 0.5rem; }
.assignments-list li { display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem; border-bottom: 1px solid #eee; }
.a-teacher { font-weight: 600; }
.a-status { color: #6c757d; font-size: 0.875rem; }
button { padding: 0.4rem 0.8rem; border: 1px solid #ccc; border-radius: 0.25rem; background-color: #fff; cursor: pointer; }
button:disabled { color: #6c757d; cursor: not-allowed; }
.form-field { display: flex; flex-direction: column; gap: 0.25rem; margin-bottom: 0.75rem; }
.form-field label { font-weight: 600; }
.form-field select { padding: 0.4rem; border: 1px solid #ccc; border-radius: 0.25rem; }
.create-form { border: 1px solid #ccc; border-radius: 0.5rem; padding: 1rem; margin-top: 0.5rem; background-color: #f8f9fa; }
.form-actions { display: flex; gap: 0.5rem; justify-content: flex-end; }
.form-error { color: #dc3545; font-size: 0.875rem; }
</style>
