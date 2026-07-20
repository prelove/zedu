<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { listAvailability, createAvailability, updateAvailability, type Availability, type AvailabilityWrite } from '../../../api/directory'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import LoadingState from '../../../components/LoadingState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import { usePaginatedList } from '../../../composables/usePaginatedList'
import AvailabilityCreateForm from './AvailabilityCreateForm.vue'
import AvailabilityEditForm from './AvailabilityEditForm.vue'
import AvailabilityTable from './AvailabilityTable.vue'

const props = defineProps<{ teacherId: number }>()
const { t } = useI18n()
const { items: availabilities, page, pageSize, total, loading, error, setData } = usePaginatedList<Availability>(20)

const weekdayOptions = [1, 2, 3, 4, 5, 6, 7]
const weekdayKeys: Record<number, string> = {
  1: 'teachers.weekdayMonday',
  2: 'teachers.weekdayTuesday',
  3: 'teachers.weekdayWednesday',
  4: 'teachers.weekdayThursday',
  5: 'teachers.weekdayFriday',
  6: 'teachers.weekdaySaturday',
  7: 'teachers.weekdaySunday',
}

const showCreateForm = ref(false)
const creating = ref(false)
const createError = ref<string | null>(null)
const clientError = ref<string | null>(null)
const editingAvail = ref<Availability | null>(null)
const editSaving = ref(false)
const editError = ref<string | null>(null)
const editClientError = ref<string | null>(null)

function validateAvail(form: AvailabilityWrite): string | null {
  if (!form.weekday || form.weekday < 1 || form.weekday > 7) return 'teachers.availabilityInvalidWeekday'
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
  } finally {
    loading.value = false
  }
}

async function handleCreate(body: AvailabilityWrite): Promise<void> {
  clientError.value = validateAvail(body)
  if (clientError.value) return
  creating.value = true
  createError.value = null
  try {
    await authStore.authedRequest((token) => createAvailability(token, props.teacherId, body))
    showCreateForm.value = false
    await loadAvailability()
  } catch (err) {
    if (err instanceof NetworkError) createError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) createError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else createError.value = 'errors.UNKNOWN'
  } finally {
    creating.value = false
  }
}

async function handleEditSave(body: AvailabilityWrite): Promise<void> {
  if (!editingAvail.value) return
  editClientError.value = validateAvail(body)
  if (editClientError.value) return
  editSaving.value = true
  editError.value = null
  try {
    await authStore.authedRequest((token) => updateAvailability(token, props.teacherId, editingAvail.value!.id, body))
    editingAvail.value = null
    await loadAvailability()
  } catch (err) {
    if (err instanceof NetworkError) editError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) editError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else editError.value = 'errors.UNKNOWN'
  } finally {
    editSaving.value = false
  }
}

function weekdayLabel(weekday: number): string {
  return weekday in weekdayKeys ? t(weekdayKeys[weekday]) : String(weekday)
}

onMounted(loadAvailability)
watch(page, loadAvailability)
</script>

<template>
  <section
    class="availability-section"
    data-testid="availability-section"
  >
    <h2>{{ t('teachers.availability') }}</h2>
    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadAvailability"
    />
    <EmptyState
      v-else-if="availabilities.length === 0"
      :message="t('teachers.availabilityEmpty')"
    />
    <AvailabilityTable
      v-else
      :availabilities="availabilities"
      :weekday-label="weekdayLabel"
      @edit="(availability) => { editingAvail = availability; editError = null; editClientError = null }"
    />
    <PaginationBar
      v-if="total > 0"
      :page="page"
      :page-size="pageSize"
      :total="total"
      data-testid="availability-pagination"
      @page-change="(nextPage) => (page = nextPage)"
    />
    <button
      type="button"
      data-testid="add-availability-btn"
      @click="showCreateForm = true; createError = null; clientError = null"
    >
      {{ t('teachers.addAvailability') }}
    </button>
    <AvailabilityCreateForm
      v-if="showCreateForm"
      :weekday-options="weekdayOptions"
      :weekday-keys="weekdayKeys"
      :creating="creating"
      :client-error="clientError"
      :create-error="createError"
      @submit="handleCreate"
      @cancel="showCreateForm = false"
    />
    <AvailabilityEditForm
      v-if="editingAvail"
      :availability="editingAvail"
      :weekday-options="weekdayOptions"
      :weekday-keys="weekdayKeys"
      :saving="editSaving"
      :client-error="editClientError"
      :error="editError"
      @submit="handleEditSave"
      @cancel="editingAvail = null"
    />
  </section>
</template>

<style scoped>
.availability-section {
  margin-top: 1.5rem;
}

.availability-section h2 {
  margin: 0 0 0.5rem;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
