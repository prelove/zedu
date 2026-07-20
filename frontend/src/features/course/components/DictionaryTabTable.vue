<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingState from '../../../components/LoadingState.vue'
import ErrorState from '../../../components/ErrorState.vue'
import EmptyState from '../../../components/EmptyState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import type { Column, DictItem } from './dictionary-types'

defineProps<{
  items: DictItem[]
  columns: Column[]
  loading: boolean
  error: string | null
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  edit: [item: DictItem]
  toggle: [item: DictItem]
  pageChange: [page: number]
  retry: []
}>()

const { t } = useI18n()
</script>

<template>
  <LoadingState v-if="loading" />
  <ErrorState
    v-else-if="error"
    :error="error"
    @retry="emit('retry')"
  />
  <EmptyState v-else-if="items.length === 0" />
  <div
    v-else
    class="dict-table-wrap"
  >
    <table data-testid="course-table">
      <thead>
        <tr>
          <th
            v-for="col in columns"
            :key="col.key"
          >
            {{ t(col.label) }}
          </th>
          <th>{{ t('common.status') }}</th>
          <th>{{ t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="item in items"
          :key="item.id"
          data-testid="course-row"
        >
          <td
            v-for="col in columns"
            :key="col.key"
          >
            {{ item[col.key] }}
          </td>
          <td>{{ item.enabled ? t('common.enabled') : t('common.disabled') }}</td>
          <td class="actions">
            <button
              type="button"
              :data-testid="`course-edit-${item.id}`"
              @click="emit('edit', item)"
            >
              {{ t('common.edit') }}
            </button>
            <button
              type="button"
              :data-testid="`course-toggle-${item.id}`"
              @click="emit('toggle', item)"
            >
              {{ item.enabled ? t('common.disable') : t('common.enable') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <PaginationBar
      :page="page"
      :page-size="pageSize"
      :total="total"
      @page-change="(nextPage) => emit('pageChange', nextPage)"
    />
  </div>
</template>

<style scoped>
.dict-table-wrap {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 0.5rem;
  border: 1px solid #dee2e6;
  text-align: left;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.actions button {
  padding: 0.25rem 0.5rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}
</style>
