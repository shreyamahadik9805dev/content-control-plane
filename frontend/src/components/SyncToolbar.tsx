import type { SyncRun } from '../types'
import { formatRelative } from '../lib/freshness'

type Props = {
  searchQuery: string
  onSearchQueryChange: (v: string) => void
  onSync: () => void
  busy: boolean
  lastSync: SyncRun | null
}

export function SyncToolbar({
  searchQuery,
  onSearchQueryChange,
  onSync,
  busy,
  lastSync,
}: Props) {
  const completed =
    lastSync?.completedAt != null
      ? formatRelative(lastSync.completedAt)
      : '—'

  return (
    <header className="sync-toolbar">
      <div className="sync-toolbar__brand">
        <span className="sync-toolbar__dot" aria-hidden />
        <div>
          <h1 className="sync-toolbar__title">Content Control Plane</h1>
          <p className="sync-toolbar__subtitle">
            Podcasts · iTunes search · normalize · curate
          </p>
        </div>
      </div>

      <div className="sync-toolbar__actions">
        <label className="field">
          <span className="field__label">Sync search query</span>
          <input
            className="input"
            value={searchQuery}
            onChange={(e) => onSearchQueryChange(e.target.value)}
            placeholder="e.g. technology, design, news"
            disabled={busy}
          />
        </label>
        <button
          type="button"
          className="btn btn--primary"
          onClick={onSync}
          disabled={busy || !searchQuery.trim()}
        >
          {busy ? 'Syncing…' : 'Run sync'}
        </button>
      </div>

      <div className="sync-toolbar__meta" role="status">
        <div className="meta-chip">
          <span className="meta-chip__k">Last run</span>
          <span className="meta-chip__v">{completed}</span>
        </div>
        {lastSync && (
          <div className="meta-chip">
            <span className="meta-chip__k">Shows</span>
            <span className="meta-chip__v mono">{lastSync.recordsProcessed}</span>
          </div>
        )}
        {lastSync && (
          <div className="meta-chip">
            <span className="meta-chip__k">Query</span>
            <span className="meta-chip__v mono">{lastSync.subject}</span>
          </div>
        )}
      </div>
    </header>
  )
}
