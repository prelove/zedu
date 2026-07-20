<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Capability } from '../../../api/directory'

defineProps<{
  capabilities: Capability[]
  domainName: (domainId: number) => string
  trackName: (trackId: number) => string
  levelName: (levelId: number) => string
  statusLabel: (status: string) => string
}>()

const emit = defineEmits<{ end: [capability: Capability] }>()
const { t } = useI18n()
</script>

<template>
  <table data-testid="capabilities-table">
    <thead>
      <tr>
        <th>{{ t('teachers.capabilityDomain') }}</th>
        <th>{{ t('teachers.capabilityTrack') }}</th>
        <th>{{ t('teachers.capabilityLevel') }}</th>
        <th>{{ t('common.status') }}</th>
        <th>{{ t('teachers.capabilityVerified') }}</th>
        <th>{{ t('common.actions') }}</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="capability in capabilities"
        :key="capability.id"
        data-testid="capability-row"
      >
        <td>{{ domainName(capability.domainId) }}</td>
        <td>{{ trackName(capability.trackId) }}</td>
        <td>{{ levelName(capability.levelId) }}</td>
        <td>{{ statusLabel(capability.status) }}</td>
        <td>{{ capability.verified ? t('common.yes') : t('common.no') }}</td>
        <td>
          <button
            v-if="capability.status === 'ACTIVE'"
            type="button"
            data-testid="capability-end-btn"
            @click="emit('end', capability)"
          >
            {{ t('teachers.capabilityEnd') }}
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
  padding: 0.3rem 0.75rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}
</style>
