<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{ errorKey: string }>()
const emit = defineEmits<{ retry: [] }>()

const { t } = useI18n()
</script>

<template>
  <div class="teacher-load-notice">
    <div
      class="teacher-load-error"
      role="alert"
      aria-live="assertive"
      data-testid="teacher-load-error"
    >
      <p>{{ t(errorKey) }}</p>
      <button
        type="button"
        data-testid="teacher-load-retry"
        @click="emit('retry')"
      >
        {{ t('common.retry') }}
      </button>
    </div>
    <p
      class="hint"
      data-testid="teacher-load-hint"
    >
      {{ t('enrollments.teacherLoadRequired') }}
    </p>
  </div>
</template>

<style scoped>
.teacher-load-notice {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.teacher-load-error,
.hint {
  margin: 0;
}

.teacher-load-error {
  color: #dc3545;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.teacher-load-error p {
  margin: 0;
}

.hint {
  color: #6c757d;
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
