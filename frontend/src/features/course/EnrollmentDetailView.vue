<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import { getEnrollment, type Enrollment } from '../../api/course'
import { listTeachers, type Teacher } from '../../api/directory'
import LoadingState from '../../components/LoadingState.vue'
import ErrorState from '../../components/ErrorState.vue'
import CourseSelectionSection from './components/CourseSelectionSection.vue'
import LevelChangeSection from './components/LevelChangeSection.vue'
import AssignmentsSection from './components/AssignmentsSection.vue'
import { useDictionary } from '../../composables/useDictionary'

const props = defineProps<{ id: string }>()
const router = useRouter()
const { t } = useI18n()

const enrollment = ref<Enrollment | null>(null)
const loading = ref(false)
const error = ref<unknown>(null)

const { domains, tracks, levels, error: dictError, load: loadDict } = useDictionary()

const teachers = ref<Teacher[]>([])

const enrollmentId = () => Number(props.id)

async function loadEnrollment(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    enrollment.value = await authStore.authedRequest((token) => getEnrollment(token, enrollmentId()))
  } catch (err) {
    error.value = err
  } finally {
    loading.value = false
  }
}

async function loadTeachers(): Promise<void> {
  try {
    const data = await authStore.authedRequest((token) => listTeachers(token, { pageSize: 100 }))
    teachers.value = data.items
  } catch {
    // Non-critical: assignment creation will show empty teacher list.
    teachers.value = []
  }
}

onMounted(() => {
  void loadEnrollment()
  void loadDict()
  void loadTeachers()
})

function onSaved(updated: Enrollment): void {
  enrollment.value = updated
}
</script>

<template>
  <div
    class="enrollment-detail-view"
    data-testid="enrollment-detail-view"
  >
    <button
      type="button"
      data-testid="back-to-enrollments"
      @click="router.back()"
    >
      {{ t('common.back') }}
    </button>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadEnrollment"
    />
    <template v-else-if="enrollment">
      <h1>{{ t('enrollments.title') }} #{{ enrollment.id }}</h1>

      <CourseSelectionSection
        :enrollment="enrollment"
        :domains="domains"
        :tracks="tracks"
        :levels="levels"
        :dict-error="dictError"
        @saved="onSaved"
        @retry-dict="loadDict"
      />

      <LevelChangeSection
        :enrollment="enrollment"
        :levels="levels"
        :dict-error="dictError"
        @saved="onSaved"
        @retry-dict="loadDict"
      />

      <AssignmentsSection
        :enrollment-id="enrollmentId()"
        :teachers="teachers"
      />
    </template>
  </div>
</template>

<style scoped>
.enrollment-detail-view {
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
