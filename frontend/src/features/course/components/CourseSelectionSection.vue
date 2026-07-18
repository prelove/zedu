<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../../stores/auth'
import { updateEnrollment, type Enrollment, type EnrollmentWrite } from '../../../api/course'
import { type CourseDomain, type Track, type Level } from '../../../api/course-dict'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'

const props = defineProps<{
  enrollment: Enrollment
  domains: CourseDomain[]
  tracks: Track[]
  levels: Level[]
  dictError: string | null
}>()
const emit = defineEmits<{ saved: [enrollment: Enrollment]; retryDict: [] }>()
const { t } = useI18n()

const selDomain = ref(0)
const selTrack = ref(0)
const selTargetLevel = ref(0)
const saving = ref(false)
const saveError = ref<string | null>(null)

const filteredTracks = computed(() => props.tracks.filter((tr) => tr.domainId === selDomain.value))
const filteredLevels = computed(() => props.levels.filter((lv) => lv.trackId === selTrack.value))

watch(
  () => props.enrollment,
  (e) => {
    if (!e) return
    selDomain.value = e.domainId
    selTrack.value = e.trackId
    selTargetLevel.value = e.targetLevelId ?? 0
  },
  { immediate: true },
)

async function handleSave(): Promise<void> {
  saving.value = true
  saveError.value = null
  const e = props.enrollment
  // Course selection PATCH: ALWAYS send domainId, trackId, targetLevelId.
  // MUST NOT include currentLevelId.
  const body: EnrollmentWrite = {
    domainId: selDomain.value,
    trackId: selTrack.value,
  }
  if (selTargetLevel.value) body.targetLevelId = selTargetLevel.value

  try {
    const updated = await authStore.authedRequest((token) => updateEnrollment(token, e.id, body))
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
    class="course-selection-section"
    data-testid="course-selection-section"
  >
    <h2>{{ t('enrollments.domain') }} / {{ t('enrollments.track') }} / {{ t('enrollments.targetLevel') }}</h2>
    <p class="hint">
      {{ t('enrollments.courseSelectionHint') }}
    </p>

    <p
      v-if="props.dictError"
      class="dict-error"
      role="alert"
      aria-live="assertive"
      data-testid="course-dict-error"
    >
      {{ t(props.dictError) }}
      <button
        type="button"
        data-testid="course-dict-retry"
        @click="emit('retryDict')"
      >
        {{ t('common.retry') }}
      </button>
    </p>

    <div class="form-grid">
      <div class="form-field">
        <label for="sel-domain">{{ t('enrollments.domain') }}</label>
        <select
          id="sel-domain"
          v-model.number="selDomain"
          data-testid="sel-domain"
          @change="selTrack = 0; selTargetLevel = 0"
        >
          <option
            v-for="d in props.domains"
            :key="d.id"
            :value="d.id"
          >
            {{ d.name }}
          </option>
        </select>
      </div>
      <div class="form-field">
        <label for="sel-track">{{ t('enrollments.track') }}</label>
        <select
          id="sel-track"
          v-model.number="selTrack"
          :disabled="!selDomain"
          data-testid="sel-track"
          @change="selTargetLevel = 0"
        >
          <option :value="0">
            —
          </option>
          <option
            v-for="tr in filteredTracks"
            :key="tr.id"
            :value="tr.id"
          >
            {{ tr.name }}
          </option>
        </select>
      </div>
      <div class="form-field">
        <label for="sel-target-level">{{ t('enrollments.targetLevel') }}</label>
        <select
          id="sel-target-level"
          v-model.number="selTargetLevel"
          :disabled="!selTrack"
          data-testid="sel-target-level"
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
    </div>

    <p
      v-if="saveError"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="course-selection-error"
    >
      {{ t(saveError) }}
    </p>

    <button
      type="button"
      :disabled="props.dictError !== null || saving || !selDomain || !selTrack"
      data-testid="save-course-selection"
      @click="handleSave"
    >
      {{ saving ? t('common.saving') : t('enrollments.saveCourseSelection') }}
    </button>
  </section>
</template>

<style scoped>
.course-selection-section { margin: 1rem 0; padding: 1rem; border: 1px solid #ddd; border-radius: 6px; }
.course-selection-section h2 { margin-top: 0; font-size: 1.1rem; }
.hint { color: #666; font-size: 0.85rem; margin-bottom: 0.75rem; }
.form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 0.75rem; }
.form-field { display: flex; flex-direction: column; }
.form-field label { margin-bottom: 0.25rem; font-weight: 600; font-size: 0.85rem; }
.form-field select { padding: 0.4rem; border: 1px solid #ccc; border-radius: 4px; }
.dict-error { color: #b00020; background: #fde8e8; padding: 0.5rem; border-radius: 4px; margin-bottom: 0.75rem; }
.dict-error button { margin-left: 0.5rem; }
.no-changes { color: #555; font-size: 0.85rem; margin: 0.5rem 0; }
.form-error { color: #b00020; font-size: 0.85rem; margin: 0.5rem 0; }
button[data-testid='save-course-btn'] { margin-top: 0.5rem; padding: 0.5rem 1rem; cursor: pointer; }
button[data-testid='save-course-btn']:disabled { cursor: not-allowed; opacity: 0.6; }
</style>
