<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { EnrollmentWrite } from '../../../api/course'
import { type CourseDomain, type Level, type Track } from '../../../api/course-dict'

const props = defineProps<{
  domains: CourseDomain[]
  tracks: Track[]
  levels: Level[]
  creating: boolean
  createError: string | null
}>()

const emit = defineEmits<{
  submit: [body: EnrollmentWrite]
  cancel: []
}>()

const { t } = useI18n()
const form = ref({
  domainId: 0,
  trackId: 0,
  currentLevelId: 0,
  targetLevelId: 0,
  enrollmentType: 'ONE_TO_ONE',
})

const filteredTracks = computed(() => props.tracks.filter((track) => track.domainId === form.value.domainId))
const filteredLevels = computed(() => props.levels.filter((level) => level.trackId === form.value.trackId))

function buildBody(): EnrollmentWrite {
  const body: EnrollmentWrite = {
    domainId: form.value.domainId,
    trackId: form.value.trackId,
    enrollmentType: form.value.enrollmentType,
  }
  if (form.value.currentLevelId) body.currentLevelId = form.value.currentLevelId
  if (form.value.targetLevelId) body.targetLevelId = form.value.targetLevelId
  return body
}
</script>

<template>
  <form
    data-testid="enrollment-create-form"
    class="create-form"
    @submit.prevent="emit('submit', buildBody())"
  >
    <div class="form-row">
      <label for="enrollment-form-domain">{{ t('enrollments.domain') }}</label>
      <select
        id="enrollment-form-domain"
        v-model.number="form.domainId"
        data-testid="enrollment-form-domain"
      >
        <option :value="0">
          {{ t('common.none') }}
        </option>
        <option
          v-for="domain in domains"
          :key="domain.id"
          :value="domain.id"
        >
          {{ domain.name }}
        </option>
      </select>
    </div>
    <div class="form-row">
      <label for="enrollment-form-track">{{ t('enrollments.track') }}</label>
      <select
        id="enrollment-form-track"
        v-model.number="form.trackId"
        data-testid="enrollment-form-track"
        :disabled="!form.domainId"
      >
        <option :value="0">
          {{ t('common.none') }}
        </option>
        <option
          v-for="track in filteredTracks"
          :key="track.id"
          :value="track.id"
        >
          {{ track.name }}
        </option>
      </select>
    </div>
    <div class="form-row">
      <label for="enrollment-form-current-level">{{ t('enrollments.currentLevel') }}</label>
      <select
        id="enrollment-form-current-level"
        v-model.number="form.currentLevelId"
        data-testid="enrollment-form-current-level"
        :disabled="!form.trackId"
      >
        <option :value="0">
          {{ t('common.none') }}
        </option>
        <option
          v-for="level in filteredLevels"
          :key="level.id"
          :value="level.id"
        >
          {{ level.name }}
        </option>
      </select>
    </div>
    <div class="form-row">
      <label for="enrollment-form-target-level">{{ t('enrollments.targetLevel') }}</label>
      <select
        id="enrollment-form-target-level"
        v-model.number="form.targetLevelId"
        data-testid="enrollment-form-target-level"
        :disabled="!form.trackId"
      >
        <option :value="0">
          {{ t('common.none') }}
        </option>
        <option
          v-for="level in filteredLevels"
          :key="level.id"
          :value="level.id"
        >
          {{ level.name }}
        </option>
      </select>
    </div>
    <div class="form-row">
      <label for="enrollment-form-type">{{ t('enrollments.enrollmentType') }}</label>
      <select
        id="enrollment-form-type"
        v-model="form.enrollmentType"
        data-testid="enrollment-form-type"
      >
        <option value="ONE_TO_ONE">
          {{ t('enrollments.typeRegular') }}
        </option>
        <option value="TRIAL">
          {{ t('enrollments.typeTrial') }}
        </option>
      </select>
    </div>
    <p
      v-if="createError"
      class="create-error"
      data-testid="enrollment-create-error"
      role="alert"
    >
      {{ t(createError) }}
    </p>
    <div class="form-actions">
      <button
        type="button"
        data-testid="enrollment-create-cancel"
        @click="emit('cancel')"
      >
        {{ t('common.cancel') }}
      </button>
      <button
        type="submit"
        data-testid="enrollment-create-submit"
        :disabled="creating || !form.domainId || !form.trackId"
      >
        {{ t('common.create') }}
      </button>
    </div>
  </form>
</template>

<style scoped>
.create-form {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem;
  border: 1px solid #dee2e6;
  border-radius: 0.25rem;
  background: #f8f9fa;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-row label {
  font-size: 0.875rem;
  font-weight: 500;
}

.form-row select {
  padding: 0.25rem 0.5rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.create-error {
  color: #dc3545;
  margin: 0;
  font-size: 0.875rem;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.form-actions button {
  padding: 0.25rem 0.75rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  background: #fff;
  cursor: pointer;
}
</style>
