import { httpRequest } from './http'
import type { ListData } from './types'

// Re-export course dictionary types and functions for backward compatibility.
export * from './course-dict'

// ---------- Enrollment & Assignment types ----------

export interface Enrollment {
  id: number
  studentId: number
  domainId: number
  trackId: number
  currentLevelId?: number
  targetLevelId?: number
  enrollmentType: string
  status: string
  startedAt?: string
  endedAt?: string
  note?: string
  createdAt: string
  updatedAt: string
}

export interface Assignment {
  id: number
  enrollmentId: number
  studentId: number
  teacherId: number
  roleType: string
  rateAmount?: number
  status: string
  startDate: string
  endDate?: string
  reason?: string
  note?: string
  createdAt: string
  updatedAt: string
}

export interface EnrollmentWrite {
  domainId?: number
  trackId?: number
  currentLevelId?: number
  targetLevelId?: number
  enrollmentType?: string
  status?: string
  startedAt?: string
  note?: string
}

export interface AssignmentWrite {
  teacherId?: number
  roleType?: string
  reason?: string
  note?: string
}

export interface EndAssignmentWrite {
  reason?: string
}

// ---------- Enrollment API ----------

export function listEnrollments(token: string, studentId: number, params: { page?: number; pageSize?: number } = {}): Promise<ListData<Enrollment>> {
  const q = new URLSearchParams()
  if (params.page) q.set('page', String(params.page))
  if (params.pageSize) q.set('pageSize', String(params.pageSize))
  const s = q.toString()
  return httpRequest<ListData<Enrollment>>(`/students/${studentId}/enrollments${s ? `?${s}` : ''}`, { token }).then((r) => r.data)
}

export function createEnrollment(token: string, studentId: number, body: EnrollmentWrite): Promise<Enrollment> {
  return httpRequest<Enrollment>(`/students/${studentId}/enrollments`, { method: 'POST', body, token }).then((r) => r.data)
}

export function getEnrollment(token: string, id: number): Promise<Enrollment> {
  return httpRequest<Enrollment>(`/enrollments/${id}`, { token }).then((r) => r.data)
}

export function updateEnrollment(token: string, id: number, body: EnrollmentWrite): Promise<Enrollment> {
  return httpRequest<Enrollment>(`/enrollments/${id}`, { method: 'PATCH', body, token }).then((r) => r.data)
}

// ---------- Assignment API ----------

export function listAssignments(token: string, enrollmentId: number, params: { page?: number; pageSize?: number } = {}): Promise<ListData<Assignment>> {
  const q = new URLSearchParams()
  if (params.page) q.set('page', String(params.page))
  if (params.pageSize) q.set('pageSize', String(params.pageSize))
  const s = q.toString()
  return httpRequest<ListData<Assignment>>(`/enrollments/${enrollmentId}/assignments${s ? `?${s}` : ''}`, { token }).then((r) => r.data)
}

export function createAssignment(token: string, enrollmentId: number, body: AssignmentWrite): Promise<Assignment> {
  return httpRequest<Assignment>(`/enrollments/${enrollmentId}/assignments`, { method: 'POST', body, token }).then((r) => r.data)
}

export function endAssignment(token: string, assignmentId: number, body: EndAssignmentWrite = {}): Promise<Assignment> {
  return httpRequest<Assignment>(`/assignments/${assignmentId}/end`, { method: 'POST', body, token }).then((r) => r.data)
}
