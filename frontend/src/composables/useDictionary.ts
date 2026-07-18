import { ref, type Ref } from 'vue'
import { authStore } from '../stores/auth'
import { listDomains, listTracks, listLevels, type CourseDomain, type Track, type Level } from '../api/course-dict'
import { ApiError, NetworkError } from '../api/http'
import { errorToI18nKey } from '../api/error-mapping'

export interface DictionaryState {
  domains: Ref<CourseDomain[]>
  tracks: Ref<Track[]>
  levels: Ref<Level[]>
  loading: Ref<boolean>
  error: Ref<string | null>
}

/**
 * Loads course dictionary (domains, tracks, levels) for views that need
 * hierarchical course selection. Exposes loading/error state so the UI can
 * show a visible error, retry, and disable dependent submit buttons when
 * the dictionary fails to load.
 */
export function useDictionary(): DictionaryState & {
  load: () => Promise<void>
} {
  const domains = ref<CourseDomain[]>([]) as Ref<CourseDomain[]>
  const tracks = ref<Track[]>([]) as Ref<Track[]>
  const levels = ref<Level[]>([]) as Ref<Level[]>
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const [d, t, l] = await Promise.all([
        authStore.authedRequest((token) => listDomains(token, { pageSize: 100 })),
        authStore.authedRequest((token) => listTracks(token, { pageSize: 100 })),
        authStore.authedRequest((token) => listLevels(token, { pageSize: 100 })),
      ])
      domains.value = d.items
      tracks.value = t.items
      levels.value = l.items
    } catch (err) {
      if (err instanceof ApiError || err instanceof NetworkError) {
        error.value = errorToI18nKey(err)
      } else {
        error.value = 'errors.UNKNOWN'
      }
    } finally {
      loading.value = false
    }
  }

  return { domains, tracks, levels, loading, error, load }
}
