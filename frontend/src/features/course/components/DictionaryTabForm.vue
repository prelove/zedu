<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { DictItem, FormField } from './dictionary-types'

const props = defineProps<{
  item: DictItem | null
  formFields: FormField[]
  saving: boolean
  formError: string | null
  createLabel: string
  editLabel: string
}>()

const emit = defineEmits<{
  submit: [body: Record<string, unknown>]
  cancel: []
}>()

const { t } = useI18n()
const form = ref<Record<string, unknown>>({})
const noChanges = ref(false)

function applyInitialState(): void {
  const nextState: Record<string, unknown> = {}
  for (const field of props.formFields) {
    if (props.item) nextState[field.key] = props.item[field.key] ?? ''
    else if (field.defaultValue !== undefined) nextState[field.key] = field.defaultValue
    else if (field.type === 'number') nextState[field.key] = 0
    else if (field.type === 'select') nextState[field.key] = field.options?.[0]?.value ?? 0
    else nextState[field.key] = ''
  }
  form.value = nextState
  noChanges.value = false
}

function buildEditBody(item: DictItem): Record<string, unknown> {
  const body: Record<string, unknown> = {}
  for (const field of props.formFields) {
    const original = item[field.key]
    const current = form.value[field.key]
    if (String(current ?? '') !== String(original ?? '')) body[field.key] = current
  }
  return body
}

function handleSubmit(): void {
  if (props.item) {
    const body = buildEditBody(props.item)
    if (Object.keys(body).length === 0) {
      noChanges.value = true
      return
    }
    emit('submit', body)
    return
  }
  emit('submit', { ...form.value })
}

watch(
  () => [props.item, props.formFields] as const,
  () => applyInitialState(),
  { immediate: true },
)
</script>

<template>
  <form
    :data-testid="item ? 'course-edit-form' : 'course-create-form'"
    class="dict-form"
    @submit.prevent="handleSubmit"
  >
    <h2>{{ item ? t(editLabel) : t(createLabel) }}</h2>
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
      v-for="field in formFields"
      :key="field.key"
      class="form-field"
    >
      <label :for="`field-${field.key}`">{{ t(field.label) }}</label>
      <input
        v-if="field.type === 'text'"
        :id="`field-${field.key}`"
        v-model="form[field.key]"
        type="text"
        :data-testid="field.testid"
      >
      <input
        v-else-if="field.type === 'number'"
        :id="`field-${field.key}`"
        v-model.number="form[field.key]"
        type="number"
        :data-testid="field.testid"
      >
      <select
        v-else
        :id="`field-${field.key}`"
        v-model="form[field.key]"
        :data-testid="field.testid"
      >
        <option
          v-for="option in field.options"
          :key="option.value"
          :value="option.value"
        >
          {{ option.label }}
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
        @click="emit('cancel')"
      >
        {{ t('common.cancel') }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.dict-form {
  border: 1px solid #dee2e6;
  border-radius: 0.25rem;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  background-color: #f8f9fa;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-field input,
.form-field select {
  padding: 0.25rem 0.5rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
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

.form-actions {
  display: flex;
  gap: 0.5rem;
}

.form-actions button {
  padding: 0.25rem 0.75rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}

.form-actions button:disabled {
  color: #6c757d;
  cursor: not-allowed;
}
</style>
