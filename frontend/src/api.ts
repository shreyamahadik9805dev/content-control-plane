// Thin fetch wrappers: maps snake_case JSON from the API into the app's types.
import type { AuditEntry, Podcast, SyncRun } from './types'

function apiBase(): string {
  const raw = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
  return raw.replace(/\/$/, '')
}

type PodcastDTO = {
  id: string
  source_id: string
  title: string
  author: string
  categories: string[]
  feed_url: string
  artwork_url: string
  track_count?: number
  pinned: boolean
  featured: boolean
  created_at: string
  updated_at: string
}

type SyncRunDTO = {
  id: string
  subject: string
  status: string
  records_processed: number
  started_at: string
  completed_at: string | null
}

type AuditLogDTO = {
  id: string
  action: string
  entity_id: string
  metadata?: Record<string, unknown> | null
  created_at: string
}

function mapPodcast(d: PodcastDTO): Podcast {
  return {
    id: d.id,
    sourceId: d.source_id,
    title: d.title,
    author: d.author,
    categories: d.categories ?? [],
    feedUrl: d.feed_url ?? '',
    artworkUrl: d.artwork_url ?? '',
    trackCount: d.track_count,
    pinned: d.pinned,
    featured: d.featured,
    updatedAt: d.updated_at,
  }
}

function mapSyncRun(d: SyncRunDTO): SyncRun {
  const st = d.status
  const status: SyncRun['status'] =
    st === 'failed' || st === 'running' || st === 'success' ? st : 'success'
  return {
    id: d.id,
    subject: d.subject,
    status,
    recordsProcessed: d.records_processed,
    startedAt: d.started_at,
    completedAt: d.completed_at,
  }
}

function mapAudit(d: AuditLogDTO): AuditEntry {
  let detail = ''
  try {
    const m = d.metadata
    detail =
      m && typeof m === 'object' && Object.keys(m).length > 0
        ? JSON.stringify(m)
        : ''
  } catch {
    detail = ''
  }
  return {
    id: d.id,
    action: d.action,
    entityId: d.entity_id,
    createdAt: d.created_at,
    detail,
  }
}

async function readError(res: Response): Promise<string> {
  const t = await res.text()
  try {
    const j = JSON.parse(t) as { error?: string }
    if (j.error) return j.error
  } catch {
    /* ignore */
  }
  return t || res.statusText
}

export async function fetchPodcasts(): Promise<Podcast[]> {
  const r = await fetch(`${apiBase()}/podcasts`)
  if (!r.ok) throw new Error(await readError(r))
  const data = (await r.json()) as PodcastDTO[]
  if (!Array.isArray(data)) return []
  return data.map(mapPodcast)
}

export async function syncPodcasts(query: string): Promise<SyncRun> {
  const q = new URLSearchParams({ query })
  const r = await fetch(`${apiBase()}/sync/podcasts?${q}`, { method: 'POST' })
  if (!r.ok) throw new Error(await readError(r))
  const data = (await r.json()) as SyncRunDTO
  return mapSyncRun(data)
}

export async function setPodcastPinned(
  id: string,
  pinned: boolean,
): Promise<Podcast> {
  const r = await fetch(`${apiBase()}/podcasts/${id}/pin`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pinned }),
  })
  if (!r.ok) throw new Error(await readError(r))
  const data = (await r.json()) as PodcastDTO
  return mapPodcast(data)
}

export async function fetchAuditLogs(limit = 100): Promise<AuditEntry[]> {
  const q = new URLSearchParams({ limit: String(limit) })
  const r = await fetch(`${apiBase()}/audit-logs?${q}`)
  if (!r.ok) throw new Error(await readError(r))
  const data = (await r.json()) as AuditLogDTO[]
  if (!Array.isArray(data)) return []
  return data.map(mapAudit)
}

export function getApiBaseDisplay(): string {
  return apiBase()
}
