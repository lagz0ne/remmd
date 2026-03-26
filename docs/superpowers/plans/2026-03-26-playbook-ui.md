# Playbook UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make playbook types, ownership, and validation visible on the React Flow canvas with floating indicators, compact edge badges, problems-first panel, and foundation fixes (legend, empty state, hover).

**Architecture:** Backend adds 2 NATS handlers (playbook + validate). Frontend fetches both, merges into enriched graph data, renders with new PlaybookNode (3 zoom levels + floating indicators) and upgraded BundledEdge (arrows + type labels + thread counts). Panel switches between GapPanel (problems-first) and ThreadPanel (edge conversations).

**Tech Stack:** Go (NATS handlers), React 19, @xyflow/react, @phosphor-icons/react (Light weight), @tanstack/react-query, Tailwind v4

---

## File Structure

```
Backend (Go):
  internal/serve/handlers.go          — MODIFY: add remmd.q.playbook + remmd.q.validate
  internal/serve/handlers_test.go     — MODIFY: add handler tests

Frontend (React):
  npm/package.json                    — MODIFY: add @phosphor-icons/react
  npm/view/src/theme/colors.ts        — MODIFY: add validation colors
  npm/view/src/hooks/use-playbook.ts  — CREATE: fetch playbook + validation
  npm/view/src/canvas/use-graph-data.ts — MODIFY: merge playbook data
  npm/view/src/canvas/PlaybookNode.tsx — CREATE: replaces DocumentNode
  npm/view/src/canvas/GhostNode.tsx   — CREATE: missing required node
  npm/view/src/canvas/BundledEdge.tsx  — MODIFY: arrows + labels + threads
  npm/view/src/canvas/Legend.tsx       — CREATE: collapsible state legend
  npm/view/src/canvas/EmptyState.tsx   — CREATE: guided empty state
  npm/view/src/panel/GapPanel.tsx      — CREATE: problems-first validation
  npm/view/src/panel/ThreadPanel.tsx   — CREATE: edge conversation view
  npm/view/src/panel/use-panel-state.ts — MODIFY: add thread mode
  npm/view/src/App.tsx                — MODIFY: wire everything
```

---

### Task 1: Backend — NATS Handlers

**Files:**
- Modify: `internal/serve/handlers.go`
- Modify: `internal/serve/handlers_test.go`

- [ ] **Step 1: Write failing test for playbook handler**

Add to `handlers_test.go`:

```go
func TestBuildPlaybookResponse(t *testing.T) {
	t.Parallel()
	pb := &playbook.Playbook{
		Types: map[string]*playbook.TypeDef{
			"component": {
				Name:        "component",
				Description: "test",
				Fields: map[string]playbook.FieldDef{
					"goal": {Type: "string", Required: true},
				},
				Sections: []playbook.SectionDef{
					{Name: "Dependencies", Required: true},
				},
				Rules: map[string]*playbook.RuleDef{
					"edit-scope": {Name: "edit-scope", Severity: "error", Expr: "self.owner in principal.teams"},
				},
			},
		},
		Edges: map[string]*playbook.EdgeDef{
			"cites": {Name: "cites", From: []string{"component"}, To: "ref", MinCard: 1, MaxCard: -1},
		},
		Rules: map[string]*playbook.RuleDef{},
	}

	resp := buildPlaybookResponse(pb)
	if resp.Name != "" {
		t.Fatalf("expected empty name, got %q", resp.Name)
	}
	if len(resp.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(resp.Types))
	}
	comp := resp.Types["component"]
	if comp.Description != "test" {
		t.Fatalf("description: %q", comp.Description)
	}
	if len(comp.Fields) != 1 || comp.Fields["goal"] != "string!" {
		t.Fatalf("fields: %v", comp.Fields)
	}
	if len(resp.Edges) != 1 {
		t.Fatalf("edges: %v", resp.Edges)
	}
}
```

