import { Graph, layout } from '@dagrejs/dagre'
import type { Node, Edge } from '@xyflow/react'

// Close-view: fixed 280px wide. Plus 20px float overhang each side = 320.
// Height varies by section count but caps at ~220px CSS.
// These are the bounding box sizes dagre uses for collision avoidance.
// Must be >= the largest rendered node at any zoom level.
// Close-view: 280px content + 40px float overhang = 320px wide
// Close-view: ~200px content + 28px float overhang = 228px tall
export const LAYOUT_NODE_WIDTH = 340
export const LAYOUT_NODE_HEIGHT = 240
const NODE_WIDTH = LAYOUT_NODE_WIDTH
const NODE_HEIGHT = LAYOUT_NODE_HEIGHT
const RANK_SEP = 140
const NODE_SEP = 100

export function computeAutoLayout(nodes: Node[], edges: Edge[]): Node[] {
  if (nodes.length === 0) return nodes

  const g = new Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({
    rankdir: 'TB',
    ranksep: RANK_SEP,
    nodesep: NODE_SEP,
    marginx: 40,
    marginy: 40,
  })

  for (const node of nodes) {
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
  }

  for (const edge of edges) {
    g.setEdge(edge.source, edge.target)
  }

  layout(g)

  return nodes.map((node) => {
    const pos = g.node(node.id)
    return {
      ...node,
      position: {
        x: pos.x! - NODE_WIDTH / 2,
        y: pos.y! - NODE_HEIGHT / 2,
      },
    }
  })
}
