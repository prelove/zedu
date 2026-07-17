/**
 * Stable business error codes from the frozen M2 contract.
 * These map 1:1 to backend `code` values in the `{ code, message, requestId }` envelope.
 * The `message` field from the backend is a STABLE KEY (e.g. "LOGIN_FAILED"), not user-facing text.
 */
export const ApiErrorCode = {
  AUTH_REQUIRED: 40101,
  LOGIN_FAILED: 40102,
  ACCOUNT_LOCKED: 40103,
  FORBIDDEN: 40301,
  NOT_FOUND: 40401,
  CONFLICT: 40901,
  INVALID_STATE: 42201,
  INTERNAL_ERROR: 50001,
  DATABASE_ERROR: 50002,
} as const

export type ApiErrorCodeValue = (typeof ApiErrorCode)[keyof typeof ApiErrorCode]

/**
 * ApiError is thrown when the backend returns a non-zero `code`.
 * `stableKey` is the backend `message` field (a stable key, e.g. "LOGIN_FAILED"),
 * used to look up the localized user-facing message.
 * `requestId` is exposed for support but never rendered as debug info.
 */
export class ApiError extends Error {
  readonly code: number
  readonly stableKey: string
  readonly requestId: string
  readonly httpStatus: number

  constructor(code: number, stableKey: string, requestId: string, httpStatus: number) {
    super(`API error ${code}: ${stableKey}`)
    this.name = 'ApiError'
    this.code = code
    this.stableKey = stableKey
    this.requestId = requestId
    this.httpStatus = httpStatus
  }
}

/**
 * NetworkError is thrown when fetch fails before reaching the backend
 * (DNS, connection refused, CORS preflight failure, etc.).
 */
export class NetworkError extends Error {
  constructor(message = 'NETWORK_ERROR') {
    super(message)
    this.name = 'NetworkError'
  }
}

/** Success envelope: { code: 0, data }. */
interface SuccessEnvelope {
  code: 0
  data: unknown
}

/** Error envelope: { code, message, requestId }. */
interface ErrorEnvelope {
  code: number
  message: string
  requestId: string
}

function isSuccessEnvelope(value: unknown): value is SuccessEnvelope {
  return (
    typeof value === 'object' &&
    value !== null &&
    'code' in value &&
    (value as { code: unknown }).code === 0 &&
    'data' in value
  )
}

function isErrorEnvelope(value: unknown): value is ErrorEnvelope {
  return (
    typeof value === 'object' &&
    value !== null &&
    'code' in value &&
    typeof (value as { code: unknown }).code === 'number' &&
    'message' in value &&
    typeof (value as { message: unknown }).message === 'string' &&
    'requestId' in value &&
    typeof (value as { requestId: unknown }).requestId === 'string'
  )
}

export interface HttpRequestOptions {
  method?: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'
  body?: unknown
  /** Bearer access token; never persisted to storage. */
  token?: string | null
  /** When true, a 40101 response is surfaced as ApiError instead of triggering refresh. */
  skipAuthRetry?: boolean
  /** Custom headers. */
  headers?: Record<string, string>
  /** Abort signal. */
  signal?: AbortSignal
}

export interface HttpResponse<T> {
  data: T
  status: number
}

/**
 * Low-level HTTP client that parses the unified envelope.
 * On a non-zero `code`, throws ApiError with the stable key.
 * On network failure, throws NetworkError.
 *
 * This function does NOT implement refresh/retry — that is the auth store's
 * responsibility, because only the auth store knows whether a refresh is
 * already in flight.
 */
export async function httpRequest<T = unknown>(
  url: string,
  options: HttpRequestOptions = {},
): Promise<HttpResponse<T>> {
  const {
    method = 'GET',
    body,
    token,
    headers = {},
    signal,
  } = options

  const finalHeaders: Record<string, string> = { ...headers }
  if (body !== undefined && !finalHeaders['Content-Type']) {
    finalHeaders['Content-Type'] = 'application/json; charset=utf-8'
  }
  if (token) {
    finalHeaders['Authorization'] = `Bearer ${token}`
  }

  let response: Response
  try {
    response = await fetch(url, {
      method,
      headers: finalHeaders,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      credentials: 'include', // send/receive refresh cookie
      signal,
    })
  } catch (err) {
    if (err instanceof Error && err.name === 'AbortError') {
      throw err
    }
    throw new NetworkError()
  }

  let payload: unknown
  try {
    payload = await response.json()
  } catch {
    // Non-JSON response (e.g. 502 from proxy) — classify by status.
    if (response.status >= 500) {
      throw new ApiError(ApiErrorCode.INTERNAL_ERROR, 'INTERNAL_ERROR', 'unknown', response.status)
    }
    throw new NetworkError()
  }

  if (isSuccessEnvelope(payload)) {
    return { data: payload.data as T, status: response.status }
  }

  if (isErrorEnvelope(payload)) {
    throw new ApiError(payload.code, payload.message, payload.requestId, response.status)
  }

  // Malformed envelope — treat as internal error, never expose raw body.
  throw new ApiError(ApiErrorCode.INTERNAL_ERROR, 'INTERNAL_ERROR', 'unknown', response.status)
}
