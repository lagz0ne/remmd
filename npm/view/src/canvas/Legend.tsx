import { useState, useEffect } from 'react'
import { Panel } from '@xyflow/react'
import { CaretDownIcon, CaretUpIcon, CircleIcon } from '@phosphor-icons/react'
import { stateColor } from '../theme/colors'

const STORAGE_KEY = 'remmd-legend-collapsed'

interface LegendProps {
  types?: string[]
  activeFilter?: string | null
  onFilterType?: (type: string | null) => void
}

const states = [
  { key: 'aligned', label: 'Aligned' },
  { key: 'stale', label: 'Stale' },
  { key: 'broken', label: 'Broken' },
  { key: 'pending', label: 'Pending' },
] as const

export function Legend({ types, activeFilter, onFilterType }: LegendProps) {
  const [collapsed, setCollapsed] = useState(() => {
    try {
      return localStorage.getItem(STORAGE_KEY) === 'true'
    } catch {
      return false
    }
  })

  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, String(collapsed))
    } catch {
      // ignore
    }
  }, [collapsed])

  if (collapsed) {
    return (
      <Panel position="bottom-left" style={{ marginBottom: 230 }}>
        <button
          onClick={() => setCollapsed(false)}
          className="flex items-center gap-1 px-2.5 py-1 text-[9px] font-medium text-zinc-500 cursor-pointer"
          style={{
            background: '#fff',
            border: '1px solid #e4e4e7',
            borderRadius: 8,
            boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
          }}
        >
          Legend
          <CaretDownIcon size={10} weight="light" />
        </button>
      </Panel>
    )
  }

  return (
    <Panel position="bottom-left" style={{ marginBottom: 230 }}>
      <div
        style={{
          background: '#fff',
          border: '1px solid #e4e4e7',
          borderRadius: 8,
          boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
          padding: '8px 10px',
          minWidth: 130,
        }}
      >
        {/* Header */}
        <div className="flex items-center justify-between mb-1.5">
          <span className="text-[9px] font-semibold text-zinc-500 uppercase tracking-wide">
            Legend
          </span>
          <button
            onClick={() => setCollapsed(true)}
            className="text-zinc-400 cursor-pointer hover:text-zinc-600"
          >
            <CaretUpIcon size={10} weight="light" />
          </button>
        </div>

        {/* State section */}
        <div className="mb-2">
          <div className="text-[9px] text-zinc-400 mb-1">State</div>
          <div className="space-y-0.5">
            {states.map(({ key, label }) => (
              <div key={key} className="flex items-center gap-1.5">
                <CircleIcon
                  size={10}
                  weight="fill"
                  style={{ color: stateColor[key] }}
                />
                <span className="text-[10px] text-zinc-600">{label}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Border section */}
        <div className="mb-2">
          <div className="text-[9px] text-zinc-400 mb-1">Border</div>
          <div className="space-y-0.5">
            <div className="flex items-center gap-1.5">
              <svg width="16" height="10">
                <line
                  x1="0" y1="5" x2="16" y2="5"
                  stroke="#a1a1aa"
                  strokeWidth="1.5"
                />
              </svg>
              <span className="text-[10px] text-zinc-600">Linked</span>
            </div>
            <div className="flex items-center gap-1.5">
              <svg width="16" height="10">
                <line
                  x1="0" y1="5" x2="16" y2="5"
                  stroke="#a1a1aa"
                  strokeWidth="1.5"
                  strokeDasharray="3 2"
                />
              </svg>
              <span className="text-[10px] text-zinc-600">Orphan</span>
            </div>
          </div>
        </div>

        {/* Types section */}
        {types && types.length > 0 && (
          <div>
            <div className="text-[9px] text-zinc-400 mb-1">Types</div>
            <div className="space-y-0.5">
              {types.map((t) => {
                const isActive = activeFilter === t
                return (
                  <button
                    key={t}
                    onClick={() => onFilterType?.(isActive ? null : t)}
                    className="flex items-center w-full text-left px-1 py-0.5 rounded cursor-pointer transition-colors"
                    style={{
                      background: isActive ? '#27272a' : 'transparent',
                      color: isActive ? '#fff' : '#52525b',
                      fontSize: 10,
                    }}
                  >
                    {t}
                  </button>
                )
              })}
            </div>
          </div>
        )}
      </div>
    </Panel>
  )
}
