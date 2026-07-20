<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Parent } from '../../../api/directory'

defineProps<{ parents: Parent[] }>()

const emit = defineEmits<{ edit: [parent: Parent] }>()
const { t } = useI18n()
</script>

<template>
  <table data-testid="parents-table">
    <thead>
      <tr>
        <th>{{ t('students.parentName') }}</th>
        <th>{{ t('students.parentEmail') }}</th>
        <th>{{ t('students.parentPhone') }}</th>
        <th>{{ t('students.parentRelationship') }}</th>
        <th>{{ t('students.parentIsPrimary') }}</th>
        <th>{{ t('common.actions') }}</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="parent in parents"
        :key="parent.id"
        data-testid="parent-row"
      >
        <td>{{ parent.name }}</td>
        <td>{{ parent.email ?? t('common.none') }}</td>
        <td>{{ parent.phone ?? t('common.none') }}</td>
        <td>{{ parent.relationship ?? t('common.none') }}</td>
        <td>{{ parent.isPrimary ? t('common.yes') : t('common.no') }}</td>
        <td>
          <button
            type="button"
            data-testid="parent-edit-btn"
            @click="emit('edit', parent)"
          >
            {{ t('common.edit') }}
          </button>
        </td>
      </tr>
    </tbody>
  </table>
</template>

<style scoped>
table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 0.5rem;
}

th,
td {
  border: 1px solid #dee2e6;
  padding: 0.5rem;
  text-align: left;
}

button {
  padding: 0.25rem 0.75rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}
</style>
