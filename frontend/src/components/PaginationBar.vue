<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  (e: 'page-change', page: number): void
}>()

const { t } = useI18n()

const totalPages = () => Math.max(1, Math.ceil(props.total / props.pageSize))

function goPrev(): void {
  if (props.page > 1) emit('page-change', props.page - 1)
}

function goNext(): void {
  if (props.page < totalPages()) emit('page-change', props.page + 1)
}
</script>

<template>
  <div
    class="pagination"
    data-testid="pagination"
  >
    <span class="pagination-info">{{ t('common.total', { count: total }) }}</span>
    <div class="pagination-controls">
      <button
        type="button"
        :disabled="page <= 1"
        data-testid="pagination-prev"
        @click="goPrev"
      >
        {{ t('common.previous') }}
      </button>
      <span class="pagination-page">{{ page }} / {{ totalPages() }}</span>
      <button
        type="button"
        :disabled="page >= totalPages()"
        data-testid="pagination-next"
        @click="goNext"
      >
        {{ t('common.next') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
  padding: 0.5rem 0;
}

.pagination-info {
  color: #6c757d;
  font-size: 0.875rem;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

button {
  padding: 0.25rem 0.5rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}

button:disabled {
  color: #6c757d;
  cursor: not-allowed;
}

.pagination-page {
  font-size: 0.875rem;
}
</style>
