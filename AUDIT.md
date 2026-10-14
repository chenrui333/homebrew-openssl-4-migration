# OpenSSL 4 Migration Audit (2026-10-14)

Tracking issue: Homebrew/homebrew-core#278366

## Summary

- Staging-scope formulae: 56
- Live pending: 4
- Live done: 52 (92.9%)
- Open staging PRs: 0
- Draft migration PRs: 0
- PRs with merge/check blockers: 0
- Pending formulae without open migration PRs: 4

## Retarget to Staging

Open staging-scope migration PRs whose base branch is not openssl-4-migration-staging.

| Formula | PR | Current Base | Expected Base | Target | Readiness |
|---|---|---|---|---|---|
| _none_ |  |  |  |  |  |

## Staging Priority

Pending staged formulae are sorted by transitive dependent count.

| Formula | Target | Depth | Impact | Status | PR | Readiness | Upstream | Issues |
|---|---|---:|---:|---|---|---|---|---|
| python@3.13 | openssl-4-migration-staging | 0 | 10 | PENDING | none | missing-pr | python |  |
| python@3.12 | openssl-4-migration-staging | 0 | 3 | PENDING | none | missing-pr | python |  |
| tcl-tk@8 | openssl-4-migration-staging | 0 | 3 | PENDING | none | missing-pr | other |  |
| python@3.11 | openssl-4-migration-staging | 0 | 0 | PENDING | none | missing-pr | python |  |

## Upstream Issue Coverage Gaps

Top 20 pending staged formulae with upstream metadata and no curated upstream issue entry.

| Formula | Depth | Impact | Upstream | Search | Readiness |
|---|---:|---:|---|---|---|
| python@3.13 | 0 | 10 | python |  | missing-pr |
| python@3.12 | 0 | 3 | python |  | missing-pr |
| python@3.11 | 0 | 0 | python |  | missing-pr |

## Curated Upstream Issues

| Formula | Upstream | State | Status | Link | Note |
|---|---|---|---|---|---|
| cryptography | github:pyca/cryptography | closed | reference | [OpenSSL 4.0.0 support](https://github.com/pyca/cryptography/issues/14656) | Closed upstream support tracker for OpenSSL 4. |
| grpc | github:grpc/grpc | open | relevant | [Fix compile with OpenSSL 4.0](https://github.com/grpc/grpc/issues/42020) | Tracks upstream OpenSSL 4 compile compatibility. |
| libfido2 | github:Yubico/libfido2 | closed | reference | [Fix library build with OpenSSL 4.x](https://github.com/Yubico/libfido2/issues/966) | Closed upstream compatibility fix to inspect when validating Homebrew failures. |
| node | github:nodejs/node | closed | reference | [[openssl-4.0.0] - failing tests](https://github.com/nodejs/node/issues/62817) | Closed upstream test-failure tracker for OpenSSL 4. |
| rust | github:rust-lang/rust | open | relevant | [Build with OpenSSL-4.0.0 fails](https://github.com/rust-lang/rust/issues/155397) | Tracks upstream Rust build compatibility with OpenSSL 4. |
| s2n | github:aws/s2n-tls | open | relevant | [OpenSSL Support Roadmap](https://github.com/aws/s2n-tls/issues/5783) | Upstream roadmap for supported OpenSSL versions. |
