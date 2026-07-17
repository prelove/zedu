<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import {
  getEnrollment, updateEnrollment,
  listAssignments, createAssignment, endAssignment,
  listDomains, listTracks, listLevels,
  type Enrollment, type Assignment, type EnrollmentWrite, type AssignmentWrite,
  type CourseDomain, type Track, type Level,
} from '../../api/course'
import { listTeachers, type Teacher } from '../../api/directory'
import { ApiError, NetworkError } from '../../api/http'
import { errorToI18nKey } from '../../api/error-mapping'
import LoadingState from '../../components/LoadingState.vue'
import ErrorState from '../../components/ErrorState.vue'
import EmptyState from '../../components/EmptyState.vue'
import ConfirmDialog from '../../components/ConfirmDialog.vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const { t } = useI18n()

const enrollment = ref<Enrollment | null>(null)
const loading = ref(false)
const error = ref<unknown>(null)

// Course selection form (domain/track/targetLevel — NOT currentLevel).
const selDomain = ref(0)
const selTrack = ref(0)
const selTargetLevel = ref(0)
const savingCourse = ref(false)
const courseError = ref<string | null>(null)

// Level change form (currentLevel only — NOT domain/track/targetLevel).
const selCurrentLevel = ref(0)
const savingLevel = ref(false)
const levelError = ref<string | null>(null)

// Assignments.
const assignments = ref<Assignment[]>([])
const assignmentsLoading = ref(false)
const assignmentsError = ref<unknown>(null)
const showAssignForm = ref(false)
const assignForm = ref<AssignmentWrite>({ teacherId: 0, roleType: 'MAIN' })
const assignSaving = ref(false)
const assignError = ref<string | null>(null)
const assignToEnd = ref<Assignment | null>(null)

// Dictionary.
const domains = ref<CourseDomain[]>([])
const tracks = ref<Track[]>([])
const levels = ref<Level[]>([])
const teachers = ref<Teacher[]>([])

const enrollmentId = () => Number(props.id)

const hasActiveAssignment = computed(() =>
  assignments.value.some((a) => a.status === 'ACTIVE'),
)

async function loadEnrollment(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) => getEnrollment(token, enrollmentId()))
    enrollment.value = data
    selDomain.value = data.domainId
    selTrack.value = data.trackId
    selTargetLevel.value = data.targetLevelId ?? 0
    selCurrentLevel.value = data.currentLevelId ?? 0
  } catch (err) {
    error.value = err
  } finally {
    loading.value = false
  }
}

async function loadAssignments(): Promise<void> {
  assignmentsLoading.value = true
  assignmentsError.value = null
  try {
    const data = await authStore.authedRequest((token) => listAssignments(token, enrollmentId(), { pageSize: 100 }))
    assignments.value = data.items ?? []
  } catch (err) {
    assignmentsError.value = err
  } finally {
    assignmentsLoading.value = false
  }
}

async function loadDictionary(): Promise<void> {
  try {
    const [dom, trk, lvl, tch] = await Promise.all([
      authStore.authedRequest((token) => listDomains(token, { pageSize: 100 })),
      authStore.authedRequest((token) => listTracks(token, { pageSize: 100 })),
      authStore.authedRequest((token) => listLevels(token, { pageSize: 100 })),
      authStore.authedRequest((token) => listTeachers(token, { pageSize: 100 })),
    ])
    domains.value = dom.items ?? []
    tracks.value = trk.items ?? []
    levels.value = lvl.items ?? []
    teachers.value = tch.items ?? []
  } catch {
    // Non-fatal.
  }
}

onMounted(() => {
  void loadEnrollment()
  void loadAssignments()
  void loadDictionary()
})

// Filtered tracks/levels for cascading selectors.
const filteredTracks = computed(() => tracks.value.filter((tr) => tr.domainId === selDomain.value))
const filteredLevels = computed(() => levels.value.filter((lv) => lv.trackId === selTrack.value))

