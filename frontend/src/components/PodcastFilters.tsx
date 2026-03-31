type Props = {
  query: string
  onQueryChange: (q: string) => void
  category: string
  onCategoryChange: (s: string) => void
  pinnedOnly: boolean
  onPinnedOnlyChange: (v: boolean) => void
  categories: string[]
}

export function PodcastFilters({
  query,
  onQueryChange,
  category,
  onCategoryChange,
  pinnedOnly,
  onPinnedOnlyChange,
  categories,
}: Props) {
  return (
    <div className="filters">
      <label className="field field--grow">
        <span className="field__label">Search</span>
        <input
          className="input"
          value={query}
          onChange={(e) => onQueryChange(e.target.value)}
          placeholder="Show title or publisher"
        />
      </label>
      <label className="field">
        <span className="field__label">Category</span>
        <select
          className="input input--select"
          value={category}
          onChange={(e) => onCategoryChange(e.target.value)}
        >
          <option value="">All categories</option>
          {categories.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
      </label>
      <label className="toggle">
        <input
          type="checkbox"
          checked={pinnedOnly}
          onChange={(e) => onPinnedOnlyChange(e.target.checked)}
        />
        <span>Pinned only</span>
      </label>
    </div>
  )
}
