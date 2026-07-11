import { describe, it, expect } from 'vitest'
import { classifyError } from '../src/utils/errors'

describe('classifyError', () => {
  it('classifies TypeError with fetch as NETWORK_ERROR', () => {
    const error = new TypeError('Failed to fetch')
    expect(classifyError(error)).toBe('NETWORK_ERROR')
  })

  it('classifies TimeoutError as TIMEOUT', () => {
    const error = new Error('Request timed out')
    error.name = 'TimeoutError'
    expect(classifyError(error)).toBe('TIMEOUT')
  })

  it('classifies error with timeout message as TIMEOUT', () => {
    const error = new Error('operation timeout exceeded')
    expect(classifyError(error)).toBe('TIMEOUT')
  })

  it('classifies 500 error as SERVER_ERROR', () => {
    const error = new Error('HTTP 500 Internal Server Error')
    expect(classifyError(error)).toBe('SERVER_ERROR')
  })

  it('classifies 503 error as SERVER_ERROR', () => {
    const error = new Error('HTTP 503 Service Unavailable')
    expect(classifyError(error)).toBe('SERVER_ERROR')
  })

  it('classifies 502 error as SERVER_ERROR', () => {
    const error = new Error('HTTP 502 Bad Gateway')
    expect(classifyError(error)).toBe('SERVER_ERROR')
  })

  it('classifies unknown error as UNKNOWN', () => {
    const error = new Error('Something unexpected happened')
    expect(classifyError(error)).toBe('UNKNOWN')
  })

  it('classifies non-Error thrown value as UNKNOWN', () => {
    expect(classifyError('string error')).toBe('UNKNOWN')
    expect(classifyError(42)).toBe('UNKNOWN')
    expect(classifyError(null)).toBe('UNKNOWN')
    expect(classifyError(undefined)).toBe('UNKNOWN')
  })
})
