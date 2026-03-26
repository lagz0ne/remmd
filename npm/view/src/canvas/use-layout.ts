import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useReactFlow, type Node, type Edge } from '@xyflow/react'
import { natsRequest } from '../nats'
import { computeAutoLayout, LAYOUT_NODE_WIDTH, LAYOUT_NODE_HEIGHT } from './use-auto-layout'

interface PositionMap {
  [nodeId: string]: { node_id: string; x: number; y: number }
}

function hasOverlaps(positions: { x: number; y: number }[]): boolean {
  const w = LAYOUT_NODE_WIDTH
  const h = LAYOUT_NODE_HEIGHT
  for (let i = 0; i < positions.length; i++) {
    for (let j = i + 1; j < positions.length; j++) {
      const a = positions[i], b = positions[j]
      if (
        a.x < b.x + w && a.x + w > b.x &&
        a.y < b.y + h && a.y + h > b.y
      ) return true
    }
  }
  return false
}

/**
 * Orchestrates node layout:
 * 1. On load: fetch saved positions via NATS
 * 2. If saved positions exist: apply them
 * 3. If not: run dagre auto-layout
 * 4. On drag stop: persist the dragged node's position
 * 5. Exposes resetLayout() to re-run dagre and clear saved positions
 */
export function useForceLayout(nodes: Node[], edges: Edge[]) {
  const { setNodes } = useReactFlow()
  const [layoutApplied, setLayoutApplied] = useState(false)
  const savedPositionsRef = useRef<PositionMap | null>(null)
  const layoutKeyRef = useRef<string>('')
  const nodesRef = useRef(nodes)
  const edgesRef = useRef(edges)
  nodesRef.current = nodes
  edgesRef.current = edges

  const graphKey = useMemo(
    () =>
      nodes.map((n) => n.id).sort().join(',') +
      '|' +
      edges.map((e) => e.id).sort().join(','),
    [nodes, edges],
  )

  useEffect(() => {
    const curNodes = nodesRef.current
    const curEdges = edgesRef.current
    if (curNodes.length === 0) return
    if (layoutKeyRef.current === graphKey && layoutApplied) return

    layoutKeyRef.current = graphKey

    let cancelled = false

    const applyDagre = () => {
      const positioned = computeAutoLayout(curNodes, curEdges)
      const posMap = new Map(positioned.map((n) => [n.id, n.position]))
      setNodes((prev) =>
        prev.map((node) => {
          const pos = posMap.get(node.id)
          if (!pos) return node
          return { ...node, position: pos }
        }),
      )
    }

    ;(async () => {
      try {
        const saved = await natsRequest<PositionMap>('remmd.q.positions')
        if (cancelled) return
        savedPositionsRef.current = saved

        const hasPositions =
          saved && Object.keys(saved).length > 0 &&
          curNodes.some((n) => saved[n.id])

        // Always run dagre for consistent non-overlapping layout.
        // Saved positions are only applied for individual node drags,
        // not for full graph layout — prevents stale overlapping positions.
        {
          applyDagre()
        }
        setLayoutApplied(true)
      } catch {
        if (cancelled) return
        applyDagre()
        setLayoutApplied(true)
      }
    })()

    return () => {
      cancelled = true
    }
  }, [graphKey, setNodes, layoutApplied])

  const onNodeDragStart: (_: unknown, node: Node) => void = useCallback(() => {
    // No-op -- just keeping the interface consistent
  }, [])

  const onNodeDrag: (_: unknown, node: Node) => void = useCallback(() => {
    // No-op -- ReactFlow handles visual position during drag
  }, [])

  const onNodeDragStop = useCallback(
    (_: unknown, node: Node) => {
      // Persist the dragged node's position
      const position = {
        node_id: node.id,
        x: node.position.x,
        y: node.position.y,
      }
      natsRequest('remmd.c.positions', [position]).catch(() => {
        // Silently ignore save failures
      })
    },
    [],
  )

  const resetLayout = useCallback(() => {
    if (nodes.length === 0) return

    natsRequest('remmd.c.positions.clear').catch(() => {})

    const positioned = computeAutoLayout(nodes, edges)
    const posMap = new Map(positioned.map((n) => [n.id, n.position]))
    setNodes((prev) =>
      prev.map((node) => {
        const pos = posMap.get(node.id)
        if (!pos) return node
        return { ...node, position: pos }
      }),
    )

    savedPositionsRef.current = null
    setLayoutApplied(false)
    layoutKeyRef.current = ''
  }, [nodes, edges, setNodes])

  return { onNodeDragStart, onNodeDrag, onNodeDragStop, resetLayout }
}
