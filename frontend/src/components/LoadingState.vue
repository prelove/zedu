<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <div
    class="state-loading"
    data-testid="state-loading"
    role="status"
    aria-live="polite"
  >
    {{ t('common.loading') }}
  </div>
</template>

<style scoped>
.state-loading {
  color: #6c757d;
  padding: 1rem 0;
}
</style>
