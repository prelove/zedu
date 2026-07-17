<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import { listStudents, createStudent, type Student, type StudentWrite } from '../../api/directory'
import { ApiError, NetworkError } from '../../api/http'
import { errorToI18nKey } from '../../api/error-mapping'
import PaginationBar from '../../components/PaginationBar.vue'
import LoadingState from '../../components/LoadingState.vue'
import EmptyState from '../../components/EmptyState.vue'
import ErrorState from '../../components/ErrorState.vue'

const router = useRouter()
const { t } = useI18n()

const students = ref<Student[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const error = ref<unknown>(null)
const search = ref('')

// Create form state.
const showCreateForm = ref(false)
const creating = ref(false)
const createError = ref<string | null>(null)
const form = ref<StudentWrite>({
  name: '',
  email: '',
  phone: '',
  timezone: 'Asia/Tokyo',
  status: 'ACTIVE',
})

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) =>
      listStudents(token, { page: page.value, pageSize: pageSize.value, search: search.value || undefined }),
    )
    students.value = data.items ?? []
    total.value = data.total ?? 0
  } catch (err) {
    error.value = err
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})

function handlePageChange(newPage: number): void {
  page.value = newPage
  void load()
}

function handleSearch(): void {
  page.value = 1
  void load()
}

function openCreateForm(): void {
  showCreateForm.value = true
  createError.value = null
  form.value = { name: '', email: '', phone: '', timezone: 'Asia/Tokyo', status: 'ACTIVE' }
}

async function handleCreate(): Promise<void> {
  creating.value = true
  createError.value = null
  try {
    // Only send non-empty email (empty means "no email").
    const body: StudentWrite = { ...form.value }
    if (body.email === '') delete body.email
    if (body.phone === '') delete body.phone
    await authStore.authedRequest((token) => createStudent(token, body))
    showCreateForm.value = false
    await load()
  } catch (err) {
    if (err instanceof NetworkError) {
      createError.value = 'errors.NETWORK_ERROR'
    } else if (err instanceof ApiError) {
      createError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    } else {
      createError.value = 'errors.UNKNOWN'
    }
  } finally {
    creating.value = false
  }
}

function viewStudent(id: number): void {
  void router.push({ name: 'student-detail', params: { id } })
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    ACTIVE: t('students.statusActive'),
    PAUSED: t('students.statusPaused'),
    ENDED: t('students.statusEnded'),
    CANCELLED: t('students.statusCancelled'),
  }
  return map[status] ?? status
}
</script>

<template>
  <div
    class="students-list-view"
    data-testid="students-list-view"
  >
    <h1>{{ t('students.title') }}</h1>

    <div class="toolbar">
      <input
        v-model="search"
        type="search"
        :placeholder="t('common.searchPlaceholder')"
        data-testid="students-search"
        @keyup.enter="handleSearch"
      >
      <button
        type="button"
        data-testid="students-search-btn"
        @click="handleSearch"
      >
        {{ t('common.search') }}
      </button>
      <button
        type="button"
        data-testid="students-create-btn"
        @click="openCreateForm"
      >
        {{ t('students.create') }}
      </button>
    </div>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="load"
    />
    <EmptyState
      v-else-if="students.length === 0"
      :message="t('students.title')"
    />
    <template v-else>
      <table data-testid="students-table">
        <thead>
          <tr>
            <th>{{ t('students.name') }}</th>
            <th>{{ t('students.email') }}</th>
            <th>{{ t('students.phone') }}</th>
            <th>{{ t('students.status') }}</th>
            <th>{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="s in students"
            :key="s.id"
            data-testid="student-row"
          >
            <td>{{ s.name }}</td>
            <td>{{ s.email ?? '—' }}</td>
            <td>{{ s.phone ?? '—' }}</td>
            <td>{{ statusLabel(s.status) }}</td>
            <td>
              <button
                type="button"
                data-testid="student-view-btn"
                @click="viewStudent(s.id)"
              >
                {{ t('common.edit') }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <PaginationBar
        :page="page"
        :page-size="pageSize"
        :total="total"
        @page-change="handlePageChange"
      />
    </template>

    <!-- Create form dialog -->
    <div
      v-if="showCreateForm"
      class="create-dialog"
      data-testid="students-create-form"
    >
      <h2>{{ t('students.createTitle') }}</h2>
      <form @submit.prevent="handleCreate">
        <div class="form-field">
          <label for="student-name">{{ t('students.name') }} *</label>
          <input
            id="student-name"
            v-model="form.name"
            type="text"
            required
            data-testid="student-form-name"
          >
        </div>
        <div class="form-field">
          <label for="student-email">{{ t('students.emailOptional') }}</label>
          <input
            id="student-email"
            v-model="form.email"
            type="email"
            data-testid="student-form-email"
          >
        </div>
        <div class="form-field">
          <label for="student-phone">{{ t('students.phone') }}</label>
          <input
            id="student-phone"
            v-model="form.phone"
            type="tel"
            data-testid="student-form-phone"
          >
        </div>
        <p
          v-if="createError"
          class="form-error"
          role="alert"
          aria-live="assertive"
          data-testid="students-create-error"
        >
          {{ t(createError) }}
        </p>
        <p
          v-if="createError === 'apiErrors.CONFLICT'"
          class="form-hint"
          data-testid="students-no-bypass"
        >
          {{ t('students.noBypass') }}
        </p>
        <div class="form-actions">
          <button
            type="button"
            :disabled="creating"
            data-testid="students-create-cancel"
            @click="showCreateForm = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="creating || !form.name"
            data-testid="students-create-submit"
          >
            {{ creating ? t('common.creating') : t('common.create') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.students-list-view {
  max-width: 900px;
  margin: 0 auto;
  padding: 1rem;
}

.toolbar {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

input[type="search"] {
  padding: 0.4rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  flex: 1;
  min-width: 200px;
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

table {
  width: 100%;
  border-collapse: collapse;
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
  margin-top: 1rem;
  background-color: #f8f9fa;
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

.form-field input {
  padding: 0.4rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
  margin: 0;
}

.form-hint {
  color: #856404;
  font-size: 0.8125rem;
  background-color: #fff3cd;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  margin: 0.25rem 0;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}
</style>