- [ ] **Step 2: Run test, verify fail**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/serve/ -run TestBuildPlaybookResponse -v`
Expected: FAIL — `buildPlaybookResponse` not defined

- [ ] **Step 3: Implement playbook + validate handlers**

Add to `handlers.go` after the graph handler:

```go
// remmd.q.playbook — active playbook schema
nc.Subscribe("remmd.q.playbook", func(msg *nats.Msg) {
	ctx := context.Background()
	pb, _, err := application.Playbooks.Latest(ctx, "default")
	if err != nil {
		replyErr(msg, err)
		return
	}
	reply(msg, buildPlaybookResponse(pb))
})

// remmd.q.validate — run playbook validation against all documents
nc.Subscribe("remmd.q.validate", func(msg *nats.Msg) {
	ctx := context.Background()
	pb, _, err := application.Playbooks.Latest(ctx, "default")
	if err != nil {
		replyErr(msg, err)
		return
	}
	docs, err := application.Docs.ListDocuments(ctx)
	if err != nil {
		replyErr(msg, err)
		return
	}
	var nodes []playbook.Node
	for _, d := range docs {
		data := map[string]any{
			"title":  d.Title,
			"status": string(d.Status),
			"source": d.Source,
			"owner":  d.OwnerID,
		}
		nodes = append(nodes, playbook.Node{Type: d.DocType, ID: d.ID, Data: data})
	}
	diags := playbook.Run(pb, nodes)
	reply(msg, buildValidationResponse(diags))
})
```

Add response types and builders at the bottom of `handlers.go`:

```go
type playbookResponse struct {
	Name    string                      `json:"name"`
	Version int                         `json:"version"`
	Types   map[string]playbookTypeResp `json:"types"`
	Edges   map[string]string           `json:"edges"`
}

type playbookTypeResp struct {
	Description string            `json:"description"`
	Fields      map[string]string `json:"fields"`
	Sections    []string          `json:"sections"`
	Rules       []string          `json:"rules"`
}

func buildPlaybookResponse(pb *playbook.Playbook) playbookResponse {
	resp := playbookResponse{
		Types: make(map[string]playbookTypeResp),
		Edges: make(map[string]string),
	}
	for name, td := range pb.Types {
		tr := playbookTypeResp{
			Description: td.Description,
			Fields:      make(map[string]string),
		}
		for fn, fd := range td.Fields {
			notation := fd.Type
			if fd.Required {
				notation += "!"
			}
			tr.Fields[fn] = notation
		}
		for _, s := range td.Sections {
			name := s.Name
			if s.Required {
				name += "!"
			}
			tr.Sections = append(tr.Sections, name)
		}
		for rn := range td.Rules {
			tr.Rules = append(tr.Rules, rn)
		}
		resp.Types[name] = tr
	}
	for name, ed := range pb.Edges {
		from := ed.From[0]
		if len(ed.From) > 1 {
			from = strings.Join(ed.From, "|")
		}
		max := "*"
		if ed.MaxCard >= 0 {
			max = fmt.Sprintf("%d", ed.MaxCard)
		}
		resp.Edges[name] = fmt.Sprintf("%s -> %s [%d..%s]", from, ed.To, ed.MinCard, max)
	}
	return resp
}

type validationResponse struct {
	Diagnostics []validationDiag `json:"diagnostics"`
}

