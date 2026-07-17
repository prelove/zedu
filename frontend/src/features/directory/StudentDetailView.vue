<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import {
  getStudent, updateStudent, listParents, createParent,
  type Student, type Parent, type Enrollment, type StudentWrite, type ParentWrite,
} from '../../api/directory'
import { listEnrollments as listEnrollmentsCourse } from '../../api/course'
import { ApiError, NetworkError } from '../../api/http'
import { errorToI18nKey } from '../../api/error-mapping'
import LoadingState from '../../components/LoadingState.vue'
import ErrorState from '../../components/ErrorState.vue'
import EmptyState from '../../components/EmptyState.vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const { t } = useI18n()

const student = ref<Student | null>(null)
const loading = ref(false)
const error = ref<unknown>(null)

// Edit form.
const editForm = ref<StudentWrite>({})
const saving = ref(false)
const saveError = ref<string | null>(null)

// Parents.
const parents = ref<Parent[]>([])
const parentsLoading = ref(false)
const parentsError = ref<unknown>(null)
const showParentForm = ref(false)
const parentForm = ref<ParentWrite>({ name: '', email: '', phone: '', relationship: '', isPrimary: false })
const parentSaving = ref(false)
const parentError = ref<string | null>(null)

// Enrollments.
const enrollments = ref<Enrollment[]>([])
const enrollmentsLoading = ref(false)
const enrollmentsError = ref<unknown>(null)

const studentId = () => Number(props.id)

async function loadStudent(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) => getStudent(token, studentId()))
    student.value = data
    // Initialize edit form from loaded data.
    editForm.value = {
      name: data.name,
      nameLocal: data.nameLocal ?? '',
      email: data.email ?? '',
      phone: data.phone ?? '',
      nationality: data.nationality ?? '',
      timezone: data.timezone,
      status: data.status,
      sourceChannel: data.sourceChannel ?? '',
      note: data.note ?? '',
    }
  } catch (err) {
    error.value = err
  } finally {
    loading.value = false
  }
}

async function loadParents(): Promise<void> {
  parentsLoading.value = true
  parentsError.value = null
  try {
    const data = await authStore.authedRequest((token) => listParents(token, studentId(), { pageSize: 100 }))
    parents.value = data.items ?? []
  } catch (err) {
    parentsError.value = err
  } finally {
    parentsLoading.value = false
  }
}

async function loadEnrollments(): Promise<void> {
  enrollmentsLoading.value = true
  enrollmentsError.value = null
  try {
    const data = await authStore.authedRequest((token) => listEnrollmentsCourse(token, studentId(), { pageSize: 100 }))
    enrollments.value = data.items ?? []
  } catch (err) {
    enrollmentsError.value = err
  } finally {
    enrollmentsLoading.value = false
  }
}

onMounted(() => {
  void loadStudent()
  void loadParents()
  void loadEnrollments()
})

async function handleSave(): Promise<void> {
  saving.value = true
  saveError.value = null
  try {
    // Only send edited (non-empty) fields; empty email = clear email.
    const body: StudentWrite = {}
    const s = student.value!
    if (editForm.value.name !== s.name) body.name = editForm.value.name
    if ((editForm.value.email ?? '') !== (s.email ?? '')) body.email = editForm.value.email ?? ''
    if ((editForm.value.nameLocal ?? '') !== (s.nameLocal ?? '')) body.nameLocal = editForm.value.nameLocal
    if ((editForm.value.phone ?? '') !== (s.phone ?? '')) body.phone = editForm.value.phone
    if ((editForm.value.nationality ?? '') !== (s.nationality ?? '')) body.nationality = editForm.value.nationality
    if ((editForm.value.timezone ?? '') !== s.timezone) body.timezone = editForm.value.timezone
    if ((editForm.value.status ?? '') !== s.status) body.status = editForm.value.status
    if ((editForm.value.sourceChannel ?? '') !== (s.sourceChannel ?? '')) body.sourceChannel = editForm.value.sourceChannel
    if ((editForm.value.note ?? '') !== (s.note ?? '')) body.note = editForm.value.note

    const updated = await authStore.authedRequest((token) => updateStudent(token, studentId(), body))
    student.value = updated
  } catch (err) {
    if (err instanceof NetworkError) saveError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) saveError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else saveError.value = 'errors.UNKNOWN'
  } finally {
    saving.value = false
  }
}

function openParentForm(): void {
  showParentForm.value = true
  parentError.value = null
  parentForm.value = { name: '', email: '', phone: '', relationship: '', isPrimary: false }
}

