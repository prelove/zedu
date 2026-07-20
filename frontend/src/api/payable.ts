import { httpRequest } from './http'
import type { ListData } from './types'

export interface TeacherPayableSummary {
  teacherId: number
  teacherName: string
  unpaidAmount: number
  lessonCount: number
}

export interface TeacherPayableEntry {
  id: number
  lessonId: number
  lessonNo: string
  amountDelta: number
  balanceAfter: number
  note?: string
  createdAt: string
}

export function listTeacherPayable(token: string, params: { page?: number; pageSize?: number } = {}): Promise<ListData<TeacherPayableSummary>> {
  const q = new URLSearchParams()
  if (params.page) q.set('page', String(params.page))
  if (params.pageSize) q.set('pageSize', String(params.pageSize))
  const s = q.toString()
  return httpRequest<ListData<TeacherPayableSummary>>(`/teachers/payable${s ? `?${s}` : ''}`, { token }).then((r) => r.data)
}

export function getTeacherPayableDetail(token: string, teacherId: number, params: { page?: number; pageSize?: number } = {}): Promise<ListData<TeacherPayableEntry>> {
  const q = new URLSearchParams()
  if (params.page) q.set('page', String(params.page))
  if (params.pageSize) q.set('pageSize', String(params.pageSize))
  const s = q.toString()
  return httpRequest<ListData<TeacherPayableEntry>>(`/teachers/${teacherId}/payable${s ? `?${s}` : ''}`, { token }).then((r) => r.data)
}
