import { useCallback, useEffect, useMemo, useRef } from 'react'
import { useReactFlow, type Node, type Edge } from '@xyflow/react'
import {
  forceSimulation,
  forceLink,
  forceManyBody,
  forceCenter,
  forceCollide,
  type SimulationNodeDatum,
  type SimulationLinkDatum,
} from 'd3-force'

interface SimNode extends SimulationNodeDatum {
  id: string
  width: number
  height: number
}

interface SimLink extends SimulationLinkDatum<SimNode> {
  id: string
}

const NODE_WIDTH = 200
const NODE_HEIGHT = 80

export function useForceLayout(nodes: Node[], edges: Edge[]) {
  const { setNodes } = useReactFlow()
  const simRef = useRef<ReturnType<typeof forceSimulation<SimNode>> | null>(null)
  const draggingRef = useRef<string | null>(null)

  // Stable identity key — changes when the graph structure actually changes
  const graphKey = useMemo(
    () => nodes.map((n) => n.id).sort().join(',') + '|' + edges.map((e) => e.id).sort().join(','),
    [nodes, edges],
  )

  useEffect(() => {
    if (nodes.length === 0) return

    const simNodes: SimNode[] = nodes.map((n) => ({
      id: n.id,
      x: n.position.x || Math.random() * 800,
      y: n.position.y || Math.random() * 600,
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
    }))

    const nodeMap = new Map(simNodes.map((sn) => [sn.id, sn]))

    const simLinks: SimLink[] = edges.map((e) => ({
      id: e.id,
      source: e.source,
      target: e.target,
    }))

    const sim = forceSimulation<SimNode>(simNodes)
      .force(
        'link',
        forceLink<SimNode, SimLink>(simLinks)
          .id((d) => d.id)
          .distance(350),
      )
      .force('charge', forceManyBody().strength(-1200))
      .force('center', forceCenter(0, 0))
      .force(
        'collide',
        forceCollide<SimNode>().radius((d) => Math.max(d.width, d.height) / 2 + 40),
      )
      .alphaDecay(0.02)
      .on('tick', () => {
        setNodes((prev) =>
          prev.map((node) => {
            if (node.id === draggingRef.current) return node
            const simNode = nodeMap.get(node.id)
            if (!simNode || simNode.x == null || simNode.y == null) return node
            return { ...node, position: { x: simNode.x, y: simNode.y } }
          }),
        )
      })

    simRef.current = sim

    return () => {
      sim.stop()
      simRef.current = null
    }
  }, [graphKey, setNodes])

  const onNodeDragStart = useCallback((_: unknown, node: Node) => {
    draggingRef.current = node.id
    simRef.current?.alphaTarget(0.3).restart()
  }, [])

  const onNodeDrag = useCallback((_: unknown, node: Node) => {
    const sim = simRef.current
    if (!sim) return
    const simNode = sim.nodes().find((sn) => sn.id === node.id)
    if (simNode) {
      simNode.fx = node.position.x
      simNode.fy = node.position.y
    }
  }, [])

  const onNodeDragStop = useCallback((_: unknown, node: Node) => {
    draggingRef.current = null
    const sim = simRef.current
    if (!sim) return
    const simNode = sim.nodes().find((sn) => sn.id === node.id)
    if (simNode) {
      simNode.fx = null
      simNode.fy = null
    }
    sim.alphaTarget(0)
  }, [])

  return { onNodeDragStart, onNodeDrag, onNodeDragStop }
}
