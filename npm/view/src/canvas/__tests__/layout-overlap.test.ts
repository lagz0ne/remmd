import { describe, it, expect } from 'bun:test'
import { computeAutoLayout } from '../use-auto-layout'
import type { Node, Edge } from '@xyflow/react'

/**
 * Node dimensions per zoom level (CSS pixels).
 * These must match the actual rendered sizes in PlaybookNode.tsx.
 * Includes floating indicator overhang (~20px each side).
 *
 * CRITICAL: close-view height varies by section count.
 * We test with the MAXIMUM observed height to catch worst-case overlaps.
 * Measured from real rendered nodes:
 *   2 sections: 127px, 3 sections: 163px, 3+ sections: 200px
 *   Plus floating indicators: +28px (14px top + 14px bottom)
 */
const NODE_DIMS = {
  far: { width: 90, height: 30 },
  medium: { width: 240, height: 100 },
  // Worst case: 200px content + 28px floats = 228px
  close: { width: 320, height: 228 },
}

function makeNodes(count: number): Node[] {
  return Array.from({ length: count }, (_, i) => ({
    id: `n${i}`,
    type: 'document',
    position: { x: 0, y: 0 },
    data: { title: `Node ${i}` },
  }))
}

function makeEdges(pairs: [number, number][]): Edge[] {
  return pairs.map(([s, t], i) => ({
    id: `e${i}`,
    source: `n${s}`,
    target: `n${t}`,
    type: 'bundled',
  }))
}

/**
 * Check if two nodes overlap given their positions and dimensions.
 * Returns the overlap area if they do, 0 otherwise.
 */
function overlapArea(
  a: { x: number; y: number },
  b: { x: number; y: number },
  dims: { width: number; height: number },
): number {
  const ax1 = a.x, ay1 = a.y, ax2 = a.x + dims.width, ay2 = a.y + dims.height
  const bx1 = b.x, by1 = b.y, bx2 = b.x + dims.width, by2 = b.y + dims.height

  const overlapX = Math.max(0, Math.min(ax2, bx2) - Math.max(ax1, bx1))
  const overlapY = Math.max(0, Math.min(ay2, by2) - Math.max(ay1, by1))

  return overlapX * overlapY
}

/**
 * Find all overlapping node pairs for a given zoom level.
 */
function findOverlaps(
  nodes: Node[],
  zoomLevel: keyof typeof NODE_DIMS,
): { a: string; b: string; area: number }[] {
  const dims = NODE_DIMS[zoomLevel]
  const overlaps: { a: string; b: string; area: number }[] = []

  for (let i = 0; i < nodes.length; i++) {
    for (let j = i + 1; j < nodes.length; j++) {
      const area = overlapArea(nodes[i].position, nodes[j].position, dims)
      if (area > 0) {
        overlaps.push({ a: nodes[i].id, b: nodes[j].id, area })
      }
    }
  }

  return overlaps
}

// --- Test fixtures ---

// Simple: 5 nodes, linear chain
const linearNodes = makeNodes(5)
const linearEdges = makeEdges([[0, 1], [1, 2], [2, 3], [3, 4]])

// C3-like: 1 context → 2 containers → 5 components each → 3 refs cited
const c3Nodes = makeNodes(21) // 1 + 2 + 10 + 8 refs
const c3Edges = makeEdges([
  // context → containers
  [0, 1], [0, 2],
  // container 1 → components
  [1, 3], [1, 4], [1, 5], [1, 6], [1, 7],
  // container 2 → components
  [2, 8], [2, 9], [2, 10], [2, 11], [2, 12],
  // components cite refs
  [3, 13], [4, 13], [5, 14], [6, 14], [7, 15],
  [8, 16], [9, 17], [10, 18], [11, 19], [12, 20],
])

