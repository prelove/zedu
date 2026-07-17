<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { authStore } from '../../stores/auth'
import {
  listDomains, createDomain, updateDomain,
  listTracks, createTrack, updateTrack,
  listLevels, createLevel, updateLevel,
  listTags, createTag, updateTag,
  type CourseDomain, type Track, type Level, type CapabilityTag,
  type DomainWrite, type TrackWrite, type LevelWrite, type TagWrite,
} from '../../api/course'
import { ApiError, NetworkError } from '../../api/http'
import { errorToI18nKey } from '../../api/error-mapping'
import LoadingState from '../../components/LoadingState.vue'
import ErrorState from '../../components/ErrorState.vue'
import EmptyState from '../../components/EmptyState.vue'

const { t } = useI18n()

type TabName = 'domains' | 'tracks' | 'levels' | 'tags'
const activeTab = ref<TabName>('domains')

const domains = ref<CourseDomain[]>([])
const tracks = ref<Track[]>([])
const levels = ref<Level[]>([])
const tags = ref<CapabilityTag[]>([])
const loading = ref(false)
const error = ref<unknown>(null)

// Create forms.
const showForm = ref(false)
const saving = ref(false)
const formError = ref<string | null>(null)

const domainForm = ref<DomainWrite>({ name: '', code: '', type: 'LANGUAGE', sortOrder: 0 })
const trackForm = ref<TrackWrite>({ domainId: 0, name: '', code: '', sortOrder: 0 })
const levelForm = ref<LevelWrite>({ trackId: 0, name: '', code: '', sortOrder: 0 })
const tagForm = ref<TagWrite>({ domainId: 0, name: '', code: '', sortOrder: 0 })

async function loadAll(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const [dom, trk, lvl, tg] = await Promise.all([
      authStore.authedRequest((token) => listDomains(token, { pageSize: 100 })),
      authStore.authedRequest((token) => listTracks(token, { pageSize: 100 })),
      authStore.authedRequest((token) => listLevels(token, { pageSize: 100 })),
      authStore.authedRequest((token) => listTags(token, { pageSize: 100 })),
    ])
    domains.value = dom.items ?? []
    tracks.value = trk.items ?? []
    levels.value = lvl.items ?? []
    tags.value = tg.items ?? []
  } catch (err) {
    error.value = err
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadAll()
})

function openForm(): void {
  showForm.value = true
  formError.value = null
  if (activeTab.value === 'domains') {
    domainForm.value = { name: '', code: '', type: 'LANGUAGE', sortOrder: 0 }
  } else if (activeTab.value === 'tracks') {
    trackForm.value = { domainId: domains.value[0]?.id ?? 0, name: '', code: '', sortOrder: 0 }
  } else if (activeTab.value === 'levels') {
    levelForm.value = { trackId: tracks.value[0]?.id ?? 0, name: '', code: '', sortOrder: 0 }
  } else {
    tagForm.value = { domainId: domains.value[0]?.id ?? 0, name: '', code: '', sortOrder: 0 }
  }
}

async function handleCreate(): Promise<void> {
  saving.value = true
  formError.value = null
  try {
    if (activeTab.value === 'domains') {
      await authStore.authedRequest((token) => createDomain(token, domainForm.value))
    } else if (activeTab.value === 'tracks') {
      if (!trackForm.value.domainId) { formError.value = 'errors.UNKNOWN'; return }
      await authStore.authedRequest((token) => createTrack(token, trackForm.value))
    } else if (activeTab.value === 'levels') {
      if (!levelForm.value.trackId) { formError.value = 'errors.UNKNOWN'; return }
      await authStore.authedRequest((token) => createLevel(token, levelForm.value))
    } else {
      if (!tagForm.value.domainId) { formError.value = 'errors.UNKNOWN'; return }
      await authStore.authedRequest((token) => createTag(token, tagForm.value))
    }
    showForm.value = false
    await loadAll()
  } catch (err) {
    if (err instanceof NetworkError) formError.value = 'errors.NETWORK_ERROR'
    else if (err instanceof ApiError) formError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    else formError.value = 'errors.UNKNOWN'
  } finally {
    saving.value = false
  }
}

async function toggleEnabled(
  item: CourseDomain | Track | Level | CapabilityTag,
  kind: TabName,
): Promise<void> {
  try {
    const body = { enabled: !item.enabled }
    if (kind === 'domains') {
      await authStore.authedRequest((token) => updateDomain(token, item.id, body))
    } else if (kind === 'tracks') {
      await authStore.authedRequest((token) => updateTrack(token, item.id, body))
    } else if (kind === 'levels') {
      await authStore.authedRequest((token) => updateLevel(token, item.id, body))
    } else {
      await authStore.authedRequest((token) => updateTag(token, item.id, body))
    }
    await loadAll()
  } catch (err) {
    // 42201 = referenced item cannot be disabled; show stable error.
    if (err instanceof ApiError) {
      formError.value = errorToI18nKey(err) ?? 'errors.UNKNOWN'
    } else if (err instanceof NetworkError) {
      formError.value = 'errors.NETWORK_ERROR'
    } else {
      formError.value = 'errors.UNKNOWN'
    }
  }
}

