<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Availability, AvailabilityWrite } from '../../../api/directory'

const props = defineProps<{
  availability: Availability
  weekdayOptions: number[]
  weekdayKeys: Record<number, string>
  saving: boolean
  clientError: string | null
  error: string | null
}>()

const emit = defineEmits<{
  submit: [body: AvailabilityWrite]
  cancel: []
}>()

const { t } = useI18n()
const form = ref<AvailabilityWrite>({})
const noChanges = ref(false)

function syncForm(): void {
  form.value = {
    weekday: props.availability.weekday,
    startTime: props.availability.startTime,
    endTime: props.availability.endTime,
  }
  noChanges.value = false
}

function handleSubmit(): void {
  const body: AvailabilityWrite = {
    weekday: form.value.weekday,
    startTime: form.value.startTime,
    endTime: form.value.endTime,
  }
  if (Object.keys(body).length === 0) {
    noChanges.value = true
    return
  }
  if (
    body.weekday === props.availability.weekday &&
    body.startTime === props.availability.startTime &&
    body.endTime === props.availability.endTime
  ) {
    noChanges.value = true
    return
  }
  emit('submit', body)
}

watch(() => props.availability, syncForm, { immediate: true })
</script>

<template>
  <form
    class="avail-form"
    data-testid="availability-edit-form"
    @submit.prevent="handleSubmit"
  >
    <div class="form-field">
      <label for="avail-edit-weekday">{{ t('teachers.availabilityWeekday') }}</label>
      <select
        id="avail-edit-weekday"
        v-model.number="form.weekday"
        data-testid="avail-edit-weekday"
      >
        <option
          v-for="weekday in weekdayOptions"
          :key="weekday"
          :value="weekday"
        >
          {{ t(weekdayKeys[weekday]) }}
        </option>
      </select>
    </div>
    <div class="form-field">
      <label for="avail-edit-start-time">{{ t('teachers.availabilityStartTime') }}</label>
      <input
        id="avail-edit-start-time"
        v-model="form.startTime"
        type="time"
        data-testid="avail-edit-start-time"
      >
    </div>
    <div class="form-field">
      <label for="avail-edit-end-time">{{ t('teachers.availabilityEndTime') }}</label>
      <input
        id="avail-edit-end-time"
        v-model="form.endTime"
        type="time"
        data-testid="avail-edit-end-time"
      >
    </div>
    <p
      v-if="clientError"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="availability-client-error"
    >
      {{ t(clientError) }}
    </p>
    <p
      v-if="noChanges"
      class="form-info"
      role="alert"
      aria-live="assertive"
      data-testid="avail-edit-no-changes"
    >
      {{ t('common.noChangesToSave') }}
    </p>
    <p
      v-if="error"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="avail-edit-error"
    >
      {{ t(error) }}
    </p>
    <div class="form-actions">
      <button
        type="button"
        :disabled="saving"
        data-testid="avail-edit-cancel"
        @click="emit('cancel')"
      >
        {{ t('common.cancel') }}
      </button>
      <button
        type="submit"
        :disabled="saving"
        data-testid="avail-edit-submit"
      >
        {{ saving ? t('common.saving') : t('common.save') }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.avail-form {
  margin-top: 1rem;
  padding: 1rem;
  border: 1px solid #dee2e6;
  border-radius: 0.375rem;
  max-width: 28rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  margin-bottom: 0.75rem;
}

.form-field label {
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
  font-weight: 600;
}

.form-field input,
.form-field select {
  padding: 0.35rem 0.5rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.form-error {
  color: #dc3545;
  margin: 0 0 0.5rem;
  font-size: 0.875rem;
}

.form-info {
  color: #856404;
  margin: 0 0 0.5rem;
  font-size: 0.875rem;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
}

button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background: #fff;
  cursor: pointer;
}
</style>
