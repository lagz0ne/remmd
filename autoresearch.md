# Autoresearch: Playbook Syntax Minimality

## Config
- **Benchmark**: `bash autoresearch.sh`
- **Target metric**: `keywords` (lower is better)
- **Constraint**: `coverage` must stay at 20 (all 20 patterns)
- **Scope**: `poc/playbook/candidate.yaml`, `poc/playbook/keywords.txt`
- **Branch**: `autoresearch/playbook-syntax`
- **Started**: 2026-03-25

## Goal
Find the most minimal YAML syntax for expressing document graph constraints.
Express types, schemas, edges, and CEL rules with the fewest reserved keywords.
Both c3-design and sft constraint patterns must be expressible.

## Rules
1. One change per experiment
2. Run benchmark after every change
3. Keep if keywords decrease AND coverage stays 20, discard otherwise
4. Log every run to autoresearch.jsonl
5. Commit kept changes with `Result:` trailer

## Patterns (20 total)
See `poc/playbook/patterns.yaml` for the full list.
P1-P7: structural, P8-P10: per-node, P11-P18: cross-node, P19-P20: meta.
