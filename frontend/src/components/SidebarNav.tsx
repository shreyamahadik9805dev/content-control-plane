type View = 'podcasts' | 'audit'

type Props = {
  active: View
  onSelect: (v: View) => void
  apiBase: string
}

export function SidebarNav({ active, onSelect, apiBase }: Props) {
  return (
    <nav className="sidebar" aria-label="Primary">
      <div className="sidebar__section">
        <p className="sidebar__label">Navigate</p>
        <button
          type="button"
          className={`sidebar__link${active === 'podcasts' ? ' is-active' : ''}`}
          onClick={() => onSelect('podcasts')}
        >
          Podcasts
        </button>
        <button
          type="button"
          className={`sidebar__link${active === 'audit' ? ' is-active' : ''}`}
          onClick={() => onSelect('audit')}
        >
          Audit log
        </button>
      </div>
      <div className="sidebar__section sidebar__hint">
        <p className="sidebar__label">API</p>
        <p className="sidebar__muted">
          UI reads/writes <span className="mono">/podcasts</span> on{' '}
          <span className="mono break-all">{apiBase}</span>. Sync uses the live
          iTunes Search endpoint when <span className="mono">ITUNES_MOCK</span>{' '}
          is off.
        </p>
      </div>
    </nav>
  )
}
