/** Unified list envelope: { items, page, pageSize, total }. */
export interface ListData<T> {
  items: T[]
  page: number
  pageSize: number
  total: number
}
