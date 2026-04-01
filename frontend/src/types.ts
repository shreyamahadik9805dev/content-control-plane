export type Freshness = 'fresh' | 'aging' | 'stale'

export interface Podcast {
  id: string
  sourceId: string
  title: string
  author: string
  categories: string[]
  feedUrl: string
  artworkUrl: string
  trackCount?: number
  pinned: boolean
  featured: boolean
  updatedAt: string
}

export interface SyncRun {
  id: string
  subject: string
  status: 'success' | 'failed' | 'running'
  recordsProcessed: number
  startedAt: string
  completedAt: string | null
}

export interface AuditEntry {
  id: string
  action: string
  entityId: string
  createdAt: string
  detail: string
}
