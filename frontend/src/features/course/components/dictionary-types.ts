export interface DictItem {
  id: number
  name: string
  code: string
  enabled: boolean
  [key: string]: unknown
}

export interface Column {
  key: string
  label: string
}

export interface FormField {
  key: string
  label: string
  type: 'text' | 'number' | 'select'
  testid: string
  options?: Array<{ value: number | string; label: string }>
  defaultValue?: unknown
}
