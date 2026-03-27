import { NodeColumn } from './NodeColumn'

interface StackEntry {
  nodeId: string
  title: string
  typeName: string
  edgeType?: string
  sourceId?: string
}

interface NodePanelProps {
  stack: StackEntry[]
  onClose: () => void
}

export function NodePanel({ stack, onClose }: NodePanelProps) {
  if (stack.length === 0) return null

  const current = stack[stack.length - 1]

  return (
    <div style={{
      position: 'absolute',
      top: 0,
      right: 0,
      bottom: 0,
      width: '45vw',
      minWidth: 360,
      background: 'white',
      borderLeft: '1px solid #e4e4e7',
      zIndex: 10,
      display: 'flex',
      fontFamily: 'system-ui, -apple-system, sans-serif',
    }}>
      <NodeColumn
        docId={current.nodeId}
        title={current.title}
        typeName={current.typeName}
        onClose={onClose}
      />
    </div>
  )
}

export type { StackEntry }
