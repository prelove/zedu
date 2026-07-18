import { ref, type Ref } from 'vue'
import type { ListData } from '../api/types'

/**
 * Generic paginated list state for nested sub-lists (parents, enrollments,
 * capabilities, availability, assignments). Tracks page/pageSize/total and
 * exposes computed hasNext/hasPrev for pagination controls.
 */
export function usePaginatedList<T>(
  pageSize: number = 20,
): {
  items: Ref<T[]>
  page: Ref<number>
  pageSize: Ref<number>
  total: Ref<number>
  loading: Ref<boolean>
  error: Ref<string | null>
  hasNext: Ref<boolean>
  hasPrev: Ref<boolean>
  setData: (data: ListData<T>) => void
  nextPage: () => void
  prevPage: () => void
  reset: () => void
} {
  const items = ref<T[]>([]) as Ref<T[]>
  const page = ref(1)
  const ps = ref(pageSize)
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const hasNext = ref(false)
  const hasPrev = ref(false)

  function setData(data: ListData<T>): void {
    items.value = data.items
    page.value = data.page
    ps.value = data.pageSize
    total.value = data.total
    hasNext.value = data.page * data.pageSize < data.total
    hasPrev.value = data.page > 1
  }

  function nextPage(): void {
    if (hasNext.value) page.value++
  }

  function prevPage(): void {
    if (hasPrev.value) page.value--
  }

  function reset(): void {
    items.value = []
    page.value = 1
    total.value = 0
    error.value = null
    hasNext.value = false
    hasPrev.value = false
  }

  return {
    items,
    page,
    pageSize: ps,
    total,
    loading,
    error,
    hasNext,
    hasPrev,
    setData,
    nextPage,
    prevPage,
    reset,
  }
}
