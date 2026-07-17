<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  message?: string
}>()

const { t } = useI18n()
</script>

<template>
  <div
    class="state-empty"
    data-testid="state-empty"
  >
    {{ message ?? t('common.noData') }}
  </div>
</template>

<style scoped>
.state-empty {
  color: #6c757d;
  padding: 1rem 0;
  font-style: italic;
}
</style>
