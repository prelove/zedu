import { httpRequest } from './http'
import type { ListData } from './types'

export interface Lesson {
  id: number
  lessonNo: string
  enrollmentId: number
  assignmentId: number
  teacherId: number
  studentId: number
  scheduledStartAt: string
  scheduledEndAt: string
  durationMin: number
  timezone: string
  meetingType: string
  meetingLink?: string
  lessonTopic?: string
  note?: string
  status: 'SCHEDULED' | 'COMPLETED' | 'CANCELLED'
  cancelReason?: string
}

export interface LessonWrite {
  enrollmentId: number
  assignmentId: number
  startAt: string
  durationMin: number
  timezone: string
  meetingType: string
  meetingLink?: string
  lessonTopic?: string
  note?: string
}

export type LessonUpdate = Omit<LessonWrite, 'enrollmentId' | 'assignmentId'>

export interface LessonFilters {
  studentId?: number
  teacherId?: number
  status?: Lesson['status']
  from?: string
  to?: string
  page?: number
  pageSize?: number
}

function query(filters: LessonFilters): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== '') params.set(key, String(value))
  }
  return params.size ? `?${params.toString()}` : ''
}

export function listLessons(token: string, filters: LessonFilters = {}): Promise<ListData<Lesson>> {
  return httpRequest<ListData<Lesson>>(`/lessons${query(filters)}`, { token }).then((response) => response.data)
}

export function getLesson(token: string, id: number): Promise<Lesson> {
  return httpRequest<Lesson>(`/lessons/${id}`, { token }).then((response) => response.data)
}

export function createLesson(token: string, body: LessonWrite): Promise<Lesson> {
  return httpRequest<Lesson>('/lessons', { method: 'POST', body, token }).then((response) => response.data)
}

export function updateLesson(token: string, id: number, body: LessonUpdate): Promise<Lesson> {
  return httpRequest<Lesson>(`/lessons/${id}`, { method: 'PATCH', body, token }).then((response) => response.data)
}

export function cancelLesson(token: string, id: number, reason: string): Promise<Lesson> {
  return httpRequest<Lesson>(`/lessons/${id}/cancel`, { method: 'POST', body: { reason }, token }).then((response) => response.data)
}

export interface LessonConfirmation { outcomeType: string; lessonDeducted: string; chargeAmount: number; teacherPayAmount: number; actualDurationMin?: number; note?: string }
export interface AttendanceOutcome { code:string; name:string; suggestedLessonDeducted:string; suggestedChargeRatio:string; suggestedTeacherPayRatio:string }
export function listAttendanceOutcomes(token:string): Promise<AttendanceOutcome[]> { return httpRequest<AttendanceOutcome[]>('/system/attendance-outcomes',{token}).then((response)=>response.data) }
export function confirmLesson(token: string, id: number, body: LessonConfirmation): Promise<{ lessonId: number }> {
  return httpRequest<{ lessonId: number }>(`/lessons/${id}/confirm`, { method: 'POST', body, token }).then((response) => response.data)
}
