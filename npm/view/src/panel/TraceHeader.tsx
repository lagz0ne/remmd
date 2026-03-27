interface TraceEntry {
  nodeId: string
  title: string
  edgeType?: string
}

interface TraceHeaderProps {
  entries: TraceEntry[]
  onJump: (index: number) => void
  onClose: () => void
}

const pillBase = {
  display: 'inline-flex',
  alignItems: 'center',
  padding: '2px 8px',
  borderRadius: 9999,
  fontSize: 11,
  fontWeight: 500 as const,
  cursor: 'pointer',
}

const prevPill = { ...pillBase, background: '#f4f4f5', color: '#71717a' }
const currentPill = { ...pillBase, background: '#eff6ff', color: '#1d4ed8' }

export function TraceHeader({ entries, onJump, onClose }: TraceHeaderProps) {
  if (entries.length === 0) return null

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      gap: 6,
      padding: '6px 14px',
      background: '#f8fafc',
      borderBottom: '1px solid #e4e4e7',
      fontSize: 11,
      color: '#a1a1aa',
      minHeight: 30,
      flexWrap: 'wrap',
    }}>
      <span style={{ fontSize: 10, marginRight: 2 }}>via</span>
      {entries.map((entry, i) => {
        const isLast = i === entries.length - 1
        return (
          <span key={entry.nodeId + i} style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            <span
              onClick={() => onJump(i)}
              style={isLast ? currentPill : prevPill}
            >
              {entry.title}
            </span>
            {!isLast && entry.edgeType && (
              <span style={{ fontSize: 10, color: '#a1a1aa' }}>
                {entry.edgeType} →
              </span>
            )}
            {!isLast && !entry.edgeType && (
              <span style={{ fontSize: 10, color: '#d4d4d8' }}>→</span>
            )}
          </span>
        )
      })}
      <span
        onClick={onClose}
        style={{
          marginLeft: 'auto',
          fontSize: 12,
          color: '#a1a1aa',
          cursor: 'pointer',
          padding: '0 4px',
          flexShrink: 0,
        }}
      >
        ✕
      </span>
    </div>
  )
}

export type { TraceEntry }
