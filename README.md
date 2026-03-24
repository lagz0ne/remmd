# remmd

Human-approved agreement between independently maintained document sections.

Documents are independent claims — a requirement, a design, an implementation, a test plan. They are supposed to match. Nobody checks whether they do. remmd does.

## How it works

```
Author creates content → sections get stable @refs automatically
Author proposes link between sections across documents
Both sides approve → link is ALIGNED
Content changes → graph walks → impacted links go STALE
Author reaffirms or withdraws → counterparty reviews → iterate
```

Links are the only reviewed entity. Content edits are immediate and versioned. The graph ripples on every change.

## Install

```bash
go install github.com/lagz0ne/remmd/cmd/remmd@latest
```

Or download a binary from [Releases](https://github.com/lagz0ne/remmd/releases).

## Quick start

```bash
# Create documents
remmd doc create "API Specification" --content "# Auth\nToken-based authentication\n# Endpoints\nGET /users"
remmd doc create "Implementation Notes" --content "# Auth Handler\nValidates JWT tokens"

# Show sections
remmd show @a1

# Link sections across documents
remmd link propose @a1 --implements @c1 --rationale "auth handler implements auth spec"

# Approve from both sides
remmd link approve <link-id> --context-hash <hash>

# See blast radius when content changes
remmd impact @a1

# External content (Notion, Figma, anything with a stable ID + hash)
remmd doc create "Design Spec" --external --system figma --external-id frame-abc --hash sha256:...
remmd edit @ext:figma/frame-abc --hash sha256:...  # push hash update
```

## Core concepts

| Concept | What it is |
|---------|-----------|
| **Section** | Minimum accountable unit. Every heading/block gets a stable `@ref`. |
| **Link** | Bilateral agreement between sections. The only reviewed entity. |
| **Thread** | Persistent review workspace on each link. Like a PR conversation. |
| **Graph walk** | Every content edit walks the graph. Impact shown before action. |
| **External content** | Anything with a stable ID + hash can participate. `@ext:system/id` refs. |

## Link states

`pending` (proposed) → `aligned` (both approved) → `stale` (content changed) → review → `aligned`

Also: `broken` (section deleted), `archived` (explicitly closed).

## Principals

**Humans** create trust — approve, reaffirm, withdraw links.
**Service principals** (AI, integrations) draft content, propose links, route work. They never approve.

## Tech

- Go, embedded SQLite (pure Go, no CGO)
- Single binary, no external dependencies
- Event-sourced, domain-driven

## License

[AGPL-3.0](LICENSE)
