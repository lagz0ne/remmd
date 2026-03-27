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
        await natsRequest<PositionMap>('remmd.q.positions')
        if (cancelled) return
        applyDagre()
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

  const onNodeDragStart: (_: unknown, node: Node) => void = useCallback(() => {}, [])
  const onNodeDrag: (_: unknown, node: Node) => void = useCallback(() => {}, [])

  const onNodeDragStop = useCallback(
    (_: unknown, node: Node) => {
      const position = {
        node_id: node.id,
        x: node.position.x,
        y: node.position.y,
      }
      natsRequest('remmd.c.positions', [position]).catch(() => {})
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

    setLayoutApplied(false)
    layoutKeyRef.current = ''
  }, [nodes, edges, setNodes])

  return { onNodeDragStart, onNodeDrag, onNodeDragStop, resetLayout }
}
