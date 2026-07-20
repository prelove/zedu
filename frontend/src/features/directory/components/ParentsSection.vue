<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { listParents, createParent, updateParent, type Parent, type ParentWrite } from '../../../api/directory'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import LoadingState from '../../../components/LoadingState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import { usePaginatedList } from '../../../composables/usePaginatedList'
import ParentCreateForm from './ParentCreateForm.vue'
import ParentEditForm from './ParentEditForm.vue'
import ParentsTable from './ParentsTable.vue'

const props = defineProps<{ studentId: number }>()
const { t } = useI18n()
const { items: parents, page, pageSize, total, loading, error, setData } = usePaginatedList<Parent>(20)

const showCreateForm = ref(false)
const creating = ref(false)
const createError = ref<string | null>(null)
const editingParent = ref<Parent | null>(null)
const editSaving = ref(false)
const editError = ref<string | null>(null)

async function loadParents() {
  loading.value = true
  error.value = null
  try {
    setData(await authStore.authedRequest((token) => listParents(token, props.studentId, { page: page.value, pageSize: pageSize.value })))
  } catch (err) {
    if (err instanceof NetworkError) error.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) error.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else error.value = 'errors.UNKNOWN'
  } finally {
    loading.value = false
  }
}

async function handleCreate(body: ParentWrite) {
  creating.value = true
  createError.value = null
  try {
    await authStore.authedRequest((token) => createParent(token, props.studentId, body))
    showCreateForm.value = false
    await loadParents()
  } catch (err) {
    if (err instanceof NetworkError) createError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) createError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else createError.value = 'errors.UNKNOWN'
  } finally {
    creating.value = false
  }
}

async function handleEditSave(body: ParentWrite) {
  if (!editingParent.value) return
  editSaving.value = true
  editError.value = null
  try {
    await authStore.authedRequest((token) => updateParent(token, props.studentId, editingParent.value!.id, body))
    editingParent.value = null
    await loadParents()
  } catch (err) {
    if (err instanceof NetworkError) editError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) editError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else editError.value = 'errors.UNKNOWN'
  } finally {
    editSaving.value = false
  }
}

onMounted(loadParents)
watch(page, loadParents)
</script>

<template>
  <section
    class="parents-section"
    data-testid="student-parents-section"
  >
    <h2>{{ t('students.parents') }}</h2>
    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadParents"
    />
    <EmptyState
      v-else-if="parents.length === 0"
      :message="t('students.parentsEmpty')"
    />
    <ParentsTable
      v-else
      :parents="parents"
      @edit="(parent) => { editingParent = parent; editError = null }"
    />
    <ParentEditForm
      v-if="editingParent"
      :parent="editingParent"
      :saving="editSaving"
      :error="editError"
      @submit="handleEditSave"
      @cancel="editingParent = null"
    />
    <PaginationBar
      v-if="!loading && !error && parents.length > 0"
      data-testid="parents-pagination"
      :page="page"
      :page-size="pageSize"
      :total="total"
      @page-change="(nextPage) => (page = nextPage)"
    />
    <button
      type="button"
      data-testid="add-parent-btn"
      @click="showCreateForm = true; createError = null"
    >
      {{ t('students.addParent') }}
    </button>
    <ParentCreateForm
      v-if="showCreateForm"
      :creating="creating"
      :create-error="createError"
      @submit="handleCreate"
      @cancel="showCreateForm = false"
    />
  </section>
</template>

<style scoped>
.parents-section {
  margin-top: 1.5rem;
}

.parents-section h2 {
  margin-bottom: 0.5rem;
}

button {
  padding: 0.25rem 0.75rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}

button:disabled {
  color: #6c757d;
  cursor: not-allowed;
}
</style>
