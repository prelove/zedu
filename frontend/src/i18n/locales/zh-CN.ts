export const zhCN = {
  app: {
    name: 'Zedu 教务管理 🎓',
    version: '版本',
    versionPlaceholder: '0.1.0',
  },
  health: {
    title: '后端健康状态',
    loading: '检查中…',
    healthy: '服务正常',
    unavailable: '服务不可用',
    retry: '重新检查',
  },
  errors: {
    NETWORK_ERROR: '网络连接失败，请检查网络后重试。',
    SERVER_ERROR: '服务器内部错误，请稍后再试。',
    TIMEOUT: '请求超时，请稍后再试。',
    UNKNOWN: '发生未知错误，请稍后再试。',
  },
  common: {
    localeLabel: '语言',
    loading: '加载中…',
  },
} as const

export type LocaleSchema = typeof zhCN
