<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
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
import CapabilitiesTable from './CapabilitiesTable.vue'
import CapabilityCreateForm from './CapabilityCreateForm.vue'

const props = defineProps<{ teacherId: number; domains: CourseDomain[]; tracks: Track[]; levels: Level[]; dictError: string | null }>()
const { t } = useI18n()
const { items: capabilities, page, pageSize, total, loading, error, setData } = usePaginatedList<Capability>(20)

const showForm = ref(false)
const saving = ref(false)
const formError = ref<string | null>(null)
const capToEnd = ref<Capability | null>(null)

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

async function handleCreate(body: CapabilityWrite) {
  saving.value = true
  formError.value = null
  try {
    await authStore.authedRequest((token) => createCapability(token, props.teacherId, body))
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
    await authStore.authedRequest((token) => updateCapability(token, props.teacherId, capToEnd.value!.id, { status: 'ENDED' }))
    capToEnd.value = null
    await loadCapabilities()
  } catch (err) {
    formError.value = err instanceof ApiError ? errorToI18nKey(err) ?? 'errors.UNKNOWN' : 'errors.UNKNOWN'
  }
}

const domainName = (id: number) => props.domains.find((domain) => domain.id === id)?.name ?? String(id)
const trackName = (id: number) => props.tracks.find((track) => track.id === id)?.name ?? String(id)
const levelName = (id: number) => props.levels.find((level) => level.id === id)?.name ?? String(id)
const capStatusLabel = (status: string) => ({ ACTIVE: t('teachers.statusActive'), ENDED: t('students.statusEnded') }[status] ?? status)

onMounted(loadCapabilities)
watch(page, loadCapabilities)
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
    <CapabilitiesTable
      v-else
      :capabilities="capabilities"
      :domain-name="domainName"
      :track-name="trackName"
      :level-name="levelName"
      :status-label="capStatusLabel"
      @end="(capability) => (capToEnd = capability)"
    />
    <PaginationBar
      v-if="!loading && !error && capabilities.length > 0"
      :page="page"
      :page-size="pageSize"
      :total="total"
      data-testid="capabilities-pagination"
      @page-change="(nextPage) => (page = nextPage)"
    />
    <button
      type="button"
      :disabled="dictError !== null"
      data-testid="add-capability-btn"
      @click="showForm = true; formError = null"
    >
      {{ t('teachers.addCapability') }}
    </button>
    <CapabilityCreateForm
      v-if="showForm"
      :domains="domains"
      :tracks="tracks"
      :levels="levels"
      :saving="saving"
      :form-error="formError"
      @submit="handleCreate"
      @cancel="showForm = false"
    />
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
.capabilities-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.dict-error {
  color: #dc3545;
  font-size: 0.875rem;
}
</style>
