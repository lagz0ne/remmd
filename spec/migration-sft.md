# Proposal: Migrate SFT to use remmd as trust backend

## Summary

SFT (behavioral spec tool) imports remmd's Go packages as a library to gain bilateral trust, change detection, and review workflows for UI specifications. SFT keeps its own domain tables (screens, regions, events, transitions, flows, fixtures, enums, data_types) and domain logic (state machines, event bubbling, validation rules, flow parsing). remmd provides the trust layer.

## Capability mapping

| SFT Concept | remmd Capability | How |
|---|---|---|
| App / Screen / Flow | Document Types (`doc_type`) | Each SFT screen → 1 remmd document with `doc_type = "sft:screen"` etc. |
| Region descriptions, event specs | Section Kinds (`kind`) | Each region → section with `kind = "region"`, events → `kind = "event"` |
| App → Screen containment | Document Hierarchy (`parent_doc_id`) | Screen doc's `parent_doc_id` = app doc's ID |
| Screen → Region containment | Relations | `relation create <screen-doc> <region-doc> --type contains` |
| Flow (cross-screen journey) | Links | Flow steps that cross screens → links between screen sections |
| State machine transitions | Section Metadata | Transition data stored as JSON metadata on "state-machine" kind sections |
| Fixtures, enums, data types | SFT domain tables | Not projected into remmd — pure domain data |
| `sft validate` (23 rules) | SFT domain logic | Validation runs against SFT's own tables, not remmd's |
| `sft impact` | Graph Walk | Change a screen section → walk finds linked flows/components |
| `sft show` (tree view) | SFT domain query | SFT queries its own tables for the tree; remmd not involved |

## What SFT gains

1. **Cross-tool trust**: An SFT screen can link to a C3 component. When the component's architecture changes, the screen's link goes stale — UI team must review whether the screen still matches.
2. **Bilateral spec review**: When a PM changes a screen spec, the designer's link to that spec goes stale. Thread-based review for iteration.
3. **Change detection for free**: Editing a screen's region description automatically recomputes content hash. All links to that section detect staleness.
4. **Unified search**: FTS5 search across all SFT screens + C3 components + native documents in one query.

## What SFT keeps

- `apps`, `screens`, `regions` tables (hierarchical entity model)
- `events`, `transitions` (state machine definitions)
- `flows`, `flow_steps` (arrow notation sequences)
- `data_types`, `enums`, `contexts`, `ambient_refs`, `region_data` (type system)
- `fixtures`, `state_fixtures` (test data with inheritance)
- `components` (UI bindings)
- `attachments` (binary content)
- `state_regions` (per-state visibility)
- `state_templates` (reusable state machine patterns)
- 4 SQL views (event_index, state_machines, tag_index, region_tree)
- 23+ validation rules
- Flow sequence parser (arrow notation → typed steps)
- Event bubbling logic

## What gets projected into remmd

Only content that benefits from **bilateral trust**:

| SFT Entity | Projected As | When |
|---|---|---|
| Screen description | remmd document (`sft:screen`) + section (`kind: description`) | Always — screen specs are reviewable |
| Region description | remmd section (`kind: region`) within screen doc | Always — region specs are reviewable |
| Flow sequence | remmd document (`sft:flow`) + section per step | When flow crosses team boundaries |
| State machine summary | remmd section (`kind: state-machine`) | When states are contractual |

NOT projected (pure domain internals):
- Individual transitions, events, fixtures, enums, data types, ambient refs, components, attachments

## Implementation sketch

```go
import (
    "github.com/lagz0ne/remmd/internal/core"
    "github.com/lagz0ne/remmd/internal/store"
)

db, _ := store.OpenDB(dbPath)
store.Migrate(db)   // remmd schema
sft.Migrate(db)     // sft domain schema (same DB)

remmdDocs := store.NewDocumentRepo(db)
sftStore := sft.NewStore(db)

// When SFT creates a screen, project into remmd
func (s *Service) AddScreen(name, description string) {
    // 1. SFT domain
    screen := s.sftStore.InsertScreen(name, description)

    // 2. Trust projection
    doc := core.NewDocument(name, "sft")
    doc.DocType = "sft:screen"
    s.remmdDocs.CreateDocument(ctx, doc)

    sec := core.Section{
        ID: core.NewID().String(), DocID: doc.ID,
        Kind: "description", Content: description,
        ContentHash: core.ContentHash(description),
    }
    s.remmdDocs.CreateSection(ctx, &sec)

    // 3. Store mapping (sft screen ID → remmd doc ID)
    s.sftStore.SetRemmdMapping(screen.ID, doc.ID)
}

// When SFT screen description changes, update remmd section
func (s *Service) UpdateScreenDescription(screenID int64, newDesc string) {
    s.sftStore.UpdateScreen(screenID, newDesc)

    // Update remmd section → triggers stale detection on all links
    docID := s.sftStore.GetRemmdMapping(screenID)
    sections, _ := s.remmdDocs.ListSections(ctx, docID)
    for _, sec := range sections {
        if sec.Kind == "description" {
            s.remmdDocs.UpdateSectionContent(ctx, sec.ID, newDesc, core.ContentHash(newDesc))
            break
        }
    }
}
```

## Open questions

1. **Projection granularity**: Should every region be its own remmd section (fine-grained trust) or should the entire screen spec be one section (simpler, but coarser change detection)?
2. **Mapping table**: SFT needs a `sft_remmd_mappings (sft_entity_type, sft_entity_id, remmd_doc_id)` table to track which SFT entities are projected. Where does this table live — SFT's schema or remmd's?
3. **Selective projection**: Not all screens need trust tracking. Should projection be opt-in (explicit command) or automatic?
4. **Event-driven sync**: Should SFT emit events that remmd subscribes to, or should the sync be imperative (SFT calls remmd APIs directly)?
