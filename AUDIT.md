# OpenSSL 4 Migration Audit (2026-05-22)

Tracking issue: Homebrew/homebrew-core#278366

## Summary

- Staging-scope formulae: 57
- Live pending: 18
- Live done: 39 (68.4%)
- Open staging PRs: 18
- Draft migration PRs: 17
- PRs with merge/check blockers: 17
- Pending formulae without open migration PRs: 0

## Retarget to Staging

Open staging-scope migration PRs whose base branch is not openssl-4-migration-staging.

| Formula | PR | Current Base | Expected Base | Target | Readiness |
|---|---|---|---|---|---|
| _none_ |  |  |  |  |  |

## Staging Priority

Pending staged formulae are sorted by transitive dependent count.

| Formula | Target | Depth | Impact | Status | PR | Readiness | Upstream | Issues |
|---|---|---:|---:|---|---|---|---|---|
| python@3.14 | openssl-4-migration-staging | 0 | 469 | PENDING | #280846 | draft, checks-blocked, merge-unstable | python |  |
| rust | openssl-4-migration-staging | 1 | 269 | PENDING | #280863 | draft, checks-blocked, merge-unstable | github:rust-lang/rust | [issues#155397](https://github.com/rust-lang/rust/issues/155397) open |
| python@3.13 | openssl-4-migration-staging | 0 | 80 | PENDING | #280845 | draft, checks-blocked, merge-unstable | python |  |
| systemd | openssl-4-migration-staging | 1 | 56 | PENDING | #280864 | draft, checks-blocked, merge-unstable | github:systemd/systemd |  |
| cargo-c | openssl-4-migration-staging | 2 | 25 | PENDING | #280867 | draft, checks-blocked, merge-unstable | github:lu-zero/cargo-c |  |
| pulseaudio | openssl-4-migration-staging | 1 | 21 | PENDING | #280861 | draft, checks-blocked, merge-unstable | gitlab:gitlab.freedesktop.org/pulseaudio/pulseaudio |  |
| cryptography | openssl-4-migration-staging | 2 | 15 | PENDING | #280868 | draft, checks-blocked, merge-unstable | github:pyca/cryptography | [issues#14656](https://github.com/pyca/cryptography/issues/14656) closed |
| grpc | openssl-4-migration-staging | 0 | 11 | PENDING | #280832 | draft, checks-blocked, merge-unstable | github:grpc/grpc | [issues#42020](https://github.com/grpc/grpc/issues/42020) open |
| qtbase | openssl-4-migration-staging | 1 | 10 | PENDING | #280862 | draft, checks-blocked, merge-unstable | qt |  |
| libfido2 | openssl-4-migration-staging | 0 | 9 | PENDING | #280836 | draft, checks-blocked, merge-unstable | github:Yubico/libfido2 | [issues#966](https://github.com/Yubico/libfido2/issues/966) closed |
| node | openssl-4-migration-staging | 1 | 7 | PENDING | #280858 | draft, checks-blocked, merge-unstable | github:nodejs/node | [issues#62817](https://github.com/nodejs/node/issues/62817) closed |
| s2n | openssl-4-migration-staging | closure | 7 | PENDING | #280912 | draft, checks-blocked, merge-unstable | github:aws/s2n-tls | [issues#5783](https://github.com/aws/s2n-tls/issues/5783) open |
| thrift | openssl-4-migration-staging | closure | 5 | PENDING | #280913 | draft, checks-blocked, merge-unstable | github:apache/thrift |  |
| apache-arrow | openssl-4-migration-staging | 1 | 4 | PENDING | #280851 | draft, checks-blocked, merge-unstable | github:apache/arrow |  |
| gstreamer | openssl-4-migration-staging | 3 | 4 | PENDING | #280873 | draft, checks-blocked, merge-unstable | gitlab:gitlab.freedesktop.org/gstreamer/gstreamer |  |
| dotnet | openssl-4-migration-staging | 0 | 2 | PENDING | #280829 | draft, checks-blocked, merge-unstable | github:dotnet/dotnet |  |
| gdal | openssl-4-migration-staging | 2 | 2 | PENDING | #280870 | draft, checks-blocked, merge-unstable | github:OSGeo/gdal |  |
| php | openssl-4-migration-staging | 2 | 0 | PENDING | #280871 | ready | github:php/php-src |  |

