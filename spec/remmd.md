# remmd — Specification

**Status:** v1 normative
**Sequence diagram:** [interactive](https://diashort.apps.quickable.co/d/840bab0b) · [embed](https://diashort.apps.quickable.co/e/840bab0b)

---

## 1. Purpose

remmd is the system of record for **human-approved agreement** between independently maintained document sections.

Documents are not derived from each other. A requirement, a design, an implementation, a test plan — each is an independent claim authored by a different person for a different purpose. They are supposed to *match*. Nobody checks whether they do.

remmd lets people say "these things should match" and enforces that both sides agree. When content changes, the system walks the graph, surfaces what's impacted, and routes review. Only humans create trust.

---

## 2. Design Axioms

1. **Don't derive, match.** Documents are independent claims. There is no hierarchy between them. Agreement is bilateral.

2. **Content is immediate, links are reviewed.** Authors own their content and edit freely. Every edit is versioned. Links — the agreements between sections — are the only entity that requires review and bilateral approval.

3. **Sections are the accountable unit.** Whole documents are too coarse. Agreement, approval, coverage, and policy all converge at the section level.

4. **Trust attaches to exact snapshots.** An agreement is approved against a specific combination of section versions and relationship definition. Stale context cannot approve current truth.

5. **AI drafts, humans approve.** Service principals may create content, propose links, suggest rationale, and route work. Service principals MUST NOT approve, reject, or waive. Only humans create trust.

6. **The mechanism is universal.** Native docs, code paths, Jira tickets, Figma frames, Confluence pages — all become sections with `@refs` and content hashes. The link/thread/approval mechanism never changes. Only node types expand.

7. **The graph is the product.** Every edit walks the graph. Impact is shown before action. Cascading changes are tracked as causal chains across threads.

---

## 3. Content Model

### 3.1 Documents

A document is a container of sections belonging to a tenant.

- Created via CLI or source adapter
- DX-first: a single call can create a full document from markdown-ish input
- The system parses structure and assigns stable `@refs` to every section automatically
- Documents have an owner, belong to a tenant, and carry metadata

### 3.2 Sections

A section is the minimum accountable unit. Every heading, block, or structural element gets a stable `@ref` from the moment it exists.

- `@refs` are system-assigned, stable across edits (like agent-browser element refs)
- Sections MAY be nested (parent-child)
- Sections are hierarchical and non-overlapping within a document
- No manual marking required — every section is linkable by default
- A section MAY carry tags for classification, policy, and discovery

### 3.3 Versioning

Content edits are immediate. There is no draft/activate lifecycle for content.

- Every edit creates a new version of the section
- Previous versions are immutable and retained
- The system tracks content hashes per section for change detection
- Version history is always accessible

### 3.4 Deletion

Deleting a section is a content operation that carries metadata.

- Deletion includes reason and optional replacement refs (like a PR description)
- Deletion impacts all links containing the section
- Counterparties review the deletion via their link threads
- A link survives with remaining sections, or breaks if none are left

---

## 4. Links

A link is the only entity that requires review. It is the agreement between sections.

### 4.1 Structure

A link connects **section(s) to section(s)**, potentially across multiple documents. Like a GitHub PR that spans multiple files.

- One link, one thread, one approval — covers the whole group
- Each link has a relationship type, rationale, scope, and exclusions
- Each link endpoint has an independent intervention setting (how loudly to signal on change)
- Links are cross-document only (v1)

### 4.2 Relationship Types

v1 ships with four relationship types:

| Type | Meaning |
|---|---|
| `agrees_with` | Both sides state the same thing (symmetric) |
| `implements` | One side implements the other's specification (directional) |
| `tests` | One side verifies the other's claims (directional) |
| `evidences` | One side provides evidence for the other's claims (directional) |

- A link MUST have exactly one relationship type
- Relationship type affects review prompts and policy, not the approval mechanism

### 4.3 Rationale

Every link MUST include:

- **claim** — why the relationship exists
- **scope** — what is covered
- **exclusions** — what is intentionally not covered

Rationale MAY be AI-drafted. Approval of the link includes approval of the rationale.

### 4.4 Proposal

Creating a link is a proposal — like opening a PR.

- `remmd link @a1 @a2 --implements @b1 @b2 --rationale "..."`
- A review thread opens immediately
- The proposer's side is pending (not auto-approved — it's a request)
- Both sides must approve for the link to become ALIGNED

### 4.5 Intervention

Each link endpoint has an independent intervention setting controlling review urgency:

