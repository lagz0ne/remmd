# Proposal: Migrate C3 to use remmd as trust backend

## Summary

C3 (architecture documentation tool) imports remmd's Go packages as a library to gain bilateral trust, change detection, and review workflows for architecture entities. C3 keeps its own domain tables (entities, nodes, relationships, code_map) and domain logic (merkle hashing, 10-phase validation, ADR lifecycle). remmd provides the trust layer.

## Migration path

[Diagram](https://diashort.apps.quickable.co/d/a49dbc3d)

## Capability mapping

| C3 Concept | remmd Capability | How |
|---|---|---|
| Entity types (component, ref, rule, adr) | Document Types (`doc_type`) | Each C3 entity → 1 remmd document with `doc_type = "c3:component"` etc. |
| Entity sections (Goal, Dependencies, Choice) | Section Kinds (`kind`) | Each C3 section → 1 remmd section with `kind = "goal"` etc. |
| Required sections per entity type | Schema Templates | `template set c3:component goal --min 1` |
| Context → Container → Component hierarchy | Document Hierarchy (`parent_doc_id`) + Relations | Structural nesting via `parent_doc_id`; cross-cutting via Relations |
| `c3x wire` (component ↔ ref) | Links (bilateral) | `link propose @s1 --implements @s2` — now with stale detection + review |
| `c3x impact` (transitive deps) | Graph Walk | BFS from changed section finds all impacted links |
| Entity version history | Section Versions + metadata | Per-section versioning; commit_hash in version metadata |
| Code-map (file → entity globs) | Section Metadata | Store glob patterns in metadata JSON on a "codemap" kind section |
| `c3x query` (full-text search) | FTS5 Search | `SearchSections(ctx, query)` against FTS5 index |
| Merkle hash (root_merkle) | Client-side computation | C3 reads section `content_hash` values, computes merkle client-side |

## What C3 gains

1. **Bilateral trust on architecture decisions**: When a component "wires" to a ref, both parties must approve. When the ref changes, the component's link goes stale — author must reaffirm, counterparty must review.
2. **Change blast radius across tools**: If C3 component links to an SFT screen (via remmd), changing the component triggers review on the screen. Cross-tool trust.
3. **Unified review threads**: All review discussion in one place (remmd threads), not scattered across tools.
4. **Content hash change detection for free**: remmd computes SHA-256 per section automatically. C3 gets stale detection without maintaining its own hash infrastructure.

## What C3 keeps

- `entities` table (type, slug, category, boundary, status)
- `nodes` table (goldmark AST → node tree with hashes)
- `relationships` table (entity-to-entity directed edges for internal use)
- `code_map` + `code_map_excludes` (glob matching engine)
- `versions` table (full markdown snapshots with commit_hash)
- `changelog` (field-level audit trail)
- 10-phase validation (`c3x check`)
- Merkle hash computation
- ADR lifecycle state machine

## Implementation sketch

```go
import (
    "github.com/lagz0ne/remmd/internal/core"
    "github.com/lagz0ne/remmd/internal/store"
)

// C3 opens the shared DB and runs both migration sets
db, _ := store.OpenDB(dbPath)
store.Migrate(db)      // remmd migrations
c3.Migrate(db)         // c3 domain migrations

// C3 creates remmd repos alongside its own
remmdDocs := store.NewDocumentRepo(db)
remmdLinks := store.NewLinkRepo(db)
c3Store := c3.NewStore(db) // c3's own tables

// When C3 creates an entity, it also creates a remmd document
func (s *Service) CreateComponent(name string, goal string) {
    // 1. Create in C3's entities table
    entity := s.c3Store.InsertEntity(...)

    // 2. Project into remmd for trust tracking
    doc := core.NewDocument(name, "c3")
    doc.DocType = "c3:component"
    s.remmdDocs.CreateDocument(ctx, doc)

    goalSection := core.Section{
        ID: core.NewID().String(), DocID: doc.ID,
        Kind: "goal", Content: goal,
        ContentHash: core.ContentHash(goal),
    }
    s.remmdDocs.CreateSection(ctx, &goalSection)
}

// When C3 wires entities, it creates a remmd link
func (s *Service) Wire(componentID, refID string) {
    // 1. C3 internal relationship
    s.c3Store.AddRelationship(componentID, refID, "implements")

    // 2. remmd bilateral link (gets stale detection + review)
    link := core.NewLink(
        []string{componentDocSectionID},
        []string{refDocSectionID},
        core.RelImplements,
        core.Rationale{Claim: "component implements ref pattern"},
        "c3",
    )
    s.remmdLinks.CreateLink(ctx, link)
}
```

## Open questions

1. **Dual storage overhead**: Each C3 entity exists in both C3's `entities` table and remmd's `documents` table. Is the duplication worth the trust benefits? Alternative: drop C3's storage entirely and use remmd as the sole store (more radical, higher risk).
2. **Sync strategy**: When C3 entity content changes, the corresponding remmd sections must be updated. Should this be synchronous (in the same transaction) or eventual?
3. **Ref mapping**: C3 uses `c3-101` style IDs; remmd uses `@s1` style refs. Need a mapping table or convention.
