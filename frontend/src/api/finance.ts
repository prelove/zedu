import { ApiError, ApiErrorCode, NetworkError, httpRequest } from './http'
import type { ListData } from './types'

export interface BaseCurrency {
  currency: 'JPY' | 'CNY' | 'USD'
  locked: boolean
}

export interface PaymentMethod {
  code: string
  name: string
  sortOrder: number
  enabled: boolean
}

export interface PaymentMethodCreate {
  code: string
  name: string
  sortOrder: number
  enabled: boolean
}

export interface PaymentMethodUpdate {
  name: string
  sortOrder: number
  enabled: boolean
}

export interface PaymentSummary {
  id: number
  paymentNo: string
  studentId: number
  enrollmentId: number
  amountBase: number
  lessonsAdded: number
  status: string
	paymentMethodCode: string
	paymentMethodName: string
}

export interface PaymentDetail extends PaymentSummary {
  originalAmount: string
  originalCurrency: string
  fxRateToBase: string
  paymentMethodCode: string
  paidAt: string
  note: string
}

export interface PaymentWrite {
  paymentNo: string
  studentId: number
  enrollmentId: number
  originalAmount: string
  originalCurrency: BaseCurrency['currency']
  fxRateToBase: string
  lessonsAdded: number
  paymentMethodCode: string
  paidAt: string
  note?: string
}

export interface PaymentVoidWrite {
  reason: string
}

export interface StudentLedgerEntry {
  id: number
  enrollmentId: number
  bizType: string
  amountDelta: number
  lessonDelta: number
  balanceAfter: number
  lessonBalanceAfter: number
  relatedPaymentId?: number
  note: string
  createdAt: string
}

export interface PaymentAttachment {
  id: number
  paymentId: number
  fileName: string
  fileType: string
  fileSize: number
  uploadedBy: number
  uploadedAt: string
}

export interface DownloadedAttachment {
  blob: Blob
  fileName: string
  fileType: string
}

interface ErrorEnvelope {
  code: number
  message: string
  requestId: string
}

interface SuccessEnvelope<T> {
  code: 0
  data: T
}

function isErrorEnvelope(value: unknown): value is ErrorEnvelope {
  return (
    typeof value === 'object' &&
    value !== null &&
    'code' in value &&
    'message' in value &&
    'requestId' in value
  )
}

function isSuccessEnvelope<T>(value: unknown): value is SuccessEnvelope<T> {
  return (
    typeof value === 'object' &&
    value !== null &&
    'code' in value &&
    (value as { code: unknown }).code === 0 &&
    'data' in value
  )
}

function buildQuery(params: Record<string, string | number | undefined>): string {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') {
      query.set(key, String(value))
    }
  }
  const search = query.toString()
  return search ? `?${search}` : ''
}

function createUnknownApiError(status: number): ApiError {
  return new ApiError(ApiErrorCode.INTERNAL_ERROR, 'INTERNAL_ERROR', 'unknown', status)
}

async function parseEnvelopeData<T>(response: Response): Promise<T> {
  let payload: unknown
  try {
    payload = await response.json()
  } catch {
    throw createUnknownApiError(response.status)
  }

  if (isSuccessEnvelope<T>(payload)) {
    return payload.data
  }

  if (isErrorEnvelope(payload)) {
    throw new ApiError(payload.code, payload.message, payload.requestId, response.status)
  }

  throw createUnknownApiError(response.status)
}

async function fetchWithToken(url: string, token: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers)
  headers.set('Authorization', `Bearer ${token}`)

  try {
    return await fetch(url, {
      ...init,
      headers,
      credentials: 'include',
    })
  } catch (error) {
    if (error instanceof Error && error.name === 'AbortError') {
      throw error
    }
    throw new NetworkError()
  }
}

function parseFileName(contentDisposition: string | null, fallback: string): string {
  if (!contentDisposition) {
    return fallback
  }
  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match?.[1]) {
    return decodeURIComponent(utf8Match[1])
  }
  const quotedMatch = contentDisposition.match(/filename="([^"]+)"/i)
  if (quotedMatch?.[1]) {
    return quotedMatch[1]
  }
  return fallback
}

