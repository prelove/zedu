<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { StudentWrite } from '../../../api/directory'

const props = defineProps<{ modelValue: StudentWrite }>()
const emit = defineEmits<{ 'update:modelValue': [value: StudentWrite] }>()
const { t } = useI18n()

function updateField<K extends keyof StudentWrite>(key: K, value: StudentWrite[K]): void {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}
</script>

<template>
  <div class="form-grid">
    <div class="form-field">
      <label for="edit-name">{{ t('students.name') }} *</label>
      <input
        id="edit-name"
        :value="modelValue.name"
        type="text"
        required
        data-testid="edit-student-name"
        @input="updateField('name', ($event.target as HTMLInputElement).value)"
      >
    </div>
    <div class="form-field">
      <label for="edit-name-local">{{ t('students.nameLocal') }}</label>
      <input
        id="edit-name-local"
        :value="modelValue.nameLocal"
        type="text"
        data-testid="edit-student-name-local"
        @input="updateField('nameLocal', ($event.target as HTMLInputElement).value)"
      >
    </div>
    <div class="form-field">
      <label for="edit-email">{{ t('students.emailOptional') }}</label>
      <input
        id="edit-email"
        :value="modelValue.email"
        type="email"
        data-testid="edit-student-email"
        @input="updateField('email', ($event.target as HTMLInputElement).value)"
      >
    </div>
    <div class="form-field">
      <label for="edit-phone">{{ t('students.phone') }}</label>
      <input
        id="edit-phone"
        :value="modelValue.phone"
        type="tel"
        data-testid="edit-student-phone"
        @input="updateField('phone', ($event.target as HTMLInputElement).value)"
      >
    </div>
    <div class="form-field">
      <label for="edit-nationality">{{ t('students.nationality') }}</label>
      <input
        id="edit-nationality"
        :value="modelValue.nationality"
        type="text"
        data-testid="edit-student-nationality"
        @input="updateField('nationality', ($event.target as HTMLInputElement).value)"
      >
    </div>
    <div class="form-field">
      <label for="edit-timezone">{{ t('students.timezone') }}</label>
      <input
        id="edit-timezone"
        :value="modelValue.timezone"
        type="text"
        data-testid="edit-student-timezone"
        @input="updateField('timezone', ($event.target as HTMLInputElement).value)"
      >
    </div>
    <div class="form-field">
      <label for="edit-status">{{ t('students.status') }}</label>
      <select
        id="edit-status"
        :value="modelValue.status"
        data-testid="edit-student-status"
        @change="updateField('status', ($event.target as HTMLSelectElement).value)"
      >
        <option value="ACTIVE">
          {{ t('students.statusActive') }}
        </option>
        <option value="PAUSED">
          {{ t('students.statusPaused') }}
        </option>
        <option value="ENDED">
          {{ t('students.statusEnded') }}
        </option>
        <option value="CANCELLED">
          {{ t('students.statusCancelled') }}
        </option>
      </select>
    </div>
    <div class="form-field">
      <label for="edit-source-channel">{{ t('students.sourceChannel') }}</label>
      <input
        id="edit-source-channel"
        :value="modelValue.sourceChannel"
        type="text"
        data-testid="edit-student-source-channel"
        @input="updateField('sourceChannel', ($event.target as HTMLInputElement).value)"
      >
    </div>
    <div class="form-field">
      <label for="edit-note">{{ t('students.note') }}</label>
      <textarea
        id="edit-note"
        :value="modelValue.note"
        data-testid="edit-student-note"
        @input="updateField('note', ($event.target as HTMLTextAreaElement).value)"
      />
    </div>
  </div>
</template>

<style scoped>
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-field label {
  font-weight: 600;
}

.form-field input,
.form-field select,
.form-field textarea {
  padding: 0.4rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}
</style>
