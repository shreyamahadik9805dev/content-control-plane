import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  fetchAuditLogs,
  fetchPodcasts,
  getApiBaseDisplay,
  setPodcastPinned,
  syncPodcasts,
} from './api'
import { AuditTable } from './components/AuditTable'
import { DetailPanel } from './components/DetailPanel'
import { PodcastFilters } from './components/PodcastFilters'
import { PodcastTable } from './components/PodcastTable'
import { SidebarNav } from './components/SidebarNav'
import { SyncToolbar } from './components/SyncToolbar'
import type { AuditEntry, Podcast, SyncRun } from './types'
import './App.css'

// Single-page operator shell: catalog + detail, sync toolbar, audit tab.

type View = 'podcasts' | 'audit'

function uniqueCategories(podcasts: Podcast[]): string[] {
  const s = new Set<string>()
  for (const p of podcasts) for (const c of p.categories) s.add(c)
  return [...s].sort((a, b) => a.localeCompare(b))
}

function filterPodcasts(
  podcasts: Podcast[],
  query: string,
  category: string,
  pinnedOnly: boolean,
): Podcast[] {
  const q = query.trim().toLowerCase()
  return podcasts.filter((p) => {
    if (pinnedOnly && !p.pinned) return false
    if (category && !p.categories.includes(category)) return false
    if (!q) return true
    return (
      p.title.toLowerCase().includes(q) ||
      p.author.toLowerCase().includes(q) ||
      p.sourceId.toLowerCase().includes(q)
    )
  })
}

export default function App() {
  const [view, setView] = useState<View>('podcasts')
  const [podcasts, setPodcasts] = useState<Podcast[]>([])
  const [audit, setAudit] = useState<AuditEntry[]>([])
  const [lastSync, setLastSync] = useState<SyncRun | null>(null)
  const [syncQuery, setSyncQuery] = useState('technology')
  const [syncBusy, setSyncBusy] = useState(false)
  const [catalogLoading, setCatalogLoading] = useState(true)
  const [catalogError, setCatalogError] = useState<string | null>(null)
  const [toast, setToast] = useState<string | null>(null)

  const [query, setQuery] = useState('')
  const [filterCategory, setFilterCategory] = useState('')
  const [pinnedOnly, setPinnedOnly] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)

  const categories = useMemo(() => uniqueCategories(podcasts), [podcasts])
  const filtered = useMemo(
    () => filterPodcasts(podcasts, query, filterCategory, pinnedOnly),
    [podcasts, query, filterCategory, pinnedOnly],
  )
  const selected = useMemo(
    () => podcasts.find((p) => p.id === selectedId) ?? null,
    [podcasts, selectedId],
  )

  const showToast = useCallback((msg: string) => {
    setToast(msg)
    window.setTimeout(() => setToast(null), 4200)
  }, [])

  const refreshCatalog = useCallback(async () => {
    setCatalogError(null)
    const list = await fetchPodcasts()
    setPodcasts(list)
    return list
  }, [])

  const refreshAudit = useCallback(async () => {
    const logs = await fetchAuditLogs(200)
    setAudit(logs)
  }, [])

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setCatalogLoading(true)
      try {
        await refreshCatalog()
        await refreshAudit()
      } catch (e) {
        if (!cancelled) {
          setCatalogError(e instanceof Error ? e.message : String(e))
          showToast('Could not reach API — is the backend running?')
        }
      } finally {
        if (!cancelled) setCatalogLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [refreshCatalog, refreshAudit, showToast])

  useEffect(() => {
    if (view !== 'audit') return
    refreshAudit().catch(() => showToast('Failed to refresh audit log'))
  }, [view, refreshAudit, showToast])

  const runSync = useCallback(async () => {
    const q = syncQuery.trim()
    if (!q || syncBusy) return
    setSyncBusy(true)
    setLastSync(null)
    try {
      const run = await syncPodcasts(q)
      setLastSync(run)
      await refreshCatalog()
      await refreshAudit()
      showToast(
        `Synced “${q}” · ${run.recordsProcessed} show(s) from iTunes Search`,
      )
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e)
      showToast(`Sync failed: ${msg}`)
    } finally {
      setSyncBusy(false)
    }
  }, [syncBusy, syncQuery, refreshCatalog, refreshAudit, showToast])

  const togglePin = useCallback(
    async (id: string) => {
      const pod = podcasts.find((p) => p.id === id)
      if (!pod) return
      const nextPinned = !pod.pinned
      try {
        const updated = await setPodcastPinned(id, nextPinned)
        setPodcasts((prev) => prev.map((p) => (p.id === id ? updated : p)))
        await refreshAudit()
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e)
        showToast(`Pin update failed: ${msg}`)
      }
    },
    [podcasts, refreshAudit, showToast],
  )

  return (
    <div className="shell">
      <SidebarNav active={view} onSelect={setView} apiBase={getApiBaseDisplay()} />

      <div className="shell__main">
        <SyncToolbar
          searchQuery={syncQuery}
          onSearchQueryChange={setSyncQuery}
          onSync={() => void runSync()}
          busy={syncBusy}
          lastSync={lastSync}
        />

        {view === 'podcasts' ? (
          <section className="panel" aria-labelledby="podcasts-heading">
            <div className="panel__bar">
              <h2 id="podcasts-heading" className="panel__title">
                Normalized catalog
              </h2>
              <span className="panel__count mono">
                {filtered.length} / {podcasts.length} shown
              </span>
            </div>
            {catalogError && (
              <p className="panel__error" role="alert">
                {catalogError}
              </p>
            )}
            {catalogLoading && (
              <p className="panel__lede">Loading catalog from API…</p>
            )}
            <PodcastFilters
              query={query}
              onQueryChange={setQuery}
              category={filterCategory}
              onCategoryChange={setFilterCategory}
              pinnedOnly={pinnedOnly}
              onPinnedOnlyChange={setPinnedOnly}
              categories={categories}
            />
            <div className="split">
              <div className="split__list">
                {!catalogLoading && podcasts.length === 0 && !catalogError ? (
                  <div className="empty-state">
                    <p>
                      No shows yet. Enter a search query above and run{' '}
                      <strong>Run sync</strong> to ingest from the iTunes Search
                      API.
                    </p>
                  </div>
                ) : (
                  <PodcastTable
                    podcasts={filtered}
                    selectedId={selectedId}
                    onSelect={setSelectedId}
                  />
                )}
              </div>
              <DetailPanel
                podcast={selected}
                onPinToggle={(id) => void togglePin(id)}
                onClose={() => setSelectedId(null)}
              />
            </div>
          </section>
        ) : (
          <section className="panel" aria-labelledby="audit-heading">
            <div className="panel__bar">
              <h2 id="audit-heading" className="panel__title">
                Audit log
              </h2>
              <span className="panel__count mono">{audit.length} events</span>
            </div>
            <p className="panel__lede">
              Append-only trail from the API: sync runs and curation actions.
            </p>
            <AuditTable entries={audit} />
          </section>
        )}
      </div>

      {toast && (
        <div className="toast" role="status">
          {toast}
        </div>
      )}
    </div>
  )
}
