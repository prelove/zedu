import { httpRequest } from './http'
import type { ListData } from './types'

// ---------- Domain types (mirror backend repository structs) ----------

export interface Student {
  id: number
  name: string
  nameLocal?: string
  email?: string
  phone?: string
  nationality?: string
  timezone: string
  status: string
  sourceChannel?: string
  note?: string
  createdAt: string
  updatedAt: string
  deletedAt?: string
}

export interface Parent {
  id: number
  studentId: number
  name: string
  email?: string
  phone?: string
  relationship?: string
  isPrimary: boolean
  note?: string
  createdAt: string
  updatedAt: string
}

export interface Teacher {
  id: number
  name: string
  nameLocal?: string
  email?: string
  phone?: string
  bio?: string
  defaultRate: number
  status: string
  note?: string
  createdAt: string
  updatedAt: string
  deletedAt?: string
}

export interface Capability {
  id: number
  teacherId: number
  domainId: number
  trackId: number
  levelId: number
  skillTagCodes?: string
  status: string
  verified: boolean
  effectiveFrom?: string
  effectiveTo?: string
  note?: string
  createdAt: string
  updatedAt: string
}

export interface Availability {
  id: number
  teacherId: number
  weekday: number
  startTime: string
  endTime: string
  effectiveFrom?: string
  effectiveTo?: string
  note?: string
  createdAt: string
  updatedAt: string
}

// ---------- Write payloads (pointer fields = optional PATCH) ----------

export interface StudentWrite {
  name?: string
  nameLocal?: string
  email?: string
  phone?: string
  nationality?: string
  timezone?: string
  status?: string
  sourceChannel?: string
  note?: string
}

export interface ParentWrite {
  name?: string
  email?: string
  phone?: string
  relationship?: string
  isPrimary?: boolean
  note?: string
}

export interface TeacherWrite {
  name?: string
  nameLocal?: string
  email?: string
  phone?: string
  bio?: string
  defaultRate?: number
  status?: string
  note?: string
}

export interface CapabilityWrite {
  domainId?: number
  trackId?: number
  levelId?: number
  skillTagCodes?: string
  status?: string
  verified?: boolean
  effectiveFrom?: string
  effectiveTo?: string
  note?: string
}

export interface AvailabilityWrite {
  weekday?: number
  startTime?: string
  endTime?: string
  effectiveFrom?: string
  effectiveTo?: string
  note?: string
}

// ---------- List query params ----------

export interface ListParams {
  page?: number
  pageSize?: number
  search?: string
  status?: string
}

function buildQuery(params: ListParams = {}): string {
  const q = new URLSearchParams()
  if (params.page) q.set('page', String(params.page))
  if (params.pageSize) q.set('pageSize', String(params.pageSize))
  if (params.search) q.set('search', params.search)
  if (params.status) q.set('status', params.status)
  const s = q.toString()
  return s ? `?${s}` : ''
}

// ---------- Student API ----------

export function listStudents(token: string, params: ListParams = {}): Promise<ListData<Student>> {
  return httpRequest<ListData<Student>>(`/students${buildQuery(params)}`, { token }).then((r) => r.data)
}

export function createStudent(token: string, body: StudentWrite): Promise<Student> {
  return httpRequest<Student>('/students', { method: 'POST', body, token }).then((r) => r.data)
}

export function getStudent(token: string, id: number): Promise<Student> {
  return httpRequest<Student>(`/students/${id}`, { token }).then((r) => r.data)
}

export function updateStudent(token: string, id: number, body: StudentWrite): Promise<Student> {
  return httpRequest<Student>(`/students/${id}`, { method: 'PATCH', body, token }).then((r) => r.data)
}

// ---------- Parent API (scoped to student) ----------

export function listParents(token: string, studentId: number, params: ListParams = {}): Promise<ListData<Parent>> {
  return httpRequest<ListData<Parent>>(`/students/${studentId}/parents${buildQuery(params)}`, { token }).then((r) => r.data)
}

export function createParent(token: string, studentId: number, body: ParentWrite): Promise<Parent> {
  return httpRequest<Parent>(`/students/${studentId}/parents`, { method: 'POST', body, token }).then((r) => r.data)
}

export function updateParent(token: string, studentId: number, parentId: number, body: ParentWrite): Promise<Parent> {
  return httpRequest<Parent>(`/students/${studentId}/parents/${parentId}`, { method: 'PATCH', body, token }).then((r) => r.data)
}

// ---------- Teacher API ----------

export function listTeachers(token: string, params: ListParams = {}): Promise<ListData<Teacher>> {
  return httpRequest<ListData<Teacher>>(`/teachers${buildQuery(params)}`, { token }).then((r) => r.data)
}

export function createTeacher(token: string, body: TeacherWrite): Promise<Teacher> {
  return httpRequest<Teacher>('/teachers', { method: 'POST', body, token }).then((r) => r.data)
}

export function getTeacher(token: string, id: number): Promise<Teacher> {
  return httpRequest<Teacher>(`/teachers/${id}`, { token }).then((r) => r.data)
}

export function updateTeacher(token: string, id: number, body: TeacherWrite): Promise<Teacher> {
  return httpRequest<Teacher>(`/teachers/${id}`, { method: 'PATCH', body, token }).then((r) => r.data)
}

// ---------- Capability API (scoped to teacher) ----------

export function listCapabilities(token: string, teacherId: number, params: ListParams = {}): Promise<ListData<Capability>> {
  return httpRequest<ListData<Capability>>(`/teachers/${teacherId}/capabilities${buildQuery(params)}`, { token }).then((r) => r.data)
}

export function createCapability(token: string, teacherId: number, body: CapabilityWrite): Promise<Capability> {
  return httpRequest<Capability>(`/teachers/${teacherId}/capabilities`, { method: 'POST', body, token }).then((r) => r.data)
}

export function updateCapability(token: string, teacherId: number, capId: number, body: CapabilityWrite): Promise<Capability> {
  return httpRequest<Capability>(`/teachers/${teacherId}/capabilities/${capId}`, { method: 'PATCH', body, token }).then((r) => r.data)
}

// ---------- Availability API (scoped to teacher) ----------

export function listAvailability(token: string, teacherId: number, params: ListParams = {}): Promise<ListData<Availability>> {
  return httpRequest<ListData<Availability>>(`/teachers/${teacherId}/availability${buildQuery(params)}`, { token }).then((r) => r.data)
}

export function createAvailability(token: string, teacherId: number, body: AvailabilityWrite): Promise<Availability> {
  return httpRequest<Availability>(`/teachers/${teacherId}/availability`, { method: 'POST', body, token }).then((r) => r.data)
}

export function updateAvailability(token: string, teacherId: number, availId: number, body: AvailabilityWrite): Promise<Availability> {
  return httpRequest<Availability>(`/teachers/${teacherId}/availability/${availId}`, { method: 'PATCH', body, token }).then((r) => r.data)
}
