import type { ReactNode } from 'react'
import { Panel } from '@xyflow/react'
import type { PanelMode } from './use-panel-state'

interface PanelShellProps {
  mode: PanelMode
  columns: 1 | 2
  onClose: () => void
  children: ReactNode
}

export function PanelShell({ mode, columns, onClose, children }: PanelShellProps) {
  if (mode === 'closed') return null

  return (
    <Panel
      position="top-right"
      className="nowheel nodrag nopan"
      style={{ margin: 0, padding: 0, height: '100%' }}
    >
      <div
        className="h-full bg-white border-l border-zinc-200 shadow-lg overflow-hidden flex flex-col transition-all duration-200"
        style={{ width: columns === 2 ? '60vw' : '35vw', minWidth: 360 }}
      >
        <div className="flex items-center justify-between px-4 py-2 border-b border-zinc-100">
          <span className="text-xs text-zinc-400 font-mono">
            {mode === 'doc' ? 'Document' : 'Link Detail'}
          </span>
          <button
            onClick={onClose}
            className="text-zinc-400 hover:text-zinc-700 text-sm px-1"
          >
            ✕
          </button>
        </div>

        <div className="flex-1 overflow-hidden flex">
          {children}
        </div>
      </div>
    </Panel>
  )
}
