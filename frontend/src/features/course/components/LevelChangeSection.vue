<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { updateEnrollment, type Enrollment, type EnrollmentWrite } from '../../../api/course'
import { type Level } from '../../../api/course-dict'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'

const props = defineProps<{ enrollment: Enrollment; levels: Level[]; dictError: string | null }>()
const emit = defineEmits<{ saved: [enrollment: Enrollment]; retryDict: [] }>()
const { t } = useI18n()

const selCurrentLevel = ref(0)
const effectiveCurrentLevel = ref(0)
const saving = ref(false)
const saveError = ref<string | null>(null)
const sameLevelRejected = ref(false)

const filteredLevels = computed(() => props.levels.filter((l) => l.trackId === props.enrollment.trackId))

watch(() => props.enrollment.id, () => {
  const e = props.enrollment
  selCurrentLevel.value = e.currentLevelId ?? 0
  effectiveCurrentLevel.value = e.currentLevelId ?? 0
  sameLevelRejected.value = false
}, { immediate: true })

async function handleSave(): Promise<void> {
  saving.value = true
  saveError.value = null
  sameLevelRejected.value = false

  // The enrollment row keeps its initial-level snapshot. Later level changes
  // are represented by events, so retain this page's effective level locally.
  if (selCurrentLevel.value === effectiveCurrentLevel.value) {
    sameLevelRejected.value = true
    saving.value = false
    return
  }

  // Level change PATCH: ONLY currentLevelId.
  // MUST NOT include domainId, trackId, or targetLevelId.
  const body: EnrollmentWrite = { currentLevelId: selCurrentLevel.value }

  try {
    const updated = await authStore.authedRequest((token) => updateEnrollment(token, props.enrollment.id, body))
    effectiveCurrentLevel.value = selCurrentLevel.value
    emit('saved', updated)
  } catch (err) {
    if (err instanceof NetworkError) saveError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) saveError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else saveError.value = 'errors.UNKNOWN'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <section
    class="level-change-section"
    data-testid="level-change-section"
  >
    <h2>{{ t('enrollments.currentLevel') }}</h2>
    <p class="hint">
      {{ t('enrollments.levelChangeHint') }}
    </p>

    <p
      v-if="props.dictError"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="level-dict-error"
    >
      {{ t(props.dictError) }}
      <button
        type="button"
        data-testid="level-dict-retry"
        @click="emit('retryDict')"
      >
        {{ t('common.retry') }}
      </button>
    </p>

    <div class="form-field">
      <label for="sel-current-level">{{ t('enrollments.currentLevel') }}</label>
      <select
        id="sel-current-level"
        v-model.number="selCurrentLevel"
        data-testid="sel-current-level"
      >
        <option :value="0">
          —
        </option>
        <option
          v-for="lv in filteredLevels"
          :key="lv.id"
          :value="lv.id"
        >
          {{ lv.name }}
        </option>
      </select>
    </div>

    <p
      v-if="sameLevelRejected"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="level-change-error"
    >
      {{ t('enrollments.sameLevelRejected') }}
    </p>

    <p
      v-if="saveError"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="level-change-error"
    >
      {{ t(saveError) }}
    </p>

    <button
      type="button"
      :disabled="!!props.dictError || saving"
      data-testid="save-level-change"
      @click="handleSave"
    >
      {{ saving ? t('common.saving') : t('enrollments.saveLevelChange') }}
    </button>
  </section>
</template>

<style scoped>
.level-change-section {
  margin-top: 1.5rem;
  padding: 1rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}
.level-change-section h2 {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
}
.hint {
  margin: 0 0 0.75rem;
  color: #666;
  font-size: 0.875rem;
}
.form-field {
  display: flex;
  flex-direction: column;
  margin-bottom: 0.75rem;
}
.form-field label {
  margin-bottom: 0.25rem;
  font-weight: 600;
  font-size: 0.875rem;
}
.form-field select {
  padding: 0.4rem 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.form-error {
  margin: 0 0 0.75rem;
  color: #c00;
  font-size: 0.875rem;
}
.form-warning {
  margin: 0 0 0.75rem;
  color: #b80;
  font-size: 0.875rem;
}
button {
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  background: #f5f5f5;
  cursor: pointer;
}
button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
