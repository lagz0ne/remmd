# Playbook UI Manifestation — Design Spec

**Goal:** Make the playbook's types, ownership, and validation rules visible on the canvas and actionable in the panel, while fixing foundation UX gaps (legend, empty states, discoverability).

**Design Principles:**
- Quiet type, loud state — problems draw the eye, taxonomy is context
- Problems-first panel — issues surface at top with actionable descriptions
- Progressive disclosure — more playbook info at closer zoom levels
- Editorial style — sparse, grayscale foundation, color only for state
- Light touch — Phosphor Light icons, 1.5px borders, whisper shadows, maximum air

**Icon Set:** Phosphor Icons (`@phosphor-icons/react`) — Light weight (1.25px stroke, native 16x16 grid)

---

## Canvas Nodes (Quiet Type, Loud State)

### Visual structure

Card content is clean — type, title, brief, team tag.
Two floating indicators orbit the card:
- **Top-left float:** self-validation state (fraction passing)
- **Bottom-right float:** port summary (outgoing, incoming, threads)

### Card styling
- Border: 1.5px solid, zinc-300 default, colored only when link state demands it
- Border-radius: 12px
- Padding: 16px vertical, 20px horizontal
- Background: white (healthy), near-white tint for errors (#fefefe)
- Shadow: none on card (floats have 0 1px 2px at 4% opacity)
- Orphan: 1.5px dashed zinc-300, bg #fafafa

### Floating validation indicator (top-left)
- Position: absolute, -14px top, -6px left (overlaps card corner)
- Background: tinted near-white (#f0fdf4 green, #fef2f2 red, #fafafa gray)
- Border: 1px solid matching tint
- Border-radius: 8px
- Content: Phosphor check-circle or alert-circle icon (11px) + fraction "4/4"
- Font: 9px, weight 500, state-colored
- Shadow: 0 1px 2px rgba(0,0,0,0.04)

### Floating port summary (bottom-right)
- Position: absolute, -14px bottom, -6px right
- Individual pills with 4px gap between them
- Each pill: white bg, 1px zinc-200 border (or state-colored border if stale/broken), 8px radius
- Content: Phosphor arrow-right (10px) + count, arrow-left + count, message-square + count
- Font: 9px, weight 500, zinc-500
- Port pill border inherits worst state color of those edges
- Orphans: no port summary shown
- Thread count only shown when > 0

### Node anatomy at each zoom level

**Far zoom** (<0.35):
- Title text only (8px, truncated)
- Border color = worst link state
- Border style: solid = linked, dashed = orphan
- Background tint for errors
- No floats, no type, no owner

**Medium zoom** (0.35–0.85):
- Type label: 9px, zinc-400, weight 400 (e.g., "component", "ref")
- Team tag: 8px gray pill, weight 500, right-aligned on same line as type
- Title: 13-15px, weight 600
- Content brief: 10px, zinc-400, first line of goal field, truncated
- Floating validation indicator (top-left)
- Floating port summary (bottom-right)
- Hover: subtle lift shadow (0 4px 12px rgba(0,0,0,0.06)), cursor pointer

**Close zoom** (>0.85):
- All of medium, plus:
- Full markdown content preview (fades out at bottom with gradient)
- Section list with left-bar state indicators
- Missing required sections shown as italic placeholder with "missing" label
- Validation error detail box (red-tinted, specific failure messages)

### Ghost nodes (missing required nodes)
- Dashed indigo border (#c7d2fe), bg #eef2ff
- Type label in indigo
- Title: "+ add component" in indigo
- Subtitle: "required by container cli"
- Clickable to create

---

## Edges & Connections

### Edge visual encoding
- **Direction:** arrowhead marker on target end
- **Color:** worst link state in the bundle (green/orange/red)
- **Width:** single link = 2px, bundled = min(2 + count, 6)px
- **Structural edges** (contains, parent-of): dashed stroke, gray, quieter
- **Trust edges** (cites, implements): solid stroke, state-colored
- **Ghost edges:** indigo dashed, to ghost nodes

### Edge labels (compact badges)
- Positioned at edge midpoint
- Two-part pill: type name | thread count
- Type part: white bg, state-colored border and text
- Thread count part: tinted bg, state-colored, only shown when > 0
- Aligned edges with 0 threads: type + checkmark (e.g., "cites ✓")
- Bundled: "3×cites" with combined thread count
- Contains edges: abbreviated "cnt" label, gray

### Edge thread indicator
- Thread count in the edge label pill (Phosphor message-square icon + number)
- Click edge opens panel with full thread view
- Thread panel shows: system events + comments + actions (Reaffirm / Comment / Withdraw)
- Thread entries show team badge + agent/user name + timestamp
- System events use gray dot, comments use team-colored avatar

---

## Panel (Problems-First Gap View)

### Header
- Type label (9px, zinc-400)
- Issue count badge (colored pill)
- Title (14px, semibold)
- Close button
- Owner + version subtitle

### Body — "Needs attention" (top, conditionally shown)
- Background: #fef2f2 (near-white red tint)
- Each issue: card with 1px #fecaca border, white bg
  - Title: 11px semibold (what's wrong)
  - Description: 10px zinc-500 (how to fix, actionable)
- Sorted: errors first, then warnings
- Hidden entirely when 0 issues

### Body — "N passing" (collapsed)
- Default: collapsed to single line "5 passing" + expand arrow + Phosphor check-circle
- Expanded: checklist with green checks per field/section/rule
- Collapse state persists in localStorage

### Body — Edge thread view (on edge click)
- Header: state badge + edge type + connected doc names
- Diff summary: yellow-tinted bar showing what changed
- Thread entries: avatar + name + timestamp + message bubble
- Actions bar: Reaffirm (primary), Comment (secondary), Withdraw (danger)

---

## Foundation Fixes

### Legend (new component)
- Position: bottom-left, above MiniMap/Controls
- Collapsible (localStorage persistence)
- Content: state colors, border styles, badge meanings
- Filter by type: click type name in legend to highlight/dim others
- Uses Phosphor icons for state indicators

### Empty state (new component)
- Centered on canvas when 0 documents
- "No documents yet" + guidance text + CLI commands in monospace
- Zinc-400 text, no decorative elements

### Hover discovery
- Nodes: subtle lift shadow on hover (0 4px 12px at 6% opacity)
- Cursor: pointer on nodes and edges
- Edges: opacity bump on hover (0.4 → 0.8)

---

## Data Flow

### New NATS subjects
- `remmd.q.playbook` — returns active playbook (types, edges, rules)
- `remmd.q.validate` — runs Guard against all nodes, returns diagnostics per node

### Client-side merge
- `use-graph-data.ts` fetches graph + playbook + validation
- Each node enriched with: playbookType, owner, validationErrors, portSummary
- Edge labels enriched with: playbook edge type name, thread count

---

## Dependencies
- `@phosphor-icons/react` — Light weight icons (add to package.json)

## File Changes

### New files
- `npm/view/src/canvas/PlaybookNode.tsx` — three-zoom node with floating indicators
- `npm/view/src/panel/GapPanel.tsx` — problems-first validation panel
- `npm/view/src/panel/ThreadPanel.tsx` — edge thread/conversation view
- `npm/view/src/canvas/Legend.tsx` — collapsible legend with type filter
- `npm/view/src/canvas/EmptyState.tsx` — guided empty state
- `npm/view/src/canvas/GhostNode.tsx` — missing required node placeholder
- `npm/view/src/hooks/use-playbook.ts` — fetch + cache playbook data
- `npm/view/src/hooks/use-validation.ts` — fetch + cache validation diagnostics

### Modified files
- `npm/view/src/canvas/use-graph-data.ts` — merge playbook + validation into nodes
- `npm/view/src/canvas/BundledEdge.tsx` — add type labels, thread indicators, arrows
- `npm/view/src/App.tsx` — add Legend, EmptyState, wire new panels
- `npm/view/src/theme/colors.ts` — add validation state colors
- `npm/view/src/panel/use-panel-state.ts` — add 'thread' mode
- `npm/package.json` — add @phosphor-icons/react
- `internal/serve/handlers.go` — add remmd.q.playbook + remmd.q.validate handlers

---

## What's NOT in scope
- Playbook editor in UI (import via CLI only)
- Workflow/scaffolding visualization (deferred)
- Multi-playbook composition
- Keyboard shortcuts (separate improvement)
- Mobile/responsive (desktop canvas tool)
