<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingState from '../../../components/LoadingState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'

interface DictItem { id: number; name: string; code: string; enabled: boolean; [k: string]: unknown }
interface Column { key: string; label: string }
interface FormField {
  key: string; label: string; type: 'text' | 'number' | 'select'; testid: string
  options?: { value: number | string; label: string }[]
  defaultValue?: unknown
}

const props = defineProps<{
  items: DictItem[]; columns: Column[]; formFields: FormField[]
  loading: boolean; error: string | null
  page: number; pageSize: number; total: number; hasNext: boolean; hasPrev: boolean
  canCreate: boolean; saving: boolean; formError: string | null
  createLabel: string; editLabel: string
}>()

const emit = defineEmits<{
  create: [body: Record<string, unknown>]
  edit: [item: DictItem, body: Record<string, unknown>]
  toggle: [item: DictItem]
  pageChange: [page: number]
  retry: []
}>()

const { t } = useI18n()
const showForm = ref(false)
const isEdit = ref(false)
const editingItem = ref<DictItem | null>(null)
const form = ref<Record<string, unknown>>({})
const noChanges = ref(false)

function initForm(source: DictItem | null): void {
  form.value = {}
  for (const f of props.formFields) {
    if (source) form.value[f.key] = source[f.key] ?? ''
    else if (f.defaultValue !== undefined) form.value[f.key] = f.defaultValue
    else if (f.type === 'number') form.value[f.key] = 0
    else if (f.type === 'select') form.value[f.key] = f.options?.[0]?.value ?? 0
    else form.value[f.key] = ''
  }
}

function openCreateForm(): void {
  isEdit.value = false; editingItem.value = null; showForm.value = true; noChanges.value = false
  initForm(null)
}

function openEditForm(item: DictItem): void {
  isEdit.value = true; editingItem.value = item; showForm.value = true; noChanges.value = false
  initForm(item)
}

function closeForm(): void {
  showForm.value = false; editingItem.value = null; noChanges.value = false
}

function handleSubmit(): void {
  if (isEdit.value && editingItem.value) {
    const body: Record<string, unknown> = {}
    for (const f of props.formFields) {
      const original = editingItem.value[f.key]
      const current = form.value[f.key]
      if (String(current ?? '') !== String(original ?? '')) body[f.key] = current
    }
    if (Object.keys(body).length === 0) { noChanges.value = true; return }
    emit('edit', editingItem.value, body)
  } else {
    emit('create', { ...form.value })
  }
}
</script>

<template>
  <div
    class="dict-tab"
    data-testid="course-dict-tab"
  >
    <button
      type="button"
      class="create-btn"
      :disabled="!canCreate"
      data-testid="course-create-btn"
      @click="openCreateForm"
    >
      {{ t(createLabel) }}
    </button>

    <p
      v-if="formError && !showForm"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="course-form-error"
    >
      {{ t(formError) }}
    </p>
    <p
      v-if="formError === 'apiErrors.INVALID_STATE' && !showForm"
      class="form-hint"
      data-testid="course-referenced-hint"
    >
      {{ t('courses.referencedCannotDisable') }}
    </p>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="emit('retry')"
    />
    <EmptyState v-else-if="items.length === 0" />
    <table
      v-else
      data-testid="course-table"
    >
      <thead>
        <tr>
          <th
            v-for="col in columns"
            :key="col.key"
          >
            {{ t(col.label) }}
          </th>
          <th>{{ t('common.status') }}</th>
          <th>{{ t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="item in items"
          :key="item.id"
          data-testid="course-row"
        >
          <td
            v-for="col in columns"
            :key="col.key"
          >
            {{ item[col.key] }}
          </td>
          <td>{{ item.enabled ? t('common.enabled') : t('common.disabled') }}</td>
          <td class="actions">
            <button
              type="button"
              :data-testid="`course-edit-${item.id}`"
              @click="openEditForm(item)"
            >
              {{ t('common.edit') }}
            </button>
            <button
              type="button"
              :data-testid="`course-toggle-${item.id}`"
              @click="emit('toggle', item)"
            >
              {{ item.enabled ? t('common.disable') : t('common.enable') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <PaginationBar
      v-if="!loading && !error && items.length > 0"
      :page="page"
      :page-size="pageSize"
      :total="total"
      @page-change="(p) => emit('pageChange', p)"
    />

    <form
      v-if="showForm"
      :data-testid="isEdit ? 'course-edit-form' : 'course-create-form'"
      class="dict-form"
      @submit.prevent="handleSubmit"
    >
      <h2>{{ isEdit ? t(editLabel) : t(createLabel) }}</h2>
      <p
        v-if="formError"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="course-form-error"
      >
        {{ t(formError) }}
      </p>
      <p
        v-if="noChanges"
        class="form-hint"
        data-testid="course-no-changes"
      >
        {{ t('common.noChangesToSave') }}
      </p>
      <p
        v-if="formError === 'apiErrors.INVALID_STATE'"
        class="form-hint"
        data-testid="course-referenced-hint"
      >
        {{ t('courses.referencedCannotDisable') }}
      </p>
      <div
        v-for="f in formFields"
        :key="f.key"
        class="form-field"
      >
        <label :for="`field-${f.key}`">{{ t(f.label) }}</label>
        <input
          v-if="f.type === 'text'"
          :id="`field-${f.key}`"
          v-model="form[f.key]"
          type="text"
          :data-testid="f.testid"
        >
        <input
          v-else-if="f.type === 'number'"
          :id="`field-${f.key}`"
          v-model.number="form[f.key]"
          type="number"
          :data-testid="f.testid"
        >
        <select
          v-else
          :id="`field-${f.key}`"
          v-model="form[f.key]"
          :data-testid="f.testid"
        >
          <option
            v-for="opt in f.options"
            :key="opt.value"
            :value="opt.value"
          >
            {{ opt.label }}
          </option>
        </select>
      </div>
      <div class="form-actions">
        <button
          type="submit"
          :disabled="saving"
          data-testid="course-submit"
        >
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
        <button
          type="button"
          data-testid="course-cancel"
          @click="closeForm"
        >
          {{ t('common.cancel') }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.dict-tab { display: flex; flex-direction: column; gap: 1rem; }
.create-btn { align-self: flex-start; padding: 0.5rem 1rem; border: 1px solid #ccc;
  border-radius: 0.25rem; background-color: #fff; cursor: pointer; }
.create-btn:disabled { color: #6c757d; cursor: not-allowed; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 0.5rem; border: 1px solid #dee2e6; text-align: left; }
.actions { display: flex; gap: 0.5rem; }
.actions button { padding: 0.25rem 0.5rem; border: 1px solid #ccc;
  border-radius: 0.25rem; background-color: #fff; cursor: pointer; }
.dict-form { border: 1px solid #dee2e6; border-radius: 0.25rem; padding: 1rem;
  display: flex; flex-direction: column; gap: 0.75rem; background-color: #f8f9fa; }
.form-field { display: flex; flex-direction: column; gap: 0.25rem; }
.form-field input, .form-field select { padding: 0.25rem 0.5rem;
  border: 1px solid #ccc; border-radius: 0.25rem; }
.form-error { color: #dc3545; margin: 0; }
.form-hint { color: #6c757d; margin: 0; font-style: italic; }
.form-actions { display: flex; gap: 0.5rem; }
.form-actions button { padding: 0.25rem 0.75rem; border: 1px solid #ccc;
  border-radius: 0.25rem; background-color: #fff; cursor: pointer; }
.form-actions button:disabled { color: #6c757d; cursor: not-allowed; }
</style>
