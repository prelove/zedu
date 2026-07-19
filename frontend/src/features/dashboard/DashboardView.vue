<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { createBackup, getDashboard, type Dashboard } from '../../api/dashboard'
import { authStore } from '../../stores/auth'
const data=ref<Dashboard|null>(null);const error=ref(false);const backup=ref('')
const { t } = useI18n()
async function load(){try{data.value=await authStore.authedRequest(token=>getDashboard(token))}catch{error.value=true}}
async function runBackup(){try{backup.value=(await authStore.authedRequest(token=>createBackup(token))).file}catch{error.value=true}}
onMounted(()=>{void load()})
</script>
<template>
  <main>
    <h1>{{ t('dashboard.title') }}</h1><p
      v-if="error"
      role="alert"
    >
      {{ t('dashboard.loadError') }}
    </p><template v-else-if="data">
      <p>{{ t('dashboard.pendingConfirmations', { count: data.pendingLessonConfirmations }) }}</p><p>{{ t('dashboard.failedNotifications', { count: data.failedNotifications }) }}</p><button
        v-if="authStore.isOwner.value"
        type="button"
        @click="runBackup"
      >
        {{ t('dashboard.createBackup') }}
      </button><p v-if="backup">
        {{ t('dashboard.backupCreated', { file: backup }) }}
      </p>
    </template>
  </main>
</template>
