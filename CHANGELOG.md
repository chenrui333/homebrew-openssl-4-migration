# Changelog

## Unreleased

- Add a Go-generated MkDocs progress site with executive summary, depth-gate tracking, upstream blocker views, and GitHub Pages deployment.

## 0.3.0 - 2026-05-10

- Improve upstream metadata parsing for formulae that declare GitHub heads inside `head do` blocks.
- Add audit report sections for branch/base mismatches and upstream issue coverage gaps.

## 0.2.0 - 2026-05-10

- Add an audit report command and generated AUDIT.md for readiness, blocker, main-track opportunity, and upstream issue review.
- Add upstream source metadata to data/dep_tree.json and a curated data/upstream_issues.json dataset.
- Update the daily sync workflow to regenerate and commit audit/report datasets with a concurrency guard.

## 0.1.1 - 2026-05-10

- Compute OpenSSL 4 target branches from the full transitive staging closure, including blockers reached through bridge formulae that do not directly depend on OpenSSL.
- Add target-branch, staging-reason, staging-required-by, and transitive OpenSSL dependency metadata to data/dep_tree.json.
- Group status/checklist output into explicit depth batches, staging closure formulae, and main-track leaves.
- Map open migration PRs by changed formula files so multi-formula PRs are tracked correctly.
- Include unlabeled migration PRs in tracking by unioning migration-label searches with an openssl@4 title fallback.
- Rename repository agent guidance from CLAUDE.md to AGENTS.md for Codex-based harness work.
