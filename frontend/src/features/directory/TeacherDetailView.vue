<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import {
  getTeacher, updateTeacher, listCapabilities, createCapability, updateCapability,
  listAvailability, createAvailability,
  type Teacher, type Capability, type Availability,
  type TeacherWrite, type CapabilityWrite, type AvailabilityWrite,
} from '../../api/directory'
import { listDomains, listTracks, listLevels, type CourseDomain, type Track, type Level } from '../../api/course'
import { ApiError, NetworkError } from '../../api/http'
import { errorToI18nKey } from '../../api/error-mapping'
import LoadingState from '../../components/LoadingState.vue'
import ErrorState from '../../components/ErrorState.vue'
import EmptyState from '../../components/EmptyState.vue'
import ConfirmDialog from '../../components/ConfirmDialog.vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const { t } = useI18n()

const teacher = ref<Teacher | null>(null)
const loading = ref(false)
const error = ref<unknown>(null)

const editForm = ref<TeacherWrite>({})
const saving = ref(false)
const saveError = ref<string | null>(null)

// Capabilities.
const capabilities = ref<Capability[]>([])
const capsLoading = ref(false)
const capsError = ref<unknown>(null)
const showCapForm = ref(false)
const capForm = ref<CapabilityWrite>({ domainId: 0, trackId: 0, levelId: 0, status: 'ACTIVE' })
const capSaving = ref(false)
const capError = ref<string | null>(null)
const capToEnd = ref<Capability | null>(null)

// Availability.
const availabilities = ref<Availability[]>([])
const availsLoading = ref(false)
const availsError = ref<unknown>(null)
const showAvailForm = ref(false)
const availForm = ref<AvailabilityWrite>({ weekday: 1, startTime: '09:00', endTime: '10:00' })
const availSaving = ref(false)
const availError = ref<string | null>(null)
const availClientError = ref<string | null>(null)

// Course dictionary for capability selectors.
const domains = ref<CourseDomain[]>([])
const tracks = ref<Track[]>([])
const levels = ref<Level[]>([])

const teacherId = () => Number(props.id)

async function loadTeacher(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) => getTeacher(token, teacherId()))
    teacher.value = data
    editForm.value = {
      name: data.name,
      nameLocal: data.nameLocal ?? '',
      email: data.email ?? '',
      phone: data.phone ?? '',
      bio: data.bio ?? '',
      defaultRate: data.defaultRate,
      status: data.status,
      note: data.note ?? '',
    }
  } catch (err) {
    error.value = err
  } finally {
    loading.value = false
  }
}

async function loadCapabilities(): Promise<void> {
  capsLoading.value = true
  capsError.value = null
  try {
    const data = await authStore.authedRequest((token) => listCapabilities(token, teacherId(), { pageSize: 100 }))
    capabilities.value = data.items ?? []
  } catch (err) {
    capsError.value = err
  } finally {
    capsLoading.value = false
  }
}

async function loadAvailability(): Promise<void> {
  availsLoading.value = true
  availsError.value = null
  try {
    const data = await authStore.authedRequest((token) => listAvailability(token, teacherId(), { pageSize: 100 }))
    availabilities.value = data.items ?? []
  } catch (err) {
    availsError.value = err
  } finally {
    availsLoading.value = false
  }
}

async function loadDictionary(): Promise<void> {
  try {
    const [dom, trk, lvl] = await Promise.all([
      authStore.authedRequest((token) => listDomains(token, { pageSize: 100 })),
      authStore.authedRequest((token) => listTracks(token, { pageSize: 100 })),
      authStore.authedRequest((token) => listLevels(token, { pageSize: 100 })),
    ])
    domains.value = dom.items ?? []
    tracks.value = trk.items ?? []
    levels.value = lvl.items ?? []
  } catch {
    // Dictionary load failure is non-fatal; selectors will be empty.
  }
}

onMounted(() => {
  void loadTeacher()
  void loadCapabilities()
  void loadAvailability()
  void loadDictionary()
})

