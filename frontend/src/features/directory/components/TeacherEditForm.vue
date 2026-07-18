<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { updateTeacher, type Teacher, type TeacherWrite } from '../../../api/directory'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import { formatJPY } from '../../../utils/formatters'
import type { Locale } from '../../../i18n/config'

const props = defineProps<{ teacher: Teacher }>()
const emit = defineEmits<{ saved: [teacher: Teacher] }>()
const { t, locale } = useI18n()

const editForm = ref<TeacherWrite>({})
const saving = ref(false)
const saveError = ref<string | null>(null)
const noChanges = ref(false)

watch(() => props.teacher, (t) => {
  if (!t) return
  editForm.value = {
    name: t.name, nameLocal: t.nameLocal ?? '', email: t.email ?? '',
    phone: t.phone ?? '', bio: t.bio ?? '', defaultRate: t.defaultRate,
    status: t.status, note: t.note ?? '',
  }
  noChanges.value = false
}, { immediate: true })

async function handleSave(): Promise<void> {
  saving.value = true
  saveError.value = null
  noChanges.value = false
  const body: TeacherWrite = {}
  const s = props.teacher
  if (editForm.value.name !== s.name) body.name = editForm.value.name
  if ((editForm.value.nameLocal ?? '') !== (s.nameLocal ?? '')) body.nameLocal = editForm.value.nameLocal
  if ((editForm.value.email ?? '') !== (s.email ?? '')) body.email = editForm.value.email ?? ''
  if ((editForm.value.phone ?? '') !== (s.phone ?? '')) body.phone = editForm.value.phone
  if ((editForm.value.bio ?? '') !== (s.bio ?? '')) body.bio = editForm.value.bio
  if (editForm.value.defaultRate !== s.defaultRate) body.defaultRate = editForm.value.defaultRate
  if ((editForm.value.status ?? '') !== s.status) body.status = editForm.value.status
  if ((editForm.value.note ?? '') !== (s.note ?? '')) body.note = editForm.value.note

  if (Object.keys(body).length === 0) {
    noChanges.value = true
    saving.value = false
    return
  }

  try {
    const updated = await authStore.authedRequest((token) => updateTeacher(token, s.id, body))
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
    data-testid="teacher-edit-section"
  >
    <h2>{{ t('teachers.editTitle') }}</h2>
    <form @submit.prevent="handleSave">
      <div class="form-grid">
        <div class="form-field">
          <label for="t-edit-name">{{ t('teachers.name') }} *</label>
          <input
            id="t-edit-name"
            v-model="editForm.name"
            type="text"
            required
            data-testid="edit-teacher-name"
          >
        </div>
        <div class="form-field">
          <label for="t-edit-name-local">{{ t('teachers.nameLocal') }}</label>
          <input
            id="t-edit-name-local"
            v-model="editForm.nameLocal"
            type="text"
            data-testid="edit-teacher-name-local"
          >
        </div>
        <div class="form-field">
          <label for="t-edit-email">{{ t('teachers.email') }}</label>
          <input
            id="t-edit-email"
            v-model="editForm.email"
            type="email"
            data-testid="edit-teacher-email"
          >
        </div>
        <div class="form-field">
          <label for="t-edit-phone">{{ t('teachers.phone') }}</label>
          <input
            id="t-edit-phone"
            v-model="editForm.phone"
            type="tel"
            data-testid="edit-teacher-phone"
          >
        </div>
        <div class="form-field">
          <label for="t-edit-rate">{{ t('teachers.defaultRate') }}</label>
          <input
            id="t-edit-rate"
            v-model.number="editForm.defaultRate"
            type="number"
            min="0"
            step="1"
            data-testid="edit-teacher-default-rate"
          >
          <span class="rate-hint">{{ formatJPY(editForm.defaultRate ?? 0, locale as Locale) }}</span>
        </div>
        <div class="form-field">
          <label for="t-edit-status">{{ t('teachers.status') }}</label>
          <select
            id="t-edit-status"
            v-model="editForm.status"
            data-testid="edit-teacher-status"
          >
            <option value="ACTIVE">
              {{ t('teachers.statusActive') }}
            </option>
            <option value="INACTIVE">
              {{ t('teachers.statusInactive') }}
            </option>
          </select>
        </div>
        <div class="form-field">
          <label for="t-edit-bio">{{ t('teachers.bio') }}</label>
          <textarea
            id="t-edit-bio"
            v-model="editForm.bio"
            data-testid="edit-teacher-bio"
          />
        </div>
        <div class="form-field">
          <label for="t-edit-note">{{ t('teachers.note') }}</label>
          <textarea
            id="t-edit-note"
            v-model="editForm.note"
            data-testid="edit-teacher-note"
          />
        </div>
      </div>
      <p
        v-if="noChanges"
        class="form-hint"
        data-testid="teacher-no-changes"
      >
        {{ t('common.noChangesToSave') }}
      </p>
      <p
        v-if="saveError"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="teacher-save-error"
      >
        {{ t(saveError) }}
      </p>
      <button
        type="submit"
        :disabled="saving"
        data-testid="teacher-save-btn"
      >
        {{ saving ? t('common.saving') : t('teachers.edit') }}
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
.rate-hint { font-size: 0.8125rem; color: #495057; }
.form-error { color: #dc3545; font-size: 0.875rem; }
.form-hint { color: #856404; font-size: 0.8125rem; background-color: #fff3cd; padding: 0.25rem 0.5rem; border-radius: 0.25rem; }
button { padding: 0.4rem 0.8rem; border: 1px solid #ccc; border-radius: 0.25rem; background-color: #fff; cursor: pointer; }
button:disabled { color: #6c757d; cursor: not-allowed; }
</style>
