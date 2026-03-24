# Autoresearch: Trim spec without reducing capability coverage

## Config
- **Benchmark**: `bash autoresearch.sh`
- **Target metric**: `words` (lower is better)
- **Constraint**: `coverage_pct` must stay at 100 (65/65 checks)
- **Scope**: `spec/remmd.md`
- **Branch**: `autoresearch/trim-spec-keep-capabilities`
- **Started**: 2026-03-24

## Rules
1. One change per experiment
2. Run benchmark after every change
3. Keep if words decrease AND coverage stays 100%, discard otherwise
4. Log every run to autoresearch.jsonl
5. Commit kept changes with `Result:` trailer
