<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Availability } from '../../../api/directory'

defineProps<{
  availabilities: Availability[]
  weekdayLabel: (weekday: number) => string
}>()

const emit = defineEmits<{ edit: [availability: Availability] }>()
const { t } = useI18n()
</script>

<template>
  <table data-testid="availability-table">
    <thead>
      <tr>
        <th>{{ t('teachers.availabilityWeekday') }}</th>
        <th>{{ t('teachers.availabilityStartTime') }}</th>
        <th>{{ t('teachers.availabilityEndTime') }}</th>
        <th>{{ t('common.actions') }}</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="availability in availabilities"
        :key="availability.id"
        data-testid="availability-row"
      >
        <td>{{ weekdayLabel(availability.weekday) }}</td>
        <td>{{ availability.startTime }}</td>
        <td>{{ availability.endTime }}</td>
        <td>
          <button
            type="button"
            data-testid="availability-edit-btn"
            @click="emit('edit', availability)"
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
}

th,
td {
  padding: 0.5rem;
  border: 1px solid #dee2e6;
  text-align: left;
}

button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background: #fff;
  cursor: pointer;
}
</style>
