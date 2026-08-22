# OpenSSL 4 Migration Audit (2026-08-22)

Tracking issue: Homebrew/homebrew-core#278366

## Summary

- Staging-scope formulae: 59
- Live pending: 50
- Live done: 8 (13.6%)
- Open staging PRs: 0
- Draft migration PRs: 0
- PRs with merge/check blockers: 0
- Pending formulae without open migration PRs: 50

## Retarget to Staging

Open staging-scope migration PRs whose base branch is not openssl-4-migration-staging.

| Formula | PR | Current Base | Expected Base | Target | Readiness |
|---|---|---|---|---|---|
| _none_ |  |  |  |  |  |

## Staging Priority

Pending staged formulae are sorted by transitive dependent count.

| Formula | Target | Depth | Impact | Status | PR | Readiness | Upstream | Issues |
|---|---|---:|---:|---|---|---|---|---|
| cmake | openssl-4-migration-staging | 0 | 671 | PENDING | none | missing-pr | gitlab:gitlab.kitware.com/cmake/cmake |  |
| python@3.14 | openssl-4-migration-staging | 0 | 472 | PENDING | none | missing-pr | python |  |
| libssh2 | openssl-4-migration-staging | 0 | 316 | PENDING | none | missing-pr | github:libssh2/libssh2 |  |
| libgit2 | openssl-4-migration-staging | closure | 293 | PENDING | none | missing-pr | github:libgit2/libgit2 |  |
| rust | openssl-4-migration-staging | 1 | 291 | PENDING | none | missing-pr | github:rust-lang/rust | [issues#155397](https://github.com/rust-lang/rust/issues/155397) open |
| python@3.13 | openssl-4-migration-staging | 0 | 84 | PENDING | none | missing-pr | python |  |
| systemd | openssl-4-migration-staging | 1 | 66 | PENDING | none | missing-pr | github:systemd/systemd |  |
| libevent | openssl-4-migration-staging | 0 | 64 | PENDING | none | missing-pr | github:libevent/libevent |  |
| libngtcp2 | openssl-4-migration-staging | closure | 34 | PENDING | none | missing-pr | github:ngtcp2/ngtcp2 |  |
| cargo-c | openssl-4-migration-staging | 2 | 25 | PENDING | none | missing-pr | github:lu-zero/cargo-c |  |
| curl | openssl-4-migration-staging | 1 | 20 | PENDING | none | missing-pr | github:curl/curl |  |
| pulseaudio | openssl-4-migration-staging | 1 | 20 | PENDING | none | missing-pr | gitlab:gitlab.freedesktop.org/pulseaudio/pulseaudio |  |
| ruby | openssl-4-migration-staging | 2 | 19 | PENDING | none | missing-pr | github:ruby/ruby |  |
| cryptography | openssl-4-migration-staging | 2 | 17 | PENDING | none | missing-pr | github:pyca/cryptography | [issues#14656](https://github.com/pyca/cryptography/issues/14656) closed |
| libpq | openssl-4-migration-staging | 1 | 17 | PENDING | none | missing-pr | other |  |
| pipewire | openssl-4-migration-staging | closure | 16 | PENDING | none | missing-pr | gitlab:gitlab.freedesktop.org/pipewire/pipewire |  |
| ffmpeg | openssl-4-migration-staging | 1 | 13 | PENDING | none | missing-pr | github:FFmpeg/FFmpeg |  |
| apr-util | openssl-4-migration-staging | 0 | 11 | PENDING | none | missing-pr | apache |  |
| grpc | openssl-4-migration-staging | 0 | 11 | PENDING | none | missing-pr | github:grpc/grpc | [issues#42020](https://github.com/grpc/grpc/issues/42020) open |
| libfido2 | openssl-4-migration-staging | 0 | 10 | PENDING | none | missing-pr | github:Yubico/libfido2 | [issues#966](https://github.com/Yubico/libfido2/issues/966) closed |
| openldap | openssl-4-migration-staging | 0 | 10 | PENDING | none | missing-pr | other |  |
| qtbase | openssl-4-migration-staging | 1 | 10 | PENDING | none | missing-pr | qt |  |
| aws-c-cal | openssl-4-migration-staging | closure | 9 | PENDING | none | missing-pr | github:awslabs/aws-c-cal |  |
| node | openssl-4-migration-staging | 1 | 9 | PENDING | none | missing-pr | github:nodejs/node | [issues#62817](https://github.com/nodejs/node/issues/62817) closed |
| s2n | openssl-4-migration-staging | closure | 9 | PENDING | none | missing-pr | github:aws/s2n-tls | [issues#5783](https://github.com/aws/s2n-tls/issues/5783) open |
| aws-c-io | openssl-4-migration-staging | closure | 8 | PENDING | none | missing-pr | github:awslabs/aws-c-io |  |
| httpd | openssl-4-migration-staging | 1 | 8 | PENDING | none | missing-pr | apache |  |
| folly | openssl-4-migration-staging | 1 | 7 | PENDING | none | missing-pr | github:facebook/folly |  |
| freetds | openssl-4-migration-staging | 0 | 7 | PENDING | none | missing-pr | github:FreeTDS/freetds |  |
| srt | openssl-4-migration-staging | 0 | 6 | PENDING | none | missing-pr | github:Haivision/srt |  |
| asio | openssl-4-migration-staging | 0 | 5 | PENDING | none | missing-pr | github:chriskohlhoff/asio |  |
| hiredis | openssl-4-migration-staging | 0 | 5 | PENDING | none | missing-pr | github:redis/hiredis |  |
| libssh | openssl-4-migration-staging | 0 | 5 | PENDING | none | missing-pr | other |  |
| mariadb-connector-c | openssl-4-migration-staging | 0 | 5 | PENDING | none | missing-pr | github:mariadb-corporation/mariadb-connector-c |  |
| thrift | openssl-4-migration-staging | closure | 5 | PENDING | none | missing-pr | github:apache/thrift |  |
| apache-arrow | openssl-4-migration-staging | 1 | 4 | PENDING | none | missing-pr | github:apache/arrow |  |
| bind | openssl-4-migration-staging | 1 | 4 | PENDING | none | missing-pr | gitlab:gitlab.isc.org/isc-projects/bind9 |  |
| erlang | openssl-4-migration-staging | 0 | 4 | PENDING | none | missing-pr | github:erlang/otp |  |
| gstreamer | openssl-4-migration-staging | 3 | 4 | PENDING | none | missing-pr | gitlab:gitlab.freedesktop.org/gstreamer/gstreamer |  |
| tcl-tk@8 | openssl-4-migration-staging | 0 | 4 | PENDING | none | missing-pr | other |  |
| unbound | openssl-4-migration-staging | 1 | 4 | PENDING | none | missing-pr | github:NLnetLabs/unbound |  |
| python@3.12 | openssl-4-migration-staging | 0 | 3 | PENDING | none | missing-pr | python |  |
| tcl-tk | openssl-4-migration-staging | 0 | 3 | PENDING | none | missing-pr | other |  |
| dotnet | openssl-4-migration-staging | 0 | 2 | PENDING | none | missing-pr | github:dotnet/dotnet |  |
| gdal | openssl-4-migration-staging | 2 | 2 | PENDING | none | missing-pr | github:OSGeo/gdal |  |
| librdkafka | openssl-4-migration-staging | 0 | 1 | PENDING | none | missing-pr | github:confluentinc/librdkafka |  |
| postgresql@17 | openssl-4-migration-staging | 1 | 1 | PENDING | none | missing-pr | other |  |
| postgresql@18 | openssl-4-migration-staging | 1 | 1 | PENDING | none | missing-pr | other |  |
| php | openssl-4-migration-staging | 2 | 0 | PENDING | none | missing-pr | github:php/php-src |  |
| python@3.11 | openssl-4-migration-staging | 0 | 0 | PENDING | none | missing-pr | python |  |

## Upstream Issue Coverage Gaps

Top 20 pending staged formulae with upstream metadata and no curated upstream issue entry.

| Formula | Depth | Impact | Upstream | Search | Readiness |
|---|---:|---:|---|---|---|
| cmake | 0 | 671 | gitlab:gitlab.kitware.com/cmake/cmake |  | missing-pr |
| python@3.14 | 0 | 472 | python |  | missing-pr |
| libssh2 | 0 | 316 | github:libssh2/libssh2 | [issues](https://github.com/search?q=repo%3Alibssh2%2Flibssh2+%22OpenSSL+4%22&type=issues) | missing-pr |
| libgit2 | closure | 293 | github:libgit2/libgit2 | [issues](https://github.com/search?q=repo%3Alibgit2%2Flibgit2+%22OpenSSL+4%22&type=issues) | missing-pr |
| python@3.13 | 0 | 84 | python |  | missing-pr |
| systemd | 1 | 66 | github:systemd/systemd | [issues](https://github.com/search?q=repo%3Asystemd%2Fsystemd+%22OpenSSL+4%22&type=issues) | missing-pr |
| libevent | 0 | 64 | github:libevent/libevent | [issues](https://github.com/search?q=repo%3Alibevent%2Flibevent+%22OpenSSL+4%22&type=issues) | missing-pr |
| libngtcp2 | closure | 34 | github:ngtcp2/ngtcp2 | [issues](https://github.com/search?q=repo%3Angtcp2%2Fngtcp2+%22OpenSSL+4%22&type=issues) | missing-pr |
| cargo-c | 2 | 25 | github:lu-zero/cargo-c | [issues](https://github.com/search?q=repo%3Alu-zero%2Fcargo-c+%22OpenSSL+4%22&type=issues) | missing-pr |
| curl | 1 | 20 | github:curl/curl | [issues](https://github.com/search?q=repo%3Acurl%2Fcurl+%22OpenSSL+4%22&type=issues) | missing-pr |
| pulseaudio | 1 | 20 | gitlab:gitlab.freedesktop.org/pulseaudio/pulseaudio |  | missing-pr |
| ruby | 2 | 19 | github:ruby/ruby | [issues](https://github.com/search?q=repo%3Aruby%2Fruby+%22OpenSSL+4%22&type=issues) | missing-pr |
| pipewire | closure | 16 | gitlab:gitlab.freedesktop.org/pipewire/pipewire |  | missing-pr |
| ffmpeg | 1 | 13 | github:FFmpeg/FFmpeg | [issues](https://github.com/search?q=repo%3AFFmpeg%2FFFmpeg+%22OpenSSL+4%22&type=issues) | missing-pr |
| apr-util | 0 | 11 | apache |  | missing-pr |
| qtbase | 1 | 10 | qt |  | missing-pr |
| aws-c-cal | closure | 9 | github:awslabs/aws-c-cal | [issues](https://github.com/search?q=repo%3Aawslabs%2Faws-c-cal+%22OpenSSL+4%22&type=issues) | missing-pr |
| aws-c-io | closure | 8 | github:awslabs/aws-c-io | [issues](https://github.com/search?q=repo%3Aawslabs%2Faws-c-io+%22OpenSSL+4%22&type=issues) | missing-pr |
| httpd | 1 | 8 | apache |  | missing-pr |
| folly | 1 | 7 | github:facebook/folly | [issues](https://github.com/search?q=repo%3Afacebook%2Ffolly+%22OpenSSL+4%22&type=issues) | missing-pr |

## Curated Upstream Issues

| Formula | Upstream | State | Status | Link | Note |
|---|---|---|---|---|---|
| cryptography | github:pyca/cryptography | closed | reference | [OpenSSL 4.0.0 support](https://github.com/pyca/cryptography/issues/14656) | Closed upstream support tracker for OpenSSL 4. |
| grpc | github:grpc/grpc | open | relevant | [Fix compile with OpenSSL 4.0](https://github.com/grpc/grpc/issues/42020) | Tracks upstream OpenSSL 4 compile compatibility. |
| libfido2 | github:Yubico/libfido2 | closed | reference | [Fix library build with OpenSSL 4.x](https://github.com/Yubico/libfido2/issues/966) | Closed upstream compatibility fix to inspect when validating Homebrew failures. |
| node | github:nodejs/node | closed | reference | [[openssl-4.0.0] - failing tests](https://github.com/nodejs/node/issues/62817) | Closed upstream test-failure tracker for OpenSSL 4. |
| rust | github:rust-lang/rust | open | relevant | [Build with OpenSSL-4.0.0 fails](https://github.com/rust-lang/rust/issues/155397) | Tracks upstream Rust build compatibility with OpenSSL 4. |
| s2n | github:aws/s2n-tls | open | relevant | [OpenSSL Support Roadmap](https://github.com/aws/s2n-tls/issues/5783) | Upstream roadmap for supported OpenSSL versions. |
