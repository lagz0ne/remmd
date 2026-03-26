#!/usr/bin/env bash
# Benchmark: playbook syntax keyword minimality + pattern coverage
# Goal: fewer keywords, 20/20 coverage
set -euo pipefail

CANDIDATE="poc/playbook/candidate.yaml"

# --- Count pattern coverage (P1..P20 markers) ---
COVERED=$(grep -oP 'P\d+' "$CANDIDATE" | sort -u | wc -l)

# --- Extract syntax keywords ---
# A syntax keyword is a YAML key that is part of the playbook language itself.
# We extract all keys, then subtract user-defined identifiers.
# User-defined = entity type names, field names, rule names, domain values
#
# Strategy: keywords file lists them explicitly (ground truth per experiment)
KEYWORDS_FILE="poc/playbook/keywords.txt"

if [[ ! -f "$KEYWORDS_FILE" ]]; then
  echo "ERROR: $KEYWORDS_FILE not found. List one syntax keyword per line." >&2
  exit 1
fi

KEYWORD_COUNT=$(grep -c '.' "$KEYWORDS_FILE" 2>/dev/null || echo 0)

# --- Verify each keyword actually appears in candidate ---
UNUSED=0
while IFS= read -r kw; do
  [[ -z "$kw" || "$kw" =~ ^# ]] && continue
  if ! grep -qP "(^|\s)${kw}(\s|:)" "$CANDIDATE" 2>/dev/null; then
    echo "UNUSED KEYWORD: $kw" >&2
    UNUSED=$((UNUSED + 1))
  fi
done < "$KEYWORDS_FILE"

# --- Output ---
echo ""
echo "METRIC keywords=$KEYWORD_COUNT"
echo "METRIC coverage=$COVERED"
echo "METRIC unused_keywords=$UNUSED"
