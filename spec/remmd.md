# remmd — Specification

**Status:** v1 normative
**Sequence diagram:** [interactive](https://diashort.apps.quickable.co/d/840bab0b) · [embed](https://diashort.apps.quickable.co/e/840bab0b)

---

## 1. Design Axioms

1. **Don't derive, match.** Documents are independent claims. No hierarchy. Agreement is bilateral.
2. **Content is immediate, links are reviewed.** Every edit is versioned. Links are the only entity requiring bilateral approval.
3. **Sections are the accountable unit.** Agreement, approval, and coverage converge at the section level.
4. **Trust attaches to exact snapshots.** Stale context cannot approve current truth.
5. **AI drafts, humans approve.** Service principals MUST NOT approve, reject, or waive. Only humans create trust.
6. **The mechanism is universal.** Native docs, code paths, Jira tickets, Figma frames — all become sections with `@refs` and content hashes. Only node types expand.
7. **The graph is the product.** Every edit walks the graph. Impact before action. Cascading changes tracked as causal chains.

---

## 2. Content Model

### 2.1 Documents and Sections

A document is a container of sections belonging to a tenant. The system parses structure and assigns stable `@refs` to every section automatically — no manual marking required.

Sections are the minimum accountable unit. `@refs` are system-assigned, stable across edits. Sections MAY be nested (parent-child), are hierarchical and non-overlapping, and MAY carry tags for classification and discovery.

### 2.2 Versioning and Deletion

Content edits are immediate — no draft/activate lifecycle. Every edit creates a new version of the section. Previous versions are immutable and retained. Content hashes per section drive change detection.

Deletion includes reason and optional replacement refs. Deletion impacts all links containing the section — a link survives with remaining sections, or breaks if none are left. Counterparties review via their link threads.

### 2.3 Content Types

Every section has four primitives regardless of where the content lives:

| Primitive | Required | Description |
|---|---|---|
| `@ref` | yes | Stable identity within remmd |
| `hash` | yes | Content hash for change detection |
| `metadata` | yes | JSON — provenance, location, system-specific context |
| `body` | native only | The actual content (e.g., markdown) |

Links, threads, approvals, graph walks, and blast radius operate identically across all content types. Content type affects storage, detection, and review UX — never the trust mechanism.

#### 2.3.1 Native Content

Content stored and managed by remmd. This is the default.

- Body is stored in remmd (markdown by default)
- Hash is derived automatically (see `ref-content-hashing`)
- `@refs` are system-assigned: `@s1`, `@s2`, etc.
- Metadata: `{ "system": "native", "format": "markdown" }`
- Full detection, diff, and rendering built-in

#### 2.3.2 External Content

Content that lives in an external system. remmd stores only identity, hash, and metadata — not the body.

- `@refs` are namespaced: `@ext:<system>/<external_id>` (e.g., `@ext:notion/page-abc`, `@ext:figma/frame-123`)
- Hash is provided by source — opaque to remmd
- Metadata is flexible JSON: `{ "system": "notion", "page_id": "abc123", "workspace": "acme" }`
- Detection: manual or push (CLI/API). Diff: optional, provided on hash push. Rendering: reviewer verifies at source.
- Granularity: one source = one section by default. MAY register multiple sections per document.
- Review basis: `external-verify` — reviewer attests they verified at the external source. Recorded in approval.

---

## 3. Links

A link is the only entity that requires review — the agreement between section(s) and section(s), potentially across multiple documents. One link, one thread, one approval. Links are cross-document only (v1).

**Relationship types** (exactly one per link): `agrees_with` (symmetric — both sides state the same thing), `implements` (directional — one side implements the other's specification), `tests` (directional — one side verifies the other's claims), `evidences` (directional — one side provides evidence for the other's claims).

**Rationale:** Every link MUST include **claim** (why), **scope** (what's covered), **exclusions** (what's not). MAY be AI-drafted. Approval includes approval of the rationale.

**Proposal:** Creating a link is a proposal (like opening a PR). Both sides must approve for the link to become ALIGNED. The proposer's side is pending, not auto-approved.

**Intervention** (per endpoint, operational, not semantic): `watch` (dashboard only), `notify` (normal-priority task), `urgent` (high-priority + notification), `blocking` (high-priority + blocks matching gates). Changing intervention does NOT invalidate approvals.

---

## 4. Review Model

Review is thread-based, like PR code review. Every link has a persistent thread — the review workspace. Comments, diffs, rationale, decisions, and system events all live in the thread. Threads accumulate across review cycles. Comments are immutable after creation.

### 4.1 Review Flow

`propose link → thread opens → comment/iterate → both approve → ALIGNED`

Same as a PR: push, review, request changes, push, approve. Edits are immediate, versioned, and diff appears in thread.

### 4.2 Content Change Review

When content changes in an aligned link's section:

1. Graph walk detects impacted links → author sees **impact preview** (grouped by counterparty, relationship type, last aligned state)
2. Author **reaffirms** ("I still stand behind this") or **withdraws** (archives immediately with reason)
3. Reaffirmed links become STALE → counterparty reviews **cumulative diff** since last ALIGNED in thread
4. Counterparty approves or requests changes (iterate in thread)

Bulk reaffirm is supported when multiple links are impacted by one edit.

### 4.3 Link States

Derived from approval status, not explicitly set: `pending` (not yet approved by both sides) → `aligned` (both approved against current snapshots) → `stale` (content changed, waiting on counterparty). Also: `broken` (section deleted or became unresolvable), `archived` (explicitly closed by a participant). No "disputed" state — rejection is just "request changes" and iterate.

**Stale-context guard:** If section snapshots change between review-load and approve-submit, the approval MUST be rejected. Reviewer must refresh before re-approving.

---

## 5. Graph

The graph is the trust network. Nodes are sections, edges are links.

Every content edit triggers a graph walk from the changed section. The system identifies all impacted links, updates their threads with change context, and shows the author the full blast radius before taking stance.

Changes propagate as causal chains: PM edits `@r2` → Eng link goes STALE → Eng updates `@i2` → QA link goes STALE. Each thread captures *why* the review was triggered. Impact preview groups affected links by counterparty, relationship type, and last aligned state — shown at edit time, not as an afterthought.

---

## 6. Hash Updates

Content hashes reach remmd through two channels: **built-in** (remmd computes hash on edit for native content) and **push** (external system calls CLI/API with `{ref, new_hash, ?diff}`). No special path for external vs internal — the graph doesn't care how the hash arrived.

Bulk import: a service principal MAY import multiple documents at once and propose links. Each link still requires human bilateral approval — nothing is trusted until approved.

---

## 7. Tag Subscriptions

Sections MAY carry tags for classification and discovery. Tags follow the same versioning as content.

A subscription is a standing notification request: "tell me when content with this tag appears." Fires on: new content with matching tag, or tag added to existing content. Notification includes the new document, matched tag, and subscriber's section for context.

Subscriptions create notifications, NOT links. The subscriber decides manually whether to propose a link.

---

## 8. Principals

**Human principals** MAY: create/edit/delete content, propose/approve/reaffirm/withdraw/archive links, comment in threads.

**Service principals** (AI, adapters, integrations) MAY: create/edit content, propose links, post comments, suggest repairs. Service principals MUST NOT approve, reject, reaffirm, or withdraw links.

Every trust action MUST record the acting human principal, the exact section snapshots reviewed, and a timestamp.

---

## 9. Error Surface

Service principals (LLMs, integrations) are first-class consumers. When something fails, the error must give the caller enough to self-correct without human intervention.

### 9.1 Error Structure

Every error carries: `code` (stable string — the only field callers match on), `entity` + `id` (what failed), `message` (human-readable, full sentence), `fields` (structured context, e.g., expected vs actual hash), `remediation` (what to do next — specific command or action).

### 9.2 Error Codes

`NOT_FOUND`, `STALE_CONTEXT`, `UNAUTHORIZED`, `CONFLICT`, `INVALID_REF`, `INVALID_METADATA`, `DUPLICATE`, `CONTENT_TYPE_MISMATCH`, `VALIDATION`. Each includes a remediation hint specific to the failure.

### 9.3 Output Modes

- **Human (default)** — `message` to stderr, structured output to stdout
- **Machine (`--json`)** — full error/success structure as JSON to stdout

Success responses are structured: `{"ok": true, "entity": "section", "id": "@s3", "action": "created", "fields": {...}}`. The caller never needs to parse human-readable output or run a follow-up query.

---

## 10. Non-Goals (v1)

- Arbitrary overlapping inline spans
- Same-document links
- Automated trust creation (AI approving links)
- N-of-M approver rules (v1 is 1-of-N only)
- Semantic equivalence inference that auto-approves
- Partial document activation
- Anonymous or unauthenticated approval

---

## 11. Invariants

The implementation is correct only if all of the following remain true:

1. Content edits are immediate and versioned
2. Links are the only reviewed entity — bilateral approval required
3. Trust attaches to exact section snapshots
4. Changed side must reaffirm or withdraw impacted links
5. Review is thread-based iteration (no special states beyond the thread)
6. Service principals never satisfy trust actions
7. Graph walks on every change — blast radius is shown before action
8. Cascading changes are tracked as causal chains in threads
9. External content participates in the graph identically to native content — the mechanism never changes
10. Every decision is immutable and auditable
11. Stale context cannot approve current truth
