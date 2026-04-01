import type { Podcast } from '../types'
import { formatRelative, freshnessForPodcast } from '../lib/freshness'
import { FreshnessBadge } from './FreshnessBadge'

type Props = {
  podcasts: Podcast[]
  selectedId: string | null
  onSelect: (id: string) => void
}

export function PodcastTable({ podcasts, selectedId, onSelect }: Props) {
  if (podcasts.length === 0) {
    return (
      <div className="empty-state">
        <p>No podcasts match these filters.</p>
      </div>
    )
  }

  return (
    <div className="table-wrap">
      <table className="data-table">
        <thead>
          <tr>
            <th scope="col">Show</th>
            <th scope="col">Publisher</th>
            <th scope="col">Categories</th>
            <th scope="col">Episodes</th>
            <th scope="col">Freshness</th>
            <th scope="col">Flags</th>
          </tr>
        </thead>
        <tbody>
          {podcasts.map((p) => {
            const fresh = freshnessForPodcast(p)
            const rel = formatRelative(p.updatedAt)
            const active = p.id === selectedId
            return (
              <tr
                key={p.id}
                className={active ? 'is-selected' : undefined}
                onClick={() => onSelect(p.id)}
              >
                <td>
                  <span className="cell-title">{p.title}</span>
                  <span className="cell-sub mono">{p.sourceId}</span>
                </td>
                <td>{p.author}</td>
                <td>
                  <div className="chips">
                    {p.categories.slice(0, 3).map((s) => (
                      <span key={s} className="chip">
                        {s}
                      </span>
                    ))}
                  </div>
                </td>
                <td className="mono">
                  {p.trackCount != null ? p.trackCount : '—'}
                </td>
                <td>
                  <FreshnessBadge value={fresh} />
                  <span className="cell-sub">{rel}</span>
                </td>
                <td>
                  <div className="flags">
                    {p.pinned && <span className="flag flag--pin">Pinned</span>}
                    {p.featured && (
                      <span className="flag flag--feat">Featured</span>
                    )}
                  </div>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
