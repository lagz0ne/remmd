# Panel Drawer — Design Spec

**Goal:** Click a node → right panel opens with full markdown content + edge list footer. Follow an edge → panel expands to 2 columns with navigation trace.

---

## Panel States

### Closed
No panel visible. Full canvas.

### Single Column (~45vw)
Click a node on canvas. Panel slides in from right, canvas shrinks.

- **Header:** type label (9px gray) + title (15px bold) + ✕ close
- **Content:** Full rendered markdown via `marked`. Scrollable. All headings, tables, code blocks, lists render.
- **Footer (edge list):** Collapsible. Shows all connections with state dot (green/orange/red/gray) + edge type + target name + chevron. Clickable.

### Two Columns (~70vw)
Click an edge in the footer. Panel expands. Canvas shrinks further.

- **Column 1:** Same as single column. Footer edge list unchanged — active edge highlighted with blue background.
- **Column 2:** Opens with trace header + target node content + its own edge list footer.

---

## Column 2 Trace Header

Gray bar at top of column 2 showing how it was reached:

```
via [cmd-root] → implements →
```

- Each node name is a clickable pill
- Previous nodes: gray pill (`#f4f4f5` bg, `#71717a` text)
- Current source: blue pill (`#eff6ff` bg, `#1d4ed8` text)
- Edge type shown as text between arrows
- ✕ at right end to close column 2

---

## Navigation

### Follow edge from column 2 footer
Columns shift left: col2 becomes col1, new target fills col2. Trace grows.

### Click breadcrumb in trace
That node becomes col1, next node in chain becomes col2. Trace truncates to that point.

### Click different edge in column 1 footer
Column 2 swaps to new target. Trace resets to 1 hop.

### Close column 2
Click ✕ on trace → collapse to single column.

### Close panel
Click ✕ on column 1 header or click canvas → close entire panel.

---

## Data Requirements

### Sections API
Column content comes from `useSections(docId)` — returns all sections with title + content. Rendered via `marked`.

### Edges
Footer edge list comes from the graph data — `edges` filtered by the current node ID. Each edge has: id, source, target, relationship_type, state.

---

## Sizing

- Single column width: `45vw`, min `360px`
- Two column total width: `70vw`, min `700px`
- Each column in 2-col mode: equal width (`50%`)
- Footer: auto height, collapsible
- Canvas: fills remaining space on the left

---

## Components

- `NodePanel.tsx` — The panel container. Manages column count, navigation stack.
- `NodeColumn.tsx` — Single column: header + markdown content + edge footer.
- `EdgeFooter.tsx` — Collapsible edge list with state dots.
- `TraceHeader.tsx` — Breadcrumb trace for column 2.

State managed in App.tsx:
```ts
interface PanelState {
  // Navigation stack: array of { nodeId, edgeType?, sourceNodeId? }
  stack: { nodeId: string; edgeType?: string; sourceId?: string }[]
}
// stack.length === 0 → closed
// stack.length === 1 → single column (stack[0] is the node)
// stack.length >= 2 → two columns (stack[last-1] is col1, stack[last] is col2)
```

---

## What's NOT in scope
- Edge detail view (rationale, thread, approval actions) — future
- Markdown editing — read-only for now
- Position persistence on panel open/close — canvas re-fits