async function handleSave(): Promise<void> {
  saving.value = true
  saveError.value = null
  try {
    const body: TeacherWrite = {}
    const tc = teacher.value!
    if (editForm.value.name !== tc.name) body.name = editForm.value.name
    if ((editForm.value.nameLocal ?? '') !== (tc.nameLocal ?? '')) body.nameLocal = editForm.value.nameLocal
    if ((editForm.value.email ?? '') !== (tc.email ?? '')) body.email = editForm.value.email
    if ((editForm.value.phone ?? '') !== (tc.phone ?? '')) body.phone = editForm.value.phone
    if ((editForm.value.bio ?? '') !== (tc.bio ?? '')) body.bio = editForm.value.bio
    if (editForm.value.defaultRate !== tc.defaultRate) body.defaultRate = editForm.value.defaultRate
    if ((editForm.value.status ?? '') !== tc.status) body.status = editForm.value.status
    if ((editForm.value.note ?? '') !== (tc.note ?? '')) body.note = editForm.value.note
    const updated = await authStore.authedRequest((token) => updateTeacher(token, teacherId(), body))
    teacher.value = updated
  } catch (err) {
    if (err instanceof NetworkError) saveError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) saveError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else saveError.value = 'errors.UNKNOWN'
  } finally {
    saving.value = false
  }
}

function openCapForm(): void {
  showCapForm.value = true
  capError.value = null
  capForm.value = { domainId: 0, trackId: 0, levelId: 0, status: 'ACTIVE' }
}

async function handleCreateCapability(): Promise<void> {
  capSaving.value = true
  capError.value = null
  try {
    if (!capForm.value.domainId || !capForm.value.trackId || !capForm.value.levelId) {
      capError.value = 'errors.UNKNOWN'
      return
    }
    const body: CapabilityWrite = {
      domainId: capForm.value.domainId,
      trackId: capForm.value.trackId,
      levelId: capForm.value.levelId,
      status: capForm.value.status,
    }
    await authStore.authedRequest((token) => createCapability(token, teacherId(), body))
    showCapForm.value = false
    await loadCapabilities()
  } catch (err) {
    if (err instanceof NetworkError) capError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) capError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else capError.value = 'errors.UNKNOWN'
  } finally {
    capSaving.value = false
  }
}

async function handleEndCapability(): Promise<void> {
  if (!capToEnd.value) return
  const cap = capToEnd.value
  capToEnd.value = null
  try {
    await authStore.authedRequest((token) =>
      updateCapability(token, teacherId(), cap.id, { status: 'ENDED', effectiveTo: new Date().toISOString() }),
    )
    await loadCapabilities()
  } catch (err) {
    if (err instanceof NetworkError) capsError.value = err
    else if (err instanceof ApiError) capsError.value = err
    else capsError.value = err
  }
}

function openAvailForm(): void {
  showAvailForm.value = true
  availError.value = null
  availClientError.value = null
  availForm.value = { weekday: 1, startTime: '09:00', endTime: '10:00' }
}

function validateAvail(): boolean {
  availClientError.value = null
  const wd = availForm.value.weekday
  if (wd === undefined || wd < 1 || wd > 7) {
    availClientError.value = 'teachers.availabilityInvalidWeekday'
    return false
  }
  const st = availForm.value.startTime ?? ''
  const et = availForm.value.endTime ?? ''
  if (!st || !et || st >= et) {
    availClientError.value = 'teachers.availabilityInvalidTime'
    return false
  }
  return true
}

async function handleCreateAvailability(): Promise<void> {
  availSaving.value = true
  availError.value = null
  try {
    if (!validateAvail()) {
      availSaving.value = false
      return
    }
    const body: AvailabilityWrite = {
      weekday: availForm.value.weekday,
      startTime: availForm.value.startTime,
      endTime: availForm.value.endTime,
    }
    await authStore.authedRequest((token) => createAvailability(token, teacherId(), body))
    showAvailForm.value = false
    await loadAvailability()
  } catch (err) {
    if (err instanceof NetworkError) availError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) availError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else availError.value = 'errors.UNKNOWN'
  } finally {
    availSaving.value = false
  }
}

