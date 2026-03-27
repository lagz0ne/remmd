import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useReactFlow, type Node, type Edge } from '@xyflow/react'
import { natsRequest } from '../nats'
import { computeAutoLayout } from './use-auto-layout'

interface PositionMap {
  [nodeId: string]: { node_id: string; x: number; y: number }
}

export function useForceLayout(nodes: Node[], edges: Edge[]) {
  const { setNodes } = useReactFlow()
  const [layoutApplied, setLayoutApplied] = useState(false)
  const layoutKeyRef = useRef('')
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

    ;(async () => {
      let saved: PositionMap = {}
      try {
        saved = await natsRequest<PositionMap>('remmd.q.positions')
      } catch {}
      if (cancelled) return

      // Split: nodes with saved positions vs nodes needing dagre
      const hasSaved = (id: string) => saved[id] && saved[id].x !== undefined
      const unsavedNodes = curNodes.filter((n) => !hasSaved(n.id))

      // Run dagre only for nodes without saved positions
      let dagrePositions = new Map<string, { x: number; y: number }>()
      if (unsavedNodes.length > 0) {
        const positioned = computeAutoLayout(unsavedNodes, curEdges)
        dagrePositions = new Map(positioned.map((n) => [n.id, n.position]))
      }

      // Apply: saved positions take priority, dagre fills the gaps
      setNodes((prev) =>
        prev.map((node) => {
          if (hasSaved(node.id)) {
            return { ...node, position: { x: saved[node.id].x, y: saved[node.id].y } }
          }
          const dagre = dagrePositions.get(node.id)
          if (dagre) {
            return { ...node, position: dagre }
          }
          return node
        }),
      )
      setLayoutApplied(true)
    })()

    return () => { cancelled = true }
  }, [graphKey, setNodes, layoutApplied])

  const onNodeDragStart = useCallback(() => {}, [])
  const onNodeDrag = useCallback(() => {}, [])

  const onNodeDragStop = useCallback(
    (_: unknown, node: Node) => {
      natsRequest('remmd.c.positions', [{
        node_id: node.id,
        x: node.position.x,
        y: node.position.y,
      }]).catch(() => {})
    },
    [],
  )

  const resetLayout = useCallback(() => {
    const curNodes = nodesRef.current
    const curEdges = edgesRef.current
    if (curNodes.length === 0) return

    natsRequest('remmd.c.positions.clear').catch(() => {})

    const positioned = computeAutoLayout(curNodes, curEdges)
    const posMap = new Map(positioned.map((n) => [n.id, n.position]))
    setNodes((prev) =>
      prev.map((node) => {
        const pos = posMap.get(node.id)
        if (!pos) return node
        return { ...node, position: pos }
      }),
    )

    setLayoutApplied(false)
    layoutKeyRef.current = ''
  }, [setNodes])

  return { onNodeDragStart, onNodeDrag, onNodeDragStop, resetLayout }
}
