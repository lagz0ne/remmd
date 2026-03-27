# Manual Layout with Position Persistence

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make manual node positioning work — drag to place, positions saved to DB, restored on reload. Dagre is the fallback for nodes without saved positions only.

**Architecture:** Fix `use-layout.ts` to actually apply loaded positions. Dagre only runs for nodes that have no saved position. On drag stop, save position. Re-layout button clears all saved positions and re-runs dagre.

**Tech Stack:** React, @xyflow/react, @dagrejs/dagre, NATS (existing handlers)

---

### Task 1: Fix layout hook to apply saved positions

**Files:**
- Modify: `npm/view/src/canvas/use-layout.ts`
- Test: `npm/view/src/canvas/__tests__/layout-overlap.test.ts` (existing, must still pass)

- [ ] **Step 1: Rewrite the layout effect**

Replace the entire `use-layout.ts` with:

```ts
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
```

Key changes from current code:
- Saved positions are actually applied (was fetched then ignored)
- Dagre only runs for nodes WITHOUT saved positions
- `resetLayout` uses refs instead of closure-captured `nodes`/`edges` (fixes stale closure)
- No dependency on `nodes`/`edges` in `resetLayout` deps (stable callback)

- [ ] **Step 2: Run existing tests**

```bash
cd /home/lagz0ne/dev/remmd/npm && bun test view/src/canvas/__tests__/layout-overlap.test.ts
```

Expected: 11 pass (dagre logic unchanged, only the orchestration changed)

- [ ] **Step 3: Run TypeScript check**

```bash
cd /home/lagz0ne/dev/remmd/npm && bunx @typescript/native-preview -p view/tsconfig.json --noEmit 2>&1 | grep -v bun:test
```

Expected: 0 errors

- [ ] **Step 4: Verify with agent-browser**

```bash
# Start server if not running
agent-browser open http://localhost:4126
# Wait for load
sleep 4
# Drag a node, reload, verify position persists
# Click Re-layout, verify all nodes reposition
```

Verification checklist:
1. Page loads — nodes appear (dagre positions if no saved, saved positions if exist)
2. Drag a node → release → reload page → node is in the dragged position
3. Click Re-layout → all nodes reposition via dagre, saved positions cleared
4. Drag several nodes → reload → all dragged positions preserved

- [ ] **Step 5: Commit**

```bash
git add npm/view/src/canvas/use-layout.ts
git commit -m "fix(view): restore manual layout — saved positions applied, dagre fills gaps only"
```

---

### Task 2: Clear stale positions on first load

The DB may have stale positions from before the dagre tuning. Clear them once so users start fresh.

**Files:**
- Modify: `npm/view/src/canvas/use-layout.ts`

- [ ] **Step 1: Add a one-time clear on first mount if positions look stale**

Actually, the simpler approach: the Re-layout button already clears positions. Instead of auto-detecting stale positions, just clear them in the DB directly:

```bash
cd /home/lagz0ne/dev/remmd
go run ./cmd/remmd --db ~/.remmd/remmd.db health  # verify DB access
```

Then add a one-liner to the position load: if there ARE saved positions but they were saved before a certain commit (we can't detect this), just let the user click Re-layout.

No code change needed — the Re-layout button already works. Skip this task.

- [ ] **Step 2: Commit**

Already handled by Task 1.
