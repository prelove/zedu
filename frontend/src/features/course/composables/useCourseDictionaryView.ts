import { computed, reactive, ref } from 'vue'
import { authStore } from '../../../stores/auth'
import {
  createDomain,
  createLevel,
  createTag,
  createTrack,
  listDomains,
  listLevels,
  listTags,
  listTracks,
  updateDomain,
  updateLevel,
  updateTag,
  updateTrack,
  type CourseDomain,
  type Track,
} from '../../../api/course-dict'
import { ApiError, NetworkError } from '../../../api/http'
import { errorToI18nKey } from '../../../api/error-mapping'
import type { DictItem } from '../components/dictionary-types'
import {
  buildColumns,
  buildCreateLabel,
  buildDisplayItems,
  buildFormFields,
  isTabItemList,
  shouldRefreshReferences,
  type DictionaryTabName,
} from './dictionary-view-helpers'

interface TabState {
  items: DictItem[]
  page: number
  pageSize: number
  total: number
  loading: boolean
  error: string | null
}

const DEFAULT_PAGE_SIZE = 20

function emptyTabState(): TabState {
  return { items: [], page: 1, pageSize: DEFAULT_PAGE_SIZE, total: 0, loading: false, error: null }
}

function toErrorKey(err: unknown): string {
  if (err instanceof NetworkError) return 'errors.NETWORK_ERROR'
  if (err instanceof ApiError) return errorToI18nKey(err) ?? 'errors.UNKNOWN'
  return 'errors.UNKNOWN'
}

export function useCourseDictionaryView() {
  const activeTab = ref<DictionaryTabName>('domains')
  const saving = ref(false)
  const formError = ref<string | null>(null)
  const domainOptions = ref<CourseDomain[]>([])
  const trackOptions = ref<Track[]>([])

  const tabs = reactive<Record<DictionaryTabName, TabState>>({
    domains: emptyTabState(),
    tracks: emptyTabState(),
    levels: emptyTabState(),
    tags: emptyTabState(),
  })

  async function fetchAllPages<T>(loadPage: (page: number) => Promise<{ items: T[]; page: number; pageSize: number; total: number }>): Promise<T[]> {
    const pageSize = 100
    let page = 1
    let total = 0
    const items: T[] = []
    do {
      const data = await loadPage(page)
      items.push(...data.items)
      total = data.total
      page += 1
      if (data.pageSize <= 0) break
    } while ((page - 1) * pageSize < total)
    return items
  }

  async function loadReferenceData(): Promise<void> {
    const [domains, tracks] = await Promise.all([
      fetchAllPages((page) => authStore.authedRequest((token) => listDomains(token, { page, pageSize: 100 }))),
      fetchAllPages((page) => authStore.authedRequest((token) => listTracks(token, { page, pageSize: 100 }))),
    ])
    domainOptions.value = domains
    trackOptions.value = tracks
  }

  async function loadTab(tabName: DictionaryTabName, page = tabs[tabName].page): Promise<void> {
    const state = tabs[tabName]
    state.loading = true
    state.error = null
    try {
      const params = { page, pageSize: state.pageSize || DEFAULT_PAGE_SIZE }
      const data = await authStore.authedRequest((token) => {
        if (tabName === 'domains') return listDomains(token, params)
        if (tabName === 'tracks') return listTracks(token, params)
        if (tabName === 'levels') return listLevels(token, params)
        return listTags(token, params)
      })
      state.items = isTabItemList(tabName, data.items)
      state.page = data.page
      state.pageSize = data.pageSize
      state.total = data.total
    } catch (err) {
      state.error = toErrorKey(err)
    } finally {
      state.loading = false
    }
  }

  async function initialize(): Promise<void> {
    await Promise.all([loadTab(activeTab.value), loadReferenceData().catch(() => undefined)])
  }

  function activateTab(tabName: DictionaryTabName): void {
    activeTab.value = tabName
    formError.value = null
    if (tabs[tabName].items.length === 0 && !tabs[tabName].loading && !tabs[tabName].error) {
      void loadTab(tabName)
    }
  }

  async function refreshAfterMutation(refreshReferences: boolean): Promise<void> {
    const tasks: Array<Promise<void>> = [loadTab(activeTab.value, tabs[activeTab.value].page)]
    if (refreshReferences) tasks.push(loadReferenceData())
    await Promise.all(tasks)
  }

  async function handleCreate(body: Record<string, unknown>): Promise<void> {
    saving.value = true
    formError.value = null
    try {
      if (activeTab.value === 'domains') {
        await authStore.authedRequest((token) => createDomain(token, body))
      } else if (activeTab.value === 'tracks') {
        await authStore.authedRequest((token) => createTrack(token, body))
      } else if (activeTab.value === 'levels') {
        await authStore.authedRequest((token) => createLevel(token, body))
      } else {
        await authStore.authedRequest((token) => createTag(token, body))
      }
      await refreshAfterMutation(shouldRefreshReferences(activeTab.value))
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
      await refreshAfterMutation(shouldRefreshReferences(activeTab.value))
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
      await refreshAfterMutation(shouldRefreshReferences(activeTab.value))
    } catch (err) {
      formError.value = toErrorKey(err)
    }
  }

  const currentState = computed(() => tabs[activeTab.value])
  const currentColumns = computed(() => buildColumns(activeTab.value))
  const displayItems = computed(() => buildDisplayItems(activeTab.value, currentState.value.items, domainOptions.value, trackOptions.value))
  const currentFormFields = computed(() => buildFormFields(activeTab.value, domainOptions.value, trackOptions.value))
  const createLabel = computed(() => buildCreateLabel(activeTab.value))

  return {
    activeTab,
    createLabel,
    currentColumns,
    currentFormFields,
    currentState,
    displayItems,
    formError,
    saving,
    activateTab,
    changePage: (page: number) => loadTab(activeTab.value, page),
    handleCreate,
    handleEdit,
    handleToggle,
    initialize,
    retryActiveTab: () => loadTab(activeTab.value),
  }
}
