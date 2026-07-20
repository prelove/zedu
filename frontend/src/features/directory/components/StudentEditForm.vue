<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { updateStudent, type Student, type StudentWrite } from '../../../api/directory'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import StudentEditFields from './StudentEditFields.vue'

const props = defineProps<{ student: Student }>()
const emit = defineEmits<{ saved: [student: Student] }>()
const { t } = useI18n()

const editForm = ref<StudentWrite>({})
const saving = ref(false)
const saveError = ref<string | null>(null)
const noChanges = ref(false)

function buildBody(student: Student, form: StudentWrite): StudentWrite {
  const body: StudentWrite = {}
  if (form.name !== student.name) body.name = form.name
  if ((form.email ?? '') !== (student.email ?? '')) body.email = form.email ?? ''
  if ((form.nameLocal ?? '') !== (student.nameLocal ?? '')) body.nameLocal = form.nameLocal
  if ((form.phone ?? '') !== (student.phone ?? '')) body.phone = form.phone
  if ((form.nationality ?? '') !== (student.nationality ?? '')) body.nationality = form.nationality
  if ((form.timezone ?? '') !== student.timezone) body.timezone = form.timezone
  if ((form.status ?? '') !== student.status) body.status = form.status
  if ((form.sourceChannel ?? '') !== (student.sourceChannel ?? '')) body.sourceChannel = form.sourceChannel
  if ((form.note ?? '') !== (student.note ?? '')) body.note = form.note
  return body
}

async function handleSave(): Promise<void> {
  saving.value = true
  saveError.value = null
  noChanges.value = false
  const body = buildBody(props.student, editForm.value)
  if (Object.keys(body).length === 0) {
    noChanges.value = true
    saving.value = false
    return
  }

  try {
    emit('saved', await authStore.authedRequest((token) => updateStudent(token, props.student.id, body)))
  } catch (err) {
    if (err instanceof NetworkError) saveError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) saveError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else saveError.value = 'errors.UNKNOWN'
  } finally {
    saving.value = false
  }
}

watch(() => props.student, (student) => {
  editForm.value = {
    name: student.name,
    nameLocal: student.nameLocal ?? '',
    email: student.email ?? '',
    phone: student.phone ?? '',
    nationality: student.nationality ?? '',
    timezone: student.timezone,
    status: student.status,
    sourceChannel: student.sourceChannel ?? '',
    note: student.note ?? '',
  }
  noChanges.value = false
}, { immediate: true })
</script>

<template>
  <section
    class="edit-section"
    data-testid="student-edit-section"
  >
    <h2>{{ t('students.editTitle') }}</h2>
    <form @submit.prevent="handleSave">
      <StudentEditFields v-model="editForm" />
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
.edit-section {
  margin-top: 1.5rem;
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
}

.form-hint {
  color: #856404;
  font-size: 0.8125rem;
  background-color: #fff3cd;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
}

button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
  margin-top: 0.5rem;
}

button:disabled {
  color: #6c757d;
  cursor: not-allowed;
}
</style>
