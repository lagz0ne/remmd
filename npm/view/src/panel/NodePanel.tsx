import { useMemo } from 'react'
import type { Node, Edge } from '@xyflow/react'
import { NodeColumn } from './NodeColumn'
import { EdgeFooter, type EdgeItem } from './EdgeFooter'

interface StackEntry {
  nodeId: string
  title: string
  typeName: string
  edgeType?: string
  sourceId?: string
}

interface NodePanelProps {
  stack: StackEntry[]
  graphEdges: Edge[]
  graphNodes: Node[]
  onClose: () => void
  onEdgeFollow: (edge: EdgeItem) => void
}

export function NodePanel({ stack, graphEdges, graphNodes, onClose, onEdgeFollow }: NodePanelProps) {
  if (stack.length === 0) return null

  const current = stack[stack.length - 1]

  const edgeItems = useMemo(() => {
    const items: EdgeItem[] = []
    for (const edge of graphEdges) {
      const isSource = edge.source === current.nodeId
      const isTarget = edge.target === current.nodeId
      if (!isSource && !isTarget) continue

      const otherId = isSource ? edge.target : edge.source
      const otherNode = graphNodes.find(n => n.id === otherId)
      const otherTitle = (otherNode?.data as any)?.title || otherId
      const edgeData = edge.data as any
      const edgeType = edgeData?.links?.[0]?.relationship_type || edgeData?.edgeType || 'link'
      const state = edgeData?.worstState || edgeData?.links?.[0]?.state || 'pending'

      items.push({
        id: edge.id,
        targetId: otherId,
        targetTitle: otherTitle,
        edgeType,
        state,
        direction: isSource ? 'outgoing' : 'incoming',
      })
    }
    return items
  }, [graphEdges, graphNodes, current.nodeId])

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
        footer={
          <EdgeFooter
            edges={edgeItems}
            onEdgeClick={onEdgeFollow}
          />
        }
      />
    </div>
  )
}

export type { StackEntry }
