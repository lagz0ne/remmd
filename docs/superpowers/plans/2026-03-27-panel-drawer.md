# Panel Drawer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Click a node → right panel opens with full markdown content + edge list footer. Follow an edge → panel expands to 2 columns with navigation trace.

**Architecture:** 4 incremental tasks. Each produces a working, verifiable UI. Panel state managed in App.tsx via a navigation stack. Content rendered with `marked`. New components: NodeColumn, EdgeFooter, TraceHeader, NodePanel.

**Tech Stack:** React 19, @xyflow/react, marked, Tailwind v4

---

## File Structure

```
npm/view/src/
├── App.tsx              — MODIFY: add panel state, onNodeClick, panel rendering
├── panel/
│   ├── NodePanel.tsx    — CREATE: panel container (1 or 2 columns)
│   ├── NodeColumn.tsx   — CREATE: single column (header + markdown + footer)
│   ├── EdgeFooter.tsx   — CREATE: collapsible edge list
│   └── TraceHeader.tsx  — CREATE: breadcrumb trace for column 2
```

---

### Task 1: Single Column Panel — Click Node → See Markdown

**Files:**
- Create: `npm/view/src/panel/NodeColumn.tsx`
- Create: `npm/view/src/panel/NodePanel.tsx`
- Modify: `npm/view/src/App.tsx`

- [ ] **Step 1: Create NodeColumn — header + markdown content**

Create `npm/view/src/panel/NodeColumn.tsx`:

```tsx
import { useSections } from '../hooks'
import { marked } from 'marked'
import { useMemo } from 'react'

marked.setOptions({ breaks: true, gfm: true })

function unescape(s: string): string {
  return s.replace(/\\n/g, '\n')
}

interface NodeColumnProps {
  docId: string
  title: string
  typeName: string
  onClose?: () => void
  header?: React.ReactNode
}

export function NodeColumn({ docId, title, typeName, onClose, header }: NodeColumnProps) {
  const { data } = useSections(docId)
  const sections = data?.sections ?? []

  const html = useMemo(() => {
    if (sections.length === 0) return ''
    const md = sections.map(s => s.content || '').join('\n\n')
    return marked.parse(unescape(md)) as string
  }, [sections])

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', minWidth: 0 }}>
      {header}
      <div style={{
        padding: '12px 18px',
        borderBottom: '1px solid #f4f4f5',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
      }}>
        <div>
          {typeName && <div style={{ fontSize: 9, color: '#a1a1aa', letterSpacing: '0.03em' }}>{typeName}</div>}
          <div style={{ fontSize: 15, fontWeight: 600, color: '#18181b', marginTop: 2 }}>{title}</div>
        </div>
        {onClose && (
          <span onClick={onClose} style={{ fontSize: 12, color: '#a1a1aa', cursor: 'pointer', padding: '2px 4px' }}>✕</span>
        )}
      </div>
      <div
        className="doc-node-md"
        style={{
          flex: 1,
          padding: '14px 18px',
          overflowY: 'auto',
          fontSize: 12,
          color: '#52525b',
          lineHeight: 1.7,
        }}
        dangerouslySetInnerHTML={{ __html: html }}
      />
    </div>
  )
}
```

- [ ] **Step 2: Create NodePanel — container that renders 1 column**

Create `npm/view/src/panel/NodePanel.tsx`:

```tsx
import { NodeColumn } from './NodeColumn'

interface StackEntry {
  nodeId: string
  title: string
  typeName: string
  edgeType?: string
  sourceId?: string
}

interface NodePanelProps {
  stack: StackEntry[]
  onClose: () => void
}

export function NodePanel({ stack, onClose }: NodePanelProps) {
  if (stack.length === 0) return null

  const current = stack[stack.length - 1]

  return (
    <div style={{
      position: 'absolute',
      top: 0,
      right: 0,
      bottom: 0,
      width: '45vw',
      minWidth: 360,
      background: 'white',
      borderLeft: '1px solid #e4e4e7',
      zIndex: 10,
      display: 'flex',
      fontFamily: 'system-ui, -apple-system, sans-serif',
    }}>
      <NodeColumn
        docId={current.nodeId}
        title={current.title}
        typeName={current.typeName}
        onClose={onClose}
      />
    </div>
  )
}

export type { StackEntry }
```

- [ ] **Step 3: Wire panel into App.tsx**

Add to App.tsx imports:
```tsx
import { NodePanel, type StackEntry } from './panel/NodePanel'
```

Add state and click handler inside Canvas():
```tsx
const [panelStack, setPanelStack] = useState<StackEntry[]>([])

const onNodeClick = (_: any, node: Node) => {
  setPanelStack([{
    nodeId: node.id,
    title: (node.data as any).title || node.id,
    typeName: (node.data as any).playbookType || '',
  }])
}

const onPaneClick = () => setPanelStack([])
```

