<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { errorToI18nKey } from '../api/error-mapping'
import { ApiError, NetworkError } from '../api/http'

const props = defineProps<{
  error: unknown
}>()

const emit = defineEmits<{
  (e: 'retry'): void
}>()

const { t } = useI18n()

function errorKey(): string {
  if (props.error instanceof NetworkError) {
    return 'errors.NETWORK_ERROR'
  }
  if (props.error instanceof ApiError) {
    return errorToI18nKey(props.error) ?? 'errors.UNKNOWN'
  }
  return 'errors.UNKNOWN'
}
</script>

<template>
  <div
    class="state-error"
    data-testid="state-error"
    role="alert"
    aria-live="assertive"
  >
    <p class="error-message">
      {{ t(errorKey()) }}
    </p>
    <button
      type="button"
      data-testid="state-error-retry"
      @click="emit('retry')"
    >
      {{ t('common.retry') }}
    </button>
  </div>
</template>

<style scoped>
.state-error {
  color: #dc3545;
  padding: 1rem 0;
}

.error-message {
  margin: 0 0 0.5rem 0;
}

button {
  padding: 0.25rem 0.5rem;
  border: 1px solid #dc3545;
  border-radius: 0.25rem;
  background-color: #fff;
  color: #dc3545;
  cursor: pointer;
}
</style>
