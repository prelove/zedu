<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import HealthStatus from './components/HealthStatus.vue'
import LocaleSwitcher from './components/LocaleSwitcher.vue'

const { t } = useI18n()
</script>

<template>
  <div
    id="app-root"
    class="app"
  >
    <header class="app-header">
      <h1>{{ t('app.name') }}</h1>
      <LocaleSwitcher />
    </header>
    <main class="app-main">
      <HealthStatus />
      <p class="version-info">
        {{ t('app.version') }}: {{ t('app.versionPlaceholder') }}
      </p>
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

.version-info {
  color: #6c757d;
  font-size: 0.875rem;
}
</style>
