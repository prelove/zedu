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

const props = defineProps<{ studentId: number }>()
const { t } = useI18n()
const { items: parents, page, pageSize, total, loading, error, setData } = usePaginatedList<Parent>(20)
const showCreateForm = ref(false)
const createForm = ref<ParentWrite>({ name: '', email: '', phone: '', relationship: '', isPrimary: false })
const creating = ref(false)
const createError = ref<string | null>(null)
const editingParent = ref<Parent | null>(null)
const editForm = ref<ParentWrite>({})
const editSaving = ref(false)
const editError = ref<string | null>(null)
const editNoChanges = ref(false)

async function loadParents() {
  loading.value = true; error.value = null
  try {
    setData(await authStore.authedRequest((tk) => listParents(tk, props.studentId, { page: page.value, pageSize: pageSize.value })))
  } catch (err) {
    if (err instanceof NetworkError) error.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) error.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else error.value = 'errors.UNKNOWN'
  } finally { loading.value = false }
}
onMounted(loadParents)
watch(page, loadParents)
function onPageChange(n: number) { page.value = n }

function openCreateForm() {
  showCreateForm.value = true; createError.value = null
  createForm.value = { name: '', email: '', phone: '', relationship: '', isPrimary: false }
}
async function handleCreate() {
  creating.value = true; createError.value = null
  try {
    const body: ParentWrite = { ...createForm.value }
    if (!body.email) delete body.email
    if (!body.phone) delete body.phone
    if (!body.relationship) delete body.relationship
    await authStore.authedRequest((tk) => createParent(tk, props.studentId, body))
    showCreateForm.value = false; await loadParents()
  } catch (err) {
    if (err instanceof NetworkError) createError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) createError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else createError.value = 'errors.UNKNOWN'
  } finally { creating.value = false }
}

