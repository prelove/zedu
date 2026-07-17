<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import { ApiError, NetworkError } from '../../api/http'
import { errorToI18nKey } from '../../api/error-mapping'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const username = ref('')
const password = ref('')
const submitting = ref(false)
const errorKey = ref<string | null>(null)

const canSubmit = computed(() => !submitting.value && username.value.trim() !== '' && password.value !== '')

async function handleSubmit(): Promise<void> {
  if (!canSubmit.value) {
    return
  }
  submitting.value = true
  errorKey.value = null
  try {
    await authStore.login(username.value.trim(), password.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await router.push(redirect)
  } catch (err) {
    // Map to a stable i18n key; never expose raw error text, requestId, or stack.
    if (err instanceof NetworkError) {
      errorKey.value = 'errors.NETWORK_ERROR'
    } else if (err instanceof ApiError) {
      const key = errorToI18nKey(err)
      // For login, AUTH_REQUIRED is not a valid outcome; map to LOGIN_FAILED.
      errorKey.value = key === 'apiErrors.AUTH_REQUIRED' ? 'apiErrors.LOGIN_FAILED' : key
    } else {
      errorKey.value = 'errors.UNKNOWN'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div
    class="login-view"
    data-testid="login-view"
  >
    <h1>{{ t('auth.loginTitle') }}</h1>
    <form
      class="login-form"
      data-testid="login-form"
      @submit.prevent="handleSubmit"
    >
      <div class="form-field">
        <label for="login-username">{{ t('auth.usernameLabel') }}</label>
        <input
          id="login-username"
          v-model="username"
          type="text"
          autocomplete="username"
          :placeholder="t('auth.usernamePlaceholder')"
          :disabled="submitting"
          required
          data-testid="login-username"
        >
      </div>
      <div class="form-field">
        <label for="login-password">{{ t('auth.passwordLabel') }}</label>
        <input
          id="login-password"
          v-model="password"
          type="password"
          autocomplete="current-password"
          :placeholder="t('auth.passwordPlaceholder')"
          :disabled="submitting"
          required
          data-testid="login-password"
        >
      </div>
      <p
        v-if="errorKey"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="login-error"
      >
        {{ t(errorKey) }}
      </p>
      <button
        type="submit"
        :disabled="!canSubmit"
        data-testid="login-submit"
      >
        {{ submitting ? t('auth.submitting') : t('auth.submit') }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.login-view {
  max-width: 400px;
  margin: 2rem auto;
  padding: 1rem;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-field label {
  font-weight: 600;
}

.form-field input {
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  font-size: 1rem;
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
  margin: 0;
}

button[type="submit"] {
  padding: 0.6rem 1rem;
  border: 1px solid transparent;
  border-radius: 0.25rem;
  background-color: #0d6efd;
  color: #fff;
  font-size: 1rem;
  cursor: pointer;
}

button[type="submit"]:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}
</style>
