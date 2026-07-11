<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { classifyError, mapApiError } from '../utils/errors'
import type { ApiErrorState } from '../utils/errors'
import type { Locale } from '../i18n/config'

type HealthState = 'loading' | 'healthy' | 'unavailable'

const { t, locale } = useI18n()

const state = ref<HealthState>('loading')
const errorMessage = ref<string>('')

async function checkHealth(): Promise<void> {
  state.value = 'loading'
  errorMessage.value = ''
  try {
    const response = await fetch('/api/healthz')
    if (response.ok) {
      state.value = 'healthy'
    } else {
      state.value = 'unavailable'
      errorMessage.value = mapApiError('SERVER_ERROR', locale.value as Locale)
    }
  } catch (error: unknown) {
    state.value = 'unavailable'
    const errorState: ApiErrorState = classifyError(error)
    errorMessage.value = mapApiError(errorState, locale.value as Locale)
  }
}

onMounted(() => {
  void checkHealth()
})
</script>

<template>
  <section
    class="health-status"
    data-testid="health-status"
  >
    <h2>{{ t('health.title') }}</h2>
    <p
      v-if="state === 'loading'"
      class="state-loading"
      data-testid="health-state-loading"
    >
      {{ t('health.loading') }}
    </p>
    <p
      v-else-if="state === 'healthy'"
      class="state-healthy"
      data-testid="health-state-healthy"
    >
      {{ t('health.healthy') }}
    </p>
    <p
      v-else
      class="state-unavailable"
      data-testid="health-state-unavailable"
    >
      {{ t('health.unavailable') }}
    </p>
    <button
      v-if="state === 'unavailable'"
      data-testid="health-retry"
      @click="checkHealth"
    >
      {{ t('health.retry') }}
    </button>
  </section>
</template>

<style scoped>
.health-status {
  padding: 1rem;
  border: 1px solid #ddd;
  border-radius: 0.5rem;
  margin: 1rem 0;
}

.state-healthy {
  color: #28a745;
  font-weight: bold;
}

.state-unavailable {
  color: #dc3545;
  font-weight: bold;
}

.state-loading {
  color: #6c757d;
}
</style>
