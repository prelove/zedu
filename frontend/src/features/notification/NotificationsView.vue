<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  listNotificationOutbox,
  processNotificationOutbox,
  retryNotification,
  type NotificationOutbox,
} from '../../api/notification'
import { authStore } from '../../stores/auth'

const { t } = useI18n()
const items = ref<NotificationOutbox[]>([])
const error = ref(false)
const busy = ref(false)

async function load(): Promise<void> {
  try {
    items.value = (await authStore.authedRequest((token) => listNotificationOutbox(token))).items
    error.value = false
  } catch {
    error.value = true
  }
}

async function process(): Promise<void> {
  busy.value = true
  try {
    await authStore.authedRequest((token) => processNotificationOutbox(token))
    await load()
  } catch {
    error.value = true
  } finally {
    busy.value = false
  }
}

async function retry(id: number): Promise<void> {
  try {
    await authStore.authedRequest((token) => retryNotification(token, id))
    await load()
  } catch {
    error.value = true
  }
}

function statusLabel(status: string): string {
  if (status === 'SENT') return t('notifications.sent')
  if (status === 'FAILED') return t('notifications.failed')
  if (status === 'PENDING') return t('notifications.pending')
  return status
}

onMounted(() => {
  void load()
})
</script>
<template>
  <main data-testid="notifications-view">
    <h1>{{ t('notifications.title') }}</h1>
    <button
      type="button"
      data-testid="notifications-process"
      :disabled="busy"
      @click="process"
    >
      {{ t('notifications.process') }}
    </button>
    <p
      v-if="error"
      role="alert"
    >
      {{ t('errors.UNKNOWN') }}
    </p>
    <p
      class="manual-retry-hint"
      data-testid="notifications-manual-retry-hint"
    >
      {{ t('notifications.manualRetryHint') }}
    </p>
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>{{ t('notifications.event') }}</th>
          <th>{{ t('notifications.status') }}</th>
          <th>{{ t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="item in items"
          :key="item.id"
          :data-testid="`notification-row-${item.id}`"
        >
          <td>{{ item.id }}</td>
          <td>{{ item.eventType === 'LESSON_REMINDER' ? t('notifications.reminder') : item.eventType }}</td>
          <td>{{ statusLabel(item.status) }}</td>
          <td>
            <button
              v-if="item.status === 'FAILED' && item.attempts < 3"
              type="button"
              data-testid="notification-retry"
              @click="retry(item.id)"
            >
              {{ t('notifications.retry') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </main>
</template>

<style scoped>
.manual-retry-hint {
  color: #6c757d;
  font-size: 0.875rem;
}
</style>