## Upstream Issue Coverage Gaps

Top 20 pending staged formulae with upstream metadata and no curated upstream issue entry.

| Formula | Depth | Impact | Upstream | Search | Readiness |
|---|---:|---:|---|---|---|
| python@3.14 | 0 | 469 | python |  | draft, checks-blocked, merge-unstable |
| python@3.13 | 0 | 80 | python |  | draft, checks-blocked, merge-unstable |
| systemd | 1 | 56 | github:systemd/systemd | [issues](https://github.com/search?q=repo%3Asystemd%2Fsystemd+%22OpenSSL+4%22&type=issues) | draft, checks-blocked, merge-unstable |
| cargo-c | 2 | 25 | github:lu-zero/cargo-c | [issues](https://github.com/search?q=repo%3Alu-zero%2Fcargo-c+%22OpenSSL+4%22&type=issues) | draft, checks-blocked, merge-unstable |
| pulseaudio | 1 | 21 | gitlab:gitlab.freedesktop.org/pulseaudio/pulseaudio |  | draft, checks-blocked, merge-unstable |
| qtbase | 1 | 10 | qt |  | draft, checks-blocked, merge-unstable |
| thrift | closure | 5 | github:apache/thrift | [issues](https://github.com/search?q=repo%3Aapache%2Fthrift+%22OpenSSL+4%22&type=issues) | draft, checks-blocked, merge-unstable |
| apache-arrow | 1 | 4 | github:apache/arrow | [issues](https://github.com/search?q=repo%3Aapache%2Farrow+%22OpenSSL+4%22&type=issues) | draft, checks-blocked, merge-unstable |
| gstreamer | 3 | 4 | gitlab:gitlab.freedesktop.org/gstreamer/gstreamer |  | draft, checks-blocked, merge-unstable |
| dotnet | 0 | 2 | github:dotnet/dotnet | [issues](https://github.com/search?q=repo%3Adotnet%2Fdotnet+%22OpenSSL+4%22&type=issues) | draft, checks-blocked, merge-unstable |
| gdal | 2 | 2 | github:OSGeo/gdal | [issues](https://github.com/search?q=repo%3AOSGeo%2Fgdal+%22OpenSSL+4%22&type=issues) | draft, checks-blocked, merge-unstable |
| php | 2 | 0 | github:php/php-src | [issues](https://github.com/search?q=repo%3Aphp%2Fphp-src+%22OpenSSL+4%22&type=issues) | ready |

## Curated Upstream Issues

| Formula | Upstream | State | Status | Link | Note |
|---|---|---|---|---|---|
| cryptography | github:pyca/cryptography | closed | reference | [OpenSSL 4.0.0 support](https://github.com/pyca/cryptography/issues/14656) | Closed upstream support tracker for OpenSSL 4. |
| grpc | github:grpc/grpc | open | relevant | [Fix compile with OpenSSL 4.0](https://github.com/grpc/grpc/issues/42020) | Tracks upstream OpenSSL 4 compile compatibility. |
| libfido2 | github:Yubico/libfido2 | closed | reference | [Fix library build with OpenSSL 4.x](https://github.com/Yubico/libfido2/issues/966) | Closed upstream compatibility fix to inspect when validating Homebrew failures. |
| node | github:nodejs/node | closed | reference | [[openssl-4.0.0] - failing tests](https://github.com/nodejs/node/issues/62817) | Closed upstream test-failure tracker for OpenSSL 4. |
| rust | github:rust-lang/rust | open | relevant | [Build with OpenSSL-4.0.0 fails](https://github.com/rust-lang/rust/issues/155397) | Tracks upstream Rust build compatibility with OpenSSL 4. |
| s2n | github:aws/s2n-tls | open | relevant | [OpenSSL Support Roadmap](https://github.com/aws/s2n-tls/issues/5783) | Upstream roadmap for supported OpenSSL versions. |
