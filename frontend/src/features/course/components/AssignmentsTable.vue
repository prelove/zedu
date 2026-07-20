<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import EmptyState from '../../../components/EmptyState.vue'
import PaginationBar from '../../../components/PaginationBar.vue'
import type { Assignment } from '../../../api/course'

defineProps<{
  assignments: Assignment[]
  page: number
  pageSize: number
  total: number
  teacherName: (teacherId: number) => string
  statusLabel: (status: string) => string
  roleLabel: (role: string) => string
}>()

const emit = defineEmits<{
  end: [assignment: Assignment]
  pageChange: [page: number]
}>()

const { t } = useI18n()
</script>

<template>
  <EmptyState
    v-if="assignments.length === 0"
    :message="t('enrollments.noTeacherEnrollment')"
  />
  <div
    v-else
    class="assignments-table-wrap"
  >
    <table data-testid="assignments-table">
      <thead>
        <tr>
          <th>{{ t('enrollments.assignmentTeacher') }}</th>
          <th>{{ t('enrollments.assignmentRoleType') }}</th>
          <th>{{ t('enrollments.assignmentStatus') }}</th>
          <th>{{ t('enrollments.assignmentStartDate') }}</th>
          <th>{{ t('enrollments.assignmentEndDate') }}</th>
          <th>{{ t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="assignment in assignments"
          :key="assignment.id"
          data-testid="assignment-row"
        >
          <td>{{ teacherName(assignment.teacherId) }}</td>
          <td>{{ roleLabel(assignment.roleType) }}</td>
          <td>{{ statusLabel(assignment.status) }}</td>
          <td>{{ assignment.startDate }}</td>
          <td>{{ assignment.endDate ?? t('common.none') }}</td>
          <td>
            <button
              v-if="assignment.status === 'ACTIVE'"
              type="button"
              data-testid="end-assignment-btn"
              @click="emit('end', assignment)"
            >
              {{ t('enrollments.endAssignment') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <PaginationBar
      :page="page"
      :page-size="pageSize"
      :total="total"
      data-testid="assignments-pagination"
      @page-change="(nextPage) => emit('pageChange', nextPage)"
    />
  </div>
</template>

<style scoped>
.assignments-table-wrap {
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

button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}
</style>
