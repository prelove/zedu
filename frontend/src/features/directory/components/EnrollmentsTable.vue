<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Enrollment } from '../../../api/directory'

defineProps<{
  enrollments: Enrollment[]
  domainName: (domainId: number) => string
  trackName: (trackId: number) => string
  statusLabel: (status: string) => string
}>()

const emit = defineEmits<{ viewEnrollment: [id: number] }>()
const { t } = useI18n()
</script>

<template>
  <table
    data-testid="enrollments-table"
    class="enrollments-table"
  >
    <thead>
      <tr>
        <th>{{ t('enrollments.domain') }}</th>
        <th>{{ t('enrollments.track') }}</th>
        <th>{{ t('enrollments.enrollmentType') }}</th>
        <th>{{ t('enrollments.status') }}</th>
        <th>{{ t('common.actions') }}</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="enrollment in enrollments"
        :key="enrollment.id"
        data-testid="enrollment-row"
      >
        <td>{{ domainName(enrollment.domainId) }}</td>
        <td>{{ trackName(enrollment.trackId) }}</td>
        <td>{{ enrollment.enrollmentType === 'R' ? t('enrollments.typeRegular') : t('enrollments.typeTrial') }}</td>
        <td>{{ statusLabel(enrollment.status) }}</td>
        <td>
          <button
            type="button"
            data-testid="view-enrollment-btn"
            @click="emit('viewEnrollment', enrollment.id)"
          >
            {{ t('students.viewEnrollment') }}
          </button>
        </td>
      </tr>
    </tbody>
  </table>
</template>

<style scoped>
.enrollments-table {
  width: 100%;
  border-collapse: collapse;
}

.enrollments-table th,
.enrollments-table td {
  padding: 0.5rem;
  border-bottom: 1px solid #dee2e6;
  text-align: left;
  font-size: 0.875rem;
}

.enrollments-table th {
  background: #f8f9fa;
  font-weight: 600;
}

.enrollments-table button {
  padding: 0.2rem 0.5rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background: #fff;
  cursor: pointer;
  font-size: 0.8125rem;
}
</style>