type validationDiag struct {
	Rule     string `json:"rule"`
	NodeID   string `json:"node_id"`
	NodeType string `json:"node_type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func buildValidationResponse(diags []playbook.Diagnostic) validationResponse {
	resp := validationResponse{Diagnostics: make([]validationDiag, 0, len(diags))}
	for _, d := range diags {
		resp.Diagnostics = append(resp.Diagnostics, validationDiag{
			Rule:     d.Rule,
			NodeID:   d.NodeID,
			NodeType: d.NodeType,
			Severity: d.Severity,
			Message:  d.Message,
		})
	}
	return resp
}
```

Add import for `"fmt"` and `"github.com/lagz0ne/remmd/internal/playbook"` at the top.

- [ ] **Step 4: Run tests**

Run: `cd /home/lagz0ne/dev/remmd && go test ./internal/serve/ -v`
Expected: all PASS

- [ ] **Step 5: Run full Go tests**

Run: `cd /home/lagz0ne/dev/remmd && go test ./... -count=1`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/serve/handlers.go internal/serve/handlers_test.go
git commit -m "feat(serve): remmd.q.playbook + remmd.q.validate NATS handlers"
```

---

### Task 2: Add Phosphor Icons + Color Tokens

**Files:**
- Modify: `npm/package.json`
- Modify: `npm/view/src/theme/colors.ts`

- [ ] **Step 1: Add Phosphor dependency**

```bash
cd /home/lagz0ne/dev/remmd/npm && bun add @phosphor-icons/react
```

- [ ] **Step 2: Extend color tokens**

Replace `npm/view/src/theme/colors.ts` with:

```ts
export const stateColor: Record<string, string> = {
  aligned: 'var(--aligned)',
  stale: 'var(--stale)',
  pending: 'var(--pending)',
  broken: 'var(--broken)',
  archived: 'var(--archived)',
}

export const validationBg: Record<string, string> = {
  pass: '#f0fdf4',
  error: '#fef2f2',
  warning: '#fffbeb',
  neutral: '#fafafa',
}

export const validationBorder: Record<string, string> = {
  pass: '#dcfce7',
  error: '#fecaca',
  warning: '#fef3c7',
  neutral: '#e4e4e7',
}

export const validationText: Record<string, string> = {
  pass: '#16a34a',
  error: '#dc2626',
  warning: '#d97706',
  neutral: '#a1a1aa',
}
```

- [ ] **Step 3: Commit**

```bash
git add npm/package.json npm/bun.lock npm/view/src/theme/colors.ts
git commit -m "feat(view): add Phosphor icons dep + validation color tokens"
```

---

### Task 3: Data Hooks — usePlaybook + useValidation

**Files:**
- Create: `npm/view/src/hooks/use-playbook.ts`

- [ ] **Step 1: Create playbook + validation hooks**

```ts
import { useQuery } from '@tanstack/react-query'
import { natsRequest } from '../nats'

interface PlaybookType {
  description: string
  fields: Record<string, string>
  sections: string[]
  rules: string[]
}

export interface PlaybookResponse {
  name: string
  version: number
  types: Record<string, PlaybookType>
  edges: Record<string, string>
}

interface ValidationDiag {
  rule: string
  node_id: string
  node_type: string
  severity: string
  message: string
}

interface ValidationResponse {
  diagnostics: ValidationDiag[]
}

export function usePlaybook() {
  return useQuery({
    queryKey: ['playbook'],
    queryFn: () => natsRequest<PlaybookResponse>('remmd.q.playbook'),
    staleTime: Infinity,
    retry: false,
  })
}

export function useValidation() {
  return useQuery({
    queryKey: ['validation'],
    queryFn: () => natsRequest<ValidationResponse>('remmd.q.validate'),
    staleTime: 30_000,
    retry: false,
  })
}

export type { PlaybookType, ValidationDiag }
```

- [ ] **Step 2: Commit**

```bash
git add npm/view/src/hooks/use-playbook.ts
git commit -m "feat(view): usePlaybook + useValidation data hooks"
```

---

### Task 4: Enrich Graph Data with Playbook + Validation

**Files:**
- Modify: `npm/view/src/canvas/use-graph-data.ts`

- [ ] **Step 1: Extend DocNodeData and BundleEdgeData interfaces**

Add to `DocNodeData`:
```ts
export interface DocNodeData extends Record<string, unknown> {
  title: string
  status: string
  source: string
  worstState: string
  sectionCount: number
  edgeCounts: Record<string, number>
  // Playbook enrichment
  playbookType: string
  owner: string
  brief: string
  validationPassing: number
  validationTotal: number
  validationErrors: { rule: string; message: string }[]
  outgoing: number
  incoming: number
  threadCount: number
  portWorstState: string
}
```

Add to `BundleEdgeData`:
```ts
export interface BundleEdgeData extends Record<string, unknown> {
  links: GraphEdge[]
  worstState: string
  count: number
  edgeType: string
  threadCount: number
  isStructural: boolean
}
```

- [ ] **Step 2: Update transformGraph to accept playbook + validation**

Update the `transformGraph` function signature and body to merge playbook type, validation diagnostics, and edge metadata. The function receives optional playbook and validation data and enriches each node/edge.

Key changes:
- Each node gets `playbookType` from matching `doc_type` to playbook types
- Each node gets `validationErrors` from matching diagnostics by `node_id`
- Each edge bundle gets `edgeType` from playbook edge definitions (match by relationship_type)
- Compute `outgoing`/`incoming`/`threadCount` per node from edges
- `isStructural` = true for "contains", "parent-of" edge types

- [ ] **Step 3: Update useGraphData to fetch playbook + validation**

```ts
export function useGraphData() {
  const { data: raw, ...graphRest } = useQuery({
    queryKey: ['graph'],
    queryFn: () => natsRequest<GraphResponse>('remmd.q.graph'),
    staleTime: Infinity,
  })

  const { data: pb } = usePlaybook()
  const { data: validation } = useValidation()

  const { nodes, edges } = useMemo(() => {
    if (!raw) return { nodes: [], edges: [] }
    return transformGraph(raw, pb, validation)
  }, [raw, pb, validation])

  return { nodes, edges, ...graphRest }
}
```

Add import for `usePlaybook` and `useValidation`.

- [ ] **Step 4: Commit**

```bash
git add npm/view/src/canvas/use-graph-data.ts
git commit -m "feat(view): enrich graph nodes with playbook type, validation, port summary"
```

---

### Task 5: PlaybookNode — Three-Zoom Node with Floating Indicators

**Files:**
- Create: `npm/view/src/canvas/PlaybookNode.tsx`

- [ ] **Step 1: Create PlaybookNode component**

Full component with:
- `FarView`: title only, 8px, border = state color, dashed = orphan, bg tint for errors
- `MediumView`: type label + team tag (inline) + title + brief + floating validation TL + floating ports BR + hover lift
- `CloseView`: all of medium + section list with state bars + validation error detail box + fade gradient
- Floating validation: Phosphor `CheckCircle` (Light) or `WarningCircle` (Light) + fraction
- Floating ports: Phosphor `ArrowRight`, `ArrowLeft`, `ChatCircle` (Light) + counts
- Styling per spec: 1.5px borders, 12px radius, 16/20px padding, whisper shadows

Use `@phosphor-icons/react` imports:
```ts
import { CheckCircle, WarningCircle, ArrowRight, ArrowLeft, ChatCircle } from '@phosphor-icons/react'
```

All icons rendered at `weight="light"` and `size={11}` (validation) or `size={10}` (ports).

- [ ] **Step 2: Commit**

```bash
git add npm/view/src/canvas/PlaybookNode.tsx
git commit -m "feat(view): PlaybookNode — three-zoom node with Phosphor Light floating indicators"
```

---

### Task 6: GhostNode — Missing Required Node Placeholder

**Files:**
- Create: `npm/view/src/canvas/GhostNode.tsx`

- [ ] **Step 1: Create GhostNode component**

```tsx
import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Plus } from '@phosphor-icons/react'

