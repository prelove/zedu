<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { type Teacher } from '../../../api/directory'
import type { AssignmentWrite } from '../../../api/course'

const props = defineProps<{
  teachers: Teacher[]
  saving: boolean
  formError: string | null
}>()

const emit = defineEmits<{
  submit: [body: AssignmentWrite]
  cancel: []
}>()

const { t } = useI18n()
const form = ref<AssignmentWrite>({ teacherId: 0, roleType: 'MAIN' })

const activeTeachers = computed(() => props.teachers.filter((teacher) => teacher.status === 'ACTIVE'))

function resetForm(): void {
  form.value = { teacherId: 0, roleType: 'MAIN' }
}

function handleSubmit(): void {
  emit('submit', { ...form.value })
}

watch(
  () => props.teachers,
  () => resetForm(),
  { immediate: true },
)
</script>

<template>
  <div
    class="create-form"
    data-testid="assignment-create-form"
  >
    <form @submit.prevent="handleSubmit">
      <div class="form-field">
        <label for="assign-form-teacher">{{ t('enrollments.assignmentTeacher') }} *</label>
        <select
          id="assign-form-teacher"
          v-model.number="form.teacherId"
          required
          data-testid="assign-form-teacher"
        >
          <option :value="0">
            {{ t('common.none') }}
          </option>
          <option
            v-for="teacher in activeTeachers"
            :key="teacher.id"
            :value="teacher.id"
          >
            {{ teacher.name }}
          </option>
        </select>
      </div>
      <div class="form-field">
        <label for="assign-form-role">{{ t('enrollments.assignmentRoleType') }}</label>
        <select
          id="assign-form-role"
          v-model="form.roleType"
          data-testid="assign-form-role"
        >
          <option value="MAIN">
            {{ t('enrollments.roleTypeMain') }}
          </option>
          <option value="ASSISTANT">
            {{ t('enrollments.roleTypeAssistant') }}
          </option>
          <option value="OBSERVER">
            {{ t('enrollments.roleTypeObserver') }}
          </option>
        </select>
      </div>
      <p
        v-if="formError"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="assignment-create-error"
      >
        {{ t(formError) }}
      </p>
      <div class="form-actions">
        <button
          type="button"
          :disabled="saving"
          @click="emit('cancel')"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          :disabled="saving || !form.teacherId"
          data-testid="assignment-create-submit"
        >
          {{ saving ? t('common.creating') : t('common.create') }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.create-form {
  border: 1px solid #ccc;
  border-radius: 0.5rem;
  padding: 1rem;
  margin-top: 0.5rem;
  background-color: #f8f9fa;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.75rem;
}

.form-field label {
  font-weight: 600;
}

.form-field select {
  padding: 0.4rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
}

button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}
</style>