export function getBaseCurrency(token: string): Promise<BaseCurrency> {
  return httpRequest<BaseCurrency>('/system/base-currency', { token }).then((response) => response.data)
}

export function updateBaseCurrency(
  token: string,
  currency: BaseCurrency['currency'],
): Promise<BaseCurrency> {
  return httpRequest<BaseCurrency>('/system/base-currency', {
    method: 'PUT',
    token,
    body: { currency },
  }).then((response) => response.data)
}

export function listPaymentMethods(token: string): Promise<PaymentMethod[]> {
  return httpRequest<PaymentMethod[]>('/system/payment-methods', { token }).then((response) => response.data)
}

export function createPaymentMethod(token: string, body: PaymentMethodCreate): Promise<PaymentMethod> {
  return httpRequest<PaymentMethod>('/system/payment-methods', {
    method: 'POST',
    token,
    body,
  }).then((response) => response.data)
}

export function updatePaymentMethod(
  token: string,
  code: string,
  body: PaymentMethodUpdate,
): Promise<PaymentMethod> {
  return httpRequest<PaymentMethod>(`/system/payment-methods/${encodeURIComponent(code)}`, {
    method: 'PATCH',
    token,
    body,
  }).then((response) => response.data)
}

export function createPayment(token: string, body: PaymentWrite): Promise<PaymentSummary> {
  return httpRequest<PaymentSummary>('/finance/payments', {
    method: 'POST',
    token,
    body,
  }).then((response) => response.data)
}

export function listPayments(
  token: string,
  params: { page?: number; pageSize?: number; paymentNo?: string; studentId?: number; enrollmentId?: number; status?: string } = {},
): Promise<ListData<PaymentSummary>> {
  return httpRequest<ListData<PaymentSummary>>(`/finance/payments${buildQuery(params)}`, {
    token,
  }).then((response) => response.data)
}

export function getPayment(token: string, id: number): Promise<PaymentDetail> {
  return httpRequest<PaymentDetail>(`/finance/payments/${id}`, { token }).then((response) => response.data)
}

export function voidPayment(token: string, id: number, body: PaymentVoidWrite): Promise<PaymentDetail> {
  return httpRequest<PaymentDetail>(`/finance/payments/${id}/void`, {
    method: 'POST',
    token,
    body,
  }).then((response) => response.data)
}

export function listStudentLedger(
  token: string,
  studentId: number,
  params: { page?: number; pageSize?: number } = {},
): Promise<ListData<StudentLedgerEntry>> {
  return httpRequest<ListData<StudentLedgerEntry>>(
    `/finance/ledger/student/${studentId}${buildQuery(params)}`,
    { token },
  ).then((response) => response.data)
}

export async function uploadPaymentAttachment(
  token: string,
  paymentId: number,
  file: File,
): Promise<PaymentAttachment> {
  const formData = new FormData()
  formData.set('file', file)
  const response = await fetchWithToken(`/finance/payments/${paymentId}/attachments`, token, {
    method: 'POST',
    body: formData,
  })
  return parseEnvelopeData<PaymentAttachment>(response)
}

export function listPaymentAttachments(
  token: string,
  paymentId: number,
  params: { page?: number; pageSize?: number } = {},
): Promise<ListData<PaymentAttachment>> {
  return httpRequest<ListData<PaymentAttachment>>(
    `/finance/payments/${paymentId}/attachments${buildQuery(params)}`,
    { token },
  ).then((response) => response.data)
}

export async function downloadPaymentAttachment(
  token: string,
  paymentId: number,
  attachmentId: number,
): Promise<DownloadedAttachment> {
  const response = await fetchWithToken(
    `/finance/payments/${paymentId}/attachments/${attachmentId}/content`,
    token,
  )

  if (!response.ok) {
    await parseEnvelopeData<never>(response)
  }

  return {
    blob: await response.blob(),
    fileName: parseFileName(response.headers.get('Content-Disposition'), `attachment-${attachmentId}`),
    fileType: response.headers.get('Content-Type') ?? 'application/octet-stream',
  }
}
