# remmd — Event Storming

Aligned to [remmd.md](./remmd.md) spec and [sequence diagram](https://diashort.apps.quickable.co/d/840bab0b).

---

## Legend

| Color | Meaning |
|---|---|
| Actor | Human role (PM, Engineer, QA, Admin) |
| Command | Human or system intent |
| Event | Thing that happened (past tense, immutable) |
| Policy | Reacts to events, triggers commands |
| System | Automated process, no human decision |

---

## Process 1: Content Creation

Content edits are immediate and versioned. Sections get stable `@refs` automatically. No manual marking.

```
Actor           Command              Event                  Policy
─────           ───────              ─────                  ──────
Author    →     Create Document  →   Document Created   →   Parse Structure
                (markdown-ish,       Sections Identified     (auto-assign @refs
                 single call)                                 to every heading/block)

Author    →     Edit Section     →   Section Edited     →   Detect Impacted Links
                (by @ref,            Version Recorded        (walk graph from changed
                 immediate)          (content hash updated)    section)

Author    →     Delete Section   →   Section Deleted    →   Walk Graph
                (with reason +       Version Recorded        (find impacted links,
                 optional                                     update threads,
                 replacement ref)                             notify counterparties)

Author    →     Tag Section      →   Section Tagged     →   Policy Evaluation
                                                             (tags may affect gate
                                                              predicates)
```

### Key transitions

- `Create Document` parses structure in one call — every heading/block gets a `@ref`
- `Edit Section` is immediate — no draft/activate lifecycle. Every edit creates a new version with updated content hash
- `Delete Section` is immediate with metadata (reason, replacement refs). Links containing the section go `broken` and counterparties are notified via threads
- Content hash is per-section, not per-document or per-commit

---

## Process 2: Tag Subscription & Discovery

Tags classify sections for policy, filtering, and discovery. Subscriptions turn tagged arrivals into notifications.

```
Actor           Command              Event                  Policy
─────           ───────              ─────                  ──────
Author    →     Subscribe to Tag →   Subscription Created
                (section @ref +
                 tag expression)

                                     (later: new content
                                      with matching tag
                                      is activated)

System    →     Evaluate Subs    →   Subscription Fired →   Notify Subscriber
                                     (tag match found)       (context: new doc,
                                                              matched tag, section)

Owner     →     Create Link      →   Link Proposed      →   Normal link flow
                (from notification)                          (Process 3)

Owner     →     Dismiss          →   Subscription
                                       Dismissed
```

### Key transitions

- Subscription fires on: new document with matching tag activated, or tag added to existing active content
- Firing creates a notification, NOT a link or candidate — subscriber decides manually
- Subscriber can: create a link (enters normal bilateral approval), or dismiss (irrelevant)
- Subscriptions evaluate only against active content and active tags
- Tag changes follow the same versioning as content — every tag mutation is recorded

---

## Process 3: Link Proposal

A link connects section(s) to section(s) across documents. Creating a link opens a review thread. Both sides start pending.

```
Actor           Command              Event                  Policy
─────           ───────              ─────                  ──────
Proposer  →     Propose Link     →   Link Proposed      →   Open Thread
                (@refs[] --type      Thread Opened           (review workspace created)
                 @refs[]
                 --rationale)    →   Both Sides Pending  →   Notify Counterparty
                                                             (review request routed)
```

### Key transitions

- A link groups section(s) → section(s), potentially across multiple documents
- Both sides start pending — proposer is NOT auto-approved (it's a request)
- Thread opens immediately — the link's review workspace
- Rationale includes claim + scope + exclusions (structured, not a single string)
- Relationship type is explicit: `agrees_with`, `implements`, `tests`, `evidences`
- Intervention level is set per endpoint independently

---

## Process 4: Review Thread

Review is iterative thread-based negotiation. Comment, edit, diff appears, comment again, approve. Like a PR code review.

```
Actor           Command              Event                  Policy
─────           ───────              ─────                  ──────
Reviewer  →     Open Thread      →   Thread Viewed

Reviewer  →     Comment          →   Comment Added      →   Notify Participants
                                                             (thread update)

Author    →     Edit Section     →   Section Edited     →   Update Thread
                (responding to       Version Recorded        (diff appears inline
                 feedback)                                    in thread)

Reviewer  →     Approve Link     →   Endpoint Approved  →   Check Both Sides
                                     (for this side)

                                                         →   If Both Approved:
                                     Link Aligned            Agreement Snapshot
                                                              Recorded

                                                         →   If One Side Only:
                                     Waiting On              Route to other side
                                       Other Side
```

### Key transitions

- Approve records: acting principal, exact section snapshots, timestamp, viewed context
- Both sides must approve the same agreement snapshot (exact section versions + relationship definition)
- If underlying content changes before submission, approval fails with stale-context — must refresh
- Thread accumulates across all review cycles — full history retained
- Content edits during review appear as diffs in the thread automatically

---

## Process 5: Content Change & Graph Ripple

When content changes in a section that belongs to aligned links, the graph ripples.

```
Actor           Command              Event                  Policy
─────           ───────              ─────                  ──────
Author    →     Edit Section     →   Section Edited     →   Walk Graph
                                     Version Recorded        (find all links containing
                                                              this section)

                                     Impacted Links     →   Show Impact Preview
                                       Identified            (blast radius before action)

Author    →     Reaffirm Link    →   Link Reaffirmed   →   Set Link Stale
                ("I still stand      (author's side)         (waiting on counterparty)
                 behind this")
                                                         →   Update Thread
                                                             (diff since last aligned
                                                              + reaffirm notice)

                                                         →   Notify Counterparty
                                                             (per intervention level)

─── or ───

Author    →     Withdraw Link    →   Link Withdrawn     →   Update Thread
                ("this no longer     (author's side)         (withdrawal reason visible)
                 holds")
                                                         →   Notify Counterparty

─── counterparty reviews ───

Counter-  →     Approve Link     →   Endpoint Approved  →   If Both Approved:
party                                                        Link Aligned
                                                              (new agreement snapshot)

─── or ───

Counter-  →     Comment          →   Comment Added      →   Notify Author
party           ("needs changes")                            (iterate in thread)
```

### Key transitions

- One edit can impact multiple links — author sees full blast radius
- Bulk reaffirm supported when multiple links impacted by same edit
- Reaffirm/withdraw is mandatory before counterparty is notified
- Counterparty sees cumulative diff since last aligned, not just the latest edit
- Thread shows causal chain: what upstream change triggered this review
- No special "disputed" state — rejected reviews are just open threads with unresolved comments

### Cascading changes

```
PM edits @r2 (requirement)
  → @l1 stale (Eng's implementation)
  → @l2 stale (QA's tests)

Eng updates @i2 to follow PM's change
  → @l1 re-reviewed by PM
  → @l3 stale (QA's tests also reference @i2)

QA updates @t2 to match
  → @l2 and @l3 re-reviewed

Each thread tracks the causal chain:
  @l3 thread shows: "triggered by @r2 edit → @i2 followed → your @t2 affected"
```

---

## Process 6: Source Adapter Events

External systems emit sections and change events through the adapter contract.

```
System          Command              Event                  Policy
──────          ───────              ─────                  ──────
Adapter   →     Import Source    →   Document Created    →   Parse Structure
                (Jira, Git,          Sections Identified      (auto @refs, content hashes)
                 Confluence,
                 Figma, ...)

AI        →     Propose Links    →   Links Proposed     →   Route Review
                (analyze against     Threads Opened           (to section owners)
                 existing graph)

Adapter   →     Sync Change      →   Section Edited     →   Walk Graph
                (external source     Version Recorded        (same ripple as native
                 updated)            (new content hash)        content change)
```

### Adapter contract

| Method | Input | Output |
|---|---|---|
| `emit_sections` | source reference | `[{id, content, hash}]` |
| `on_change` | source reference | `[{section_id, old_hash, new_hash}]` |
| `render_preview` | section reference | displayable content for review UI |

### Key transitions

- External content = auto-created sections with `@refs` and content hashes (same as native)
- AI MAY propose links with rationale — humans approve. AI MUST NOT approve.
- Bulk import = documents + sections + draft link proposals in one operation
- External changes flow through the same graph ripple mechanism
- Code adapter: hash per file/function path, NOT git commit hash

---

## Process 7: Policy & Gate Enforcement

Policies evaluate the trust graph at decision points.

```
Actor           Command              Event                  Policy
─────           ───────              ─────                  ──────
Admin     →     Create Policy    →   Policy Created
                (selector +
                 predicate +
                 surface +
                 severity)

Admin/    →     Trigger Gate     →   Gate Evaluated     →   If Pass:
System          (merge, release,                             Gate Passed
                 publish, audit)

                                                         →   If Fail:
                                     Violation Created       Route to responsible
                                     (exact failing           parties
                                      objects named)

Admin     →     Grant Waiver     →   Waiver Granted     →   Gate Re-evaluated
                (reason + expiry)    (violation waived,       (passes with waiver)
                                      still visible)

System    →     Waiver Expired   →   Waiver Expired     →   Violation Re-opened
                (automatic)
```

### Key transitions

- Gate failures are always explainable — exact failing sections, what's missing, who needs to act
- No aggregate "trust score" — gates return pass/fail with drilldown
- Waivers do NOT create trust — they bypass gates temporarily
- Only human principals may grant waivers
- Expired waivers automatically re-open violations

---

## Aggregate Boundaries

```
┌─ Document Aggregate ────────────────────────────────────────┐
│  Document                                                    │
│  ├── Sections[] (auto @refs, content hashes, nested)         │
│  ├── Versions[] (per section, immutable)                     │
│  ├── Tags[]                                                  │
│  └── Source (native | adapter type + reference)              │
└──────────────────────────────────────────────────────────────┘

┌─ Link Aggregate ────────────────────────────────────────────┐
│  Link                                                        │
│  ├── Left sections[] (@refs)                                 │
│  ├── Right sections[] (@refs)                                │
│  ├── Relationship type (agrees_with|implements|tests|evidences)│
│  ├── Rationale (claim + scope + exclusions)                  │
│  ├── Left intervention (watch|notify|urgent|blocking)        │
│  ├── Right intervention (watch|notify|urgent|blocking)       │
│  ├── Agreement snapshot (current section versions + link def)│
│  ├── Thread[] (comments, diffs, system events — append-only) │
│  └── State (derived: pending|aligned|stale|broken|archived) │
└──────────────────────────────────────────────────────────────┘

┌─ Policy Aggregate ──────────────────────────────────────────┐
│  Policy                                                      │
│  ├── Selector (by tag, type, document)                       │
│  ├── Predicate (e.g., aligned(tests) >= 1)                   │
│  ├── Surface (merge|publish|release|audit|dashboard)         │
│  ├── Severity (blocking|warning)                             │
│  └── Violations[] + Waivers[]                                │
└──────────────────────────────────────────────────────────────┘

┌─ Subscription Aggregate ─────────────────────────────────────┐
│  Subscription                                                 │
│  ├── Subscriber section @ref                                  │
│  ├── Tag expression                                           │
│  ├── Status (active|archived)                                 │
│  └── Fire Events[] (audit trail)                              │
└───────────────────────────────────────────────────────────────┘

┌─ Source Adapter (boundary, not aggregate) ───────────────────┐
│  Adapter                                                      │
│  ├── Type (git|jira|confluence|figma|custom)                  │
│  ├── emit_sections() → [{id, content, hash}]                 │
│  ├── on_change() → [{section_id, old_hash, new_hash}]        │
│  └── render_preview() → displayable content                   │
└───────────────────────────────────────────────────────────────┘
```

---

## Event Catalog

| Event | Trigger | Data | Downstream |
|---|---|---|---|
| Document Created | Human or adapter: create | doc_id, tenant, owner, source_type | Parse Structure |
| Sections Identified | System: parse | doc_id, sections[{id, content, hash}] | — |
| Section Edited | Human or adapter: edit | section_id, old_hash, new_hash, version | Walk Graph |
| Version Recorded | System: version | section_id, version_num, content_hash | — |
| Section Deleted | Human: delete | section_id, reason, replacement_refs | Walk Graph, Break Links |
| Section Tagged | Human: tag | section_id, tag | Policy Evaluation |
| Link Proposed | Human or AI: propose | link_id, left_refs[], right_refs[], type, rationale | Open Thread, Notify Counterparty |
| Thread Opened | System: on link | link_id, thread_id | — |
| Comment Added | Human: comment | thread_id, principal_id, body | Notify Participants |
| Endpoint Approved | Human: approve | link_id, endpoint, principal_id, snapshot_ids, context_hash | Check Both Sides |
| Link Aligned | System: both approved | link_id, agreement_snapshot | — |
| Link Reaffirmed | Human: reaffirm | link_id, principal_id | Set Link Stale, Notify Counterparty |
| Link Withdrawn | Human: withdraw | link_id, principal_id, reason | Archive Link Immediately |
| Link Stale | System: content changed | link_id, changed_section_id, waiting_on | Notify per Intervention |
| Link Broken | System: section deleted/unresolvable | link_id, reason | Notify Both Sides |
| Link Archived | Human: archive | link_id, reason | — |
| Impacted Links Identified | System: graph walk | section_id, link_ids[] | Show Impact Preview |
| Source Imported | Adapter: import | adapter_type, doc_ids[], section_count | — |
| Source Synced | Adapter: change | adapter_type, changed_sections[] | Walk Graph |
| Links Batch Proposed | AI: analyze | link_ids[], rationale[] | Route Review |
| Policy Created | Admin: create | policy_id, selector, predicate, surface, severity | — |
| Gate Evaluated | System or admin: trigger | gate_id, surface, pass/fail, failing_objects[] | Route Violations |
| Violation Created | System: gate fail | violation_id, policy_id, failing_ref, message | Notify Responsible |
| Waiver Granted | Admin: waive | waiver_id, violation_id, reason, expires_at | Re-evaluate Gate |
| Waiver Expired | System: time | waiver_id | Re-open Violation |
| Subscription Created | Human: subscribe | sub_id, section_ref, tag_expression | — |
| Subscription Fired | System: tag match | sub_id, triggering_doc_id, matched_tag | Notify Subscriber |
| Subscription Dismissed | Human: dismiss | sub_id, triggering_doc_id | — |
| Intervention Changed | Human: update | link_id, endpoint, old_level, new_level | — |
| Approval Rejected Stale | System: snapshot mismatch | link_id, endpoint, expected_snapshot, current_snapshot | Notify Reviewer (refresh required) |