function domainName(id: number): string {
  return domains.value.find((d) => d.id === id)?.name ?? String(id)
}
function trackName(id: number): string {
  return tracks.value.find((tr) => tr.id === id)?.name ?? String(id)
}

const currentItems = computed(() => {
  if (activeTab.value === 'domains') return domains.value
  if (activeTab.value === 'tracks') return tracks.value
  if (activeTab.value === 'levels') return levels.value
  return tags.value
})
</script>

<template>
  <div
    class="course-dict-view"
    data-testid="course-dict-view"
  >
    <h1>{{ t('courses.title') }}</h1>

    <div
      class="tabs"
      role="tablist"
    >
      <button
        role="tab"
        :aria-selected="activeTab === 'domains'"
        :class="{ active: activeTab === 'domains' }"
        data-testid="tab-domains"
        @click="activeTab = 'domains'"
      >
        {{ t('courses.domains') }}
      </button>
      <button
        role="tab"
        :aria-selected="activeTab === 'tracks'"
        :class="{ active: activeTab === 'tracks' }"
        data-testid="tab-tracks"
        @click="activeTab = 'tracks'"
      >
        {{ t('courses.tracks') }}
      </button>
      <button
        role="tab"
        :aria-selected="activeTab === 'levels'"
        :class="{ active: activeTab === 'levels' }"
        data-testid="tab-levels"
        @click="activeTab = 'levels'"
      >
        {{ t('courses.levels') }}
      </button>
      <button
        role="tab"
        :aria-selected="activeTab === 'tags'"
        :class="{ active: activeTab === 'tags' }"
        data-testid="tab-tags"
        @click="activeTab = 'tags'"
      >
        {{ t('courses.tags') }}
      </button>
    </div>

    <button
      type="button"
      data-testid="course-create-btn"
      @click="openForm"
    >
      {{
        activeTab === 'domains' ? t('courses.createDomain') :
        activeTab === 'tracks' ? t('courses.createTrack') :
        activeTab === 'levels' ? t('courses.createLevel') :
        t('courses.createTag')
      }}
    </button>

    <p
      v-if="formError"
      class="form-error"
      role="alert"
      aria-live="assertive"
      data-testid="course-form-error"
    >
      {{ t(formError) }}
    </p>
    <p
      v-if="formError === 'apiErrors.INVALID_STATE'"
      class="form-hint"
      data-testid="course-referenced-hint"
    >
      {{ t('courses.referencedCannotDisable') }}
    </p>

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadAll"
    />
    <EmptyState
      v-else-if="currentItems.length === 0"
    />
    <table
      v-else
      data-testid="course-table"
    >
      <thead>
        <tr>
          <th>{{ t('common.name') }}</th>
          <th>{{ t('common.code') }}</th>
          <th v-if="activeTab === 'domains'">
            {{ t('courses.domainType') }}
          </th>
          <th v-if="activeTab === 'tracks'">
            {{ t('courses.trackDomain') }}
          </th>
          <th v-if="activeTab === 'levels'">
            {{ t('courses.levelTrack') }}
          </th>
          <th v-if="activeTab === 'tags'">
            {{ t('courses.tagDomain') }}
          </th>
          <th>{{ t('common.status') }}</th>
          <th>{{ t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="item in currentItems"
          :key="item.id"
          data-testid="course-row"
        >
          <td>{{ item.name }}</td>
          <td>{{ item.code }}</td>
          <td v-if="activeTab === 'domains'">
            {{ (item as CourseDomain).type }}
          </td>
          <td v-if="activeTab === 'tracks'">
            {{ domainName((item as Track).domainId) }}
          </td>
          <td v-if="activeTab === 'levels'">
            {{ trackName((item as Level).trackId) }}
          </td>
          <td v-if="activeTab === 'tags'">
            {{ domainName((item as CapabilityTag).domainId) }}
          </td>
          <td>{{ item.enabled ? t('common.enabled') : t('common.disabled') }}</td>
          <td>
            <button
              type="button"
              :data-testid="`course-toggle-${item.id}`"
              @click="toggleEnabled(item, activeTab)"
            >
              {{ item.enabled ? t('common.disable') : t('common.enable') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Create form -->
    <div
      v-if="showForm"
      class="create-dialog"
      data-testid="course-create-form"
    >
      <form @submit.prevent="handleCreate">
        <!-- Domain form -->
        <template v-if="activeTab === 'domains'">
          <div class="form-field">
            <label for="d-name">{{ t('courses.domainName') }} *</label>
            <input
              id="d-name"
              v-model="domainForm.name"
              type="text"
              required
              data-testid="d-form-name"
            >
          </div>
          <div class="form-field">
            <label for="d-code">{{ t('courses.domainCode') }} *</label>
            <input
              id="d-code"
              v-model="domainForm.code"
              type="text"
              required
              data-testid="d-form-code"
            >
          </div>
          <div class="form-field">
            <label for="d-type">{{ t('courses.domainType') }}</label>
            <input
              id="d-type"
              v-model="domainForm.type"
              type="text"
              data-testid="d-form-type"
            >
          </div>
        </template>
        <!-- Track form -->
        <template v-if="activeTab === 'tracks'">
          <div class="form-field">
            <label for="t-domain">{{ t('courses.trackDomain') }} *</label>
            <select
              id="t-domain"
              v-model.number="trackForm.domainId"
              required
              data-testid="t-form-domain"
            >
              <option
                v-for="d in domains"
                :key="d.id"
                :value="d.id"
              >
                {{ d.name }}
              </option>
            </select>
          </div>
          <div class="form-field">
            <label for="t-name">{{ t('courses.trackName') }} *</label>
            <input
              id="t-name"
              v-model="trackForm.name"
              type="text"
              required
              data-testid="t-form-name"
            >
          </div>
          <div class="form-field">
            <label for="t-code">{{ t('courses.trackCode') }} *</label>
            <input
              id="t-code"
              v-model="trackForm.code"
              type="text"
              required
              data-testid="t-form-code"
            >
          </div>
        </template>
        <!-- Level form -->
        <template v-if="activeTab === 'levels'">
          <div class="form-field">
            <label for="l-track">{{ t('courses.levelTrack') }} *</label>
            <select
              id="l-track"
              v-model.number="levelForm.trackId"
              required
              data-testid="l-form-track"
            >
              <option
                v-for="tr in tracks"
                :key="tr.id"
                :value="tr.id"
              >
                {{ tr.name }}
              </option>
            </select>
          </div>
          <div class="form-field">
            <label for="l-name">{{ t('courses.levelName') }} *</label>
            <input
              id="l-name"
              v-model="levelForm.name"
              type="text"
              required
              data-testid="l-form-name"
            >
          </div>
          <div class="form-field">
            <label for="l-code">{{ t('courses.levelCode') }} *</label>
            <input
              id="l-code"
              v-model="levelForm.code"
              type="text"
              required
              data-testid="l-form-code"
            >
          </div>
        </template>
        <!-- Tag form -->
        <template v-if="activeTab === 'tags'">
          <div class="form-field">
            <label for="g-domain">{{ t('courses.tagDomain') }} *</label>
            <select
              id="g-domain"
              v-model.number="tagForm.domainId"
              required
              data-testid="g-form-domain"
            >
              <option
                v-for="d in domains"
                :key="d.id"
                :value="d.id"
              >
                {{ d.name }}
              </option>
            </select>
          </div>
          <div class="form-field">
            <label for="g-name">{{ t('courses.tagName') }} *</label>
            <input
              id="g-name"
              v-model="tagForm.name"
              type="text"
              required
              data-testid="g-form-name"
            >
          </div>
          <div class="form-field">
            <label for="g-code">{{ t('courses.tagCode') }} *</label>
            <input
              id="g-code"
              v-model="tagForm.code"
              type="text"
              required
              data-testid="g-form-code"
            >
          </div>
        </template>
        <div class="form-actions">
          <button
            type="button"
            :disabled="saving"
            @click="showForm = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="saving"
            data-testid="course-create-submit"
          >
            {{ saving ? t('common.creating') : t('common.create') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.course-dict-view {
  max-width: 900px;
  margin: 0 auto;
  padding: 1rem;
}

.tabs {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
  border-bottom: 2px solid #eee;
}

.tabs button {
  padding: 0.5rem 1rem;
  border: none;
  border-bottom: 2px solid transparent;
  background-color: transparent;
  cursor: pointer;
}

.tabs button.active {
  border-bottom-color: #0d6efd;
  color: #0d6efd;
  font-weight: 600;
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

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 0.5rem;
}

th, td {
  padding: 0.5rem;
  text-align: left;
  border-bottom: 1px solid #eee;
}

th {
  font-weight: 600;
  background-color: #f8f9fa;
}

.form-error {
  color: #dc3545;
  font-size: 0.875rem;
  margin: 0.5rem 0;
}

.form-hint {
  color: #856404;
  font-size: 0.8125rem;
  background-color: #fff3cd;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  margin: 0.25rem 0;
}

.create-dialog {
  border: 1px solid #ccc;
  border-radius: 0.5rem;
  padding: 1rem;
  margin-top: 1rem;
  background-color: #f8f9fa;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.75rem;
}

.form-field label {
  font-weight: 600;
}

.form-field input, .form-field select {
  padding: 0.4rem;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}
</style>
