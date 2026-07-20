<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { type CapabilityWrite } from '../../../api/directory'
import { type CourseDomain, type Level, type Track } from '../../../api/course-dict'

const props = defineProps<{
  domains: CourseDomain[]
  tracks: Track[]
  levels: Level[]
  saving: boolean
  formError: string | null
}>()

const emit = defineEmits<{
  submit: [body: CapabilityWrite]
  cancel: []
}>()

const { t } = useI18n()
const form = ref<CapabilityWrite>({ domainId: 0, trackId: 0, levelId: 0 })

const filteredTracks = computed(() => props.tracks.filter((track) => track.domainId === form.value.domainId))
const filteredLevels = computed(() => props.levels.filter((level) => level.trackId === form.value.trackId))

function onDomainChange(): void {
  form.value.trackId = 0
  form.value.levelId = 0
}

function onTrackChange(): void {
  form.value.levelId = 0
}
</script>

<template>
  <div
    class="create-dialog"
    data-testid="capability-create-form"
  >
    <form @submit.prevent="emit('submit', { ...form })">
      <div class="form-field">
        <label for="cap-domain">{{ t('teachers.capabilityDomain') }}</label>
        <select
          id="cap-domain"
          v-model.number="form.domainId"
          data-testid="cap-form-domain"
          @change="onDomainChange"
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
      <div class="form-field">
        <label for="cap-track">{{ t('teachers.capabilityTrack') }}</label>
        <select
          id="cap-track"
          v-model.number="form.trackId"
          :disabled="!form.domainId"
          data-testid="cap-form-track"
          @change="onTrackChange"
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
      <div class="form-field">
        <label for="cap-level">{{ t('teachers.capabilityLevel') }}</label>
        <select
          id="cap-level"
          v-model.number="form.levelId"
          :disabled="!form.trackId"
          data-testid="cap-form-level"
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
      <p
        v-if="formError"
        class="form-error"
        role="alert"
        aria-live="assertive"
        data-testid="capability-create-error"
      >
        {{ t(formError) }}
      </p>
      <p
        v-if="formError === 'apiErrors.CONFLICT'"
        class="form-hint"
        data-testid="capability-duplicate-hint"
      >
        {{ t('teachers.capabilityDuplicate') }}
      </p>
      <div class="form-actions">
        <button
          type="button"
          :disabled="saving"
          @click="emit('cancel')"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          :disabled="saving || !form.domainId || !form.trackId || !form.levelId"
          data-testid="capability-create-submit"
        >
          {{ saving ? t('common.creating') : t('common.create') }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.create-dialog {
  border: 1px solid #ccc;
  border-radius: 0.5rem;
  padding: 1rem;
  background-color: #f9f9f9;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.75rem;
}

.form-field label {
  font-size: 0.8rem;
  color: #495057;
}

.form-field select {
  padding: 0.3rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.form-error {
  color: #dc3545;
  font-size: 0.8rem;
  margin: 0 0 0.5rem;
}

.form-hint {
  color: #856404;
  font-size: 0.8rem;
  margin: 0 0 0.5rem;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.form-actions button {
  padding: 0.3rem 0.75rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  cursor: pointer;
}
</style>
