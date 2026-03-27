# C3→remmd Convergence Design

## Goal

Use remmd to develop and review remmd's own architecture docs. C3 continues to work in parallel during transition. Eventually remmd replaces C3 as the primary authoring surface.

## Three Layers

### Layer 1: Fix Canvas Foundation

Five fixes that make remmd usable for daily architecture work:

1. **Dedup docs** — add `remmd doc delete` CLI command. Clean duplicate imports.
2. **Table column headers** — section parser preserves header row; canvas/panel renders actual column names.
3. **Meaningful briefs** — DocNode shows `goal` field from playbook type, not raw first section content.
4. **Edge labels** — BundledEdge renders relationship type (contains, cites, governs, etc.).
5. **Position persistence** — restore saved positions on reload instead of dagre overriding every time.

### Layer 2: C3 Adapter (External Matching)

Subprocess adapter following `ref-adapter-protocol`. C3 entities appear as external sections in remmd.

**Adapter binary:** `remmd-adapter-c3` (Go, same repo, `cmd/remmd-adapter-c3/`)

**Operations:**
- `ready` — adapter initialized, reports available systems: `["c3"]`
- `fetch @ext:c3/<entity-id>` — runs `c3x read <id>`, returns content + hash
- `watch @ext:c3/*` — polls `.c3/c3.db` mtime, emits `changed` on delta
- `content` — returns markdown body, SHA-256 hash, format: markdown
- `stop` — graceful shutdown

**Ref format:** `@ext:c3/c3-207` (maps to C3 entity ID)

**Canvas rendering:** External nodes get a subtle badge/border indicating "sourced from C3". Clicking opens panel with full content fetched via adapter.

### Layer 3: Doc Matching (Playbook Parity)

Verify `c3.playbook.yaml` captures everything `c3x check` enforces:

| c3x check | Playbook equivalent | Status |
|---|---|---|
| Missing required sections | `sections` with `required: true` | ✅ exists |
| Bad entity references | `edges` with valid endpoint types | ✅ exists |
| Orphan refs | `orphan-ref` rule | ✅ exists |
| Codemap coverage | New rule: `codemap-coverage` | ❌ needs adding |
| Scope crosscheck | `scope-crosscheck` rule | ✅ exists |
| No owner steal | `no-owner-steal` guard (CEL) | ✅ exists |

**Missing concept: Codemap.** C3 maps files→components. In remmd this becomes:
- A section kind `codemap` on component documents
- Content: glob patterns that map to this component
- `remmd lookup <file>` command that searches codemap sections
- Playbook rule: `codemap-coverage` — warn if component has no codemap section

## Architecture Decisions

- **No bidirectional sync.** C3→remmd is read-only via adapter. New authoring happens in remmd.
- **Adapter is a separate binary** in same repo — keeps remmd core clean.
- **External refs are first-class** — links can connect native sections to external sections.
- **Gradual migration** — old C3 content stays external until manually converted to native docs.

## Testing Strategy

- **Canvas fixes:** agent-browser visual verification before each commit
- **Adapter:** TDD — test fetch/watch/stop protocol with mock c3x output
- **Doc matching:** `remmd playbook check` + `remmd validate` must pass on imported C3 docs
- **Integration:** Start server, import C3 playbook, verify canvas renders all entities with correct types and edges

## Success Criteria

1. `remmd serve` shows all C3 architecture entities on canvas with correct hierarchy
2. Clicking a node shows full content with proper table rendering
3. Edge labels show relationship types
4. Positions persist across page reloads
5. `remmd validate` catches the same issues as `c3x check`
6. New architecture docs can be authored in remmd and linked to existing C3 content
