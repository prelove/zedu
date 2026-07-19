import { httpRequest } from './http'
import type { ListData } from './types'
export interface NotificationOutbox { id:number; lessonId:number; eventType:string; recipientEmail:string; status:string; attempts:number; lastError?:string }
export const listNotificationOutbox=(token:string)=>httpRequest<ListData<NotificationOutbox>>('/notifications/outbox',{token}).then(r=>r.data)
export const retryNotification=(token:string,id:number)=>httpRequest<{id:number}>(`/notifications/outbox/${id}/retry`,{method:'POST',token}).then(r=>r.data)
export const processNotificationOutbox=(token:string)=>httpRequest<{status:string}>('/notifications/outbox/process',{method:'POST',token}).then(r=>r.data)