export interface GhostNodeData extends Record<string, unknown> {
  typeName: string
  requiredBy: string
}

function GhostNodeInner({ data }: NodeProps & { data: GhostNodeData }) {
  return (
    <div
      className="rounded-xl px-5 py-4 cursor-pointer transition-shadow hover:shadow-[0_4px_12px_rgba(99,102,241,0.08)]"
      style={{
        border: '1.5px dashed #c7d2fe',
        background: '#eef2ff',
        minWidth: 170,
      }}
    >
      <Handle type="target" position={Position.Left} className="opacity-0" />
      <div className="text-[9px] text-indigo-400 font-normal">{data.typeName}</div>
      <div className="flex items-center gap-1 mt-1.5">
        <Plus size={12} weight="light" className="text-indigo-500" />
        <span className="text-[13px] font-medium text-indigo-500">add {data.typeName}</span>
      </div>
      <div className="text-[9px] text-indigo-300 mt-1">required by {data.requiredBy}</div>
      <Handle type="source" position={Position.Right} className="opacity-0" />
    </div>
  )
}

export const GhostNode = memo(GhostNodeInner)
```

- [ ] **Step 2: Commit**

```bash
git add npm/view/src/canvas/GhostNode.tsx
git commit -m "feat(view): GhostNode — indigo dashed placeholder for missing required nodes"
```

---

### Task 7: BundledEdge — Arrows + Type Labels + Thread Counts

**Files:**
- Modify: `npm/view/src/canvas/BundledEdge.tsx`

- [ ] **Step 1: Rewrite BundledEdge**

Replace the entire file with the upgraded version:
- Add SVG `<marker>` for arrowheads (colored per state)
- Replace count-only label with compact two-part badge: type name | thread count
- Structural edges (contains, parent-of): dashed stroke, gray
- Ghost edges: indigo dashed
- Hover: opacity 0.4 → 0.8 transition
- Use `markerEnd` prop on the path

The edge label uses HTML via `EdgeLabelRenderer`:
- Left part: white bg, state-colored border, edge type name
- Right part (if threads > 0): tinted bg, state-colored, thread count
- Aligned + 0 threads: type name + Phosphor CheckCircle checkmark

- [ ] **Step 2: Commit**

```bash
git add npm/view/src/canvas/BundledEdge.tsx
git commit -m "feat(view): BundledEdge — directed arrows, type labels, thread count badges"
```

---

### Task 8: Legend — Collapsible State Legend with Type Filter

**Files:**
- Create: `npm/view/src/canvas/Legend.tsx`

- [ ] **Step 1: Create Legend component**

Collapsible panel positioned bottom-left (above Controls via style offset):
- Shows state colors (aligned, stale, broken, pending) with Phosphor circle icons
- Shows border styles (solid = linked, dashed = orphan)
- Shows badge meanings
- Click type name = filter (emit callback, parent dims non-matching nodes)
- Collapsed state persists in localStorage key `remmd-legend-collapsed`
- Uses Phosphor `CaretDown`/`CaretUp` for collapse toggle

- [ ] **Step 2: Commit**

```bash
git add npm/view/src/canvas/Legend.tsx
git commit -m "feat(view): Legend — collapsible state/type legend with filter"
```

---

### Task 9: EmptyState — Guided Empty Canvas

**Files:**
- Create: `npm/view/src/canvas/EmptyState.tsx`

- [ ] **Step 1: Create EmptyState component**

```tsx
export function EmptyState() {
  return (
    <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
      <div className="text-center max-w-[400px]">
        <div className="text-sm font-semibold text-zinc-800">No documents yet</div>
        <div className="text-[11px] text-zinc-400 mt-2 leading-relaxed">
          Import a playbook to define your graph structure,
          then create documents that follow it.
        </div>
        <div className="mt-3 space-y-1.5">
          <code className="block text-[10px] bg-zinc-100 px-3 py-1.5 rounded text-zinc-600 font-mono">
            remmd playbook import c3.playbook.yaml
          </code>
          <code className="block text-[10px] bg-zinc-100 px-3 py-1.5 rounded text-zinc-600 font-mono">
            remmd doc create "My First Doc"
          </code>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add npm/view/src/canvas/EmptyState.tsx
git commit -m "feat(view): EmptyState — guided empty canvas with CLI commands"
```

---

### Task 10: GapPanel — Problems-First Validation Panel

**Files:**
- Create: `npm/view/src/panel/GapPanel.tsx`

- [ ] **Step 1: Create GapPanel component**

Props: `docId`, `docTitle`, `playbookType`, `owner`, `validationErrors`, `validationPassing`, `validationTotal`

Structure:
- Header: type label + issue count badge + title + close button + owner subtitle
- "Needs attention" section (conditionally shown): red-tinted bg, issue cards with title + description
- "N passing" section (collapsed by default): expand/collapse with localStorage persistence (`remmd-passing-collapsed`)
- Uses Phosphor `CheckCircle`, `WarningCircle`, `CaretDown`, `CaretUp` (Light)

- [ ] **Step 2: Commit**

```bash
git add npm/view/src/panel/GapPanel.tsx
git commit -m "feat(view): GapPanel — problems-first validation panel"
```

---

### Task 11: ThreadPanel — Edge Conversation View

**Files:**
- Create: `npm/view/src/panel/ThreadPanel.tsx`

- [ ] **Step 1: Create ThreadPanel component**

Props: `links` (the edge's link data), `currentDocId`, `connectedDocTitle`

Structure:
- Header: state badge + edge type + connected doc names + "changed N ago"
- Diff summary bar (yellow tint): what content changed
- Thread entries: team avatar (colored initials circle) + name + timestamp + message bubble
- System events: gray dot + "Link went stale" type messages
- Action bar: Reaffirm (black button), Comment (outline), Withdraw (red outline)
- Uses Phosphor `ChatCircle`, `ArrowClockwise`, `X` (Light)

- [ ] **Step 2: Commit**

```bash
git add npm/view/src/panel/ThreadPanel.tsx
git commit -m "feat(view): ThreadPanel — edge conversation with actions"
```

---

### Task 12: Wire Everything in App.tsx

**Files:**
- Modify: `npm/view/src/panel/use-panel-state.ts`
- Modify: `npm/view/src/App.tsx`

- [ ] **Step 1: Add 'thread' mode to panel state**

Update `PanelMode` type:
```ts
export type PanelMode = 'closed' | 'doc' | 'edge' | 'thread'
```

Add `selectThread` callback that sets mode to 'thread' with edge data.

- [ ] **Step 2: Update App.tsx**

Replace `DocumentNode` with `PlaybookNode` and `GhostNode` in node types:
```ts
const nodeTypes = { document: PlaybookNode, ghost: GhostNode }
```

Add Legend, EmptyState, and wire GapPanel + ThreadPanel into the panel overlay:
- When `mode === 'doc'`: show GapPanel (replaces DocPanel)
- When `mode === 'edge'` or `mode === 'thread'`: show ThreadPanel + GapPanel side by side

Add EmptyState when `nodes.length === 0`:
```tsx
{nodes.length === 0 && <EmptyState />}
```

Add Legend component above Controls:
```tsx
<Legend playbookTypes={...} onFilterType={...} />
```

- [ ] **Step 3: Commit**

```bash
git add npm/view/src/panel/use-panel-state.ts npm/view/src/App.tsx
git commit -m "feat(view): wire PlaybookNode, GapPanel, ThreadPanel, Legend, EmptyState"
```

---

### Task 13: Build Verification

- [ ] **Step 1: Run Go tests**

```bash
cd /home/lagz0ne/dev/remmd && go test ./... -count=1
```
Expected: all PASS

- [ ] **Step 2: Run frontend build**

```bash
cd /home/lagz0ne/dev/remmd/npm && bun run build
```
Expected: no TypeScript errors, build succeeds

- [ ] **Step 3: Final commit if fixups needed**

```bash
git add -A && git commit -m "chore: build fixes"
```
