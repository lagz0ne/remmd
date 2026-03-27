import { Graph, layout } from '@dagrejs/dagre'
import type { Node, Edge } from '@xyflow/react'

export const LAYOUT_NODE_WIDTH = 180
export const LAYOUT_NODE_HEIGHT = 50
const RANK_SEP = 60
const NODE_SEP = 30

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
    g.setNode(node.id, { width: LAYOUT_NODE_WIDTH, height: LAYOUT_NODE_HEIGHT })
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
        x: pos.x! - LAYOUT_NODE_WIDTH / 2,
        y: pos.y! - LAYOUT_NODE_HEIGHT / 2,
      },
    }
  })
}
