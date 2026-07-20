<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
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
import EnrollmentCreateForm from './EnrollmentCreateForm.vue'
import EnrollmentsTable from './EnrollmentsTable.vue'

const props = defineProps<{ studentId: number; studentStatus: string }>()
const emit = defineEmits<{ viewEnrollment: [id: number] }>()
const { t } = useI18n()

const { items: enrollments, page, pageSize, total, loading, error, setData } = usePaginatedList<Enrollment>(20)
const { domains, tracks, levels, error: dictError, load: loadDict } = useDictionary()

const showCreateForm = ref(false)
const creating = ref(false)
const createError = ref<string | null>(null)
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

async function handleCreate(body: EnrollmentWrite) {
  creating.value = true
  createError.value = null
  try {
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
    ACTIVE: t('enrollments.statusActive'),
    PAUSED: t('enrollments.statusPaused'),
    COMPLETED: t('enrollments.statusCompleted'),
    CANCELLED: t('enrollments.statusCancelled'),
  }
  return map[status] ?? status
}

const domainName = (id: number) => domains.value.find((domain) => domain.id === id)?.name ?? String(id)
const trackName = (id: number) => tracks.value.find((track) => track.id === id)?.name ?? String(id)

onMounted(() => {
  void loadEnrollments()
  void loadDict()
})
watch(page, loadEnrollments)
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
        @click="showCreateForm = true; createError = null"
      >
        {{ t('enrollments.create') }}
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

    <EnrollmentCreateForm
      v-if="showCreateForm"
      :domains="domains"
      :tracks="tracks"
      :levels="levels"
      :creating="creating"
      :create-error="createError"
      @submit="handleCreate"
      @cancel="showCreateForm = false"
    />

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadEnrollments"
    />
    <template v-else>
      <EmptyState v-if="!enrollments.length" />
      <EnrollmentsTable
        v-else
        :enrollments="enrollments"
        :domain-name="domainName"
        :track-name="trackName"
        :status-label="statusLabel"
        @view-enrollment="(id) => emit('viewEnrollment', id)"
      />
      <PaginationBar
        v-if="total > 0"
        data-testid="enrollments-pagination"
        :page="page"
        :page-size="pageSize"
        :total="total"
        @page-change="(nextPage) => (page = nextPage)"
      />
    </template>
  </section>
</template>

<style scoped>
.enrollments-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-header h2 {
  margin: 0;
  font-size: 1.125rem;
}

.dict-error {
  color: #dc3545;
}

.dict-error button {
  border: 1px solid #dc3545;
  background: #fff;
  color: #dc3545;
  border-radius: 0.25rem;
  padding: 0.25rem 0.5rem;
  cursor: pointer;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
