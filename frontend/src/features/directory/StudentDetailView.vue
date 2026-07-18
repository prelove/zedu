<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import { getStudent, type Student } from '../../api/directory'
import LoadingState from '../../components/LoadingState.vue'
import ErrorState from '../../components/ErrorState.vue'
import StudentEditForm from './components/StudentEditForm.vue'
import ParentsSection from './components/ParentsSection.vue'
import EnrollmentsSection from './components/EnrollmentsSection.vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const { t } = useI18n()

const student = ref<Student | null>(null)
const loading = ref(false)
const error = ref<unknown>(null)

const studentId = () => Number(props.id)

async function loadStudent(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    student.value = await authStore.authedRequest((token) => getStudent(token, studentId()))
  } catch (err) {
    error.value = err
  } finally {
    loading.value = false
  }
}

onMounted(loadStudent)

function onSaved(updated: Student): void {
  student.value = updated
}
</script>

<template>
  <div
    class="student-detail-view"
    data-testid="student-detail-view"
  >
    <button
      type="button"
      data-testid="back-to-students"
      @click="router.push({ name: 'students' })"
    >
      {{ t('common.back') }}
    </button>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadStudent"
    />
    <template v-else-if="student">
      <h1>{{ t('students.detailTitle') }}: {{ student.name }}</h1>

      <StudentEditForm
        :student="student"
        @saved="onSaved"
      />

      <ParentsSection :student-id="studentId()" />

      <EnrollmentsSection
        :student-id="studentId()"
        :student-status="student.status"
        @view-enrollment="(id: number) => router.push({ name: 'enrollment-detail', params: { id } })"
      />
    </template>
  </div>
</template>

<style scoped>
.student-detail-view {
  max-width: 900px;
  margin: 0 auto;
  padding: 1rem;
}

button {
  padding: 0.4rem 0.8rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background-color: #fff;
  cursor: pointer;
}

button:disabled {
  color: #6c757d;
  cursor: not-allowed;
}

section {
  margin-top: 1.5rem;
}

h1 {
  margin-top: 0.5rem;
}
</style>
