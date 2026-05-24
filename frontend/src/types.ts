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
  summary: string
  operatorTags: string[]
  pinned: boolean
  featured: boolean
  updatedAt: string
}

/** Structured AI suggestion row; catalog updates only after accept (see docs/PRODUCT_AI.md). */
export interface AIProposal {
  id: string
  podcast_id: string
  status: string
  kind: string
  payload: {
    summary?: string
    operator_tags?: string[]
    language?: string
    confidence?: number
  }
  model: string
  provider: string
  latency_ms?: number
  created_at: string
  resolved_at?: string | null
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
