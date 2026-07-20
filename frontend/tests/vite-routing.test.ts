import { describe, expect, it } from 'vitest'
import { isSpaNavigationRequest } from '../vite.config'

describe('isSpaNavigationRequest', () => {
  it('keeps a browser deep link to a protected page inside the SPA', () => {
    expect(isSpaNavigationRequest('/students/42', 'text/html,application/xhtml+xml')).toBe(true)
  })

  it('does not divert an API fetch for the same resource path', () => {
    expect(isSpaNavigationRequest('/students/42', '*/*')).toBe(false)
  })

  it('does not divert a JSON API request even when the path is a frontend route', () => {
    expect(isSpaNavigationRequest('/lessons', 'application/json')).toBe(false)
  })

  it('keeps unknown HTML paths outside the known SPA routes', () => {
    expect(isSpaNavigationRequest('/not-a-route', 'text/html')).toBe(false)
  })
})