| Level | Behavior |
|---|---|
| `watch` | Visible in dashboards, no routed task |
| `notify` | Normal-priority review task |
| `urgent` | High-priority task + immediate notification |
| `blocking` | High-priority task + blocks matching gates until resolved |

- Intervention is operational, not semantic
- Changing intervention does NOT invalidate approvals
- Same section can have `watch` on one link and `blocking` on another

---

## 5. Review Model

Review is thread-based, exactly like code review on a PR. There is no special ceremony.

### 5.1 Threads

Every link has a persistent thread. The thread is the review workspace.

- Comments, diffs, rationale, and decisions all live in the thread
- Threads accumulate across all review cycles — when a link goes STALE six months later, the reviewer sees the full negotiation history
- System events (content changes, version diffs, status transitions) appear in the thread alongside human comments
- Comments are immutable after creation

### 5.2 Review Flow

```
propose link → thread opens → comment/iterate → both approve → ALIGNED
```

1. Proposer opens link with rationale
2. Counterparty reviews in thread — sees section content, rationale, scope
3. Either side comments, author edits content if needed (immediate, versioned, diff appears in thread)
4. When both sides approve, the link is ALIGNED against current section snapshots

No special states. Just keep iterating in the thread until both sides approve. Same as a PR: push, review, request changes, push, approve.

### 5.3 Content Change Review

When content changes in a section that belongs to an aligned link:

1. System detects impacted links via graph walk
2. Author sees **impact preview** — all impacted links grouped by counterparty, relationship type, and last aligned state. Shown at edit time, not as an afterthought.
3. Author reaffirms or withdraws each impacted link
   - **reaffirm** — "I still stand behind this relationship after my change"
   - **withdraw** — archives the link immediately with reason. Counterparty is notified but cannot block. Human made the call.
4. Reaffirmed links become STALE (waiting on counterparty)
5. Counterparty reviews in thread — sees **cumulative diff** since last ALIGNED (not individual edits), full thread history
6. Counterparty approves or requests changes (iterate in thread)

Bulk reaffirm is supported when multiple links are impacted by one edit.

### 5.4 Link States

Derived from approval status, not explicitly set:

| State | Meaning |
|---|---|
| `pending` | Proposed, not yet approved by both sides |
| `aligned` | Both sides approved against current snapshots |
| `stale` | Content changed since last aligned — waiting on counterparty |
| `broken` | A section was deleted or became unresolvable |
| `archived` | Explicitly closed by a participant |

No "disputed" state. A rejection is just "request changes" — the thread stays open, the author iterates, the counterparty re-reviews. Same as a PR.

### 5.5 Stale-Context Guard

If the underlying section snapshots change between when a reviewer loads the review workspace and when they submit approval, the approval MUST be rejected. The reviewer sees a stale-context notice and must refresh before re-approving. This prevents approving content the reviewer has not actually seen.

---

## 6. Graph

The graph is the trust network. Nodes are sections, edges are links.

### 6.1 Graph Walk

Every content edit triggers a graph walk from the changed section.

- The system identifies all links containing the changed section
- Each impacted link's thread is updated with the change context
- The author sees the full blast radius before taking stance

### 6.2 Cascading Changes

Changes propagate through the graph as a causal chain.

Example:
1. PM edits `@r2` (requirement) → links to Eng and QA go STALE
2. Eng updates `@i2` (implementation follows the requirement change) → links to QA go STALE again
3. QA updates `@t2` (tests follow the implementation change)

Each thread captures the causal context: *why* this review was triggered, what upstream change caused it.

### 6.3 Impact Preview

Before reaffirming, the author sees:

- All impacted links grouped by counterparty
- Relationship type and last aligned state per link
- Which counterparties will need to review

This is shown at edit time, not as an afterthought.

---

## 7. Source Adapters

Any external system becomes a node type via a source adapter. The mechanism (links, threads, approvals, graph walks) stays identical.

### 7.1 Adapter Contract

An adapter MUST implement:

| Method | Input | Output |
|---|---|---|
| `emit_sections` | source reference | `[{id, content, hash}]` |
| `on_change` | source reference | `[{section_id, old_hash, new_hash}]` |
| `render_preview` | section reference | displayable content for review UI |

That's it. Links, threads, approvals, graph walks, ripple, and blast radius are all inherited from core.

### 7.2 Code Adapter

The git/code adapter watches file paths and emits sections per file, function, or code block.

