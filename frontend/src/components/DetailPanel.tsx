import type { Podcast } from '../types'
import { formatRelative, freshnessForPodcast } from '../lib/freshness'
import { FreshnessBadge } from './FreshnessBadge'

type Props = {
  podcast: Podcast | null
  onPinToggle: (id: string) => void
  onClose: () => void
}

export function DetailPanel({ podcast, onPinToggle, onClose }: Props) {
  if (!podcast) {
    return (
      <aside className="detail-panel detail-panel--empty" aria-label="Podcast detail">
        <p className="detail-panel__placeholder">
          Select a show to inspect feed URL, iTunes collection id, categories, and
          curation flags.
        </p>
      </aside>
    )
  }

  const fresh = freshnessForPodcast(podcast)

  return (
    <aside className="detail-panel" aria-label="Podcast detail">
      <div className="detail-panel__head">
        <div className="detail-panel__title-row">
          {podcast.artworkUrl ? (
            <img
              className="detail-panel__art"
              src={podcast.artworkUrl}
              alt=""
              width={56}
              height={56}
            />
          ) : null}
          <h2 className="detail-panel__title">{podcast.title}</h2>
        </div>
        <button type="button" className="btn btn--ghost" onClick={onClose}>
          Close
        </button>
      </div>
      <dl className="detail-dl">
        <div>
          <dt>Publisher</dt>
          <dd>{podcast.author}</dd>
        </div>
        <div>
          <dt>Episode count (iTunes)</dt>
          <dd className="mono">
            {podcast.trackCount != null ? podcast.trackCount : '—'}
          </dd>
        </div>
        <div>
          <dt>Collection ID</dt>
          <dd className="mono">{podcast.sourceId}</dd>
        </div>
        <div>
          <dt>Internal ID</dt>
          <dd className="mono break-all">{podcast.id}</dd>
        </div>
        <div>
          <dt>Feed URL</dt>
          <dd className="mono break-all">
            {podcast.feedUrl ? (
              <a href={podcast.feedUrl} target="_blank" rel="noreferrer">
                {podcast.feedUrl}
              </a>
            ) : (
              '—'
            )}
          </dd>
        </div>
        <div>
          <dt>Last updated</dt>
          <dd>
            {formatRelative(podcast.updatedAt)}{' '}
            <FreshnessBadge value={fresh} />
          </dd>
        </div>
        <div>
          <dt>Categories</dt>
          <dd>
            <div className="chips">
              {podcast.categories.map((s) => (
                <span key={s} className="chip">
                  {s}
                </span>
              ))}
            </div>
          </dd>
        </div>
      </dl>
      <div className="detail-actions">
        <button
          type="button"
          className={podcast.pinned ? 'btn btn--warn' : 'btn btn--secondary'}
          onClick={() => onPinToggle(podcast.id)}
        >
          {podcast.pinned ? 'Unpin show' : 'Pin show'}
        </button>
        <p className="detail-hint">
          Pinning calls <span className="mono">POST /podcasts/:id/pin</span> with{' '}
          <span className="mono">{'{"pinned":true|false}'}</span> and is stored in
          Postgres with an audit entry.
        </p>
      </div>
    </aside>
  )
}
