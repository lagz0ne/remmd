# Lessons — 2026-03-26/27 Session

## What shipped
- Playbook engine (6-keyword YAML, CEL checker, Guard, embedded storage)
- Canvas UI rebuilt from scratch (React Flow + dagre + panel drawer)
- 2 production playbooks (c3, sft)
- C3 architecture docs imported as remmd documents with 49 links

## Hard lessons

### Never commit unverified UI changes
The first round of UI work shipped 6 bugs in one commit — overlapping nodes, empty panel, broken selection, broken re-layout. All would have been caught by a single visual verification pass. **Before ANY git commit on UI changes, verify with agent-browser.**

### React Flow v12 is a controlled component
`useNodesState` + `onNodesChange` is the ONLY pattern that makes drag work. The earlier approach (passing `nodes` prop from external state + calling `setNodes` from `useReactFlow`) creates a fight between controlled external state and internal React Flow state. **Start with `useNodesState`/`useEdgesState` from React Flow's own API.**

### When stuck on patches, scrap and start fresh
The drag bug took 5+ attempts to fix by patching. Scrapping App.tsx and writing 80 lines from scratch (bare `useNodesState` + `onNodesChange`) worked immediately. **When 3 patches fail, the architecture is wrong — rebuild from the minimum that works.**

### Dagre node dimensions must match rendered size
Dagre allocates space based on NODE_WIDTH/NODE_HEIGHT. If the rendered component is larger, nodes overlap. When the component changes (adding markdown, changing zoom behavior), dagre params must be re-measured. **Measure actual rendered dimensions, don't guess.**

### agent-browser mouse events don't work for d3-drag
`agent-browser mouse down` fires at (0,0) regardless of prior `mouse move`. Synthetic `dispatchEvent` with PointerEvent is ignored by React Flow (untrusted events). **Can't test drag programmatically via agent-browser — need user to verify.**

### Section parser splits table rows into separate sections
remmd's section parser treats each `|` line as a potential section boundary. When importing markdown with tables, each table row becomes its own section. The panel must reconstruct tables by joining contiguous pipe-delimited sections. **Import format matters — test with tables/code blocks, not just headings.**

### Port derivation is unpredictable
`remmd serve` derives the port from the project name hash. Different working directories produce different ports (4126 vs 4161). Old processes on old ports serve stale content. **Always check `ss -tlnp | grep PORT` and kill old processes before starting new ones.**

## Architecture decisions to carry forward

### Playbook is a standalone spec format
6 keywords: `description`, `sections`, `rules`, `severity`, `expr`, `example`. Shape-based parsing (string with `->` = edge, map with `expr` = rule, other map = type). **Not remmd-specific — any tool can consume it.**

### Guard enforces via 4 CEL bindings
`self` (trusted, from store), `proposed` (untrusted, from agent), `principal` (actor), `action` (mutation verb). **Owner comes from store, not from the mutation payload — prevents self-escalation.**

### Playbook in DB, YAML is import/export
Embedded in SQLite (`playbooks`, `pb_types`, `pb_fields`, `pb_sections`, `pb_edges`, `pb_rules`, `pb_examples`). Versioned as a unit with SHA-256 hash idempotency. **DB is runtime source of truth, YAML is authoring format.**

### Canvas layout: dagre for initial, manual for ongoing
Square-ish zone layout: connected subgraphs laid out with dagre (hierarchy preserved), then clusters arranged in rows. Position persistence via `node_positions` table. **Auto-layout is the starting point, not the ongoing constraint.**

### Panel drawer: stack-based navigation
Navigation stack: `[{nodeId, title, edgeType, sourceId}]`. Stack length 0 = closed, 1 = single column, 2+ = two columns with trace. Column 1 footer always shows edge list. Column 2 has trace header showing navigation path. **Shift columns on follow, breadcrumb to jump back.**

### Workflow = remmd's existing link system
Playbooks define artifact types + ownership (static). Workflows are remmd links across playbook boundaries (dynamic). When PM's story changes, linked designer's screen goes stale. **No separate workflow engine needed — links ARE the workflow mechanism.**

## What to do next session
- Fix duplicate doc imports (need a `remmd doc delete` command)
- Table headers should use actual column names from the C3 schema, not "Col 1, Col 2"
- The brief on canvas nodes should be more meaningful (Goal text, not first section raw content)
- Edge labels on canvas (currently no labels showing the relationship type)
- Position persistence — drag positions save but dagre overrides on every reload (need to restore saved position loading)
