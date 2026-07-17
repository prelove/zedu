<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import { listTeachers, createTeacher, type Teacher, type TeacherWrite } from '../../api/directory'
import { ApiError, NetworkError } from '../../api/http'
import { errorToI18nKey } from '../../api/error-mapping'
import { formatJPY } from '../../utils/formatters'
import type { Locale } from '../../i18n/config'
import PaginationBar from '../../components/PaginationBar.vue'
import LoadingState from '../../components/LoadingState.vue'
import EmptyState from '../../components/EmptyState.vue'
import ErrorState from '../../components/ErrorState.vue'

const router = useRouter()
const { t, locale } = useI18n()

const teachers = ref<Teacher[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const error = ref<unknown>(null)
const search = ref('')

const showCreateForm = ref(false)
const creating = ref(false)
const createError = ref<string | null>(null)
const form = ref<TeacherWrite>({
  name: '',
  email: '',
  phone: '',
  bio: '',
  defaultRate: 0,
  status: 'ACTIVE',
})

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) =>
      listTeachers(token, { page: page.value, pageSize: pageSize.value, search: search.value || undefined }),
    )
    teachers.value = data.items ?? []
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
  form.value = { name: '', email: '', phone: '', bio: '', defaultRate: 0, status: 'ACTIVE' }
}

async function handleCreate(): Promise<void> {
  creating.value = true
  createError.value = null
  try {
    const body: TeacherWrite = { ...form.value }
    if (!body.email) delete body.email
    if (!body.phone) delete body.phone
    if (!body.bio) delete body.bio
    await authStore.authedRequest((token) => createTeacher(token, body))
    showCreateForm.value = false
    await load()
  } catch (err) {
    if (err instanceof NetworkError) createError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) createError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else createError.value = 'errors.UNKNOWN'
  } finally {
    creating.value = false
  }
}

function viewTeacher(id: number): void {
  void router.push({ name: 'teacher-detail', params: { id } })
}

function statusLabel(status: string): string {
  return status === 'ACTIVE' ? t('teachers.statusActive') : t('teachers.statusInactive')
}

function rateLabel(rate: number): string {
  return formatJPY(rate, locale.value as Locale)
}
</script>

<template>
  <div
    class="teachers-list-view"
    data-testid="teachers-list-view"
  >
    <h1>{{ t('teachers.title') }}</h1>

    <div class="toolbar">
      <input
        v-model="search"
        type="search"
        :placeholder="t('common.searchPlaceholder')"
        data-testid="teachers-search"
        @keyup.enter="handleSearch"
      >
      <button
        type="button"
        data-testid="teachers-search-btn"
        @click="handleSearch"
      >
        {{ t('common.search') }}
      </button>
      <button
        type="button"
        data-testid="teachers-create-btn"
        @click="openCreateForm"
      >
        {{ t('teachers.create') }}
      </button>
    </div>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="load"
    />
    <EmptyState
      v-else-if="teachers.length === 0"
      :message="t('teachers.title')"
    />
    <template v-else>
      <table data-testid="teachers-table">
        <thead>
          <tr>
            <th>{{ t('teachers.name') }}</th>
            <th>{{ t('teachers.email') }}</th>
            <th>{{ t('teachers.phone') }}</th>
            <th>{{ t('teachers.defaultRate') }}</th>
            <th>{{ t('teachers.status') }}</th>
            <th>{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="tc in teachers"
            :key="tc.id"
            data-testid="teacher-row"
          >
            <td>{{ tc.name }}</td>
            <td>{{ tc.email ?? '—' }}</td>
            <td>{{ tc.phone ?? '—' }}</td>
            <td>{{ rateLabel(tc.defaultRate) }}</td>
            <td>{{ statusLabel(tc.status) }}</td>
            <td>
              <button
                type="button"
                data-testid="teacher-view-btn"
                @click="viewTeacher(tc.id)"
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

    <div
      v-if="showCreateForm"
      class="create-dialog"
      data-testid="teachers-create-form"
    >
      <h2>{{ t('teachers.createTitle') }}</h2>
      <form @submit.prevent="handleCreate">
        <div class="form-field">
          <label for="teacher-name">{{ t('teachers.name') }} *</label>
          <input
            id="teacher-name"
            v-model="form.name"
            type="text"
            required
            data-testid="teacher-form-name"
          >
        </div>
        <div class="form-field">
          <label for="teacher-email">{{ t('teachers.email') }}</label>
          <input
            id="teacher-email"
            v-model="form.email"
            type="email"
            data-testid="teacher-form-email"
          >
        </div>
        <div class="form-field">
          <label for="teacher-phone">{{ t('teachers.phone') }}</label>
          <input
            id="teacher-phone"
            v-model="form.phone"
            type="tel"
            data-testid="teacher-form-phone"
          >
        </div>
        <div class="form-field">
          <label for="teacher-rate">{{ t('teachers.defaultRate') }}</label>
          <input
            id="teacher-rate"
            v-model.number="form.defaultRate"
            type="number"
            min="0"
            step="1"
            data-testid="teacher-form-rate"
          >
        </div>
        <div class="form-field">
          <label for="teacher-bio">{{ t('teachers.bio') }}</label>
          <textarea
            id="teacher-bio"
            v-model="form.bio"
            data-testid="teacher-form-bio"
          />
        </div>
        <p
          v-if="createError"
          class="form-error"
          role="alert"
          aria-live="assertive"
          data-testid="teachers-create-error"
        >
          {{ t(createError) }}
        </p>
        <div class="form-actions">
          <button
            type="button"
            :disabled="creating"
            @click="showCreateForm = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="creating || !form.name"
            data-testid="teachers-create-submit"
          >
            {{ creating ? t('common.creating') : t('common.create') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.teachers-list-view {
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

.form-field input, .form-field textarea {
  padding: 0.4rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
  margin: 0;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}
</style>