async function handleCreateParent(): Promise<void> {
  parentSaving.value = true
  parentError.value = null
  try {
    const body: ParentWrite = { ...parentForm.value }
    if (!body.email) delete body.email
    if (!body.phone) delete body.phone
    if (!body.relationship) delete body.relationship
    await authStore.authedRequest((token) => createParent(token, studentId(), body))
    showParentForm.value = false
    await loadParents()
  } catch (err) {
    if (err instanceof NetworkError) parentError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) parentError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else parentError.value = 'errors.UNKNOWN'
  } finally {
    parentSaving.value = false
  }
}

function viewEnrollment(id: number): void {
  void router.push({ name: 'enrollment-detail', params: { id } })
}

function enrollmentStatusLabel(status: string): string {
  const map: Record<string, string> = {
    ACTIVE: t('enrollments.statusActive'),
    PAUSED: t('enrollments.statusPaused'),
    COMPLETED: t('enrollments.statusCompleted'),
    CANCELLED: t('enrollments.statusCancelled'),
  }
  return map[status] ?? status
}
</script>

<template>
  <div
    class="student-detail-view"
    data-testid="student-detail-view"
  >
    <button
      type="button"
      data-testid="back-to-students"
      @click="router.push({ name: 'students' })"
    >
      {{ t('common.back') }}
    </button>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadStudent"
    />
    <template v-else-if="student">
      <h1>{{ t('students.detailTitle') }}: {{ student.name }}</h1>

      <!-- Edit form -->
      <section
        class="edit-section"
        data-testid="student-edit-section"
      >
        <h2>{{ t('students.editTitle') }}</h2>
        <form @submit.prevent="handleSave">
          <div class="form-grid">
            <div class="form-field">
              <label for="edit-name">{{ t('students.name') }} *</label>
              <input
                id="edit-name"
                v-model="editForm.name"
                type="text"
                required
                data-testid="edit-student-name"
              >
            </div>
            <div class="form-field">
              <label for="edit-name-local">{{ t('students.nameLocal') }}</label>
              <input
                id="edit-name-local"
                v-model="editForm.nameLocal"
                type="text"
                data-testid="edit-student-name-local"
              >
            </div>
            <div class="form-field">
              <label for="edit-email">{{ t('students.emailOptional') }}</label>
              <input
                id="edit-email"
                v-model="editForm.email"
                type="email"
                data-testid="edit-student-email"
              >
            </div>
            <div class="form-field">
              <label for="edit-phone">{{ t('students.phone') }}</label>
              <input
                id="edit-phone"
                v-model="editForm.phone"
                type="tel"
                data-testid="edit-student-phone"
              >
            </div>
            <div class="form-field">
              <label for="edit-nationality">{{ t('students.nationality') }}</label>
              <input
                id="edit-nationality"
                v-model="editForm.nationality"
                type="text"
                data-testid="edit-student-nationality"
              >
            </div>
            <div class="form-field">
              <label for="edit-timezone">{{ t('students.timezone') }}</label>
              <input
                id="edit-timezone"
                v-model="editForm.timezone"
                type="text"
                data-testid="edit-student-timezone"
              >
            </div>
            <div class="form-field">
              <label for="edit-status">{{ t('students.status') }}</label>
              <select
                id="edit-status"
                v-model="editForm.status"
                data-testid="edit-student-status"
              >
                <option value="ACTIVE">
                  {{ t('students.statusActive') }}
                </option>
                <option value="PAUSED">
                  {{ t('students.statusPaused') }}
                </option>
                <option value="ENDED">
                  {{ t('students.statusEnded') }}
                </option>
                <option value="CANCELLED">
                  {{ t('students.statusCancelled') }}
                </option>
              </select>
            </div>
            <div class="form-field">
              <label for="edit-note">{{ t('students.note') }}</label>
              <textarea
                id="edit-note"
                v-model="editForm.note"
                data-testid="edit-student-note"
              />
            </div>
          </div>
          <p
            v-if="saveError"
            class="form-error"
            role="alert"
            aria-live="assertive"
            data-testid="student-save-error"
          >
            {{ t(saveError) }}
          </p>
          <p
            v-if="saveError === 'apiErrors.CONFLICT'"
            class="form-hint"
            data-testid="student-email-no-bypass"
          >
            {{ t('students.noBypass') }}
          </p>
          <button
            type="submit"
            :disabled="saving"
            data-testid="student-save-btn"
          >
            {{ saving ? t('common.saving') : t('students.edit') }}
          </button>
        </form>
      </section>

      <!-- Parents section -->
      <section
        class="parents-section"
        data-testid="student-parents-section"
      >
        <h2>{{ t('students.parents') }}</h2>
        <LoadingState v-if="parentsLoading" />
        <ErrorState
          v-else-if="parentsError"
          :error="parentsError"
          @retry="loadParents"
        />
        <EmptyState
          v-else-if="parents.length === 0"
          :message="t('students.parentsEmpty')"
        />
        <table
          v-else
          data-testid="parents-table"
        >
          <thead>
            <tr>
              <th>{{ t('students.parentName') }}</th>
              <th>{{ t('students.parentEmail') }}</th>
              <th>{{ t('students.parentPhone') }}</th>
              <th>{{ t('students.parentRelationship') }}</th>
              <th>{{ t('students.parentIsPrimary') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="p in parents"
              :key="p.id"
              data-testid="parent-row"
            >
              <td>{{ p.name }}</td>
              <td>{{ p.email ?? '—' }}</td>
              <td>{{ p.phone ?? '—' }}</td>
              <td>{{ p.relationship ?? '—' }}</td>
              <td>{{ p.isPrimary ? t('common.yes') : t('common.no') }}</td>
            </tr>
          </tbody>
        </table>
        <button
          type="button"
          data-testid="add-parent-btn"
          @click="openParentForm"
        >
          {{ t('students.addParent') }}
        </button>
        <div
          v-if="showParentForm"
          class="create-dialog"
          data-testid="parent-create-form"
        >
          <form @submit.prevent="handleCreateParent">
            <div class="form-field">
              <label for="parent-name">{{ t('students.parentName') }} *</label>
              <input
                id="parent-name"
                v-model="parentForm.name"
                type="text"
                required
                data-testid="parent-form-name"
              >
            </div>
            <div class="form-field">
              <label for="parent-email">{{ t('students.parentEmail') }}</label>
              <input
                id="parent-email"
                v-model="parentForm.email"
                type="email"
                data-testid="parent-form-email"
              >
            </div>
            <div class="form-field">
              <label for="parent-phone">{{ t('students.parentPhone') }}</label>
              <input
                id="parent-phone"
                v-model="parentForm.phone"
                type="tel"
                data-testid="parent-form-phone"
              >
            </div>
            <div class="form-field">
              <label for="parent-relationship">{{ t('students.parentRelationship') }}</label>
              <input
                id="parent-relationship"
                v-model="parentForm.relationship"
                type="text"
                data-testid="parent-form-relationship"
              >
            </div>
            <div class="form-field">
              <label>
                <input
                  v-model="parentForm.isPrimary"
                  type="checkbox"
                  data-testid="parent-form-primary"
                >
                {{ t('students.parentIsPrimary') }}
              </label>
            </div>
            <p
              v-if="parentError"
              class="form-error"
              role="alert"
              aria-live="assertive"
              data-testid="parent-create-error"
            >
              {{ t(parentError) }}
            </p>
            <div class="form-actions">
              <button
                type="button"
                :disabled="parentSaving"
                data-testid="parent-create-cancel"
                @click="showParentForm = false"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="submit"
                :disabled="parentSaving || !parentForm.name"
                data-testid="parent-create-submit"
              >
                {{ parentSaving ? t('common.creating') : t('common.create') }}
              </button>
            </div>
          </form>
        </div>
      </section>

      <!-- Enrollments section -->
      <section
        class="enrollments-section"
        data-testid="student-enrollments-section"
      >
        <h2>{{ t('students.enrollments') }}</h2>
        <LoadingState v-if="enrollmentsLoading" />
        <ErrorState
          v-else-if="enrollmentsError"
          :error="enrollmentsError"
          @retry="loadEnrollments"
        />
        <EmptyState
          v-else-if="enrollments.length === 0"
          :message="t('students.enrollmentsEmpty')"
        />
        <table
          v-else
          data-testid="enrollments-table"
        >
          <thead>
            <tr>
              <th>ID</th>
              <th>{{ t('enrollments.status') }}</th>
              <th>{{ t('enrollments.enrollmentType') }}</th>
              <th>{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="e in enrollments"
              :key="e.id"
              data-testid="enrollment-row"
            >
              <td>{{ e.id }}</td>
              <td>{{ enrollmentStatusLabel(e.status) }}</td>
              <td>{{ e.enrollmentType }}</td>
              <td>
                <button
                  type="button"
                  data-testid="view-enrollment-btn"
                  @click="viewEnrollment(e.id)"
                >
                  {{ t('students.viewEnrollment') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </section>
    </template>
  </div>
</template>

<style scoped>
.student-detail-view {
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

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-field label {
  font-weight: 600;
}

.form-field input, .form-field select, .form-field textarea {
  padding: 0.4rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
}

.form-hint {
  color: #856404;
  font-size: 0.8125rem;
  background-color: #fff3cd;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
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