function openEditForm(p: Parent) {
  editingParent.value = p; editError.value = null; editNoChanges.value = false
  editForm.value = { name: p.name, email: p.email ?? '', phone: p.phone ?? '', relationship: p.relationship ?? '', isPrimary: p.isPrimary }
}
async function handleEditSave() {
  if (!editingParent.value) return
  editSaving.value = true; editError.value = null; editNoChanges.value = false
  const p = editingParent.value
  const body: ParentWrite = {}
  if (editForm.value.name !== p.name) body.name = editForm.value.name
  if ((editForm.value.email ?? '') !== (p.email ?? '')) body.email = editForm.value.email ?? ''
  if ((editForm.value.phone ?? '') !== (p.phone ?? '')) body.phone = editForm.value.phone
  if ((editForm.value.relationship ?? '') !== (p.relationship ?? '')) body.relationship = editForm.value.relationship
  if (editForm.value.isPrimary !== p.isPrimary) body.isPrimary = editForm.value.isPrimary
  if (Object.keys(body).length === 0) { editNoChanges.value = true; editSaving.value = false; return }
  try {
    await authStore.authedRequest((tk) => updateParent(tk, props.studentId, p.id, body))
    editingParent.value = null; await loadParents()
  } catch (err) {
    if (err instanceof NetworkError) editError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) editError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else editError.value = 'errors.UNKNOWN'
  } finally { editSaving.value = false }
}
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
    <table
      v-else
      data-testid="parents-table"
    >
      <thead>
        <tr>
          <th>{{ t('students.parentName') }}</th><th>{{ t('students.parentEmail') }}</th>
          <th>{{ t('students.parentPhone') }}</th><th>{{ t('students.parentRelationship') }}</th>
          <th>{{ t('students.parentIsPrimary') }}</th><th>{{ t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <template
          v-for="p in parents"
          :key="p.id"
        >
          <tr data-testid="parent-row">
            <td>{{ p.name }}</td><td>{{ p.email ?? '—' }}</td><td>{{ p.phone ?? '—' }}</td>
            <td>{{ p.relationship ?? '—' }}</td>
            <td>{{ p.isPrimary ? t('common.yes') : t('common.no') }}</td>
            <td>
              <button
                type="button"
                data-testid="parent-edit-btn"
                @click="openEditForm(p)"
              >
                {{ t('common.edit') }}
              </button>
            </td>
          </tr>
          <tr v-if="editingParent?.id === p.id">
            <td colspan="6">
              <form
                data-testid="parent-edit-form"
                @submit.prevent="handleEditSave"
              >
                <div class="form-field">
                  <label for="parent-edit-name">{{ t('students.parentName') }} *</label>
                  <input
                    id="parent-edit-name"
                    v-model="editForm.name"
                    type="text"
                    required
                    data-testid="parent-edit-name"
                  >
                </div>
                <div class="form-field">
                  <label for="parent-edit-email">{{ t('students.parentEmail') }}</label>
                  <input
                    id="parent-edit-email"
                    v-model="editForm.email"
                    type="email"
                    data-testid="parent-edit-email"
                  >
                </div>
                <div class="form-field">
                  <label for="parent-edit-phone">{{ t('students.parentPhone') }}</label>
                  <input
                    id="parent-edit-phone"
                    v-model="editForm.phone"
                    type="tel"
                    data-testid="parent-edit-phone"
                  >
                </div>
                <div class="form-field">
                  <label for="parent-edit-relationship">{{ t('students.parentRelationship') }}</label>
                  <input
                    id="parent-edit-relationship"
                    v-model="editForm.relationship"
                    type="text"
                    data-testid="parent-edit-relationship"
                  >
                </div>
                <div class="form-field">
                  <label for="parent-edit-primary">
                    <input
                      id="parent-edit-primary"
                      v-model="editForm.isPrimary"
                      type="checkbox"
                      data-testid="parent-edit-primary"
                    >
                    {{ t('students.parentIsPrimary') }}
                  </label>
                </div>
                <p
                  v-if="editNoChanges"
                  class="form-hint"
                  role="alert"
                  data-testid="parent-edit-no-changes"
                >
                  {{ t('common.noChangesToSave') }}
                </p>
                <p
                  v-if="editError"
                  class="form-error"
                  role="alert"
                  aria-live="assertive"
                  data-testid="parent-edit-error"
                >
                  {{ t(editError) }}
                </p>
                <div class="form-actions">
                  <button
                    type="button"
                    :disabled="editSaving"
                    data-testid="parent-edit-cancel"
                    @click="editingParent = null"
                  >
                    {{ t('common.cancel') }}
                  </button>
                  <button
                    type="submit"
                    :disabled="editSaving || !editForm.name"
                    data-testid="parent-edit-submit"
                  >
                    {{ editSaving ? t('common.saving') : t('common.save') }}
                  </button>
                </div>
              </form>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
    <PaginationBar
      v-if="!loading && !error && parents.length > 0"
      data-testid="parents-pagination"
      :page="page"
      :page-size="pageSize"
      :total="total"
      @page-change="onPageChange"
    />
    <button
      type="button"
      data-testid="add-parent-btn"
      @click="openCreateForm"
    >
      {{ t('students.addParent') }}
    </button>
    <div
      v-if="showCreateForm"
      class="create-dialog"
      data-testid="parent-create-form"
    >
      <form @submit.prevent="handleCreate">
        <div class="form-field">
          <label for="parent-form-name">{{ t('students.parentName') }} *</label>
          <input
            id="parent-form-name"
            v-model="createForm.name"
            type="text"
            required
            data-testid="parent-form-name"
          >
        </div>
        <div class="form-field">
          <label for="parent-form-email">{{ t('students.parentEmail') }}</label>
          <input
            id="parent-form-email"
            v-model="createForm.email"
            type="email"
            data-testid="parent-form-email"
          >
        </div>
        <div class="form-field">
          <label for="parent-form-phone">{{ t('students.parentPhone') }}</label>
          <input
            id="parent-form-phone"
            v-model="createForm.phone"
            type="tel"
            data-testid="parent-form-phone"
          >
        </div>
        <div class="form-field">
          <label for="parent-form-relationship">{{ t('students.parentRelationship') }}</label>
          <input
            id="parent-form-relationship"
            v-model="createForm.relationship"
            type="text"
            data-testid="parent-form-relationship"
          >
        </div>
        <div class="form-field">
          <label for="parent-form-primary">
            <input
              id="parent-form-primary"
              v-model="createForm.isPrimary"
              type="checkbox"
              data-testid="parent-form-primary"
            >
            {{ t('students.parentIsPrimary') }}
          </label>
        </div>
        <p
          v-if="createError"
          class="form-error"
          role="alert"
          aria-live="assertive"
          data-testid="parent-create-error"
        >
          {{ t(createError) }}
        </p>
        <div class="form-actions">
          <button
            type="button"
            :disabled="creating"
            data-testid="parent-create-cancel"
            @click="showCreateForm = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="creating || !createForm.name"
            data-testid="parent-create-submit"
          >
            {{ creating ? t('common.creating') : t('common.create') }}
          </button>
        </div>
      </form>
    </div>
  </section>
</template>

<style scoped>
.parents-section { margin-top: 1.5rem; }
.parents-section h2 { margin-bottom: 0.5rem; }
table { width: 100%; border-collapse: collapse; margin-bottom: 0.5rem; }
th, td { border: 1px solid #dee2e6; padding: 0.5rem; text-align: left; }
.form-field { margin-bottom: 0.5rem; }
.form-field label { display: block; font-weight: 600; margin-bottom: 0.25rem; }
.form-field input { width: 100%; padding: 0.25rem 0.5rem; border: 1px solid #ccc; border-radius: 0.25rem; }
.form-error { color: #dc3545; margin: 0.5rem 0; }
.form-hint { color: #6c757d; margin: 0.5rem 0; }
.form-actions { display: flex; gap: 0.5rem; margin-top: 0.5rem; }
.create-dialog { margin-top: 1rem; padding: 1rem; border: 1px solid #dee2e6; border-radius: 0.375rem; background-color: #f8f9fa; }
button { padding: 0.25rem 0.75rem; border: 1px solid #ccc; border-radius: 0.25rem; background-color: #fff; cursor: pointer; }
button:disabled { color: #6c757d; cursor: not-allowed; }
</style>
