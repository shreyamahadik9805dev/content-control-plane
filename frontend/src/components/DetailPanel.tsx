import type { AIProposal, Podcast } from '../types'
import { formatRelative, freshnessForPodcast } from '../lib/freshness'
import { FreshnessBadge } from './FreshnessBadge'

type Props = {
  podcast: Podcast | null
  onPinToggle: (id: string) => void
  onClose: () => void
  pendingProposals: AIProposal[]
  proposalBusy: boolean
  onGenerateSuggestion: () => void
  onAcceptProposal: (proposalId: string) => void
  onRejectProposal: (proposalId: string) => void
}

export function DetailPanel({
  podcast,
  onPinToggle,
  onClose,
  pendingProposals,
  proposalBusy,
  onGenerateSuggestion,
  onAcceptProposal,
  onRejectProposal,
}: Props) {
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
  const pending = pendingProposals[0]

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
        {podcast.summary ? (
          <div>
            <dt>Operator summary</dt>
            <dd>{podcast.summary}</dd>
          </div>
        ) : null}
        {podcast.operatorTags.length > 0 ? (
          <div>
            <dt>Operator tags</dt>
            <dd>
              <div className="chips">
                {podcast.operatorTags.map((s) => (
                  <span key={s} className="chip chip--accent">
                    {s}
                  </span>
                ))}
              </div>
            </dd>
          </div>
        ) : null}
      </dl>

      <div className="detail-ai">
        <h3 className="detail-ai__title">AI metadata (human-in-the-loop)</h3>
        <p className="detail-ai__lede">
          Generates a <strong>pending</strong> proposal only. Accept applies{' '}
          <span className="mono">summary</span> +{' '}
          <span className="mono">operator_tags</span> to this row and writes{' '}
          <span className="mono">audit</span> + <span className="mono">proposal</span>{' '}
          records.
        </p>
        <button
          type="button"
          className="btn btn--secondary"
          disabled={proposalBusy}
          onClick={onGenerateSuggestion}
        >
          {proposalBusy ? 'Working…' : 'Generate suggestion'}
        </button>
        {pending ? (
          <div className="detail-ai__proposal">
            <div className="detail-ai__meta mono">
              {pending.provider}/{pending.model}
              {pending.latency_ms != null ? ` · ${pending.latency_ms}ms` : ''}
            </div>
            {pending.payload.summary ? (
              <p className="detail-ai__summary">{pending.payload.summary}</p>
            ) : null}
            {pending.payload.operator_tags &&
            pending.payload.operator_tags.length > 0 ? (
              <div className="chips">
                {pending.payload.operator_tags.map((t) => (
                  <span key={t} className="chip">
                    {t}
                  </span>
                ))}
              </div>
            ) : null}
            <div className="detail-ai__flags">
              {pending.payload.language ? (
                <span className="detail-ai__pill">
                  lang {pending.payload.language}
                </span>
              ) : null}
              {pending.payload.confidence != null ? (
                <span className="detail-ai__pill">
                  confidence {pending.payload.confidence.toFixed(2)}
                </span>
              ) : null}
            </div>
            <div className="detail-ai__actions">
              <button
                type="button"
                className="btn btn--secondary"
                disabled={proposalBusy}
                onClick={() => onAcceptProposal(pending.id)}
              >
                Accept
              </button>
              <button
                type="button"
                className="btn btn--ghost"
                disabled={proposalBusy}
                onClick={() => onRejectProposal(pending.id)}
              >
                Reject
              </button>
            </div>
          </div>
        ) : (
          <p className="detail-ai__empty">No pending proposals for this show.</p>
        )}
      </div>

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
