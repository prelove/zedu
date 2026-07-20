<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { getTeacherPayableDetail, type TeacherPayableEntry } from '../../../api/payable'
import LoadingState from '../../../components/LoadingState.vue'
import ErrorState from '../../../components/ErrorState.vue'

const props = defineProps<{ teacherId: number }>()
const { t } = useI18n()
const items = ref<TeacherPayableEntry[]>([])
const loading = ref(false)
const error = ref<unknown>(null)

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const result = await authStore.authedRequest((token) => getTeacherPayableDetail(token, props.teacherId))
    items.value = result?.items ?? []
  } catch (err) {
    error.value = err
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <section
    class="payable-section"
    data-testid="payable-section"
  >
    <h2>{{ t('teachers.payableTitle') }}</h2>
    <p class="payable-hint">
      {{ t('teachers.payableHint') }}
    </p>
    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="load"
    />
    <p
      v-else-if="items.length === 0"
      data-testid="payable-empty"
    >
      {{ t('teachers.payableEmpty') }}
    </p>
    <table v-else>
      <thead>
        <tr>
          <th>{{ t('teachers.payableLessonNo') }}</th>
          <th>{{ t('teachers.payableAmount') }}</th>
          <th>{{ t('teachers.payableBalanceAfter') }}</th>
          <th>{{ t('teachers.payableCreatedAt') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="item in items"
          :key="item.id"
        >
          <td>{{ item.lessonNo }}</td>
          <td>{{ item.amountDelta }}</td>
          <td>{{ item.balanceAfter }}</td>
          <td>{{ item.createdAt }}</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<style scoped>
.payable-section {
  margin-top: 1.5rem;
}
.payable-hint {
  color: #6c757d;
  font-size: 0.875rem;
}
table {
  width: 100%;
  border-collapse: collapse;
}
th, td {
  padding: 0.5rem;
  border: 1px solid #dee2e6;
  text-align: left;
}
</style>
