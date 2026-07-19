<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const logoutError = ref(false)
const loggingOut = ref(false)
const ownerAccessDenied = computed(() => route.query.denied === 'owner')

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
        <nav
          class="home-nav"
          aria-label="main navigation"
          data-testid="home-nav"
        >
          <RouterLink
            to="/dashboard"
            data-testid="nav-dashboard"
          >
            {{ t('nav.dashboard') }}
          </RouterLink>
          <RouterLink
            to="/students"
            data-testid="nav-students"
          >
            {{ t('nav.students') }}
          </RouterLink>
          <RouterLink
            to="/teachers"
            data-testid="nav-teachers"
          >
            {{ t('nav.teachers') }}
          </RouterLink>
          <RouterLink
            to="/courses"
            data-testid="nav-courses"
          >
            {{ t('nav.courses') }}
          </RouterLink>
          <RouterLink
            to="/finance/payments"
            data-testid="nav-finance-payments"
          >
            {{ t('nav.financePayments') }}
          </RouterLink>
          <RouterLink
            to="/lessons"
            data-testid="nav-lessons"
          >
            {{ t('nav.lessons') }}
          </RouterLink>
          <RouterLink to="/notifications">
            {{ t('nav.notifications') }}
          </RouterLink>
          <RouterLink
            v-if="authStore.isOwner.value"
            to="/finance/config"
            data-testid="nav-finance-config"
          >
            {{ t('nav.financeConfig') }}
          </RouterLink>
          <RouterLink
            v-if="authStore.isOwner.value"
            to="/onboarding"
            data-testid="nav-onboarding"
          >
            {{ t('nav.onboarding') }}
          </RouterLink>
        </nav>
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
      v-if="ownerAccessDenied"
      class="logout-error"
      role="alert"
      aria-live="assertive"
      data-testid="home-owner-required"
    >
      {{ t('routeGuard.ownerRequired') }}
    </p>
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
  flex-wrap: wrap;
}

.home-nav {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.home-nav a {
  padding: 0.3rem 0.6rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  text-decoration: none;
  color: #0d6efd;
}

.home-nav a.router-link-active {
  background-color: #0d6efd;
  color: #fff;
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
