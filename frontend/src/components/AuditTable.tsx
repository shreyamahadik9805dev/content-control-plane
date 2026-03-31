import type { AuditEntry } from '../types'
import { formatRelative } from '../lib/freshness'

type Props = {
  entries: AuditEntry[]
}

export function AuditTable({ entries }: Props) {
  return (
    <div className="table-wrap">
      <table className="data-table">
        <thead>
          <tr>
            <th scope="col">When</th>
            <th scope="col">Action</th>
            <th scope="col">Entity</th>
            <th scope="col">Detail</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((e) => (
            <tr key={e.id}>
              <td className="mono">{formatRelative(e.createdAt)}</td>
              <td>
                <span className="mono audit-action">{e.action}</span>
              </td>
              <td className="mono break-all">{e.entityId}</td>
              <td>{e.detail}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
