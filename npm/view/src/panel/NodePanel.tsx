import { useMemo, useCallback } from 'react'
import type { Node, Edge } from '@xyflow/react'
import { NodeColumn } from './NodeColumn'
import { EdgeFooter, type EdgeItem } from './EdgeFooter'
import { TraceHeader, type TraceEntry } from './TraceHeader'

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
  onStackChange: (stack: StackEntry[]) => void
}

function useEdgeItems(nodeId: string, graphEdges: Edge[], graphNodes: Node[]): EdgeItem[] {
  return useMemo(() => {
    const items: EdgeItem[] = []
    for (const edge of graphEdges) {
      const isSource = edge.source === nodeId
      const isTarget = edge.target === nodeId
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
  }, [graphEdges, graphNodes, nodeId])
}

export function NodePanel({ stack, graphEdges, graphNodes, onClose, onEdgeFollow, onStackChange }: NodePanelProps) {
  if (stack.length === 0) return null

  const isTwoCol = stack.length >= 2
  const col1Entry = isTwoCol ? stack[stack.length - 2] : stack[stack.length - 1]
  const col2Entry = isTwoCol ? stack[stack.length - 1] : undefined

  const col1Edges = useEdgeItems(col1Entry.nodeId, graphEdges, graphNodes)
  const col2Edges = useEdgeItems(col2Entry?.nodeId ?? '', graphEdges, graphNodes)

  // Active edge in col1 = the edge connecting col1 to col2
  const activeEdgeId = useMemo(() => {
    if (!col2Entry) return undefined
    const edge = graphEdges.find(e =>
      (e.source === col1Entry.nodeId && e.target === col2Entry.nodeId) ||
      (e.target === col1Entry.nodeId && e.source === col2Entry.nodeId)
    )
    return edge?.id
  }, [graphEdges, col1Entry.nodeId, col2Entry?.nodeId])

  // Trace entries: all stack entries except the last (col2)
  const traceEntries: TraceEntry[] = useMemo(() => {
    if (!isTwoCol) return []
    return stack.slice(0, -1).map(s => ({
      nodeId: s.nodeId,
      title: s.title,
      edgeType: s.edgeType,
    }))
  }, [stack, isTwoCol])

  const onTraceJump = useCallback((index: number) => {
    // Keep entries 0..index+1 (the jumped-to node as col1, next as col2)
    const newStack = stack.slice(0, index + 2)
    if (newStack.length <= 1) {
      // Single column mode
      onStackChange(newStack.length === 0 ? [] : newStack)
    } else {
      onStackChange(newStack)
    }
  }, [stack, onStackChange])

  const onTraceClose = useCallback(() => {
    // Pop last entry, collapse to 1 col if stack becomes length 1
    onStackChange(stack.slice(0, -1))
  }, [stack, onStackChange])

  return (
    <div style={{
      position: 'absolute',
      top: 0,
      right: 0,
      bottom: 0,
      width: isTwoCol ? '60vw' : '35vw',
      minWidth: isTwoCol ? 600 : 320,
      background: 'white',
      borderLeft: '1px solid #e4e4e7',
      zIndex: 10,
      display: 'flex',
      fontFamily: 'system-ui, -apple-system, sans-serif',
    }}>
      {/* Column 1 */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        <NodeColumn
          docId={col1Entry.nodeId}
          title={col1Entry.title}
          typeName={col1Entry.typeName}
          onClose={!isTwoCol ? onClose : undefined}
          footer={
            <EdgeFooter
              edges={col1Edges}
              activeEdgeId={activeEdgeId}
              onEdgeClick={onEdgeFollow}
            />
          }
        />
      </div>

      {/* Divider + Column 2 */}
      {isTwoCol && col2Entry && (
        <>
          <div style={{ width: 1, background: '#e4e4e7', flexShrink: 0 }} />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
            <NodeColumn
              docId={col2Entry.nodeId}
              title={col2Entry.title}
              typeName={col2Entry.typeName}
              header={
                <TraceHeader
                  entries={traceEntries}
                  onJump={onTraceJump}
                  onClose={onTraceClose}
                />
              }
              footer={
                <EdgeFooter
                  edges={col2Edges}
                  onEdgeClick={onEdgeFollow}
                />
              }
            />
          </div>
        </>
      )}
    </div>
  )
}

export type { StackEntry }
