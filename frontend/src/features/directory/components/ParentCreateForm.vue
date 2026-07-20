<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ParentWrite } from '../../../api/directory'

defineProps<{
  creating: boolean
  createError: string | null
}>()

const emit = defineEmits<{
  submit: [body: ParentWrite]
  cancel: []
}>()

const { t } = useI18n()
const form = ref<ParentWrite>({ name: '', email: '', phone: '', relationship: '', isPrimary: false })

function sanitize(body: ParentWrite): ParentWrite {
  const nextBody: ParentWrite = { ...body }
  if (!nextBody.email) delete nextBody.email
  if (!nextBody.phone) delete nextBody.phone
  if (!nextBody.relationship) delete nextBody.relationship
  return nextBody
}
</script>

<template>
  <div
    class="create-dialog"
    data-testid="parent-create-form"
  >
    <form @submit.prevent="emit('submit', sanitize(form))">
      <div class="form-field">
        <label for="parent-form-name">{{ t('students.parentName') }} *</label>
        <input
          id="parent-form-name"
          v-model="form.name"
          type="text"
          required
          data-testid="parent-form-name"
        >
      </div>
      <div class="form-field">
        <label for="parent-form-email">{{ t('students.parentEmail') }}</label>
        <input
          id="parent-form-email"
          v-model="form.email"
          type="email"
          data-testid="parent-form-email"
        >
      </div>
      <div class="form-field">
        <label for="parent-form-phone">{{ t('students.parentPhone') }}</label>
        <input
          id="parent-form-phone"
          v-model="form.phone"
          type="tel"
          data-testid="parent-form-phone"
        >
      </div>
      <div class="form-field">
        <label for="parent-form-relationship">{{ t('students.parentRelationship') }}</label>
        <input
          id="parent-form-relationship"
          v-model="form.relationship"
          type="text"
          data-testid="parent-form-relationship"
        >
      </div>
      <div class="form-field">
        <label for="parent-form-primary">
          <input
            id="parent-form-primary"
            v-model="form.isPrimary"
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
          @click="emit('cancel')"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          :disabled="creating || !form.name"
          data-testid="parent-create-submit"
        >
          {{ creating ? t('common.creating') : t('common.create') }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.create-dialog {
  margin-top: 1rem;
  padding: 1rem;
  border: 1px solid #dee2e6;
  border-radius: 0.375rem;
  background-color: #f8f9fa;
}

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