- Content hash is per code path (file hash, function hash) — NOT the git commit hash
- A commit touching `src/refund/handler.ts` changes that section's hash → graph ripple
- Code sections are reviewable in link threads with syntax-highlighted diffs

### 7.3 Built-in Adapters (v1)

| Adapter | Section granularity | Hash source |
|---|---|---|
| Native (remmd docs) | headings/blocks | content hash |
| Git/code | file or function path | file/function content hash |
| Jira | ticket fields (description, acceptance criteria, subtasks) | field content hash |

Additional adapters (Confluence, Figma, Notion, etc.) follow the same contract.

### 7.4 Bulk Import

An adapter MAY import multiple documents at once. An AI/service principal MAY propose links across the imported content and existing graph in the same operation.

- Bulk import creates documents + sections + draft link proposals
- Each link still requires human bilateral approval
- Nothing is trusted until approved

### 7.5 External Updates

When an external source changes:

- The adapter emits new hashes for affected sections
- New versions are recorded in remmd
- Graph walks and thread updates follow the standard mechanism
- No special path for external vs internal changes

---

## 8. Tag Subscriptions

Tags classify sections for policy, filtering, and discovery. Subscriptions notify owners when new matching content appears.

### 8.1 Tags

A section MAY carry tags for classification, policy selection, and discovery.

- Tags follow the same versioning as content — every tag mutation is recorded
- Tags that influence policy MUST belong to the active version

### 8.2 Subscriptions

A subscription is a standing notification request: "tell me when content with this tag appears."

- `remmd subscribe @section --tag "payment"`
- When new content with a matching tag is activated, the subscriber is notified
- Notification includes: new document, matched tag, subscriber's section for context
- Subscriber decides: create a link (enters normal bilateral flow) or dismiss
- Subscriptions fire on: new document activation with matching tag, or tag added to existing content
- Subscriptions evaluate only against active content and active tags

Subscriptions create notifications, NOT links. The subscriber decides manually whether to propose a link.

---

## 9. Principals

### 9.1 Human Principals

Human principals MAY:

- Create, edit, and delete content
- Propose, approve, reaffirm, withdraw, and archive links
- Comment in threads
- Grant waivers

### 9.2 Service Principals

Service principals (AI, adapters, integrations) MAY:

- Create and edit content (auto-created docs, imports, generated drafts)
- Propose links with rationale
- Post system comments and summaries
- Suggest repairs and route work

Service principals MUST NOT:

- Approve or reject links
- Reaffirm or withdraw links
- Grant waivers

### 9.3 Invariant

Every approval, rejection, reaffirmation, withdrawal, and waiver MUST record the acting human principal, the exact section snapshots reviewed, and a timestamp.

---

## 10. Policy and Gates

Policies turn the trust graph into operational control.

### 10.1 Policy

A policy defines a predicate that must hold at a gate.

- **selector** — which sections or links the policy applies to (by tag, type, document)
- **predicate** — what must be true (e.g., "at least one aligned `tests` link")
- **surface** — where it's enforced (merge, publish, release, audit, dashboard)
- **severity** — blocking or warning

### 10.2 Gate Check

A gate check evaluates policies against the current graph and returns:

- pass or fail
- exact failing sections/links
- what's missing and who needs to act
- waiver status if present

Gate failures are always explainable. No aggregate "trust score low" messages.

### 10.3 Waivers

A waiver is a human decision to bypass a policy violation temporarily.

- Waivers have a reason and expiry
- Waivers are visible as waived (not hidden)
- Expired waivers re-open violations automatically
- Waivers do NOT create trust — they bypass gates

Only human principals MAY grant waivers.

---

## 11. Non-Goals (v1)

- Arbitrary overlapping inline spans
- Same-document links
- Automated trust creation (AI approving links)
- N-of-M approver rules (v1 is 1-of-N only)
- Semantic equivalence inference that auto-approves
- Partial document activation
- Anonymous or unauthenticated approval

---

## 12. Invariants

The implementation is correct only if all of the following remain true:

1. Content edits are immediate and versioned
2. Links are the only reviewed entity — bilateral approval required
3. Trust attaches to exact section snapshots
4. Changed side must reaffirm or withdraw impacted links
5. Review is thread-based iteration (no special states beyond the thread)
6. Service principals never satisfy trust actions
7. Graph walks on every change — blast radius is shown before action
8. Cascading changes are tracked as causal chains in threads
9. Source adapters add node types without changing the mechanism
10. Every decision is immutable and auditable
11. Stale context cannot approve current truth
