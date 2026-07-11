import type { LocaleSchema } from './zh-CN'

export const jaJP: LocaleSchema = {
  app: {
    name: 'Zedu 学務管理 🎓',
    version: 'バージョン',
    versionPlaceholder: '0.1.0',
  },
  health: {
    title: 'バックエンドヘルス状態',
    loading: '確認中…',
    healthy: 'サービス正常',
    unavailable: 'サービス利用不可',
    retry: '再確認',
  },
  errors: {
    NETWORK_ERROR: 'ネットワーク接続に失敗しました。ネットワークを確認して再試行してください。',
    SERVER_ERROR: 'サーバー内部エラーが発生しました。後でもう一度お試しください。',
    TIMEOUT: 'リクエストがタイムアウトしました。後でもう一度お試しください。',
    UNKNOWN: '不明なエラーが発生しました。後でもう一度お試しください。',
  },
  common: {
    localeLabel: '言語',
    loading: '読み込み中…',
  },
}
