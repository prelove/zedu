<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import DictionaryTabForm from './DictionaryTabForm.vue'
import DictionaryTabTable from './DictionaryTabTable.vue'
import type { Column, DictItem, FormField } from './dictionary-types'

const props = defineProps<{
  items: DictItem[]
  columns: Column[]
  formFields: FormField[]
  loading: boolean
  error: string | null
  page: number
  pageSize: number
  total: number
  canCreate: boolean
  saving: boolean
  formError: string | null
  createLabel: string
  editLabel: string
}>()

const emit = defineEmits<{
  create: [body: Record<string, unknown>]
  edit: [item: DictItem, body: Record<string, unknown>]
  toggle: [item: DictItem]
  pageChange: [page: number]
  retry: []
}>()

const { t } = useI18n()
const editingItem = ref<DictItem | null>(null)
const showForm = ref(false)

const activeItem = computed(() => editingItem.value)

function openCreateForm(): void {
  editingItem.value = null
  showForm.value = true
}

function openEditForm(item: DictItem): void {
  editingItem.value = item
  showForm.value = true
}

function closeForm(): void {
  editingItem.value = null
  showForm.value = false
}

function handleSubmit(body: Record<string, unknown>): void {
  if (activeItem.value) emit('edit', activeItem.value, body)
  else emit('create', body)
}

// Do not carry a completed create/edit form into another dictionary tab while
// its reference options are being refreshed.
watch(
  () => props.saving,
  (saving, wasSaving) => {
    if (wasSaving && !saving && !props.formError) closeForm()
  },
)
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

    <DictionaryTabTable
      :items="items"
      :columns="columns"
      :loading="loading"
      :error="error"
      :page="page"
      :page-size="pageSize"
      :total="total"
      @edit="openEditForm"
      @toggle="(item) => emit('toggle', item)"
      @page-change="(pageNumber) => emit('pageChange', pageNumber)"
      @retry="emit('retry')"
    />

    <DictionaryTabForm
      v-if="showForm"
      :item="activeItem"
      :form-fields="formFields"
      :saving="saving"
      :form-error="formError"
      :create-label="createLabel"
      :edit-label="editLabel"
      @submit="handleSubmit"
      @cancel="closeForm"
    />
  </div>
</template>

<style scoped>
.dict-tab {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.create-btn {
  align-self: flex-start;
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}

.create-btn:disabled {
  color: #6c757d;
  cursor: not-allowed;
}

.form-error {
  color: #dc3545;
  margin: 0;
}

.form-hint {
  color: #6c757d;
  margin: 0;
  font-style: italic;
}
</style>
