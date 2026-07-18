<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { type Enrollment } from '../../../api/directory'
import { listEnrollments, createEnrollment, type EnrollmentWrite } from '../../../api/course'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import LoadingState from '../../../components/LoadingState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import { usePaginatedList } from '../../../composables/usePaginatedList'
import { useDictionary } from '../../../composables/useDictionary'

const props = defineProps<{ studentId: number; studentStatus: string }>()
const emit = defineEmits<{ viewEnrollment: [id: number] }>()
const { t } = useI18n()

const { items: enrollments, page, pageSize, total, loading, error, setData } = usePaginatedList<Enrollment>(20)
const { domains, tracks, levels, error: dictError, load: loadDict } = useDictionary()

const showCreateForm = ref(false)
const selDomain = ref<number>(0)
const selTrack = ref<number>(0)
const selCurrentLevel = ref<number>(0)
const selTargetLevel = ref<number>(0)
const selType = ref<string>('R')
const creating = ref(false)
const createError = ref<string | null>(null)

const filteredTracks = computed(() => tracks.value.filter((tr) => tr.domainId === selDomain.value))
const filteredLevels = computed(() => levels.value.filter((l) => l.trackId === selTrack.value))
const canCreate = computed(() => props.studentStatus === 'ACTIVE' && !dictError.value)

async function loadEnrollments() {
  loading.value = true
  error.value = null
  try {
    const data = await authStore.authedRequest((token) => listEnrollments(token, props.studentId, { page: page.value, pageSize: pageSize.value }))
    setData(data)
  } catch (err) {
    if (err instanceof NetworkError) error.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) error.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else error.value = 'errors.UNKNOWN'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadEnrollments()
  void loadDict()
})
watch(page, loadEnrollments)

function openCreateForm() {
  showCreateForm.value = true
  createError.value = null
  selDomain.value = 0
  selTrack.value = 0
  selCurrentLevel.value = 0
  selTargetLevel.value = 0
  selType.value = 'R'
}

async function handleCreate() {
  creating.value = true
  createError.value = null
  try {
    const body: EnrollmentWrite = {
      domainId: selDomain.value,
      trackId: selTrack.value,
      enrollmentType: selType.value,
    }
    if (selCurrentLevel.value) body.currentLevelId = selCurrentLevel.value
    if (selTargetLevel.value) body.targetLevelId = selTargetLevel.value
    await authStore.authedRequest((token) => createEnrollment(token, props.studentId, body))
    showCreateForm.value = false
    await loadEnrollments()
  } catch (err) {
    if (err instanceof NetworkError) createError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) createError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else createError.value = 'errors.UNKNOWN'
  } finally {
    creating.value = false
  }
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    ACTIVE: t('enrollments.statusActive'), PAUSED: t('enrollments.statusPaused'),
    COMPLETED: t('enrollments.statusCompleted'), CANCELLED: t('enrollments.statusCancelled'),
  }
  return map[status] ?? status
}

function domainName(id: number): string {
  return domains.value.find((d) => d.id === id)?.name ?? String(id)
}
function trackName(id: number): string {
  return tracks.value.find((tr) => tr.id === id)?.name ?? String(id)
}
</script>

