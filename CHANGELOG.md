# Changelog

## 0.1.1 - 2026-05-10

- Compute OpenSSL 4 target branches from the full transitive staging closure, including blockers reached through bridge formulae that do not directly depend on OpenSSL.
- Add target-branch, staging-reason, staging-required-by, and transitive OpenSSL dependency metadata to data/dep_tree.json.
- Group status/checklist output into explicit depth batches, staging closure formulae, and main-track leaves.
- Map open migration PRs by changed formula files so multi-formula PRs are tracked correctly.
- Rename repository agent guidance from CLAUDE.md to AGENTS.md for Codex-based harness work.