// Course selection PATCH: domainId, trackId, targetLevelId — NO currentLevelId.
async function handleSaveCourseSelection(): Promise<void> {
  savingCourse.value = true
  courseError.value = null
  try {
    const body: EnrollmentWrite = {
      domainId: selDomain.value,
      trackId: selTrack.value,
    }
    if (selTargetLevel.value) body.targetLevelId = selTargetLevel.value
    // Deliberately NOT including currentLevelId — course selection and level
    // change must be separate PATCHes per the domain invariant.
    const updated = await authStore.authedRequest((token) => updateEnrollment(token, enrollmentId(), body))
    enrollment.value = updated
  } catch (err) {
    if (err instanceof NetworkError) courseError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) courseError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else courseError.value = 'errors.UNKNOWN'
  } finally {
    savingCourse.value = false
  }
}

// Level change PATCH: currentLevelId only — NO domainId/trackId/targetLevelId.
async function handleSaveLevelChange(): Promise<void> {
  savingLevel.value = true
  levelError.value = null
  try {
    if (selCurrentLevel.value === (enrollment.value?.currentLevelId ?? 0)) {
      levelError.value = 'enrollments.sameLevelRejected'
      return
    }
    const body: EnrollmentWrite = { currentLevelId: selCurrentLevel.value }
    // Deliberately NOT including domainId/trackId/targetLevelId.
    const updated = await authStore.authedRequest((token) => updateEnrollment(token, enrollmentId(), body))
    enrollment.value = updated
  } catch (err) {
    if (err instanceof NetworkError) levelError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) levelError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else levelError.value = 'errors.UNKNOWN'
  } finally {
    savingLevel.value = false
  }
}

function openAssignForm(): void {
  showAssignForm.value = true
  assignError.value = null
  assignForm.value = { teacherId: 0, roleType: 'MAIN' }
}

// Create assignment — atomic replace if an ACTIVE assignment exists.
async function handleCreateAssignment(): Promise<void> {
  assignSaving.value = true
  assignError.value = null
  try {
    if (!assignForm.value.teacherId) {
      assignError.value = 'errors.UNKNOWN'
      return
    }
    // POST /enrollments/{id}/assignments handles both first assignment and
    // atomic replacement (ending old + creating new in one transaction).
    await authStore.authedRequest((token) => createAssignment(token, enrollmentId(), assignForm.value))
    showAssignForm.value = false
    await loadAssignments()
  } catch (err) {
    if (err instanceof NetworkError) assignError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) assignError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else assignError.value = 'errors.UNKNOWN'
  } finally {
    assignSaving.value = false
  }
}

async function handleEndAssignment(): Promise<void> {
  if (!assignToEnd.value) return
  const a = assignToEnd.value
  assignToEnd.value = null
  try {
    await authStore.authedRequest((token) => endAssignment(token, a.id, {}))
    await loadAssignments()
  } catch (err) {
    if (err instanceof ApiError) {
      assignmentsError.value = err
    } else if (err instanceof NetworkError) {
      assignmentsError.value = err
    } else {
      assignmentsError.value = err
    }
  }
}

function teacherName(id: number): string {
  return teachers.value.find((tc) => tc.id === id)?.name ?? String(id)
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    ACTIVE: t('enrollments.statusActive'),
    PAUSED: t('enrollments.statusPaused'),
    COMPLETED: t('enrollments.statusCompleted'),
    CANCELLED: t('enrollments.statusCancelled'),
  }
  return map[status] ?? status
}

function assignmentStatusLabel(status: string): string {
  return status === 'ACTIVE' ? t('enrollments.statusActive') : status
}

function roleLabel(role: string): string {
  const map: Record<string, string> = {
    MAIN: t('enrollments.roleTypeMain'),
    ASSISTANT: t('enrollments.roleTypeAssistant'),
    OBSERVER: t('enrollments.roleTypeObserver'),
  }
  return map[role] ?? role
}
</script>