<template>
  <section
    data-testid="student-enrollments-section"
    class="enrollments-section"
  >
    <header class="section-header">
      <h2>{{ t('enrollments.title') }}</h2>
      <button
        type="button"
        data-testid="add-enrollment-btn"
        :disabled="!canCreate"
        @click="openCreateForm"
      >
        {{ t('enrollments.add') }}
      </button>
    </header>

    <div
      v-if="dictError"
      class="dict-error"
      data-testid="enrollment-dict-error"
      role="alert"
    >
      <p>{{ t(dictError) }}</p>
      <button
        type="button"
        data-testid="enrollment-dict-retry"
        @click="loadDict"
      >
        {{ t('common.retry') }}
      </button>
    </div>

    <form
      v-if="showCreateForm"
      data-testid="enrollment-create-form"
      class="create-form"
      @submit.prevent="handleCreate"
    >
      <div class="form-row">
        <label for="enrollment-form-domain">{{ t('enrollments.domain') }}</label>
        <select
          id="enrollment-form-domain"
          v-model.number="selDomain"
          data-testid="enrollment-form-domain"
        >
          <option :value="0">
            {{ t('common.select') }}
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
      <div class="form-row">
        <label for="enrollment-form-track">{{ t('enrollments.track') }}</label>
        <select
          id="enrollment-form-track"
          v-model.number="selTrack"
          data-testid="enrollment-form-track"
          :disabled="!selDomain"
        >
          <option :value="0">
            {{ t('common.select') }}
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
      <div class="form-row">
        <label for="enrollment-form-current-level">{{ t('enrollments.currentLevel') }}</label>
        <select
          id="enrollment-form-current-level"
          v-model.number="selCurrentLevel"
          data-testid="enrollment-form-current-level"
          :disabled="!selTrack"
        >
          <option :value="0">
            {{ t('common.none') }}
          </option>
          <option
            v-for="l in filteredLevels"
            :key="l.id"
            :value="l.id"
          >
            {{ l.name }}
          </option>
        </select>
      </div>
      <div class="form-row">
        <label for="enrollment-form-target-level">{{ t('enrollments.targetLevel') }}</label>
        <select
          id="enrollment-form-target-level"
          v-model.number="selTargetLevel"
          data-testid="enrollment-form-target-level"
          :disabled="!selTrack"
        >
          <option :value="0">
            {{ t('common.none') }}
          </option>
          <option
            v-for="l in filteredLevels"
            :key="l.id"
            :value="l.id"
          >
            {{ l.name }}
          </option>
        </select>
      </div>
      <div class="form-row">
        <label for="enrollment-form-type">{{ t('enrollments.type') }}</label>
        <select
          id="enrollment-form-type"
          v-model="selType"
          data-testid="enrollment-form-type"
        >
          <option value="R">
            {{ t('enrollments.typeRegular') }}
          </option>
          <option value="T">
            {{ t('enrollments.typeTrial') }}
          </option>
        </select>
      </div>
      <p
        v-if="createError"
        class="create-error"
        data-testid="enrollment-create-error"
        role="alert"
      >
        {{ t(createError) }}
      </p>
      <div class="form-actions">
        <button
          type="button"
          data-testid="enrollment-create-cancel"
          @click="showCreateForm = false"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          data-testid="enrollment-create-submit"
          :disabled="creating || !selDomain || !selTrack"
        >
          {{ t('common.create') }}
        </button>
      </div>
    </form>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadEnrollments"
    />
    <template v-else>
      <EmptyState v-if="!enrollments.length" />
      <table
        v-else
        data-testid="enrollments-table"
        class="enrollments-table"
      >
        <thead>
          <tr><th>{{ t('enrollments.domain') }}</th><th>{{ t('enrollments.track') }}</th><th>{{ t('enrollments.type') }}</th><th>{{ t('enrollments.status') }}</th><th>{{ t('common.actions') }}</th></tr>
        </thead>
        <tbody>
          <tr
            v-for="e in enrollments"
            :key="e.id"
            data-testid="enrollment-row"
          >
            <td>{{ domainName(e.domainId) }}</td>
            <td>{{ trackName(e.trackId) }}</td>
            <td>{{ e.enrollmentType === 'R' ? t('enrollments.typeRegular') : t('enrollments.typeTrial') }}</td>
            <td>{{ statusLabel(e.status) }}</td>
            <td>
              <button
                type="button"
                data-testid="view-enrollment-btn"
                @click="emit('viewEnrollment', e.id)"
              >
                {{ t('common.view') }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <PaginationBar
        v-if="total > 0"
        data-testid="enrollments-pagination"
        :page="page"
        :page-size="pageSize"
        :total="total"
        @page-change="(p) => { page = p }"
      />
    </template>
  </section>
</template>

<style scoped>
.enrollments-section { display: flex; flex-direction: column; gap: 0.75rem; }
.section-header { display: flex; justify-content: space-between; align-items: center; }
.section-header h2 { margin: 0; font-size: 1.125rem; }
.dict-error { color: #dc3545; }
.dict-error button { border: 1px solid #dc3545; background: #fff; color: #dc3545; border-radius: 0.25rem; padding: 0.25rem 0.5rem; cursor: pointer; }
.create-form { display: flex; flex-direction: column; gap: 0.5rem; padding: 0.75rem; border: 1px solid #dee2e6; border-radius: 0.25rem; background: #f8f9fa; }
.form-row { display: flex; flex-direction: column; gap: 0.25rem; }
.form-row label { font-size: 0.875rem; font-weight: 500; }
.form-row select { padding: 0.25rem 0.5rem; border: 1px solid #ccc; border-radius: 0.25rem; }
.create-error { color: #dc3545; margin: 0; font-size: 0.875rem; }
.form-actions { display: flex; gap: 0.5rem; justify-content: flex-end; }
.form-actions button { padding: 0.25rem 0.75rem; border: 1px solid #ccc; border-radius: 0.25rem; background: #fff; cursor: pointer; }
.form-actions button[type='submit'] { background: #0d6efd; color: #fff; border-color: #0d6efd; }
.form-actions button:disabled { opacity: 0.6; cursor: not-allowed; }
.enrollments-table { width: 100%; border-collapse: collapse; }
.enrollments-table th, .enrollments-table td { padding: 0.5rem; border-bottom: 1px solid #dee2e6; text-align: left; font-size: 0.875rem; }
.enrollments-table th { background: #f8f9fa; font-weight: 600; }
.enrollments-table button { padding: 0.2rem 0.5rem; border: 1px solid #ccc; border-radius: 0.25rem; background: #fff; cursor: pointer; font-size: 0.8125rem; }
button:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
