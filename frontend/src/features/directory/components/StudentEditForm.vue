<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { updateStudent, type Student, type StudentWrite } from '../../../api/directory'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'

const props = defineProps<{ student: Student }>()
const emit = defineEmits<{ saved: [student: Student] }>()
const { t } = useI18n()

const editForm = ref<StudentWrite>({})
const saving = ref(false)
const saveError = ref<string | null>(null)
const noChanges = ref(false)

// Initialize form from student data.
watch(() => props.student, (s) => {
  if (!s) return
  editForm.value = {
    name: s.name, nameLocal: s.nameLocal ?? '', email: s.email ?? '',
    phone: s.phone ?? '', nationality: s.nationality ?? '',
    timezone: s.timezone, status: s.status,
    sourceChannel: s.sourceChannel ?? '', note: s.note ?? '',
  }
  noChanges.value = false
}, { immediate: true })

async function handleSave(): Promise<void> {
  saving.value = true
  saveError.value = null
  noChanges.value = false
  // Build body with only changed fields.
  const body: StudentWrite = {}
  const s = props.student
  if (editForm.value.name !== s.name) body.name = editForm.value.name
  if ((editForm.value.email ?? '') !== (s.email ?? '')) body.email = editForm.value.email ?? ''
  if ((editForm.value.nameLocal ?? '') !== (s.nameLocal ?? '')) body.nameLocal = editForm.value.nameLocal
  if ((editForm.value.phone ?? '') !== (s.phone ?? '')) body.phone = editForm.value.phone
  if ((editForm.value.nationality ?? '') !== (s.nationality ?? '')) body.nationality = editForm.value.nationality
  if ((editForm.value.timezone ?? '') !== s.timezone) body.timezone = editForm.value.timezone
  if ((editForm.value.status ?? '') !== s.status) body.status = editForm.value.status
  if ((editForm.value.sourceChannel ?? '') !== (s.sourceChannel ?? '')) body.sourceChannel = editForm.value.sourceChannel
  if ((editForm.value.note ?? '') !== (s.note ?? '')) body.note = editForm.value.note

  if (Object.keys(body).length === 0) {
    noChanges.value = true
    saving.value = false
    return
  }

  try {
    const updated = await authStore.authedRequest((token) => updateStudent(token, s.id, body))
    emit('saved', updated)
  } catch (err) {
    if (err instanceof NetworkError) saveError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) saveError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else saveError.value = 'errors.UNKNOWN'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <section
    class="edit-section"
    data-testid="student-edit-section"
  >
    <h2>{{ t('students.editTitle') }}</h2>
    <form @submit.prevent="handleSave">
      <div class="form-grid">
        <div class="form-field">
          <label for="edit-name">{{ t('students.name') }} *</label>
          <input
            id="edit-name"
            v-model="editForm.name"
            type="text"
            required
            data-testid="edit-student-name"
          >
        </div>
        <div class="form-field">
          <label for="edit-name-local">{{ t('students.nameLocal') }}</label>
          <input
            id="edit-name-local"
            v-model="editForm.nameLocal"
            type="text"
            data-testid="edit-student-name-local"
          >
        </div>
        <div class="form-field">
          <label for="edit-email">{{ t('students.emailOptional') }}</label>
          <input
            id="edit-email"
            v-model="editForm.email"
            type="email"
            data-testid="edit-student-email"
          >
        </div>
        <div class="form-field">
          <label for="edit-phone">{{ t('students.phone') }}</label>
          <input
            id="edit-phone"
            v-model="editForm.phone"
            type="tel"
            data-testid="edit-student-phone"
          >
        </div>
        <div class="form-field">
          <label for="edit-nationality">{{ t('students.nationality') }}</label>
          <input
            id="edit-nationality"
            v-model="editForm.nationality"
            type="text"
            data-testid="edit-student-nationality"
          >
        </div>
        <div class="form-field">
          <label for="edit-timezone">{{ t('students.timezone') }}</label>
          <input
            id="edit-timezone"
            v-model="editForm.timezone"
            type="text"
            data-testid="edit-student-timezone"
          >
        </div>
        <div class="form-field">
          <label for="edit-status">{{ t('students.status') }}</label>
          <select
            id="edit-status"
            v-model="editForm.status"
            data-testid="edit-student-status"
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
            v-model="editForm.sourceChannel"
            type="text"
            data-testid="edit-student-source-channel"
          >
        </div>
        <div class="form-field">
          <label for="edit-note">{{ t('students.note') }}</label>
          <textarea
            id="edit-note"
            v-model="editForm.note"
            data-testid="edit-student-note"
          />
        </div>
      </div>
      <p
        v-if="noChanges"
        class="form-hint"
        data-testid="student-no-changes"
      >
        {{ t('common.noChangesToSave') }}
      </p>
      <p
        v-if="saveError"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="student-save-error"
      >
        {{ t(saveError) }}
      </p>
      <p
        v-if="saveError === 'apiErrors.CONFLICT'"
        class="form-hint"
        data-testid="student-email-no-bypass"
      >
        {{ t('students.noBypass') }}
      </p>
      <button
        type="submit"
        :disabled="saving"
        data-testid="student-save-btn"
      >
        {{ saving ? t('common.saving') : t('students.edit') }}
      </button>
    </form>
  </section>
</template>

<style scoped>
.edit-section { margin-top: 1.5rem; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.form-field { display: flex; flex-direction: column; gap: 0.25rem; }
.form-field label { font-weight: 600; }
.form-field input, .form-field select, .form-field textarea {
  padding: 0.4rem; border: 1px solid #ccc; border-radius: 0.25rem;
}
.form-error { color: #dc3545; font-size: 0.875rem; }
.form-hint {
  color: #856404; font-size: 0.8125rem; background-color: #fff3cd;
  padding: 0.25rem 0.5rem; border-radius: 0.25rem;
}
button {
  padding: 0.4rem 0.8rem; border: 1px solid #ccc; border-radius: 0.25rem;
  background-color: #fff; cursor: pointer; margin-top: 0.5rem;
}
button:disabled { color: #6c757d; cursor: not-allowed; }
</style>
