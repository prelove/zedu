<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import { getTeacher, type Teacher } from '../../api/directory'
import LoadingState from '../../components/LoadingState.vue'
import ErrorState from '../../components/ErrorState.vue'
import TeacherEditForm from './components/TeacherEditForm.vue'
import CapabilitiesSection from './components/CapabilitiesSection.vue'
import AvailabilitySection from './components/AvailabilitySection.vue'
import PayableSection from './components/PayableSection.vue'
import { useDictionary } from '../../composables/useDictionary'

const props = defineProps<{ id: string }>()
const router = useRouter()
const { t } = useI18n()

const teacher = ref<Teacher | null>(null)
const loading = ref(false)
const error = ref<unknown>(null)

const { domains, tracks, levels, loading: dictLoading, error: dictError, load: loadDict } = useDictionary()

const teacherId = () => Number(props.id)

async function loadTeacher(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    teacher.value = await authStore.authedRequest((token) => getTeacher(token, teacherId()))
  } catch (err) {
    error.value = err
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadTeacher()
  void loadDict()
})

function onSaved(updated: Teacher): void {
  teacher.value = updated
}
</script>

<template>
  <div
    class="teacher-detail-view"
    data-testid="teacher-detail-view"
  >
    <button
      type="button"
      data-testid="back-to-teachers"
      @click="router.push({ name: 'teachers' })"
    >
      {{ t('common.back') }}
    </button>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadTeacher"
    />
    <template v-else-if="teacher">
      <h1>{{ t('teachers.detailTitle') }}: {{ teacher.name }}</h1>

      <TeacherEditForm
        :teacher="teacher"
        @saved="onSaved"
      />

      <!-- Dictionary load error: visible, retryable, disables dependent submit -->
      <div
        v-if="dictError"
        class="dict-error-banner"
        role="alert"
        aria-live="assertive"
        data-testid="teacher-dict-error"
      >
        <p>{{ t('common.dictionaryLoadError') }}</p>
        <button
          type="button"
          data-testid="teacher-dict-retry"
          :disabled="dictLoading"
          @click="loadDict"
        >
          {{ t('common.retryDictionary') }}
        </button>
      </div>

      <CapabilitiesSection
        :teacher-id="teacherId()"
        :domains="domains"
        :tracks="tracks"
        :levels="levels"
        :dict-error="dictError"
      />

      <AvailabilitySection :teacher-id="teacherId()" />

      <PayableSection :teacher-id="teacherId()" />
    </template>
  </div>
</template>

<style scoped>
.teacher-detail-view {
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

.dict-error-banner {
  border: 1px solid #dc3545;
  border-radius: 0.25rem;
  padding: 0.75rem;
  margin-top: 1rem;
  background-color: #f8d7da;
  color: #721c24;
}

.dict-error-banner p {
  margin: 0 0 0.5rem 0;
  font-weight: 600;
}
</style>
