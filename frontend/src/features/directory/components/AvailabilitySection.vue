<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { listAvailability, createAvailability, updateAvailability, type Availability, type AvailabilityWrite } from '../../../api/directory'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import LoadingState from '../../../components/LoadingState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import { usePaginatedList } from '../../../composables/usePaginatedList'

const props = defineProps<{ teacherId: number }>()
const { t } = useI18n()

const { items: availabilities, page, pageSize, total, loading, error, setData } = usePaginatedList<Availability>(20)

const weekdayOptions = [1, 2, 3, 4, 5, 6, 7]
const weekdayKeys: Record<number, string> = {
  1: 'teachers.weekdayMonday', 2: 'teachers.weekdayTuesday', 3: 'teachers.weekdayWednesday',
  4: 'teachers.weekdayThursday', 5: 'teachers.weekdayFriday', 6: 'teachers.weekdaySaturday', 7: 'teachers.weekdaySunday',
}
function weekdayLabel(wd: number): string {
  return wd in weekdayKeys ? t(weekdayKeys[wd]) : String(wd)
}

const showCreateForm = ref(false)
const createForm = ref<AvailabilityWrite>({ weekday: 1, startTime: '', endTime: '' })
const creating = ref(false)
const createError = ref<string | null>(null)
const clientError = ref<string | null>(null)

const editingAvail = ref<Availability | null>(null)
const editForm = ref<AvailabilityWrite>({})
const editSaving = ref(false)
const editError = ref<string | null>(null)
const editClientError = ref<string | null>(null)
const editNoChanges = ref(false)

function validateAvail(form: AvailabilityWrite): string | null {
  const wd = form.weekday
  if (!wd || wd < 1 || wd > 7) return 'teachers.availabilityInvalidWeekday'
  if (!form.startTime || !form.endTime || form.startTime >= form.endTime) return 'teachers.availabilityInvalidTime'
  return null
}

async function loadAvailability(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    setData(await authStore.authedRequest((token) => listAvailability(token, props.teacherId, { page: page.value, pageSize: pageSize.value })))
  } catch (err) {
    if (err instanceof NetworkError) error.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) error.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else error.value = 'errors.UNKNOWN'
  } finally { loading.value = false }
}

onMounted(loadAvailability)
watch(page, loadAvailability)

function openCreateForm(): void {
  showCreateForm.value = true
  createError.value = null
  clientError.value = null
  createForm.value = { weekday: 1, startTime: '', endTime: '' }
}

async function handleCreate(): Promise<void> {
  clientError.value = validateAvail(createForm.value)
  if (clientError.value) return
  creating.value = true
  createError.value = null
  try {
    await authStore.authedRequest((token) => createAvailability(token, props.teacherId, createForm.value))
    showCreateForm.value = false
    await loadAvailability()
  } catch (err) {
    if (err instanceof NetworkError) createError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) createError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else createError.value = 'errors.UNKNOWN'
  } finally { creating.value = false }
}

function openEditForm(a: Availability): void {
  editingAvail.value = a
  editError.value = null
  editClientError.value = null
  editNoChanges.value = false
  editForm.value = { weekday: a.weekday, startTime: a.startTime, endTime: a.endTime }
}

async function handleEditSave(): Promise<void> {
  if (!editingAvail.value) return
  editClientError.value = validateAvail(editForm.value)
  if (editClientError.value) return
  editSaving.value = true
  editError.value = null
  editNoChanges.value = false
  const a = editingAvail.value
  const body: AvailabilityWrite = {}
  if (editForm.value.weekday !== a.weekday) body.weekday = editForm.value.weekday
  if (editForm.value.startTime !== a.startTime) body.startTime = editForm.value.startTime
  if (editForm.value.endTime !== a.endTime) body.endTime = editForm.value.endTime
  if (Object.keys(body).length === 0) { editNoChanges.value = true; editSaving.value = false; return }
  try {
    await authStore.authedRequest((token) => updateAvailability(token, props.teacherId, a.id, body))
    editingAvail.value = null
    await loadAvailability()
  } catch (err) {
    if (err instanceof NetworkError) editError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) editError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else editError.value = 'errors.UNKNOWN'
  } finally { editSaving.value = false }
}
</script>