<template>
  <div
    class="enrollment-detail-view"
    data-testid="enrollment-detail-view"
  >
    <button
      type="button"
      data-testid="back-from-enrollment"
      @click="router.back()"
    >
      {{ t('common.back') }}
    </button>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadEnrollment"
    />
    <template v-else-if="enrollment">
      <h1>{{ t('enrollments.title') }} #{{ enrollment.id }}</h1>

      <!-- Enrollment info -->
      <section
        class="info-section"
        data-testid="enrollment-info-section"
      >
        <p><strong>{{ t('enrollments.status') }}:</strong> {{ statusLabel(enrollment.status) }}</p>
        <p><strong>{{ t('enrollments.enrollmentType') }}:</strong> {{ enrollment.enrollmentType }}</p>
        <p v-if="enrollment.startedAt">
          <strong>{{ t('enrollments.startedAt') }}:</strong> {{ enrollment.startedAt }}
        </p>
        <p v-if="enrollment.note">
          <strong>{{ t('enrollments.note') }}:</strong> {{ enrollment.note }}
        </p>
      </section>

      <!-- Course selection section (domain/track/targetLevel — NOT currentLevel) -->
      <section
        class="course-section"
        data-testid="course-selection-section"
      >
        <h2>{{ t('enrollments.domain') }} / {{ t('enrollments.track') }} / {{ t('enrollments.targetLevel') }}</h2>
        <p class="hint">
          {{ t('enrollments.courseSelectionHint') }}
        </p>
        <div class="form-grid">
          <div class="form-field">
            <label for="sel-domain">{{ t('enrollments.domain') }}</label>
            <select
              id="sel-domain"
              v-model.number="selDomain"
              data-testid="sel-domain"
              @change="selTrack = 0; selTargetLevel = 0"
            >
              <option
                v-for="d in domains"
                :key="d.id"
                :value="d.id"
              >
                {{ d.name }}
              </option>
            </select>
          </div>
          <div class="form-field">
            <label for="sel-track">{{ t('enrollments.track') }}</label>
            <select
              id="sel-track"
              v-model.number="selTrack"
              :disabled="!selDomain"
              data-testid="sel-track"
              @change="selTargetLevel = 0"
            >
              <option :value="0">
                —
              </option>
              <option
                v-for="tr in filteredTracks"
                :key="tr.id"
                :value="tr.id"
              >
                {{ tr.name }}
              </option>
            </select>
          </div>
          <div class="form-field">
            <label for="sel-target">{{ t('enrollments.targetLevel') }}</label>
            <select
              id="sel-target"
              v-model.number="selTargetLevel"
              :disabled="!selTrack"
              data-testid="sel-target-level"
            >
              <option :value="0">
                —
              </option>
              <option
                v-for="lv in filteredLevels"
                :key="lv.id"
                :value="lv.id"
              >
                {{ lv.name }}
              </option>
            </select>
          </div>
        </div>
        <p
          v-if="courseError"
          class="form-error"
          role="alert"
          aria-live="assertive"
          data-testid="course-selection-error"
        >
          {{ t(courseError) }}
        </p>
        <button
          type="button"
          :disabled="savingCourse || !selDomain || !selTrack"
          data-testid="save-course-selection"
          @click="handleSaveCourseSelection"
        >
          {{ savingCourse ? t('common.saving') : t('enrollments.saveCourseSelection') }}
        </button>
      </section>

      <!-- Level change section (currentLevel only — NOT domain/track/targetLevel) -->
      <section
        class="level-section"
        data-testid="level-change-section"
      >
        <h2>{{ t('enrollments.currentLevel') }}</h2>
        <p class="hint">
          {{ t('enrollments.levelChangeHint') }}
        </p>
        <div class="form-field">
          <label for="sel-current">{{ t('enrollments.currentLevel') }}</label>
          <select
            id="sel-current"
            v-model.number="selCurrentLevel"
            data-testid="sel-current-level"
          >
            <option :value="0">
              —
            </option>
            <option
              v-for="lv in levels.filter((l) => l.trackId === enrollment.trackId)"
              :key="lv.id"
              :value="lv.id"
            >
              {{ lv.name }}
            </option>
          </select>
        </div>
        <p
          v-if="levelError"
          class="form-error"
          role="alert"
          aria-live="assertive"
          data-testid="level-change-error"
        >
          {{ t(levelError) }}
        </p>
        <button
          type="button"
          :disabled="savingLevel"
          data-testid="save-level-change"
          @click="handleSaveLevelChange"
        >
          {{ savingLevel ? t('common.saving') : t('enrollments.saveLevelChange') }}
        </button>
      </section>

      <!-- Assignments section -->
      <section
        class="assignments-section"
        data-testid="assignments-section"
      >
        <h2>{{ t('enrollments.assignments') }}</h2>
        <LoadingState v-if="assignmentsLoading" />
        <ErrorState
          v-else-if="assignmentsError"
          :error="assignmentsError"
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
              <td>{{ assignmentStatusLabel(a.status) }}</td>
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
        <button
          type="button"
          data-testid="add-assignment-btn"
          @click="openAssignForm"
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
          v-if="showAssignForm"
          class="create-dialog"
          data-testid="assignment-create-form"
        >
          <form @submit.prevent="handleCreateAssignment">
            <div class="form-field">
              <label for="assign-teacher">{{ t('enrollments.assignmentTeacher') }} *</label>
              <select
                id="assign-teacher"
                v-model.number="assignForm.teacherId"
                required
                data-testid="assign-form-teacher"
              >
                <option :value="0">
                  —
                </option>
                <option
                  v-for="tc in teachers.filter((t2) => t2.status === 'ACTIVE')"
                  :key="tc.id"
                  :value="tc.id"
                >
                  {{ tc.name }}
                </option>
              </select>
            </div>
            <div class="form-field">
              <label for="assign-role">{{ t('enrollments.assignmentRoleType') }}</label>
              <select
                id="assign-role"
                v-model="assignForm.roleType"
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
              v-if="assignError"
              class="form-error"
              role="alert"
              aria-live="assertive"
              data-testid="assignment-create-error"
            >
              {{ t(assignError) }}
            </p>
            <div class="form-actions">
              <button
                type="button"
                :disabled="assignSaving"
                @click="showAssignForm = false"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="submit"
                :disabled="assignSaving || !assignForm.teacherId"
                data-testid="assignment-create-submit"
              >
                {{ assignSaving ? t('common.creating') : t('common.create') }}
              </button>
            </div>
          </form>
        </div>
      </section>
    </template>

    <!-- End assignment confirmation -->
    <ConfirmDialog
      v-if="assignToEnd"
      :title="t('enrollments.endAssignment')"
      :message="t('enrollments.endAssignmentConfirm')"
      @confirm="handleEndAssignment"
      @cancel="assignToEnd = null"
    />
  </div>
</template>

<style scoped>
.enrollment-detail-view {
  max-width: 900px;
  margin: 0 auto;
  padding: 1rem;
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

section {
  margin-top: 1.5rem;
}

.info-section p {
  margin: 0.25rem 0;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 0.75rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.75rem;
}

.form-field label {
  font-weight: 600;
}

.form-field select {
  padding: 0.4rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.hint {
  color: #6c757d;
  font-size: 0.8125rem;
  font-style: italic;
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 0.5rem;
}

th, td {
  padding: 0.5rem;
  text-align: left;
  border-bottom: 1px solid #eee;
}

th {
  font-weight: 600;
  background-color: #f8f9fa;
}

.create-dialog {
  border: 1px solid #ccc;
  border-radius: 0.5rem;
  padding: 1rem;
  margin-top: 0.5rem;
  background-color: #f8f9fa;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}
</style>
