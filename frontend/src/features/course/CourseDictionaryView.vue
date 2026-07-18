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
} from '../../api/course-dict'
import { ApiError, NetworkError } from '../../api/http'
import { errorToI18nKey } from '../../api/error-mapping'
import LoadingState from '../../components/LoadingState.vue'
import ErrorState from '../../components/ErrorState.vue'
import DictionaryTab from './components/DictionaryTab.vue'

interface DictItem { id: number; name: string; code: string; enabled: boolean; [k: string]: unknown }
interface Column { key: string; label: string }
interface FormField {
  key: string; label: string; type: 'text' | 'number' | 'select'; testid: string
  options?: { value: number | string; label: string }[]
  defaultValue?: unknown
}

const { t } = useI18n()

type TabName = 'domains' | 'tracks' | 'levels' | 'tags'
const activeTab = ref<TabName>('domains')

const domains = ref<CourseDomain[]>([])
const tracks = ref<Track[]>([])
const levels = ref<Level[]>([])
const tags = ref<CapabilityTag[]>([])
const loading = ref(false)
const error = ref<unknown>(null)
const formError = ref<string | null>(null)
const saving = ref(false)

function domainName(id: number): string {
  return domains.value.find((d) => d.id === id)?.name ?? String(id)
}
function trackName(id: number): string {
  return tracks.value.find((tr) => tr.id === id)?.name ?? String(id)
}

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

function toErrorKey(err: unknown): string {
  if (err instanceof NetworkError) return 'errors.NETWORK_ERROR'
  if (err instanceof ApiError) return errorToI18nKey(err) ?? 'errors.UNKNOWN'
  return 'errors.UNKNOWN'
}

async function handleCreate(body: Record<string, unknown>): Promise<void> {
  saving.value = true
  formError.value = null
  try {
    if (activeTab.value === 'domains') {
      await authStore.authedRequest((token) => createDomain(token, body))
    } else if (activeTab.value === 'tracks') {
      if (!body.domainId) { formError.value = 'errors.UNKNOWN'; return }
      await authStore.authedRequest((token) => createTrack(token, body))
    } else if (activeTab.value === 'levels') {
      if (!body.trackId) { formError.value = 'errors.UNKNOWN'; return }
      await authStore.authedRequest((token) => createLevel(token, body))
    } else {
      if (!body.domainId) { formError.value = 'errors.UNKNOWN'; return }
      await authStore.authedRequest((token) => createTag(token, body))
    }
    await loadAll()
  } catch (err) {
    formError.value = toErrorKey(err)
  } finally {
    saving.value = false
  }
}

async function handleEdit(item: DictItem, body: Record<string, unknown>): Promise<void> {
  saving.value = true
  formError.value = null
  try {
    if (activeTab.value === 'domains') {
      await authStore.authedRequest((token) => updateDomain(token, item.id, body))
    } else if (activeTab.value === 'tracks') {
      await authStore.authedRequest((token) => updateTrack(token, item.id, body))
    } else if (activeTab.value === 'levels') {
      await authStore.authedRequest((token) => updateLevel(token, item.id, body))
    } else {
      await authStore.authedRequest((token) => updateTag(token, item.id, body))
    }
    await loadAll()
  } catch (err) {
    formError.value = toErrorKey(err)
  } finally {
    saving.value = false
  }
}

async function handleToggle(item: DictItem): Promise<void> {
  formError.value = null
  try {
    const body = { enabled: !item.enabled }
    if (activeTab.value === 'domains') {
      await authStore.authedRequest((token) => updateDomain(token, item.id, body))
    } else if (activeTab.value === 'tracks') {
      await authStore.authedRequest((token) => updateTrack(token, item.id, body))
    } else if (activeTab.value === 'levels') {
      await authStore.authedRequest((token) => updateLevel(token, item.id, body))
    } else {
      await authStore.authedRequest((token) => updateTag(token, item.id, body))
    }
    await loadAll()
  } catch (err) {
    formError.value = toErrorKey(err)
  }
}

