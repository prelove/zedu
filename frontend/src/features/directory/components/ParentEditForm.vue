<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Parent, ParentWrite } from '../../../api/directory'

const props = defineProps<{
  parent: Parent
  saving: boolean
  error: string | null
}>()

const emit = defineEmits<{
  submit: [body: ParentWrite]
  cancel: []
}>()

const { t } = useI18n()
const form = ref<ParentWrite>({})
const noChanges = ref(false)

function syncForm(): void {
  form.value = {
    name: props.parent.name,
    email: props.parent.email ?? '',
    phone: props.parent.phone ?? '',
    relationship: props.parent.relationship ?? '',
    isPrimary: props.parent.isPrimary,
  }
  noChanges.value = false
}

function buildBody(): ParentWrite {
  const body: ParentWrite = {}
  if (form.value.name !== props.parent.name) body.name = form.value.name
  if ((form.value.email ?? '') !== (props.parent.email ?? '')) body.email = form.value.email ?? ''
  if ((form.value.phone ?? '') !== (props.parent.phone ?? '')) body.phone = form.value.phone
  if ((form.value.relationship ?? '') !== (props.parent.relationship ?? '')) body.relationship = form.value.relationship
  if (form.value.isPrimary !== props.parent.isPrimary) body.isPrimary = form.value.isPrimary
  return body
}

function handleSubmit(): void {
  const body = buildBody()
  if (Object.keys(body).length === 0) {
    noChanges.value = true
    return
  }
  emit('submit', body)
}

watch(() => props.parent, syncForm, { immediate: true })
</script>

<template>
  <form
    data-testid="parent-edit-form"
    @submit.prevent="handleSubmit"
  >
    <div class="form-field">
      <label for="parent-edit-name">{{ t('students.parentName') }} *</label>
      <input
        id="parent-edit-name"
        v-model="form.name"
        type="text"
        required
        data-testid="parent-edit-name"
      >
    </div>
    <div class="form-field">
      <label for="parent-edit-email">{{ t('students.parentEmail') }}</label>
      <input
        id="parent-edit-email"
        v-model="form.email"
        type="email"
        data-testid="parent-edit-email"
      >
    </div>
    <div class="form-field">
      <label for="parent-edit-phone">{{ t('students.parentPhone') }}</label>
      <input
        id="parent-edit-phone"
        v-model="form.phone"
        type="tel"
        data-testid="parent-edit-phone"
      >
    </div>
    <div class="form-field">
      <label for="parent-edit-relationship">{{ t('students.parentRelationship') }}</label>
      <input
        id="parent-edit-relationship"
        v-model="form.relationship"
        type="text"
        data-testid="parent-edit-relationship"
      >
    </div>
    <div class="form-field">
      <label for="parent-edit-primary">
        <input
          id="parent-edit-primary"
          v-model="form.isPrimary"
          type="checkbox"
          data-testid="parent-edit-primary"
        >
        {{ t('students.parentIsPrimary') }}
      </label>
    </div>
    <p
      v-if="noChanges"
      class="form-hint"
      role="alert"
      data-testid="parent-edit-no-changes"
    >
      {{ t('common.noChangesToSave') }}
    </p>
    <p
      v-if="error"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="parent-edit-error"
    >
      {{ t(error) }}
    </p>
    <div class="form-actions">
      <button
        type="button"
        :disabled="saving"
        data-testid="parent-edit-cancel"
        @click="emit('cancel')"
      >
        {{ t('common.cancel') }}
      </button>
      <button
        type="submit"
        :disabled="saving || !form.name"
        data-testid="parent-edit-submit"
      >
        {{ saving ? t('common.saving') : t('common.save') }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.form-field {
  margin-bottom: 0.5rem;
}

.form-field label {
  display: block;
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.form-field input {
  width: 100%;
  padding: 0.25rem 0.5rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.form-error {
  color: #dc3545;
  margin: 0.5rem 0;
}

.form-hint {
  color: #6c757d;
  margin: 0.5rem 0;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

button {
  padding: 0.25rem 0.75rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}
</style>