<template>
  <section
    class="availability-section"
    data-testid="availability-section"
  >
    <h2>{{ t('teachers.availability') }}</h2>
    <LoadingState v-if="loading" />
    <p
      v-else-if="error"
      class="state-error"
      role="alert"
      aria-live="assertive"
      data-testid="state-error"
    >
      {{ t(error) }}
      <button
        type="button"
        data-testid="state-error-retry"
        @click="loadAvailability"
      >
        {{ t('common.retry') }}
      </button>
    </p>
    <EmptyState
      v-else-if="availabilities.length === 0"
      :message="t('teachers.availabilityEmpty')"
    />
    <table
      v-else
      data-testid="availability-table"
    >
      <thead>
        <tr>
          <th>{{ t('teachers.availabilityWeekday') }}</th>
          <th>{{ t('teachers.availabilityStartTime') }}</th>
          <th>{{ t('teachers.availabilityEndTime') }}</th>
          <th>{{ t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="a in availabilities"
          :key="a.id"
          data-testid="availability-row"
        >
          <td>{{ weekdayLabel(a.weekday) }}</td>
          <td>{{ a.startTime }}</td>
          <td>{{ a.endTime }}</td>
          <td>
            <button
              type="button"
              data-testid="availability-edit-btn"
              @click="openEditForm(a)"
            >
              {{ t('common.edit') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
    <PaginationBar
      v-if="total > 0"
      :page="page"
      :page-size="pageSize"
      :total="total"
      data-testid="availability-pagination"
      @page-change="(p) => (page = p)"
    />

    <button
      type="button"
      data-testid="add-availability-btn"
      @click="openCreateForm"
    >
      {{ t('teachers.addAvailability') }}
    </button>

    <div
      v-if="showCreateForm"
      class="create-dialog"
      data-testid="availability-create-form"
    >
      <form @submit.prevent="handleCreate">
        <div class="form-field">
          <label for="avail-form-weekday">{{ t('teachers.availabilityWeekday') }}</label>
          <select
            id="avail-form-weekday"
            v-model.number="createForm.weekday"
            data-testid="avail-form-weekday"
          >
            <option
              v-for="wd in weekdayOptions"
              :key="wd"
              :value="wd"
            >
              {{ t(weekdayKeys[wd]) }}
            </option>
          </select>
        </div>
        <div class="form-field">
          <label for="avail-form-start">{{ t('teachers.availabilityStartTime') }}</label>
          <input
            id="avail-form-start"
            v-model="createForm.startTime"
            type="time"
            data-testid="avail-form-start"
          >
        </div>
        <div class="form-field">
          <label for="avail-form-end">{{ t('teachers.availabilityEndTime') }}</label>
          <input
            id="avail-form-end"
            v-model="createForm.endTime"
            type="time"
            data-testid="avail-form-end"
          >
        </div>
        <p
          v-if="clientError"
          class="form-error"
          role="alert"
          aria-live="assertive"
          data-testid="availability-client-error"
        >
          {{ t(clientError) }}
        </p>
        <p
          v-if="createError"
          class="form-error"
          role="alert"
          aria-live="assertive"
          data-testid="availability-create-error"
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
            :disabled="creating"
            data-testid="availability-create-submit"
          >
            {{ creating ? t('common.creating') : t('common.create') }}
          </button>
        </div>
      </form>
    </div>

    <form
      v-if="editingAvail"
      class="avail-form"
      data-testid="availability-edit-form"
      @submit.prevent="handleEditSave"
    >
      <div class="form-field">
        <label for="avail-edit-weekday">{{ t('teachers.availabilityWeekday') }}</label>
        <select
          id="avail-edit-weekday"
          v-model.number="editForm.weekday"
          data-testid="avail-edit-weekday"
        >
          <option
            v-for="wd in weekdayOptions"
            :key="wd"
            :value="wd"
          >
            {{ t(weekdayKeys[wd]) }}
          </option>
        </select>
      </div>
      <div class="form-field">
        <label for="avail-edit-start-time">{{ t('teachers.availabilityStartTime') }}</label>
        <input
          id="avail-edit-start-time"
          v-model="editForm.startTime"
          type="time"
          data-testid="avail-edit-start-time"
        >
      </div>
      <div class="form-field">
        <label for="avail-edit-end-time">{{ t('teachers.availabilityEndTime') }}</label>
        <input
          id="avail-edit-end-time"
          v-model="editForm.endTime"
          type="time"
          data-testid="avail-edit-end-time"
        >
      </div>
      <p
        v-if="editClientError"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="availability-client-error"
      >
        {{ t(editClientError) }}
      </p>
      <p
        v-if="editNoChanges"
        class="form-info"
        role="alert"
        aria-live="assertive"
        data-testid="avail-edit-no-changes"
      >
        {{ t('teachers.availabilityNoChanges') }}
      </p>
      <p
        v-if="editError"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="avail-edit-error"
      >
        {{ t(editError) }}
      </p>
      <div class="form-actions">
        <button
          type="button"
          :disabled="editSaving"
          data-testid="avail-edit-cancel"
          @click="editingAvail = null"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          :disabled="editSaving"
          data-testid="avail-edit-submit"
        >
          {{ editSaving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </form>
  </section>
</template>

<style scoped>
.availability-section { margin-top: 1.5rem; }
.availability-section h2 { margin: 0 0 0.5rem; }
.availability-list { list-style: none; padding: 0; margin: 0 0 0.5rem; }
.availability-row { display: flex; align-items: center; gap: 0.75rem; padding: 0.4rem 0; border-bottom: 1px solid #eee; }
.avail-weekday { min-width: 6rem; font-weight: 600; }
.avail-time { min-width: 10rem; color: #495057; }
.state-error { color: #dc3545; padding: 0.5rem 0; }
.state-error button { margin-left: 0.5rem; padding: 0.2rem 0.5rem; border: 1px solid #dc3545; border-radius: 0.25rem; background: #fff; color: #dc3545; cursor: pointer; }
.avail-form { margin-top: 1rem; padding: 1rem; border: 1px solid #dee2e6; border-radius: 0.375rem; max-width: 28rem; }
.form-field { display: flex; flex-direction: column; margin-bottom: 0.75rem; }
.form-field label { margin-bottom: 0.25rem; font-size: 0.875rem; font-weight: 600; }
.form-field input, .form-field select { padding: 0.35rem 0.5rem; border: 1px solid #ccc; border-radius: 0.25rem; }
.form-error { color: #dc3545; margin: 0 0 0.5rem; font-size: 0.875rem; }
.form-info { color: #856404; margin: 0 0 0.5rem; font-size: 0.875rem; }
.form-actions { display: flex; gap: 0.5rem; }
.form-actions button { padding: 0.4rem 0.8rem; border: 1px solid #ccc; border-radius: 0.25rem; background: #fff; cursor: pointer; }
.form-actions button[type='submit'] { background: #0d6efd; color: #fff; border-color: #0d6efd; }
button:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
