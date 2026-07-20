import type { CapabilityTag, CourseDomain, Level, Track } from '../../../api/course-dict'
import type { Column, DictItem, FormField } from '../components/dictionary-types'

export type DictionaryTabName = 'domains' | 'tracks' | 'levels' | 'tags'

export function buildColumns(tab: DictionaryTabName): Column[] {
  const baseColumns: Column[] = [
    { key: 'name', label: 'common.name' },
    { key: 'code', label: 'common.code' },
  ]
  if (tab === 'domains') return [...baseColumns, { key: 'type', label: 'courses.domainType' }]
  if (tab === 'tracks') return [...baseColumns, { key: 'domainName', label: 'courses.trackDomain' }]
  if (tab === 'levels') return [...baseColumns, { key: 'trackName', label: 'courses.levelTrack' }]
  return [...baseColumns, { key: 'domainName', label: 'courses.tagDomain' }]
}

export function buildCreateLabel(tab: DictionaryTabName): string {
  if (tab === 'domains') return 'courses.createDomain'
  if (tab === 'tracks') return 'courses.createTrack'
  if (tab === 'levels') return 'courses.createLevel'
  return 'courses.createTag'
}

export function buildDisplayItems(
  tab: DictionaryTabName,
  items: DictItem[],
  domains: CourseDomain[],
  tracks: Track[],
): DictItem[] {
  const domainName = (domainId: number) => domains.find((domain) => domain.id === domainId)?.name ?? String(domainId)
  const trackName = (trackId: number) => tracks.find((track) => track.id === trackId)?.name ?? String(trackId)

  return items.map((item) => {
    if (tab === 'tracks' || tab === 'tags') return { ...item, domainName: domainName(Number(item.domainId)) }
    if (tab === 'levels') return { ...item, trackName: trackName(Number(item.trackId)) }
    return item
  })
}

export function buildFormFields(
  tab: DictionaryTabName,
  domains: CourseDomain[],
  tracks: Track[],
): FormField[] {
  if (tab === 'domains') {
    return [
      { key: 'name', label: 'courses.domainName', type: 'text', testid: 'd-form-name' },
      { key: 'code', label: 'courses.domainCode', type: 'text', testid: 'd-form-code' },
      { key: 'type', label: 'courses.domainType', type: 'text', testid: 'd-form-type', defaultValue: 'LANGUAGE' },
    ]
  }
  if (tab === 'tracks') {
    return [
      { key: 'domainId', label: 'courses.trackDomain', type: 'select', testid: 't-form-domain', options: domains.map((d) => ({ value: d.id, label: d.name })) },
      { key: 'name', label: 'courses.trackName', type: 'text', testid: 't-form-name' },
      { key: 'code', label: 'courses.trackCode', type: 'text', testid: 't-form-code' },
    ]
  }
  if (tab === 'levels') {
    return [
      { key: 'trackId', label: 'courses.levelTrack', type: 'select', testid: 'l-form-track', options: tracks.map((t) => ({ value: t.id, label: t.name })) },
      { key: 'name', label: 'courses.levelName', type: 'text', testid: 'l-form-name' },
      { key: 'code', label: 'courses.levelCode', type: 'text', testid: 'l-form-code' },
    ]
  }
  return [
    { key: 'domainId', label: 'courses.tagDomain', type: 'select', testid: 'g-form-domain', options: domains.map((d) => ({ value: d.id, label: d.name })) },
    { key: 'name', label: 'courses.tagName', type: 'text', testid: 'g-form-name' },
    { key: 'code', label: 'courses.tagCode', type: 'text', testid: 'g-form-code' },
  ]
}

export function shouldRefreshReferences(tab: DictionaryTabName): boolean {
  return tab === 'domains' || tab === 'tracks'
}

export function isTabItemList(
  tab: DictionaryTabName,
  items: CourseDomain[] | Track[] | Level[] | CapabilityTag[],
): DictItem[] {
  return items as DictItem[]
}
