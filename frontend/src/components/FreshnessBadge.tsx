import type { Freshness } from '../types'

const labels: Record<Freshness, string> = {
  fresh: 'Fresh',
  aging: 'Aging',
  stale: 'Stale',
}

export function FreshnessBadge({ value }: { value: Freshness }) {
  return <span className={`freshness freshness--${value}`}>{labels[value]}</span>
}
