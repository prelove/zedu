<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { listCapabilities, createCapability, updateCapability, type Capability, type CapabilityWrite } from '../../../api/directory'
import { type CourseDomain, type Track, type Level } from '../../../api/course-dict'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import LoadingState from '../../../components/LoadingState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import ConfirmDialog from '../../../components/ConfirmDialog.vue'
import { usePaginatedList } from '../../../composables/usePaginatedList'

const props = defineProps<{ teacherId: number; domains: CourseDomain[]; tracks: Track[]; levels: Level[]; dictError: string | null }>()
const { t } = useI18n()
const { items: capabilities, page, pageSize, total, loading, error, setData } = usePaginatedList<Capability>(20)

const showForm = ref(false)
const form = ref<CapabilityWrite>({ domainId: 0, trackId: 0, levelId: 0 })
const saving = ref(false)
const formError = ref<string | null>(null)
const capToEnd = ref<Capability | null>(null)

const filteredTracks = computed(() => props.tracks.filter((tr) => tr.domainId === form.value.domainId))
const filteredLevels = computed(() => props.levels.filter((l) => l.trackId === form.value.trackId))

async function loadCapabilities() {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) => listCapabilities(token, props.teacherId, { page: page.value, pageSize: pageSize.value }))
    setData(data)
  } catch (err) {
    if (err instanceof NetworkError) error.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) error.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else error.value = 'errors.UNKNOWN'
  } finally {
    loading.value = false
  }
}

onMounted(loadCapabilities)
watch(page, loadCapabilities)

function openForm() {
  showForm.value = true
  formError.value = null
  form.value = { domainId: 0, trackId: 0, levelId: 0 }
}
function onDomainChange() { form.value.trackId = 0; form.value.levelId = 0 }
function onTrackChange() { form.value.levelId = 0 }

async function handleCreate() {
  saving.value = true
  formError.value = null
  try {
    await authStore.authedRequest((token) => createCapability(token, props.teacherId, form.value))
    showForm.value = false
    await loadCapabilities()
  } catch (err) {
    if (err instanceof NetworkError) formError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) formError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else formError.value = 'errors.UNKNOWN'
  } finally {
    saving.value = false
  }
}

async function handleEndCapability() {
  if (!capToEnd.value) return
  try {
    await authStore.authedRequest((token) => updateCapability(token, props.teacherId, capToEnd.value.id, { status: 'ENDED' }))
    capToEnd.value = null
    await loadCapabilities()
  } catch (err) {
    if (err instanceof ApiError) formError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else formError.value = 'errors.UNKNOWN'
  }
}

function domainName(id: number): string { return props.domains.find((d) => d.id === id)?.name ?? String(id) }
function trackName(id: number): string { return props.tracks.find((tr) => tr.id === id)?.name ?? String(id) }
function levelName(id: number): string { return props.levels.find((l) => l.id === id)?.name ?? String(id) }
function capStatusLabel(status: string): string {
  const map: Record<string, string> = { ACTIVE: t('teachers.statusActive'), ENDED: t('students.statusEnded') }
  return map[status] ?? status
}
</script>

<template>
  <section
    class="capabilities-section"
    data-testid="capabilities-section"
  >
    <h2>{{ t('teachers.capabilities') }}</h2>
    <p
      v-if="dictError"
      class="dict-error"
      role="alert"
      aria-live="assertive"
    >
      {{ t(dictError) }}
    </p>
    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
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
          <th>{{ t('common.status') }}</th>
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
          <td>{{ capStatusLabel(c.status) }}</td>
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
    <PaginationBar
      v-if="!loading && !error && capabilities.length > 0"
      :page="page"
      :page-size="pageSize"
      :total="total"
      data-testid="capabilities-pagination"
      @page-change="(p) => (page = p)"
    />
    <button
      type="button"
      :disabled="dictError !== null"
      data-testid="add-capability-btn"
      @click="openForm"
    >
      {{ t('teachers.addCapability') }}
    </button>
    <div
      v-if="showForm"
      class="create-dialog"
      data-testid="capability-create-form"
    >
      <form @submit.prevent="handleCreate">
        <div class="form-field">
          <label for="cap-domain">{{ t('teachers.capabilityDomain') }}</label>
          <select
            id="cap-domain"
            v-model.number="form.domainId"
            data-testid="cap-form-domain"
            @change="onDomainChange"
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
            v-model.number="form.trackId"
            :disabled="!form.domainId"
            data-testid="cap-form-track"
            @change="onTrackChange"
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
          <label for="cap-level">{{ t('teachers.capabilityLevel') }}</label>
          <select
            id="cap-level"
            v-model.number="form.levelId"
            :disabled="!form.trackId"
            data-testid="cap-form-level"
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
        <p
          v-if="formError"
          class="form-error"
          role="alert"
          aria-live="assertive"
          data-testid="capability-create-error"
        >
          {{ t(formError) }}
        </p>
        <p
          v-if="formError === 'apiErrors.CONFLICT'"
          class="form-hint"
          data-testid="capability-duplicate-hint"
        >
          {{ t('teachers.capabilityDuplicate') }}
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
            :disabled="saving || !form.domainId || !form.trackId || !form.levelId"
            data-testid="capability-create-submit"
          >
            {{ saving ? t('common.creating') : t('common.create') }}
          </button>
        </div>
      </form>
    </div>
    <ConfirmDialog
      v-if="capToEnd"
      :title="t('teachers.capabilityEnd')"
      :message="t('teachers.capabilityEndConfirm')"
      data-testid="confirm-dialog"
      @confirm="handleEndCapability"
      @cancel="capToEnd = null"
    />
  </section>
</template>

<style scoped>
.capabilities-section { display: flex; flex-direction: column; gap: 0.75rem; }
.dict-error { color: #dc3545; font-size: 0.875rem; }
.capabilities-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 0.25rem; }
.capability-row { display: grid; grid-template-columns: 1.2fr 1.2fr 1fr 0.8fr 0.6fr auto; gap: 0.5rem; padding: 0.4rem 0.5rem; border-bottom: 1px solid #eee; font-size: 0.875rem; }
.cap-actions button { padding: 0.2rem 0.5rem; border: 1px solid #ccc; border-radius: 0.25rem; background-color: #fff; cursor: pointer; }
.create-dialog { border: 1px solid #ccc; border-radius: 0.5rem; padding: 1rem; background-color: #f9f9f9; }
.form-field { display: flex; flex-direction: column; gap: 0.25rem; margin-bottom: 0.75rem; }
.form-field label { font-size: 0.8rem; color: #495057; }
.form-field select { padding: 0.3rem; border: 1px solid #ccc; border-radius: 0.25rem; }
.form-error { color: #dc3545; font-size: 0.8rem; margin: 0 0 0.5rem; }
.form-hint { color: #856404; font-size: 0.8rem; margin: 0 0 0.5rem; }
.form-actions { display: flex; gap: 0.5rem; justify-content: flex-end; }
.form-actions button { padding: 0.3rem 0.75rem; border: 1px solid #ccc; border-radius: 0.25rem; cursor: pointer; }
.form-actions button[type='submit'] { background-color: #0d6efd; color: #fff; border-color: #0d6efd; }
.form-actions button:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
