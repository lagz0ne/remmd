#!/usr/bin/env bash
set -euo pipefail

SPEC="spec/remmd.md"
PASS=0
FAIL=0
TOTAL=0

check() {
  TOTAL=$((TOTAL + 1))
  if grep -qiP "$1" "$SPEC"; then
    PASS=$((PASS + 1))
  else
    FAIL=$((FAIL + 1))
    echo "MISSING: $2" >&2
  fi
}

# --- Content Model ---
check 'document.*container.*sections|sections.*belonging' "documents contain sections"
check '@ref' "stable @refs"
check 'nested.*parent.child|parent.child|MAY be nested' "sections nested"
check 'tags.*classif|classif.*discovery' "section tags"
check 'immediate.*no draft|no draft.*lifecycle' "edits immediate, no draft"
check 'new version.*immutable|immutable.*retained|every edit.*new version' "versioning immutable"
check 'content hash.*change detection|hash.*change detection' "content hash for detection"
check 'delet.*reason.*replacement|reason.*replacement.*ref' "deletion with reason+replacement"
check 'survives.*remaining.*breaks|breaks.*none.*left' "deletion impact on links"
check 'native.*content|body.*stored.*remmd' "native content type"
check '@ext:' "external content @ext: refs"
check 'metadata.*JSON|JSON.*provenance' "external content metadata"
check 'push.*hash|hash.*push' "push hash updates"
check 'external.verify|basis.*external' "external review basis"

# --- Links ---
check 'section\(s\).*section\(s\)|sections.*to.*sections' "links section-to-section"
check 'cross.document' "links cross-document only"
check 'one link.*one thread|one thread.*one approval' "one link one thread"
check 'agrees_with' "relationship: agrees_with"
check 'implements' "relationship: implements"
check 'tests.*verif|verif.*claims' "relationship: tests"
check 'evidences' "relationship: evidences"
check '\*\*claim\*\*|\*\*scope\*\*|\*\*exclusions\*\*' "rationale structure"
check 'bilateral.*approval|both sides.*approve' "bilateral approval"
check 'watch.*notify.*urgent.*blocking|watch.*dashboard' "intervention levels"
check 'intervention.*operational|operational.*not semantic' "intervention is operational"

# --- Review ---
check 'thread.based|persistent thread' "thread-based review"
check 'immutable after creation|comments.*immutable' "comments immutable"
check 'system events.*thread|thread.*system events' "system events in thread"
check 'propose.*thread.*approve|propose.*iterate.*approve' "review flow"
check 'reaffirm.*withdraw' "reaffirm/withdraw"
check 'cumulative diff|cumulative.*since.*aligned' "cumulative diff"
check 'bulk reaffirm' "bulk reaffirm"
check 'pending.*approved|pending.*aligned|aligned.*stale' "link states"
check 'broken.*deleted|deleted.*unresolvable' "broken state"
check 'archived.*closed' "archived state"
check 'stale.context.*guard|stale.*context.*reject' "stale-context guard"

# --- Graph ---
check 'graph walk|walks the graph' "graph walk"
check 'blast radius' "blast radius"
check 'cascad.*causal|causal.*chain' "cascading causal chains"
check 'impact preview|impact.*preview' "impact preview"

# --- Hash Updates ---
check 'built.in.*native|remmd computes hash' "built-in hash channel"
check 'external.*calls.*CLI.*API|CLI.*API.*new hash' "push hash channel"
check 'bulk import' "bulk import"

# --- Tag Subscriptions ---
check 'subscription.*standing.*notification|standing.*notification' "subscriptions"
check 'subscriptions.*create.*notifications.*NOT.*links|notification.*NOT.*link' "subscriptions create notifications not links"

# --- Principals ---
check 'human principal' "human principals"
check 'service principal' "service principals"
check 'MUST NOT.*approve|MUST NOT.*reject' "service cannot approve"
check 'record.*principal.*snapshot.*timestamp|principal.*snapshots.*timestamp' "trust action audit"

# --- Error Surface ---
check 'error.*code.*entity.*message|code.*stable.*string' "structured errors"
check 'NOT_FOUND' "error: NOT_FOUND"
check 'STALE_CONTEXT' "error: STALE_CONTEXT"
check 'UNAUTHORIZED' "error: UNAUTHORIZED"
check 'CONFLICT' "error: CONFLICT"
check 'INVALID_REF' "error: INVALID_REF"
check 'INVALID_METADATA' "error: INVALID_METADATA"
check 'DUPLICATE' "error: DUPLICATE"
check 'CONTENT_TYPE_MISMATCH' "error: CONTENT_TYPE_MISMATCH"
check 'VALIDATION' "error: VALIDATION"
check '\-\-json|machine.*json|JSON.*stdout' "json output mode"
check 'remediation' "error remediation hints"
check '"ok".*true|success.*structured' "structured success responses"

# --- Non-Goals ---
check 'same.document links' "non-goal: same-doc links"
check 'AI.*approving|automated trust' "non-goal: automated trust"

# --- Invariants ---
check 'invariants' "invariants section exists"

# --- Metrics ---
WORDS=$(wc -w < "$SPEC")
LINES=$(wc -l < "$SPEC")
COVERAGE=$((PASS * 100 / TOTAL))

echo ""
echo "METRIC words=$WORDS"
echo "METRIC lines=$LINES"
echo "METRIC coverage_checks=$PASS/$TOTAL"
echo "METRIC coverage_pct=$COVERAGE"
echo "METRIC missing=$FAIL"
