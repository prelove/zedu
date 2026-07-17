<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from './stores/auth'

const { t } = useI18n()

// On app initialization, if there is an in-memory token (e.g. from a
// same-session navigation), attempt to restore the user profile.
// This does NOT persist across page reloads — the token is in-memory only.
onMounted(() => {
  void authStore.restore()
})
</script>

<template>
  <div
    id="app-root"
    class="app"
  >
    <header class="app-header">
      <h1>{{ t('app.name') }}</h1>
    </header>
    <main class="app-main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app {
  max-width: 800px;
  margin: 0 auto;
  padding: 1rem;
  font-family: system-ui, -apple-system, sans-serif;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.app-main {
  padding: 1rem 0;
}
</style>
