# remmd — Spec Gaps

Jobs implied by remmd.md that have no JTBD or UX design.

---

## Gap 1: Reaffirm / Withdraw After Editing

**Who:** Author (content owner)
**Spec ref:** remmd.md §5.3

When an author edits a section that belongs to aligned links, the system requires an explicit stance on each impacted link before the counterparty is notified.

**Job:** When I edit my section and it's part of active links, I need to quickly see which links are impacted and declare whether I still stand behind each one — so counterparties only review what's genuinely changed in meaning, not every whitespace edit.

**Missing interaction:**
- Impact preview at edit time (blast radius)
- Inline reaffirm/withdraw per link
- Bulk reaffirm when editing affects many links
- Withdraw flow (what happens to the link thread, what counterparty sees)

---

## Gap 2: Thread-Based Review

**Who:** Author + Reviewer
**Spec ref:** remmd.md §5.1, §5.2

Review is not a 3-button transaction (verify/dismiss/flag). It's iterative thread-based negotiation — comment, edit, diff appears, comment again, approve. Like a PR code review.

**Job (Author):** When my link proposal gets comments, I need to respond in-thread, push content updates that show as diffs, and iterate until both sides are satisfied.

**Job (Reviewer):** When I'm reviewing a link, I need to comment with specific concerns, see the author's content updates as diffs in the thread, and approve only when the thread is resolved.

**Missing interaction:**
- Thread UI within link workspace
- Inline diff rendering when linked sections are edited during review
- Thread history spanning multiple review cycles (stale → review → aligned → stale again)
- Comment-to-section reference (pointing at specific content within the thread)

---

## Gap 3: Multi-Section Links

**Who:** Author
**Spec ref:** remmd.md §4.1

A link connects section(s) to section(s) across documents — like a PR spanning multiple files. One link, one thread, one approval.

**Job:** When my implementation spans three files and maps to two requirement sections, I need to express that as one relationship — not six individual links.

**Missing interaction:**
- Multi-select sections during link creation
- Visual representation of grouped sections in a link
- Adding/removing sections from an existing link
- How deletion of one section in a group affects the link (shrink vs break)

---

## Gap 4: Structured Rationale

**Who:** Author
**Spec ref:** remmd.md §4.3

Every link requires claim + scope + exclusions — not a single "reason" text field.

**Job:** When I create a link, I need to articulate what's covered, what's intentionally excluded, and why this relationship matters — so reviewers can evaluate against the scope, not guess at intent.

**Missing interaction:**
- Structured rationale form (claim, scope, exclusions) vs single reason field
- How rationale is displayed in review threads
- AI-assisted rationale drafting
- Rationale updates when link scope changes

---

## Gap 5: Intervention Setting Per Endpoint

**Who:** Author + Reviewer
**Spec ref:** remmd.md §4.5

Each link endpoint has an independent intervention level: watch, notify, urgent, blocking.

**Job:** When I'm loosely tracking a cross-team dependency vs tightly coupled to a release-critical spec, I need different notification urgency on different links — without affecting the other side's settings.

**Missing interaction:**
- Where/how to set intervention level (link creation? link panel? thread?)
- Visual indicator of intervention level per link
- Changing intervention without invalidating approvals

---

## Gap 6: Source Adapters (Code, Jira, Figma)

**Who:** Author + Admin
**Spec ref:** remmd.md §7

External systems become graph nodes via adapters. Code paths have content hashes. Jira tickets become sections. All use the same link/thread/approval mechanism.

**Job (Author):** When my requirement links to `src/refund/handler.ts`, I need to see code changes as diffs in my link thread and approve/request-changes the same way I would for any other section.

**Job (Admin):** When onboarding a new source (Jira project, Confluence space), I need to bulk-import content as sections and have AI propose links to the existing graph — then route review to the right people.

**Missing interaction:**
- Adapter configuration UI
- Code section rendering in review threads (syntax-highlighted diff)
- Bulk import flow with AI-proposed links
- External source sync status and error handling

---

## Gap 7: Policy Gates and Waivers

**Who:** Admin
**Spec ref:** remmd.md §10

Policies evaluate the trust graph at decision points (merge, release, audit). Gate failures name exact failing objects. Waivers bypass temporarily.

**Job (Admin):** When a release gate fails, I need to see exactly which sections are non-compliant, what's missing (e.g., "no aligned tests link"), who needs to act, and whether to waive or fix.

**Missing interaction:**
- Policy creation/management UI
- Gate check results screen (pass/fail with drilldown)
- Waiver granting flow (reason, expiry, visibility)
- Dashboard: gate failures, active waivers, expiring waivers

---

## Gap 8: Causal Chain Visibility

**Who:** Reviewer
**Spec ref:** remmd.md §6.2

When changes cascade (PM edits requirement → Eng updates implementation → QA updates tests), each thread should show the causal chain — why this review was triggered.

**Job:** When I'm asked to re-review a link, I need to understand the upstream cause — was it a direct edit, or a cascade from someone else's change? Context determines how carefully I review.

**Missing interaction:**
- Causal chain indicator in thread ("triggered by @r2 change by PM, 2h ago")
- Ability to trace upstream through the chain
- Grouped notifications when multiple links are impacted by the same root cause

---

## Gap 9: Admin Trust Health (without aggregate scores)

**Who:** Admin
**Spec ref:** remmd.md §10.2

The spec explicitly rejects aggregate trust scores. But admins still need organizational health visibility.

**Job:** When I'm responsible for org alignment, I need to see which areas are healthy and which need attention — without a misleading single number.

**Replaces:** "Trust: 87%" → drillable dashboard showing:
- Stale links by age and counterparty
- Broken links with context
- Gate failures by surface (merge, release, audit)
- Active waivers approaching expiry
- Links waiting on specific people

---

## Gap 10: Blast Radius Preview (standalone surface)

**Who:** Author
**Spec ref:** remmd.md §5.3, §6.3

Impact preview is the first thing an author sees after any edit that touches linked sections. It's a precondition to reaffirm/withdraw — a distinct UX surface, not a detail within the reaffirm flow.

**Missing interaction:** Dedicated impact preview showing all impacted links grouped by counterparty, with relationship type and last aligned state, before any stance is taken.

---

## Gap 11: Cumulative Diff Rendering

**Who:** Reviewer
**Spec ref:** remmd.md §5.3

When a reviewer missed multiple changes, they see one merged diff since last ALIGNED — not three sequential diffs. Non-trivial rendering problem.

**Missing interaction:** Diff view that collapses multiple versions into one cumulative change set, with option to expand individual version steps.

---

## Gap 12: Broken Link UX

**Who:** Reviewer
**Spec ref:** remmd.md §5.4

When a section is deleted, links go `broken`. The counterparty needs to understand what happened and act.

**Missing interaction:** Broken link view showing: what was deleted, when, reason, replacement refs if provided. Actions: archive the link, or create new link to replacement.

---

## Gap 13: Stale-Context Approval Guard

**Who:** Reviewer
**Spec ref:** remmd.md §5.5

Reviewer clicks approve, but content changed since they loaded the review. Approval is rejected. What do they see?

**Missing interaction:** Stale-context notice with refresh action. Shows what changed since they loaded. One click to reload current snapshots and re-review.