Add to `<ReactFlow>` props:
```tsx
onNodeClick={onNodeClick}
onPaneClick={onPaneClick}
```

Add panel after `</ReactFlow>`:
```tsx
<NodePanel stack={panelStack} onClose={() => setPanelStack([])} />
```

- [ ] **Step 4: Verify — TS check + rebuild + visual test**

```bash
cd /home/lagz0ne/dev/remmd/npm && bunx @typescript/native-preview -p view/tsconfig.json --noEmit
```

Rebuild server, open browser, click a node. Panel should open on the right with full markdown content. Click canvas to close. Drag should still work.

- [ ] **Step 5: Commit**

```bash
git add npm/view/src/panel/NodeColumn.tsx npm/view/src/panel/NodePanel.tsx npm/view/src/App.tsx
git commit -m "feat(view): single column panel — click node to see full markdown"
```

---

### Task 2: Edge Footer — Show Connections Below Content

**Files:**
- Create: `npm/view/src/panel/EdgeFooter.tsx`
- Modify: `npm/view/src/panel/NodeColumn.tsx`

- [ ] **Step 1: Create EdgeFooter**

Create `npm/view/src/panel/EdgeFooter.tsx`:

```tsx
import { useState } from 'react'

interface EdgeItem {
  id: string
  targetId: string
  targetTitle: string
  edgeType: string
  state: string
  direction: 'outgoing' | 'incoming'
}

interface EdgeFooterProps {
  edges: EdgeItem[]
  activeEdgeId?: string
  onEdgeClick: (edge: EdgeItem) => void
}

const stateColors: Record<string, string> = {
  aligned: '#2e7d32',
  stale: '#f57c00',
  pending: '#616161',
  broken: '#c62828',
}

export function EdgeFooter({ edges, activeEdgeId, onEdgeClick }: EdgeFooterProps) {
  const [collapsed, setCollapsed] = useState(false)

  if (edges.length === 0) return null

  return (
    <div style={{ borderTop: '1px solid #e4e4e7' }}>
      <div
        onClick={() => setCollapsed(!collapsed)}
        style={{
          padding: '8px 18px',
          fontSize: 10,
          color: '#71717a',
          fontWeight: 500,
          cursor: 'pointer',
          display: 'flex',
          justifyContent: 'space-between',
        }}
      >
        <span>{edges.length} connection{edges.length !== 1 ? 's' : ''}</span>
        <span style={{ fontSize: 9, color: '#a1a1aa' }}>{collapsed ? '▸' : '▾'}</span>
      </div>
      {!collapsed && (
        <div style={{ padding: '0 18px 10px' }}>
          {edges.map(edge => (
            <div
              key={edge.id}
              onClick={() => onEdgeClick(edge)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '6px 8px',
                borderRadius: 4,
                cursor: 'pointer',
                marginBottom: 2,
                background: edge.id === activeEdgeId ? '#eff6ff' : 'transparent',
              }}
            >
              <div style={{
                width: 5, height: 5, borderRadius: '50%',
                background: stateColors[edge.state] || '#a1a1aa',
              }} />
              <div style={{ fontSize: 11, color: edge.id === activeEdgeId ? '#1d4ed8' : '#52525b', flex: 1 }}>
                <span style={{ color: edge.id === activeEdgeId ? '#93c5fd' : '#a1a1aa' }}>
                  {edge.direction === 'outgoing' ? `${edge.edgeType} →` : `← ${edge.edgeType}`}
                </span>{' '}
                <span style={{ fontWeight: 500 }}>{edge.targetTitle}</span>
              </div>
              <span style={{ fontSize: 9, color: stateColors[edge.state] || '#a1a1aa' }}>{edge.state}</span>
              <span style={{ fontSize: 10, color: '#d4d4d8' }}>›</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export type { EdgeItem }
```

- [ ] **Step 2: Add EdgeFooter to NodeColumn**

Update `NodeColumn` props to accept edges and an edge click handler:

```tsx
interface NodeColumnProps {
  docId: string
  title: string
  typeName: string
  onClose?: () => void
  header?: React.ReactNode
  footer?: React.ReactNode
}
```

Add `{footer}` after the content div, before the closing `</div>`.

- [ ] **Step 3: Wire edges in App.tsx**

Build the edge list for the selected node from the React Flow edges array. Pass it to NodePanel, which passes it to NodeColumn as a footer.

In `NodePanel`, compute edges for the current node from the graph edges, build `EdgeItem[]`, and render `EdgeFooter` as the `footer` prop of `NodeColumn`.

