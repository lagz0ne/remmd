# remmd — UX Specification

Aligned to [remmd.md](./remmd.md). Gaps in interaction design are tracked in [gaps.md](./gaps.md).

---

## Design Philosophy

**Sparse. Canvas-oriented. The document is the workspace.**

remmd is not a dashboard with a document viewer. It's a canvas where documents live, links are visible, and trust is ambient. The UI should feel closer to Figma's infinite canvas than to Confluence's page stack.

Every screen should pass the squint test: if you squint, you should see content and link state — nothing else.

---

## Industry Constraints

These are table stakes for a document platform in 2026. Non-negotiable, but not differentiating.

| Constraint | Standard | remmd stance |
|---|---|---|
| Real-time save | Auto-save expected post-Google Docs era | Autosave 30s + manual save. Content edits are immediate and versioned. |
| Version history | Users expect to rewind | Full version history per section with any-to-any diff. |
| Search | Instant, full-text | Search by title, content, tag, section @ref, owner. |
| Access control | Viewer/editor/owner roles | Per-document visibility + per-user edit grants. |
| Keyboard shortcuts | Power users expect them | Essential shortcuts: save, search, navigate links. Not a shortcut-heavy app. |
| Responsive | Desktop-first, mobile-usable | Primary: desktop (1200px+). Secondary: tablet. Mobile: read-only + review queue. |
| Import/export | Users need escape hatches | Import: markdown, paste from anywhere, source adapters. Export: markdown, PDF. |
| Collaboration | Real-time co-editing expected | Phase 1: single-author per document. Phase 2: collaborative editing. |
| Accessibility | WCAG 2.1 AA | Color not sole indicator (link states use shape + position + color). Keyboard navigable. Screen reader labels. |
| Performance | Sub-second interactions | Save + change detection < 500ms. Notification delivery < 200ms. Dashboard load < 200ms. |

---

## Canvas Orientation

### Principle: The document IS the workspace

There is no "home screen" → "document list" → "open document" → "see links" hierarchy. The canvas holds documents, sections, and links as visible, spatial objects. You zoom in to author. You zoom out to see the network.

### Three zoom levels

**Close: Author mode**
- Full editor canvas. Sections auto-identified with `@refs`.
- Link panel visible as a right sidebar (collapsible).
- Link state shown as section border indicators.
- This is where 80% of time is spent.

**Medium: Document mode**
- Document header with link state summary.
- Sections visible as labeled blocks (content partially collapsed).
- Links visible as lines between sections.
- This is where you scan your links and resolve stale items.

**Far: Network mode**
- Nodes = documents. Edges = links.
- Link state indicators on edges.
- Filter by team, tag, link state, relationship type.
- This is the admin's primary view. Authors visit occasionally.

### Canvas behaviors

| Behavior | Interaction |
|---|---|
| Pan | Click + drag on empty canvas, or two-finger trackpad |
| Zoom | Scroll wheel or pinch |
| Select document | Click a document node → zooms to document mode |
| Open document | Double-click → zooms to author mode |
| Navigate link | Click a link edge → opens thread workspace |
| Return to network | Pinch out or keyboard shortcut (Cmd/Ctrl + 0) |

---

## Sparse Design Rules

| Rule | Rationale |
|---|---|
| No chrome by default | Editor canvas is 100% content. Link panel is hidden until the document has links. |
| Indicators before panels | Link state is a 2px border, not a dashboard widget. If the border shows aligned, there's nothing to do. |
| One thing at a time | Thread workspace shows one link. Queue shows one list. Network shows one graph. No split-screen multitasking. |
| Progressive disclosure | Sections start as just content. Links appear when created. State indicators appear when links exist. Network appears when enough links exist. |
| No empty states that teach | If you have no links, the document looks like a normal document. No onboarding cards, no "create your first link!" prompts. The feature is discoverable through the section hover menu. |
| Density on demand | Link panel is collapsible. Review queue is expandable. Network is zoomable. Sparse is the default. Dense is opt-in. |

---

## Typography & Visual Hierarchy (Application-Level)

| Element | Treatment |
|---|---|
| Document title | 24px, serif, bold. The loudest element. |
| Section `@ref` label | 10px, mono, muted. Ambient — not competing with content. |
| Link state indicator | 2px left border on section. 6px square in link panel. States: `aligned` (good), `stale` (attention), `pending` (neutral), `broken` (error), `archived` (dimmed). |
| Link rationale | 13px, regular weight. Readable when focused, invisible when scanning. |
| Diff | 12px, mono. Green = added, accent = removed. Scoped to section. |
| Queue item | 14px title, 12px metadata. Flat card. No borders between items — just spacing. |
| Network node | Label only. No description. Shape/indicator for state. |

---

## Responsive Breakpoints

| Breakpoint | Layout | Key changes |
|---|---|---|
| 1200px+ (desktop) | Full canvas. Editor + link panel side by side. Network mode available. | Primary experience. |
| 768–1199px (tablet) | Editor full width. Link panel as bottom sheet. Network mode simplified. | Thread workspace stacks vertically. |
| < 768px (mobile) | Read-only document view. Review queue accessible. No editor. | Approve/comment from queue only. No authoring on mobile. |

---

*This UX spec defines constraints, not wireframes. Screen layouts, interaction patterns, and role-specific workflows are tracked as gaps in [gaps.md](./gaps.md) until designed against the normative spec.*
