import { httpRequest } from './http'
export interface Dashboard {
  todayLessons: number
  pendingLessonConfirmations: number
  renewalNeededStudents: number
  teacherPayableAggregate: number
  failedNotifications: number
}
export const getDashboard=(token:string)=>httpRequest<Dashboard>('/dashboard',{token}).then(r=>r.data)
export const createBackup=(token:string)=>httpRequest<{file:string}>('/system/backups',{method:'POST',token}).then(r=>r.data)