// Filtered tracks/levels for cascading selectors.
const filteredTracks = () => tracks.value.filter((tr) => tr.domainId === capForm.value.domainId)
const filteredLevels = () => levels.value.filter((lv) => lv.trackId === capForm.value.trackId)

function domainName(id: number): string {
  return domains.value.find((d) => d.id === id)?.name ?? String(id)
}
function trackName(id: number): string {
  return tracks.value.find((tr) => tr.id === id)?.name ?? String(id)
}
function levelName(id: number): string {
  return levels.value.find((lv) => lv.id === id)?.name ?? String(id)
}

function weekdayLabel(wd: number): string {
  const map: Record<number, string> = {
    1: t('teachers.weekdayMonday'),
    2: t('teachers.weekdayTuesday'),
    3: t('teachers.weekdayWednesday'),
    4: t('teachers.weekdayThursday'),
    5: t('teachers.weekdayFriday'),
    6: t('teachers.weekdaySaturday'),
    7: t('teachers.weekdaySunday'),
  }
  return map[wd] ?? String(wd)
}
</script>

<template>
  <div
    class="teacher-detail-view"
    data-testid="teacher-detail-view"
  >
    <button
      type="button"
      data-testid="back-to-teachers"
      @click="router.push({ name: 'teachers' })"
    >
      {{ t('common.back') }}
    </button>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadTeacher"
    />
    <template v-else-if="teacher">
      <h1>{{ t('teachers.detailTitle') }}: {{ teacher.name }}</h1>

      <!-- Edit form -->
      <section
        class="edit-section"
        data-testid="teacher-edit-section"
      >
        <h2>{{ t('teachers.editTitle') }}</h2>
        <form @submit.prevent="handleSave">
          <div class="form-grid">
            <div class="form-field">
              <label for="t-edit-name">{{ t('teachers.name') }} *</label>
              <input
                id="t-edit-name"
                v-model="editForm.name"
                type="text"
                required
                data-testid="edit-teacher-name"
              >
            </div>
            <div class="form-field">
              <label for="t-edit-name-local">{{ t('teachers.nameLocal') }}</label>
              <input
                id="t-edit-name-local"
                v-model="editForm.nameLocal"
                type="text"
                data-testid="edit-teacher-name-local"
              >
            </div>
            <div class="form-field">
              <label for="t-edit-email">{{ t('teachers.email') }}</label>
              <input
                id="t-edit-email"
                v-model="editForm.email"
                type="email"
                data-testid="edit-teacher-email"
              >
            </div>
            <div class="form-field">
              <label for="t-edit-phone">{{ t('teachers.phone') }}</label>
              <input
                id="t-edit-phone"
                v-model="editForm.phone"
                type="tel"
                data-testid="edit-teacher-phone"
              >
            </div>
            <div class="form-field">
              <label for="t-edit-rate">{{ t('teachers.defaultRate') }}</label>
              <input
                id="t-edit-rate"
                v-model.number="editForm.defaultRate"
                type="number"
                min="0"
                step="1"
                data-testid="edit-teacher-rate"
              >
            </div>
            <div class="form-field">
              <label for="t-edit-status">{{ t('teachers.status') }}</label>
              <select
                id="t-edit-status"
                v-model="editForm.status"
                data-testid="edit-teacher-status"
              >
                <option value="ACTIVE">
                  {{ t('teachers.statusActive') }}
                </option>
                <option value="INACTIVE">
                  {{ t('teachers.statusInactive') }}
                </option>
              </select>
            </div>
            <div class="form-field">
              <label for="t-edit-bio">{{ t('teachers.bio') }}</label>
              <textarea
                id="t-edit-bio"
                v-model="editForm.bio"
                data-testid="edit-teacher-bio"
              />
            </div>
            <div class="form-field">
              <label for="t-edit-note">{{ t('teachers.note') }}</label>
              <textarea
                id="t-edit-note"
                v-model="editForm.note"
                data-testid="edit-teacher-note"
              />
            </div>
          </div>
          <p
            v-if="saveError"
            class="form-error"
            role="alert"
            aria-live="assertive"
            data-testid="teacher-save-error"
          >
            {{ t(saveError) }}
          </p>
          <button
            type="submit"
            :disabled="saving"
            data-testid="teacher-save-btn"
          >
            {{ saving ? t('common.saving') : t('teachers.edit') }}
          </button>
        </form>
      </section>

      <!-- Capabilities section -->
      <section
        class="capabilities-section"
        data-testid="teacher-capabilities-section"
      >
        <h2>{{ t('teachers.capabilities') }}</h2>
        <LoadingState v-if="capsLoading" />
        <ErrorState
          v-else-if="capsError"
          :error="capsError"
          @retry="loadCapabilities"
        />
        <EmptyState
          v-else-if="capabilities.length === 0"
          :message="t('teachers.capabilitiesEmpty')"
        />
        <table
          v-else
          data-testid="capabilities-table"
        >
          <thead>
            <tr>
              <th>{{ t('teachers.capabilityDomain') }}</th>
              <th>{{ t('teachers.capabilityTrack') }}</th>
              <th>{{ t('teachers.capabilityLevel') }}</th>
              <th>{{ t('teachers.capabilityStatus') }}</th>
              <th>{{ t('teachers.capabilityVerified') }}</th>
              <th>{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="c in capabilities"
              :key="c.id"
              data-testid="capability-row"
            >
              <td>{{ domainName(c.domainId) }}</td>
              <td>{{ trackName(c.trackId) }}</td>
              <td>{{ levelName(c.levelId) }}</td>
              <td>{{ c.status }}</td>
              <td>{{ c.verified ? t('common.yes') : t('common.no') }}</td>
              <td>
                <button
                  v-if="c.status === 'ACTIVE'"
                  type="button"
                  data-testid="capability-end-btn"
                  @click="capToEnd = c"
                >
                  {{ t('teachers.capabilityEnd') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <button
          type="button"
          data-testid="add-capability-btn"
          @click="openCapForm"
        >
          {{ t('teachers.addCapability') }}
        </button>
        <div
          v-if="showCapForm"
          class="create-dialog"
          data-testid="capability-create-form"
        >
          <form @submit.prevent="handleCreateCapability">
            <div class="form-field">
              <label for="cap-domain">{{ t('teachers.capabilityDomain') }}</label>
              <select
                id="cap-domain"
                v-model.number="capForm.domainId"
                data-testid="cap-form-domain"
                @change="capForm.trackId = 0; capForm.levelId = 0"
              >
                <option :value="0">
                  —
                </option>
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
              <label for="cap-track">{{ t('teachers.capabilityTrack') }}</label>
              <select
                id="cap-track"
                v-model.number="capForm.trackId"
                :disabled="!capForm.domainId"
                data-testid="cap-form-track"
                @change="capForm.levelId = 0"
              >
                <option :value="0">
                  —
                </option>
                <option
                  v-for="tr in filteredTracks()"
                  :key="tr.id"
                  :value="tr.id"
                >
                  {{ tr.name }}
                </option>
              </select>
            </div>
            <div class="form-field">
              <label for="cap-level">{{ t('teachers.capabilityLevel') }}</label>
              <select
                id="cap-level"
                v-model.number="capForm.levelId"
                :disabled="!capForm.trackId"
                data-testid="cap-form-level"
              >
                <option :value="0">
                  —
                </option>
                <option
                  v-for="lv in filteredLevels()"
                  :key="lv.id"
                  :value="lv.id"
                >
                  {{ lv.name }}
                </option>
              </select>
            </div>
            <p
              v-if="capError"
              class="form-error"
              role="alert"
              aria-live="assertive"
              data-testid="capability-create-error"
            >
              {{ t(capError) }}
            </p>
            <p
              v-if="capError === 'apiErrors.CONFLICT'"
              class="form-hint"
              data-testid="capability-duplicate-hint"
            >
              {{ t('teachers.capabilityDuplicate') }}
            </p>
            <div class="form-actions">
              <button
                type="button"
                :disabled="capSaving"
                @click="showCapForm = false"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="submit"
                :disabled="capSaving || !capForm.domainId || !capForm.trackId || !capForm.levelId"
                data-testid="capability-create-submit"
              >
                {{ capSaving ? t('common.creating') : t('common.create') }}
              </button>
            </div>
          </form>
        </div>
      </section>

      <!-- Availability section -->
      <section
        class="availability-section"
        data-testid="teacher-availability-section"
      >
        <h2>{{ t('teachers.availability') }}</h2>
        <LoadingState v-if="availsLoading" />
        <ErrorState
          v-else-if="availsError"
          :error="availsError"
          @retry="loadAvailability"
        />
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
              <th>{{ t('teachers.availabilityNote') }}</th>
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
              <td>{{ a.note ?? '—' }}</td>
            </tr>
          </tbody>
        </table>
        <button
          type="button"
          data-testid="add-availability-btn"
          @click="openAvailForm"
        >
          {{ t('teachers.addAvailability') }}
        </button>
        <div
          v-if="showAvailForm"
          class="create-dialog"
          data-testid="availability-create-form"
        >
          <form @submit.prevent="handleCreateAvailability">
            <div class="form-field">
              <label for="avail-weekday">{{ t('teachers.availabilityWeekday') }}</label>
              <select
                id="avail-weekday"
                v-model.number="availForm.weekday"
                data-testid="avail-form-weekday"
              >
                <option :value="1">
                  {{ t('teachers.weekdayMonday') }}
                </option>
                <option :value="2">
                  {{ t('teachers.weekdayTuesday') }}
                </option>
                <option :value="3">
                  {{ t('teachers.weekdayWednesday') }}
                </option>
                <option :value="4">
                  {{ t('teachers.weekdayThursday') }}
                </option>
                <option :value="5">
                  {{ t('teachers.weekdayFriday') }}
                </option>
                <option :value="6">
                  {{ t('teachers.weekdaySaturday') }}
                </option>
                <option :value="7">
                  {{ t('teachers.weekdaySunday') }}
                </option>
              </select>
            </div>
            <div class="form-field">
              <label for="avail-start">{{ t('teachers.availabilityStartTime') }}</label>
              <input
                id="avail-start"
                v-model="availForm.startTime"
                type="time"
                required
                data-testid="avail-form-start"
              >
            </div>
            <div class="form-field">
              <label for="avail-end">{{ t('teachers.availabilityEndTime') }}</label>
              <input
                id="avail-end"
                v-model="availForm.endTime"
                type="time"
                required
                data-testid="avail-form-end"
              >
            </div>
            <p
              v-if="availClientError"
              class="form-error"
              role="alert"
              aria-live="assertive"
              data-testid="availability-client-error"
            >
              {{ t(availClientError) }}
            </p>
            <p
              v-if="availError"
              class="form-error"
              role="alert"
              aria-live="assertive"
              data-testid="availability-create-error"
            >
              {{ t(availError) }}
            </p>
            <div class="form-actions">
              <button
                type="button"
                :disabled="availSaving"
                @click="showAvailForm = false"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="submit"
                :disabled="availSaving"
                data-testid="availability-create-submit"
              >
                {{ availSaving ? t('common.creating') : t('common.create') }}
              </button>
            </div>
          </form>
        </div>
      </section>
    </template>

    <!-- End capability confirmation -->
    <ConfirmDialog
      v-if="capToEnd"
      :title="t('teachers.capabilityEnd')"
      :message="t('teachers.capabilityEndConfirm')"
      @confirm="handleEndCapability"
      @cancel="capToEnd = null"
    />
  </div>
</template>

<style scoped>
.teacher-detail-view {
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
