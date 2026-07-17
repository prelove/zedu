<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import LocaleSwitcher from '../../components/LocaleSwitcher.vue'

const router = useRouter()
const { t } = useI18n()

const logoutError = ref(false)
const loggingOut = ref(false)

async function handleLogout(): Promise<void> {
  loggingOut.value = true
  logoutError.value = false
  try {
    const cleared = await authStore.logout()
    if (cleared) {
      await router.push({ name: 'login' })
    } else {
      // Logout request failed (network/server); session may still be valid.
      logoutError.value = true
    }
  } finally {
    loggingOut.value = false
  }
}

function roleLabel(role: string | null): string {
  if (role === 'OWNER') {
    return t('auth.roleOwner')
  }
  if (role === 'OPERATOR') {
    return t('auth.roleOperator')
  }
  return role ?? ''
}
</script>

<template>
  <div
    class="home-view"
    data-testid="home-view"
  >
    <header class="home-header">
      <div class="user-info">
        <p data-testid="home-welcome">
          {{ t('auth.welcome') }}, <span data-testid="home-username">{{ authStore.currentUser.value?.displayName ?? authStore.currentUser.value?.username ?? '' }}</span>
        </p>
        <p data-testid="home-role">
          {{ t('auth.role') }}: <span>{{ roleLabel(authStore.state.role) }}</span>
        </p>
      </div>
      <div class="home-controls">
        <LocaleSwitcher />
        <button
          type="button"
          :disabled="loggingOut"
          data-testid="home-logout"
          @click="handleLogout"
        >
          {{ loggingOut ? t('common.loading') : t('auth.logout') }}
        </button>
      </div>
    </header>
    <p
      v-if="logoutError"
      class="logout-error"
      role="alert"
      aria-live="assertive"
      data-testid="home-logout-error"
    >
      {{ t('auth.logoutFailed') }}
    </p>
  </div>
</template>

<style scoped>
.home-view {
  max-width: 800px;
  margin: 0 auto;
  padding: 1rem;
}

.home-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
}

.user-info p {
  margin: 0.25rem 0;
}

.home-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logout-error {
  color: #dc3545;
  font-size: 0.875rem;
  margin-top: 1rem;
}

button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}

button:disabled {
  color: #6c757d;
  cursor: not-allowed;
}
</style>
