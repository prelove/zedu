import { httpRequest } from './http'

/** Approved template names, frozen by the M2 contract. */
export type OnboardingTemplate = 'japanese' | 'k12' | 'blank'

/** POST /onboarding/initialize and /onboarding/reset response data: { template, reused }. */
export interface OnboardingResultData {
  template: string
  reused: boolean
}

/**
 * POST /onboarding/initialize — Owner-only. Applies the selected template
 * exactly once; a repeat request returns the existing result with reused=true.
 */
export function initialize(
  template: OnboardingTemplate,
  token: string,
): Promise<OnboardingResultData> {
  return httpRequest<OnboardingResultData>('/onboarding/initialize', {
    method: 'POST',
    body: { template },
    token,
  }).then((res) => res.data)
}

/**
 * POST /onboarding/reset — Owner-only. Replaces template data only when
 * no protected business record exists; otherwise returns 42201/RESET_NOT_ALLOWED.
 */
export function reset(
  template: OnboardingTemplate,
  token: string,
): Promise<OnboardingResultData> {
  return httpRequest<OnboardingResultData>('/onboarding/reset', {
    method: 'POST',
    body: { template },
    token,
  }).then((res) => res.data)
}
