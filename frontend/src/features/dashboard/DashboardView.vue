<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { createBackup, getDashboard, type Dashboard } from '../../api/dashboard'
import { authStore } from '../../stores/auth'

const data = ref<Dashboard | null>(null)
const error = ref(false)
const backup = ref('')
const backupError = ref(false)
const backing = ref(false)
const { t } = useI18n()

async function load(): Promise<void> {
  try {
    data.value = await authStore.authedRequest((token) => getDashboard(token))
    error.value = false
  } catch {
    error.value = true
  }
}

async function runBackup(): Promise<void> {
  backing.value = true
  backup.value = ''
  backupError.value = false
  try {
    backup.value = (await authStore.authedRequest((token) => createBackup(token))).file
  } catch {
    backupError.value = true
  } finally {
    backing.value = false
  }
}

onMounted(() => {
  void load()
})
</script>
<template>
  <main data-testid="dashboard-view">
    <h1>{{ t('dashboard.title') }}</h1><p
      v-if="error"
      role="alert"
    >
      {{ t('dashboard.loadError') }}
    </p>
    <template v-else-if="data">
      <p data-testid="dashboard-today-lessons">
        {{ t('dashboard.todayLessons', { count: data.todayLessons }) }}
      </p>
      <p data-testid="dashboard-pending-confirmations">
        {{ t('dashboard.pendingConfirmations', { count: data.pendingLessonConfirmations }) }}
      </p>
      <p data-testid="dashboard-renewal-needed">
        {{ t('dashboard.renewalNeeded', { count: data.renewalNeededStudents }) }}
      </p>
      <p data-testid="dashboard-teacher-payable">
        {{ t('dashboard.teacherPayable', { amount: data.teacherPayableAggregate }) }}
      </p>
      <p data-testid="dashboard-failed-notifications">
        {{ t('dashboard.failedNotifications', { count: data.failedNotifications }) }}
      </p>
      <button
        v-if="authStore.isOwner.value"
        type="button"
        data-testid="dashboard-create-backup"
        :disabled="backing"
        @click="runBackup"
      >
        {{ backing ? t('dashboard.backupCreating') : t('dashboard.createBackup') }}
      </button>
      <p
        v-if="backup"
        data-testid="dashboard-backup-created"
      >
        {{ t('dashboard.backupCreated', { file: backup }) }}
      </p>
      <p
        v-if="backupError"
        role="alert"
        data-testid="dashboard-backup-error"
      >
        {{ t('dashboard.backupError') }}
      </p>
    </template>
  </main>
</template>