This requires NodePanel to receive the full `edges` and `nodes` arrays so it can resolve target titles.

- [ ] **Step 4: Verify + commit**

Click a node → panel opens with markdown. Below the content, the footer shows connections with state dots. Click canvas to close.

```bash
git add npm/view/src/panel/EdgeFooter.tsx npm/view/src/panel/NodeColumn.tsx npm/view/src/panel/NodePanel.tsx npm/view/src/App.tsx
git commit -m "feat(view): edge footer — connections list below markdown content"
```

---

### Task 3: Two Columns — Follow an Edge

**Files:**
- Modify: `npm/view/src/panel/NodePanel.tsx`
- Create: `npm/view/src/panel/TraceHeader.tsx`

- [ ] **Step 1: Create TraceHeader**

Create `npm/view/src/panel/TraceHeader.tsx`:

```tsx
interface TraceEntry {
  nodeId: string
  title: string
  edgeType?: string
}

interface TraceHeaderProps {
  entries: TraceEntry[]
  onJump: (index: number) => void
  onClose: () => void
}

export function TraceHeader({ entries, onJump, onClose }: TraceHeaderProps) {
  return (
    <div style={{
      padding: '6px 18px',
      background: '#f8fafc',
      borderBottom: '1px solid #e2e8f0',
      display: 'flex',
      alignItems: 'center',
      gap: 5,
      fontSize: 9,
      color: '#64748b',
      flexWrap: 'wrap',
    }}>
      <span style={{ color: '#a1a1aa' }}>via</span>
      {entries.map((entry, i) => (
        <span key={i} style={{ display: 'contents' }}>
          <span
            onClick={() => onJump(i)}
            style={{
              background: i === entries.length - 1 ? '#eff6ff' : '#f4f4f5',
              color: i === entries.length - 1 ? '#1d4ed8' : '#71717a',
              padding: '1px 6px',
              borderRadius: 3,
              fontWeight: 500,
              cursor: 'pointer',
            }}
          >
            {entry.title}
          </span>
          {entry.edgeType && (
            <>
              <span style={{ color: '#d4d4d8' }}>→</span>
              <span>{entry.edgeType}</span>
            </>
          )}
          <span style={{ color: '#d4d4d8' }}>→</span>
        </span>
      ))}
      <span onClick={onClose} style={{ marginLeft: 'auto', cursor: 'pointer', color: '#a1a1aa' }}>✕</span>
    </div>
  )
}
```

- [ ] **Step 2: Update NodePanel for 2 columns**

When `stack.length >= 2`, render two columns side by side. Panel width expands to `70vw`.

Column 1 = `stack[stack.length - 2]` with its own EdgeFooter (active edge highlighted).
Column 2 = `stack[stack.length - 1]` with TraceHeader + its own EdgeFooter.

The `onEdgeClick` in column 1's footer pushes a new entry onto the stack.
The `onEdgeClick` in column 2's footer shifts: remove last, push new source + new target.

- [ ] **Step 3: Wire navigation**

In EdgeFooter click handler:
- Column 1 edge click: push new entry to stack (expand to 2 cols or swap col 2)
- Column 2 edge click: shift columns (col2 becomes col1, new target fills col2)

TraceHeader jump: truncate stack to the jumped index + 1, then push the next entry.

TraceHeader close: pop last entry from stack (collapse to 1 col).

- [ ] **Step 4: Verify + commit**

Click node → 1 column. Click edge in footer → expands to 2 columns with trace. Click breadcrumb → jumps back. Click ✕ on trace → collapses. Drag still works.

```bash
git add npm/view/src/panel/TraceHeader.tsx npm/view/src/panel/NodePanel.tsx
git commit -m "feat(view): two-column panel with edge navigation trace"
```

---

### Task 4: Polish — Panel Width, Edge Highlight, Canvas Interaction

**Files:**
- Modify: `npm/view/src/panel/NodePanel.tsx`
- Modify: `npm/view/src/App.tsx`

- [ ] **Step 1: Highlight selected node on canvas**

When panel is open, the selected node should have a visual highlight. In App.tsx, use React Flow's `selected` prop or add a class.

- [ ] **Step 2: Click different node while panel is open**

Clicking a different node on the canvas should reset the stack to that node (single column).

- [ ] **Step 3: Transition animation**

Add CSS transition on panel width changes (45vw ↔ 70vw):
```css
transition: width 0.2s ease;
```

- [ ] **Step 4: Verify everything + commit**

Full test: click node (1 col), click edge (2 col), follow edge in col 2 (shift), click breadcrumb (jump back), click canvas (close), click different node (reset), drag nodes (still works).

```bash
git add npm/view/src/
git commit -m "feat(view): panel polish — node highlight, transitions, canvas interaction"
```
