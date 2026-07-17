import { httpRequest } from './http'
import type { ListData } from './types'

// ---------- Domain types ----------

export interface CourseDomain {
  id: number
  name: string
  code: string
  type: string
  sortOrder: number
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface Track {
  id: number
  domainId: number
  name: string
  code: string
  sortOrder: number
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface Level {
  id: number
  trackId: number
  name: string
  code: string
  sortOrder: number
  minAge?: number
  maxAge?: number
  minLessonHours?: number
  recommendedLessonHours?: number
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface CapabilityTag {
  id: number
  domainId: number
  name: string
  code: string
  sortOrder: number
  enabled: boolean
  createdAt: string
  updatedAt: string
}

// ---------- Write payloads ----------

export interface DomainWrite {
  name?: string
  code?: string
  type?: string
  sortOrder?: number
  enabled?: boolean
}

export interface TrackWrite {
  domainId?: number
  name?: string
  code?: string
  sortOrder?: number
  enabled?: boolean
}

export interface LevelWrite {
  trackId?: number
  name?: string
  code?: string
  sortOrder?: number
  minAge?: number
  maxAge?: number
  minLessonHours?: number
  recommendedLessonHours?: number
  enabled?: boolean
}

export interface TagWrite {
  domainId?: number
  name?: string
  code?: string
  sortOrder?: number
  enabled?: boolean
}

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

// ---------- List query params ----------

export interface CourseListParams {
  page?: number
  pageSize?: number
  search?: string
  domainId?: number
  trackId?: number
}

function buildQuery(params: CourseListParams = {}): string {
  const q = new URLSearchParams()
  if (params.page) q.set('page', String(params.page))
  if (params.pageSize) q.set('pageSize', String(params.pageSize))
  if (params.search) q.set('search', params.search)
  if (params.domainId) q.set('domainId', String(params.domainId))
  if (params.trackId) q.set('trackId', String(params.trackId))
  const s = q.toString()
  return s ? `?${s}` : ''
}

// ---------- Domain API ----------

export function listDomains(token: string, params: CourseListParams = {}): Promise<ListData<CourseDomain>> {
  return httpRequest<ListData<CourseDomain>>(`/course-domains${buildQuery(params)}`, { token }).then((r) => r.data)
}

export function createDomain(token: string, body: DomainWrite): Promise<CourseDomain> {
  return httpRequest<CourseDomain>('/course-domains', { method: 'POST', body, token }).then((r) => r.data)
}

export function updateDomain(token: string, id: number, body: DomainWrite): Promise<CourseDomain> {
  return httpRequest<CourseDomain>(`/course-domains/${id}`, { method: 'PATCH', body, token }).then((r) => r.data)
}

// ---------- Track API ----------

export function listTracks(token: string, params: CourseListParams = {}): Promise<ListData<Track>> {
  return httpRequest<ListData<Track>>(`/tracks${buildQuery(params)}`, { token }).then((r) => r.data)
}

export function createTrack(token: string, body: TrackWrite): Promise<Track> {
  return httpRequest<Track>('/tracks', { method: 'POST', body, token }).then((r) => r.data)
}

export function updateTrack(token: string, id: number, body: TrackWrite): Promise<Track> {
  return httpRequest<Track>(`/tracks/${id}`, { method: 'PATCH', body, token }).then((r) => r.data)
}

// ---------- Level API ----------

export function listLevels(token: string, params: CourseListParams = {}): Promise<ListData<Level>> {
  return httpRequest<ListData<Level>>(`/levels${buildQuery(params)}`, { token }).then((r) => r.data)
}

export function createLevel(token: string, body: LevelWrite): Promise<Level> {
  return httpRequest<Level>('/levels', { method: 'POST', body, token }).then((r) => r.data)
}

export function updateLevel(token: string, id: number, body: LevelWrite): Promise<Level> {
  return httpRequest<Level>(`/levels/${id}`, { method: 'PATCH', body, token }).then((r) => r.data)
}

// ---------- Tag API ----------

export function listTags(token: string, params: CourseListParams = {}): Promise<ListData<CapabilityTag>> {
  return httpRequest<ListData<CapabilityTag>>(`/capability-tags${buildQuery(params)}`, { token }).then((r) => r.data)
}

export function createTag(token: string, body: TagWrite): Promise<CapabilityTag> {
  return httpRequest<CapabilityTag>('/capability-tags', { method: 'POST', body, token }).then((r) => r.data)
}

export function updateTag(token: string, id: number, body: TagWrite): Promise<CapabilityTag> {
  return httpRequest<CapabilityTag>(`/capability-tags/${id}`, { method: 'PATCH', body, token }).then((r) => r.data)
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
