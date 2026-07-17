<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import { initialize, reset, type OnboardingTemplate } from '../../api/onboarding'
import { ApiError, NetworkError } from '../../api/http'
import { errorToI18nKey } from '../../api/error-mapping'

const { t } = useI18n()

type TemplateOption = {
  value: OnboardingTemplate
  labelKey: string
}

const templates: TemplateOption[] = [
  { value: 'japanese', labelKey: 'onboarding.templateJapanese' },
  { value: 'k12', labelKey: 'onboarding.templateK12' },
  { value: 'blank', labelKey: 'onboarding.templateBlank' },
]

const selectedTemplate = ref<OnboardingTemplate | null>(null)
const initializing = ref(false)
const resetting = ref(false)
const resultMessage = ref<string | null>(null)
const errorKey = ref<string | null>(null)

// Reset confirmation dialog state.
const showResetConfirm = ref(false)

const canInitialize = computed(() => !initializing.value && !resetting.value && selectedTemplate.value !== null)

function templateLabel(value: OnboardingTemplate): string {
  const opt = templates.find((o) => o.value === value)
  return opt ? t(opt.labelKey) : value
}

function clearErrors(): void {
  errorKey.value = null
  resultMessage.value = null
}

async function handleInitialize(): Promise<void> {
  if (!selectedTemplate.value || !canInitialize.value) {
    return
  }
  initializing.value = true
  clearErrors()
  try {
    const result = await authStore.authedRequest((token) => initialize(selectedTemplate.value!, token))
    if (result.reused) {
      resultMessage.value = t('onboarding.resultReused', { template: templateLabel(result.template as OnboardingTemplate) })
    } else {
      resultMessage.value = t('onboarding.resultInitialized', { template: templateLabel(result.template as OnboardingTemplate) })
    }
  } catch (err) {
    errorKey.value = mapError(err)
  } finally {
    initializing.value = false
  }
}

function openResetConfirm(): void {
  if (!selectedTemplate.value || initializing.value || resetting.value) {
    return
  }
  showResetConfirm.value = true
  clearErrors()
}

function cancelReset(): void {
  showResetConfirm.value = false
}

async function handleReset(): Promise<void> {
  if (!selectedTemplate.value) {
    return
  }
  showResetConfirm.value = false
  resetting.value = true
  clearErrors()
  try {
    const result = await authStore.authedRequest((token) => reset(selectedTemplate.value!, token))
    resultMessage.value = t('onboarding.resultReset', { template: templateLabel(result.template as OnboardingTemplate) })
  } catch (err) {
    errorKey.value = mapError(err)
  } finally {
    resetting.value = false
  }
}

function mapError(err: unknown): string {
  if (err instanceof NetworkError) {
    return 'errors.NETWORK_ERROR'
  }
  if (err instanceof ApiError) {
    return errorToI18nKey(err) ?? 'errors.UNKNOWN'
  }
  return 'errors.UNKNOWN'
}
</script>

<template>
  <div
    class="onboarding-view"
    data-testid="onboarding-view"
  >
    <h1>{{ t('onboarding.title') }}</h1>
    <p class="description">
      {{ t('onboarding.description') }}
    </p>

    <div
      v-if="!authStore.isOwner.value"
      class="owner-only-notice"
      role="alert"
      data-testid="onboarding-owner-only"
    >
      {{ t('onboarding.ownerOnly') }}
    </div>

    <template v-else>
      <fieldset class="template-fieldset">
        <legend>{{ t('onboarding.templateLabel') }}</legend>
        <div
          v-for="opt in templates"
          :key="opt.value"
          class="template-option"
        >
          <input
            :id="`template-${opt.value}`"
            v-model="selectedTemplate"
            type="radio"
            :value="opt.value"
            :disabled="initializing || resetting"
            name="template"
            data-testid="onboarding-template-input"
          >
          <label :for="`template-${opt.value}`">{{ t(opt.labelKey) }}</label>
        </div>
      </fieldset>

      <p
        v-if="errorKey"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="onboarding-error"
      >
        {{ t(errorKey) }}
      </p>

      <p
        v-if="resultMessage"
        class="form-success"
        role="status"
        aria-live="polite"
        data-testid="onboarding-result"
      >
        {{ resultMessage }}
      </p>

      <div class="actions">
        <button
          type="button"
          :disabled="!canInitialize"
          data-testid="onboarding-initialize"
          @click="handleInitialize"
        >
          {{ initializing ? t('onboarding.initializing') : t('onboarding.initialize') }}
        </button>
        <button
          type="button"
          :disabled="!canInitialize"
          data-testid="onboarding-reset"
          @click="openResetConfirm"
        >
          {{ resetting ? t('onboarding.resetting') : t('onboarding.reset') }}
        </button>
      </div>
    </template>

    <!-- Reset confirmation dialog -->
    <div
      v-if="showResetConfirm"
      class="reset-confirm"
      role="dialog"
      aria-modal="true"
      aria-labelledby="reset-confirm-title"
      data-testid="onboarding-reset-confirm"
    >
      <div class="reset-confirm-content">
        <h2 id="reset-confirm-title">
          {{ t('onboarding.resetConfirmTitle') }}
        </h2>
        <p>{{ t('onboarding.resetConfirmMessage') }}</p>
        <div class="reset-confirm-actions">
          <button
            type="button"
            data-testid="onboarding-reset-confirm-cancel"
            @click="cancelReset"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="button"
            class="danger"
            data-testid="onboarding-reset-confirm-ok"
            @click="handleReset"
          >
            {{ t('common.confirm') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.onboarding-view {
  max-width: 600px;
  margin: 0 auto;
  padding: 1rem;
}

.description {
  color: #6c757d;
  margin-bottom: 1.5rem;
}

.owner-only-notice {
  color: #dc3545;
  font-weight: 600;
  padding: 1rem;
  border: 1px solid #dc3545;
  border-radius: 0.5rem;
}

.template-fieldset {
  border: 1px solid #ccc;
  border-radius: 0.5rem;
  padding: 1rem;
  margin-bottom: 1.5rem;
}

.template-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.5rem 0;
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
}

.form-success {
  color: #28a745;
  font-size: 0.875rem;
}

.actions {
  display: flex;
  gap: 1rem;
}

.actions button {
  padding: 0.6rem 1rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}

.actions button:disabled {
  color: #6c757d;
  cursor: not-allowed;
}

.actions button.danger {
  background-color: #dc3545;
  color: #fff;
  border-color: #dc3545;
}

.reset-confirm {
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

.reset-confirm-content {
  background-color: #fff;
  padding: 1.5rem;
  border-radius: 0.5rem;
  max-width: 400px;
}

.reset-confirm-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1rem;
}

.reset-confirm-actions button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  cursor: pointer;
}
</style>
