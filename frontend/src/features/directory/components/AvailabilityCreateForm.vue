<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AvailabilityWrite } from '../../../api/directory'

defineProps<{
  weekdayOptions: number[]
  weekdayKeys: Record<number, string>
  creating: boolean
  clientError: string | null
  createError: string | null
}>()

const emit = defineEmits<{
  submit: [body: AvailabilityWrite]
  cancel: []
}>()

const { t } = useI18n()
const form = ref<AvailabilityWrite>({ weekday: 1, startTime: '', endTime: '' })
</script>

<template>
  <div
    class="create-dialog"
    data-testid="availability-create-form"
  >
    <form @submit.prevent="emit('submit', { ...form })">
      <div class="form-field">
        <label for="avail-form-weekday">{{ t('teachers.availabilityWeekday') }}</label>
        <select
          id="avail-form-weekday"
          v-model.number="form.weekday"
          data-testid="avail-form-weekday"
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
        <label for="avail-form-start">{{ t('teachers.availabilityStartTime') }}</label>
        <input
          id="avail-form-start"
          v-model="form.startTime"
          type="time"
          data-testid="avail-form-start"
        >
      </div>
      <div class="form-field">
        <label for="avail-form-end">{{ t('teachers.availabilityEndTime') }}</label>
        <input
          id="avail-form-end"
          v-model="form.endTime"
          type="time"
          data-testid="avail-form-end"
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
        v-if="createError"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="availability-create-error"
      >
        {{ t(createError) }}
      </p>
      <div class="form-actions">
        <button
          type="button"
          :disabled="creating"
          @click="emit('cancel')"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          :disabled="creating"
          data-testid="availability-create-submit"
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
