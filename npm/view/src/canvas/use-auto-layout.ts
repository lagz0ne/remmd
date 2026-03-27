import { Graph, layout } from '@dagrejs/dagre'
import type { Node, Edge } from '@xyflow/react'

export const LAYOUT_NODE_WIDTH = 230
export const LAYOUT_NODE_HEIGHT = 120
const RANK_SEP = 40
const NODE_SEP = 20
const CLUSTER_GAP = 80

export function computeAutoLayout(nodes: Node[], edges: Edge[]): Node[] {
  if (nodes.length === 0) return nodes

  // Step 1: Find connected components (subgraphs)
  const adj = new Map<string, Set<string>>()
  for (const n of nodes) adj.set(n.id, new Set())
  for (const e of edges) {
    adj.get(e.source)?.add(e.target)
    adj.get(e.target)?.add(e.source)
  }

  const visited = new Set<string>()
  const clusters: string[][] = []
  for (const n of nodes) {
    if (visited.has(n.id)) continue
    const cluster: string[] = []
    const stack = [n.id]
    while (stack.length > 0) {
      const id = stack.pop()!
      if (visited.has(id)) continue
      visited.add(id)
      cluster.push(id)
      for (const neighbor of adj.get(id) || []) {
        if (!visited.has(neighbor)) stack.push(neighbor)
      }
    }
    clusters.push(cluster)
  }

  // Step 2: Layout each cluster with dagre
  const positioned = new Map<string, { x: number; y: number }>()

  interface ClusterBounds { minX: number; minY: number; maxX: number; maxY: number; w: number; h: number }
  const clusterBounds: ClusterBounds[] = []

  for (const cluster of clusters) {
    if (cluster.length === 1) {
      positioned.set(cluster[0], { x: 0, y: 0 })
      clusterBounds.push({ minX: 0, minY: 0, maxX: LAYOUT_NODE_WIDTH, maxY: LAYOUT_NODE_HEIGHT, w: LAYOUT_NODE_WIDTH, h: LAYOUT_NODE_HEIGHT })
      continue
    }

    const g = new Graph()
    g.setDefaultEdgeLabel(() => ({}))
    g.setGraph({ rankdir: 'TB', ranksep: RANK_SEP, nodesep: NODE_SEP, marginx: 20, marginy: 20 })

    const clusterSet = new Set(cluster)
    for (const id of cluster) {
      g.setNode(id, { width: LAYOUT_NODE_WIDTH, height: LAYOUT_NODE_HEIGHT })
    }
    for (const e of edges) {
      if (clusterSet.has(e.source) && clusterSet.has(e.target)) {
        g.setEdge(e.source, e.target)
      }
    }

    layout(g)

    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity
    for (const id of cluster) {
      const pos = g.node(id)
      const x = pos.x! - LAYOUT_NODE_WIDTH / 2
      const y = pos.y! - LAYOUT_NODE_HEIGHT / 2
      positioned.set(id, { x, y })
      minX = Math.min(minX, x)
      minY = Math.min(minY, y)
      maxX = Math.max(maxX, x + LAYOUT_NODE_WIDTH)
      maxY = Math.max(maxY, y + LAYOUT_NODE_HEIGHT)
    }
    clusterBounds.push({ minX, minY, maxX, maxY, w: maxX - minX, h: maxY - minY })
  }

  // Step 3: Arrange clusters in a square-ish grid
  const totalArea = clusterBounds.reduce((sum, b) => sum + b.w * b.h, 0)
  const targetSide = Math.sqrt(totalArea) * 1.3

  let curX = 0
  let curY = 0
  let rowMaxH = 0

  for (let i = 0; i < clusters.length; i++) {
    const cluster = clusters[i]
    const bounds = clusterBounds[i]

    if (curX > 0 && curX + bounds.w > targetSide) {
      curX = 0
      curY += rowMaxH + CLUSTER_GAP
      rowMaxH = 0
    }

    const offsetX = curX - bounds.minX
    const offsetY = curY - bounds.minY

    for (const id of cluster) {
      const pos = positioned.get(id)!
      positioned.set(id, { x: pos.x + offsetX, y: pos.y + offsetY })
    }

    curX += bounds.w + CLUSTER_GAP
    rowMaxH = Math.max(rowMaxH, bounds.h)
  }

  return nodes.map(node => ({
    ...node,
    position: positioned.get(node.id) || { x: 0, y: 0 },
  }))
}
