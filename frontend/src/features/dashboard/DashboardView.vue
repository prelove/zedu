<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { createBackup, getDashboard, type Dashboard } from '../../api/dashboard'
import { authStore } from '../../stores/auth'
const data=ref<Dashboard|null>(null);const error=ref(false);const backup=ref('')
async function load(){try{data.value=await authStore.authedRequest(token=>getDashboard(token))}catch{error.value=true}}
async function runBackup(){try{backup.value=(await authStore.authedRequest(token=>createBackup(token))).file}catch{error.value=true}}
onMounted(()=>{void load()})
</script>
<template>
  <main>
    <h1>Dashboard</h1><p
      v-if="error"
      role="alert"
    >
      Unable to load operations data.
    </p><template v-else-if="data">
      <p>Pending confirmations: {{ data.pendingLessonConfirmations }}</p><p>Failed notifications: {{ data.failedNotifications }}</p><button
        v-if="authStore.isOwner.value"
        type="button"
        @click="runBackup"
      >
        Create backup
      </button><p v-if="backup">
        Backup: {{ backup }}
      </p>
    </template>
  </main>
</template>