// Wide: 50 nodes, sparse edges (like the real c3 import)
const wideNodes = makeNodes(50)
const wideEdges = makeEdges([
  [0, 1], [0, 2], [0, 3], [0, 4],
  [1, 5], [1, 6], [1, 7], [1, 8], [1, 9],
  [2, 10], [2, 11], [2, 12], [2, 13], [2, 14],
  [3, 15], [3, 16], [3, 17],
  [4, 18], [4, 19], [4, 20],
  [5, 21], [6, 21], [7, 22], [8, 23],
  [10, 24], [11, 25], [12, 26],
])

// Realistic: 52 nodes, only 11 edges (many orphans — matches actual c3 import)
const realisticNodes = makeNodes(52)
const realisticEdges = makeEdges([
  // Container contains components
  [1, 3], [1, 4],
  // Container 2 contains components
  [5, 6], [5, 7], [5, 8],
  // Components cite refs
  [3, 40], [8, 40],
  [6, 41], [7, 42],
  // Cross links
  [3, 41], [6, 42],
])

// --- Tests ---

describe('layout overlap detection', () => {
  describe('linear chain (5 nodes)', () => {
    const positioned = computeAutoLayout(linearNodes, linearEdges)

    it('no overlap at close zoom', () => {
      const overlaps = findOverlaps(positioned, 'close')
      expect(overlaps).toEqual([])
    })

    it('no overlap at medium zoom', () => {
      const overlaps = findOverlaps(positioned, 'medium')
      expect(overlaps).toEqual([])
    })

    it('no overlap at far zoom', () => {
      const overlaps = findOverlaps(positioned, 'far')
      expect(overlaps).toEqual([])
    })
  })

  describe('c3-like hierarchy (21 nodes)', () => {
    const positioned = computeAutoLayout(c3Nodes, c3Edges)

    it('no overlap at close zoom', () => {
      const overlaps = findOverlaps(positioned, 'close')
      if (overlaps.length > 0) {
        console.log('CLOSE overlaps:', overlaps.slice(0, 5))
      }
      expect(overlaps).toEqual([])
    })

    it('no overlap at medium zoom', () => {
      const overlaps = findOverlaps(positioned, 'medium')
      if (overlaps.length > 0) {
        console.log('MEDIUM overlaps:', overlaps.slice(0, 5))
      }
      expect(overlaps).toEqual([])
    })

    it('no overlap at far zoom', () => {
      const overlaps = findOverlaps(positioned, 'far')
      expect(overlaps).toEqual([])
    })
  })

  describe('realistic graph (52 nodes, 11 edges, many orphans)', () => {
    const positioned = computeAutoLayout(realisticNodes, realisticEdges)

    it('no overlap at close zoom', () => {
      const overlaps = findOverlaps(positioned, 'close')
      if (overlaps.length > 0) {
        console.log(`REALISTIC CLOSE: ${overlaps.length} overlaps, first 5:`)
        overlaps.slice(0, 5).forEach(o => console.log(`  ${o.a} x ${o.b} area=${o.area}`))
      }
      expect(overlaps).toEqual([])
    })

    it('no overlap at medium zoom', () => {
      const overlaps = findOverlaps(positioned, 'medium')
      if (overlaps.length > 0) {
        console.log(`REALISTIC MEDIUM: ${overlaps.length} overlaps`)
      }
      expect(overlaps).toEqual([])
    })
  })

  describe('wide graph (50 nodes, sparse edges)', () => {
    const positioned = computeAutoLayout(wideNodes, wideEdges)

    it('no overlap at close zoom', () => {
      const overlaps = findOverlaps(positioned, 'close')
      if (overlaps.length > 0) {
        console.log('WIDE CLOSE overlaps:', overlaps.slice(0, 5))
      }
      expect(overlaps).toEqual([])
    })

    it('no overlap at medium zoom', () => {
      const overlaps = findOverlaps(positioned, 'medium')
      if (overlaps.length > 0) {
        console.log('WIDE MEDIUM overlaps:', overlaps.slice(0, 5))
      }
      expect(overlaps).toEqual([])
    })

    it('no overlap at far zoom', () => {
      const overlaps = findOverlaps(positioned, 'far')
      expect(overlaps).toEqual([])
    })
  })
})
