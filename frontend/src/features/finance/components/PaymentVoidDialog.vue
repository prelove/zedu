<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  open: boolean
  pending: boolean
  errorKey: string | null
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'confirm', reason: string): void
}>()

const { t } = useI18n()
const reason = ref('')
const localError = ref<string | null>(null)
const visibleError = computed(() => props.errorKey ?? localError.value)

watch(
  () => props.open,
  (open) => {
    if (open) {
      reason.value = ''
      localError.value = null
    }
  },
)

function confirm(): void {
  if (reason.value.trim() === '') {
    localError.value = 'apiErrors.INVALID_STATE'
    return
  }
  localError.value = null
  emit('confirm', reason.value.trim())
}
</script>

<template>
  <div
    v-if="open"
    class="void-overlay"
    role="dialog"
    aria-modal="true"
    :aria-label="t('financePayments.voidTitle')"
    data-testid="payment-void-dialog"
  >
    <div class="void-content">
      <h2>{{ t('financePayments.voidTitle') }}</h2>
      <p>{{ t('financePayments.voidDescription') }}</p>

      <label for="payment-void-reason">{{ t('financePayments.voidReason') }}</label>
      <textarea
        id="payment-void-reason"
        v-model="reason"
        data-testid="payment-void-reason"
        :disabled="pending"
      />

      <p
        v-if="visibleError"
        class="status-error"
        role="alert"
        aria-live="assertive"
        data-testid="payment-void-error"
      >
        {{ t(visibleError) }}
      </p>

      <div class="void-actions">
        <button
          type="button"
          data-testid="payment-void-cancel"
          :disabled="pending"
          @click="emit('cancel')"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          class="danger"
          data-testid="payment-void-confirm"
          :disabled="pending"
          @click="confirm"
        >
          {{ pending ? t('common.saving') : t('financePayments.voidConfirm') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.void-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.void-content {
  width: min(28rem, calc(100vw - 2rem));
  background: #fff;
  padding: 1.25rem;
  border-radius: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.void-content h2,
.void-content p {
  margin: 0;
}

textarea {
  min-height: 6rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
}

.void-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

button {
  padding: 0.45rem 0.8rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.375rem;
  background: #fff;
  cursor: pointer;
}

.danger {
  background: #dc2626;
  border-color: #dc2626;
  color: #fff;
}

.status-error {
  color: #dc2626;
}
</style>
