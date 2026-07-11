import type { LocaleSchema } from './zh-CN'

export const enUS: LocaleSchema = {
  app: {
    name: 'Zedu School Management 🎓',
    version: 'Version',
    versionPlaceholder: '0.1.0',
  },
  health: {
    title: 'Backend Health Status',
    loading: 'Checking…',
    healthy: 'Service Healthy',
    unavailable: 'Service Unavailable',
    retry: 'Retry',
  },
  errors: {
    NETWORK_ERROR: 'Network connection failed. Please check your network and try again.',
    SERVER_ERROR: 'Internal server error. Please try again later.',
    TIMEOUT: 'Request timed out. Please try again later.',
    UNKNOWN: 'An unknown error occurred. Please try again later.',
  },
  common: {
    localeLabel: 'Language',
    loading: 'Loading…',
  },
}
