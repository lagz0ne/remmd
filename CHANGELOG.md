# Changelog

## [0.2.0-rc.1] — 2026-03-25

### `remmd serve` — Realtime Document Viewer

New command that launches a live view of your document network in the browser.

- **Embedded NATS** server (in-process, WebSocket for browser, no external deps)
- **Vite 8** as the content transform engine (MDX, mermaid, diagrams ready)
- **Live updates**: edit via CLI → NATS event → browser refreshes automatically
- **Inline diff highlighting**: changed sections flash with green/strikethrough, fades after 3s
- **Section click → clipboard**: click any section to copy its `@ref` for LLM agent workflows
- **Document selector**: dropdown to switch between documents

Architecture: Go serves HTTP, reverse-proxies Vite, all data flows through NATS request-reply. Browser connects via `nats.ws`. No REST API.

### npm Distribution

The npm package now ships both the Go binary AND the Vite frontend source.

```bash
npm install -g @lagz0ne/remmd   # installs everything
cd your-project && remmd serve  # just works
```

- `REMMD_PACKAGE_DIR` env var passed automatically by the Node shim
- Vite + React + nats.ws + React Query installed as npm dependencies
- No separate setup step for the view

### Project-Local Database Discovery (#3)

remmd now walks up from the current directory looking for `.remmd/remmd.db`, like git discovers `.git/`.

```
1. --db flag        (explicit, highest priority)
2. .remmd/remmd.db  (walk up from cwd)
3. ~/.remmd/remmd.db (global fallback)
```

No more `--db /path/to/project/.remmd/remmd.db` on every command.

### Global Ref Uniqueness Fix (#6)

Section refs (`@a1`, `@b2`, etc.) are now globally unique across all documents. Previously, every document started at `@a1`, causing ambiguity in `show`, `link propose`, and `impact`.

- **`remmd migrate-refs`**: one-time migration command that re-numbers all existing sections with globally unique refs
- Parser body text capture: heading sections now include body paragraphs in their `Content` field (was title-only)

### CLI Improvements

- **`doc delete`** (#5): cascading delete of document + sections + tags + versions
- **`doc archive`** (#5): soft-delete, hidden from `doc list` by default (`--all` to show)
- **`relation list`** (#4): shows document titles instead of raw ULIDs
- **`search`** (#7): shows `[Doc Title] @ref "Section Title"` with content snippet
- **`section add`** (#8): add sections to existing documents incrementally
- **`--json` output** (#8): machine-readable JSON output for tool integration
- **`import --json`** (#8): bulk import from JSONL
- **`find`** (#8): find sections by metadata/tags
- **`template set/check`** (#8): schema templates for document types
- **`relation create/list/delete`** (#8): document-level dependency graph

### Domain Model

- `Document.DocType`: user-defined classification (spec, design, runbook)
- `Document.ParentDocID`: document hierarchy
- `Section.Kind`: semantic classification (goal, requirement, decision)
- `Relation`: document-to-document dependencies (grounds, derives, implements)
- `SchemaTemplate`: enforce required section kinds per document type

### Code Quality

- N+1 link queries → batch `LinksContainingSections` method
- Subject parsing helper in NATS handlers
- `useMemo` for diff computation in React
- Bounded section cache in Vite plugin (500 entries max)
- `remmd-local` dev script at `~/.local/bin/` for auto-rebuild

### Closed Issues

- #3 — Support project-local database via .remmd/ directory discovery
- #4 — relation list should show document titles alongside IDs
- #5 — Add doc delete/archive command
- #6 — link propose should support doc-scoped refs (resolved: refs now globally unique)
- #7 — search results should include document title and section ref
- #8 — CLI gaps for standalone tool integration

## [0.1.3] — 2026-03-24

- Global ref uniqueness and stale link transitions (#1, #2)
- CLI help with grouped commands, workflow examples
