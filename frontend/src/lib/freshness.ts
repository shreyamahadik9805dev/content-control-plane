import type { Freshness, Podcast } from '../types'

const FRESH_MS = 6 * 60 * 60 * 1000
const AGING_MS = 48 * 60 * 60 * 1000

export function freshnessForPodcast(pod: Podcast, now = Date.now()): Freshness {
  const t = new Date(pod.updatedAt).getTime()
  const age = now - t
  if (age <= FRESH_MS) return 'fresh'
  if (age <= AGING_MS) return 'aging'
  return 'stale'
}

export function formatRelative(iso: string, now = Date.now()): string {
  const t = new Date(iso).getTime()
  const sec = Math.round((now - t) / 1000)
  if (sec < 60) return 'just now'
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`
  const h = Math.floor(min / 60)
  if (h < 48) return `${h}h ago`
  const d = Math.floor(h / 24)
  return `${d}d ago`
}
