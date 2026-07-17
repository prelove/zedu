<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  title: string
  message: string
}>()

const emit = defineEmits<{
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

const { t } = useI18n()
</script>

<template>
  <div
    class="confirm-overlay"
    role="dialog"
    aria-modal="true"
    :aria-label="props.title"
    data-testid="confirm-dialog"
  >
    <div class="confirm-content">
      <h2>{{ props.title }}</h2>
      <p>{{ props.message }}</p>
      <div class="confirm-actions">
        <button
          type="button"
          data-testid="confirm-cancel"
          @click="emit('cancel')"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          class="danger"
          data-testid="confirm-ok"
          @click="emit('confirm')"
        >
          {{ t('common.confirm') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.confirm-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.confirm-content {
  background-color: #fff;
  padding: 1.5rem;
  border-radius: 0.5rem;
  max-width: 400px;
}

.confirm-content h2 {
  margin-top: 0;
}

.confirm-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1rem;
}

button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  cursor: pointer;
}

button.danger {
  background-color: #dc3545;
  color: #fff;
  border-color: #dc3545;
}
</style>