const currentItems = computed<DictItem[]>(() => {
  if (activeTab.value === 'domains') return domains.value
  if (activeTab.value === 'tracks') return tracks.value
  if (activeTab.value === 'levels') return levels.value
  return tags.value
})

const currentColumns = computed<Column[]>(() => {
  const base: Column[] = [
    { key: 'name', label: 'common.name' },
    { key: 'code', label: 'common.code' },
  ]
  if (activeTab.value === 'domains') return [...base, { key: 'type', label: 'courses.domainType' }]
  if (activeTab.value === 'tracks') return [...base, { key: 'domainName', label: 'courses.trackDomain' }]
  if (activeTab.value === 'levels') return [...base, { key: 'trackName', label: 'courses.levelTrack' }]
  return [...base, { key: 'domainName', label: 'courses.tagDomain' }]
})

const displayItems = computed<DictItem[]>(() => {
  return currentItems.value.map((item) => {
    const enriched: DictItem = { ...item }
    if (activeTab.value === 'tracks' || activeTab.value === 'tags') {
      enriched.domainName = domainName((item as Track | CapabilityTag).domainId)
    } else if (activeTab.value === 'levels') {
      enriched.trackName = trackName((item as Level).trackId)
    }
    return enriched
  })
})

const currentFormFields = computed<FormField[]>(() => {
  if (activeTab.value === 'domains') {
    return [
      { key: 'name', label: 'courses.domainName', type: 'text', testid: 'd-form-name' },
      { key: 'code', label: 'courses.domainCode', type: 'text', testid: 'd-form-code' },
      { key: 'type', label: 'courses.domainType', type: 'text', testid: 'd-form-type', defaultValue: 'LANGUAGE' },
    ]
  }
  if (activeTab.value === 'tracks') {
    return [
      { key: 'domainId', label: 'courses.trackDomain', type: 'select', testid: 't-form-domain',
        options: domains.value.map((d) => ({ value: d.id, label: d.name })) },
      { key: 'name', label: 'courses.trackName', type: 'text', testid: 't-form-name' },
      { key: 'code', label: 'courses.trackCode', type: 'text', testid: 't-form-code' },
    ]
  }
  if (activeTab.value === 'levels') {
    return [
      { key: 'trackId', label: 'courses.levelTrack', type: 'select', testid: 'l-form-track',
        options: tracks.value.map((tr) => ({ value: tr.id, label: tr.name })) },
      { key: 'name', label: 'courses.levelName', type: 'text', testid: 'l-form-name' },
      { key: 'code', label: 'courses.levelCode', type: 'text', testid: 'l-form-code' },
    ]
  }
  return [
    { key: 'domainId', label: 'courses.tagDomain', type: 'select', testid: 'g-form-domain',
      options: domains.value.map((d) => ({ value: d.id, label: d.name })) },
    { key: 'name', label: 'courses.tagName', type: 'text', testid: 'g-form-name' },
    { key: 'code', label: 'courses.tagCode', type: 'text', testid: 'g-form-code' },
  ]
})

const createLabel = computed(() => {
  if (activeTab.value === 'domains') return 'courses.createDomain'
  if (activeTab.value === 'tracks') return 'courses.createTrack'
  if (activeTab.value === 'levels') return 'courses.createLevel'
  return 'courses.createTag'
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

    <LoadingState v-if="loading" />
    <ErrorState
      v-else-if="error"
      :error="error"
      @retry="loadAll"
    />
    <DictionaryTab
      v-else
      :items="displayItems"
      :columns="currentColumns"
      :form-fields="currentFormFields"
      :loading="false"
      :error="null"
      :page="1"
      :page-size="100"
      :total="currentItems.length"
      :has-next="false"
      :has-prev="false"
      :can-create="true"
      :saving="saving"
      :form-error="formError"
      :create-label="createLabel"
      :edit-label="createLabel"
      @create="handleCreate"
      @edit="handleEdit"
      @toggle="handleToggle"
      @retry="loadAll"
    />
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

h1 {
  margin: 0 0 1rem 0;
}
</style>
