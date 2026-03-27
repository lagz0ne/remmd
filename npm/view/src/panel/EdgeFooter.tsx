import { useState } from 'react'

interface EdgeItem {
  id: string
  targetId: string
  targetTitle: string
  edgeType: string
  state: string
  direction: 'outgoing' | 'incoming'
}

interface EdgeFooterProps {
  edges: EdgeItem[]
  activeEdgeId?: string
  onEdgeClick: (edge: EdgeItem) => void
}

const stateColors: Record<string, string> = {
  aligned: '#2e7d32',
  stale: '#f57c00',
  pending: '#616161',
  broken: '#c62828',
}

export function EdgeFooter({ edges, activeEdgeId, onEdgeClick }: EdgeFooterProps) {
  const [collapsed, setCollapsed] = useState(false)

  if (edges.length === 0) return null

  return (
    <div style={{ borderTop: '1px solid #e4e4e7' }}>
      <div
        onClick={() => setCollapsed(!collapsed)}
        style={{
          padding: '8px 18px',
          fontSize: 10,
          color: '#71717a',
          fontWeight: 500,
          cursor: 'pointer',
          display: 'flex',
          justifyContent: 'space-between',
        }}
      >
        <span>{edges.length} connection{edges.length !== 1 ? 's' : ''}</span>
        <span style={{ fontSize: 9, color: '#a1a1aa' }}>{collapsed ? '\u25b8' : '\u25be'}</span>
      </div>
      {!collapsed && (
        <div style={{ padding: '0 18px 10px' }}>
          {edges.map(edge => (
            <div
              key={edge.id}
              onClick={() => onEdgeClick(edge)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '6px 8px',
                borderRadius: 4,
                cursor: 'pointer',
                marginBottom: 2,
                background: edge.id === activeEdgeId ? '#eff6ff' : 'transparent',
              }}
            >
              <div style={{
                width: 5, height: 5, borderRadius: '50%',
                background: stateColors[edge.state] || '#a1a1aa',
              }} />
              <div style={{ fontSize: 11, color: edge.id === activeEdgeId ? '#1d4ed8' : '#52525b', flex: 1 }}>
                <span style={{ color: edge.id === activeEdgeId ? '#93c5fd' : '#a1a1aa' }}>
                  {edge.direction === 'outgoing' ? `${edge.edgeType} \u2192` : `\u2190 ${edge.edgeType}`}
                </span>{' '}
                <span style={{ fontWeight: 500 }}>{edge.targetTitle}</span>
              </div>
              <span style={{ fontSize: 9, color: stateColors[edge.state] || '#a1a1aa' }}>{edge.state}</span>
              <span style={{ fontSize: 10, color: '#d4d4d8' }}>{'\u203a'}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export type { EdgeItem }
