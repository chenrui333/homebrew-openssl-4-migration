# OpenSSL 4 Migration Audit (2026-10-08)

Tracking issue: Homebrew/homebrew-core#278366

## Summary

- Staging-scope formulae: 55
- Live pending: 52
- Live done: 3 (5.5%)
- Open staging PRs: 0
- Draft migration PRs: 0
- PRs with merge/check blockers: 0
- Pending formulae without open migration PRs: 52

## Retarget to Staging

Open staging-scope migration PRs whose base branch is not openssl-4-migration-staging.

| Formula | PR | Current Base | Expected Base | Target | Readiness |
|---|---|---|---|---|---|
| _none_ |  |  |  |  |  |

## Staging Priority

Pending staged formulae are sorted by transitive dependent count.

| Formula | Target | Depth | Impact | Status | PR | Readiness | Upstream | Issues |
|---|---|---:|---:|---|---|---|---|---|
| python@3.14 | openssl-4-migration-staging | 0 | 441 | PENDING | none | missing-pr | python |  |
| libssh2 | openssl-4-migration-staging | 0 | 289 | PENDING | none | missing-pr | github:libssh2/libssh2 |  |
| libgit2 | openssl-4-migration-staging | closure | 270 | PENDING | none | missing-pr | github:libgit2/libgit2 |  |
| rust | openssl-4-migration-staging | 1 | 268 | PENDING | none | missing-pr | github:rust-lang/rust | [issues#155397](https://github.com/rust-lang/rust/issues/155397) open |
| krb5 | openssl-4-migration-staging | 0 | 75 | PENDING | none | missing-pr | other |  |
| systemd | openssl-4-migration-staging | 1 | 65 | PENDING | none | missing-pr | github:systemd/systemd |  |
| libevent | openssl-4-migration-staging | 0 | 59 | PENDING | none | missing-pr | github:libevent/libevent |  |
| python@3.13 | openssl-4-migration-staging | 0 | 39 | PENDING | none | missing-pr | python |  |
| libngtcp2 | openssl-4-migration-staging | closure | 21 | PENDING | none | missing-pr | github:ngtcp2/ngtcp2 |  |
| pulseaudio | openssl-4-migration-staging | 1 | 20 | PENDING | none | missing-pr | gitlab:gitlab.freedesktop.org/pulseaudio/pulseaudio |  |
| ruby | openssl-4-migration-staging | 2 | 20 | PENDING | none | missing-pr | github:ruby/ruby |  |
| cargo-c | openssl-4-migration-staging | 2 | 18 | PENDING | none | missing-pr | github:lu-zero/cargo-c |  |
| curl | openssl-4-migration-staging | 1 | 17 | PENDING | none | missing-pr | github:curl/curl |  |
| pipewire | openssl-4-migration-staging | closure | 16 | PENDING | none | missing-pr | gitlab:gitlab.freedesktop.org/pipewire/pipewire |  |
| libpq | openssl-4-migration-staging | 1 | 14 | PENDING | none | missing-pr | other |  |
| ffmpeg | openssl-4-migration-staging | 1 | 12 | PENDING | none | missing-pr | github:FFmpeg/FFmpeg |  |
| libfido2 | openssl-4-migration-staging | 0 | 12 | PENDING | none | missing-pr | github:Yubico/libfido2 | [issues#966](https://github.com/Yubico/libfido2/issues/966) closed |
| grpc | openssl-4-migration-staging | 0 | 11 | PENDING | none | missing-pr | github:grpc/grpc | [issues#42020](https://github.com/grpc/grpc/issues/42020) open |
| openldap | openssl-4-migration-staging | 0 | 11 | PENDING | none | missing-pr | other |  |
| apr-util | openssl-4-migration-staging | 0 | 10 | PENDING | none | missing-pr | apache |  |
| libzip | openssl-4-migration-staging | closure | 9 | PENDING | none | missing-pr | other |  |
| qtbase | openssl-4-migration-staging | 1 | 9 | PENDING | none | missing-pr | qt |  |
| cryptography | openssl-4-migration-staging | 2 | 8 | PENDING | none | missing-pr | github:pyca/cryptography | [issues#14656](https://github.com/pyca/cryptography/issues/14656) closed |
| folly | openssl-4-migration-staging | 1 | 7 | PENDING | none | missing-pr | github:facebook/folly |  |
| freetds | openssl-4-migration-staging | 0 | 7 | PENDING | none | missing-pr | github:FreeTDS/freetds |  |
| httpd | openssl-4-migration-staging | 1 | 7 | PENDING | none | missing-pr | apache |  |
| node | openssl-4-migration-staging | 1 | 7 | PENDING | none | missing-pr | github:nodejs/node | [issues#62817](https://github.com/nodejs/node/issues/62817) closed |
| libssh | openssl-4-migration-staging | 0 | 6 | PENDING | none | missing-pr | other |  |
| rtmpdump | openssl-4-migration-staging | closure | 6 | PENDING | none | missing-pr | other |  |
| srt | openssl-4-migration-staging | 0 | 6 | PENDING | none | missing-pr | github:Haivision/srt |  |
| aws-c-cal | openssl-4-migration-staging | closure | 5 | PENDING | none | missing-pr | github:awslabs/aws-c-cal |  |
| hiredis | openssl-4-migration-staging | 0 | 5 | PENDING | none | missing-pr | github:redis/hiredis |  |
| libshout | openssl-4-migration-staging | closure | 5 | PENDING | none | missing-pr | other |  |
| net-snmp | openssl-4-migration-staging | closure | 5 | PENDING | none | missing-pr | github:net-snmp/net-snmp |  |
| s2n | openssl-4-migration-staging | closure | 5 | PENDING | none | missing-pr | github:aws/s2n-tls | [issues#5783](https://github.com/aws/s2n-tls/issues/5783) open |
| srtp | openssl-4-migration-staging | closure | 5 | PENDING | none | missing-pr | github:cisco/libsrtp |  |
| thrift | openssl-4-migration-staging | closure | 5 | PENDING | none | missing-pr | github:apache/thrift |  |
| apache-arrow | openssl-4-migration-staging | 1 | 4 | PENDING | none | missing-pr | github:apache/arrow |  |
| gstreamer | openssl-4-migration-staging | 3 | 4 | PENDING | none | missing-pr | gitlab:gitlab.freedesktop.org/gstreamer/gstreamer |  |
| bind | openssl-4-migration-staging | 1 | 3 | PENDING | none | missing-pr | gitlab:gitlab.isc.org/isc-projects/bind9 |  |
| mariadb-connector-c | openssl-4-migration-staging | 0 | 3 | PENDING | none | missing-pr | github:mariadb-corporation/mariadb-connector-c |  |
| python@3.12 | openssl-4-migration-staging | 0 | 3 | PENDING | none | missing-pr | python |  |
| tcl-tk@8 | openssl-4-migration-staging | 0 | 3 | PENDING | none | missing-pr | other |  |
| unbound | openssl-4-migration-staging | 1 | 3 | PENDING | none | missing-pr | github:NLnetLabs/unbound |  |
| gdal | openssl-4-migration-staging | 2 | 2 | PENDING | none | missing-pr | github:OSGeo/gdal |  |
| dotnet | openssl-4-migration-staging | 0 | 1 | PENDING | none | missing-pr | github:dotnet/dotnet |  |
| librdkafka | openssl-4-migration-staging | 0 | 1 | PENDING | none | missing-pr | github:confluentinc/librdkafka |  |
| postgresql@17 | openssl-4-migration-staging | 1 | 1 | PENDING | none | missing-pr | other |  |
| postgresql@18 | openssl-4-migration-staging | 1 | 1 | PENDING | none | missing-pr | other |  |
| opusfile | openssl-4-migration-staging | 0 | 0 | PENDING | none | missing-pr | gitlab:gitlab.xiph.org/xiph/opusfile |  |
| php | openssl-4-migration-staging | 2 | 0 | PENDING | none | missing-pr | github:php/php-src |  |
| python@3.11 | openssl-4-migration-staging | 0 | 0 | PENDING | none | missing-pr | python |  |

## Upstream Issue Coverage Gaps

Top 20 pending staged formulae with upstream metadata and no curated upstream issue entry.

| Formula | Depth | Impact | Upstream | Search | Readiness |
|---|---:|---:|---|---|---|
| python@3.14 | 0 | 441 | python |  | missing-pr |
| libssh2 | 0 | 289 | github:libssh2/libssh2 | [issues](https://github.com/search?q=repo%3Alibssh2%2Flibssh2+%22OpenSSL+4%22&type=issues) | missing-pr |
| libgit2 | closure | 270 | github:libgit2/libgit2 | [issues](https://github.com/search?q=repo%3Alibgit2%2Flibgit2+%22OpenSSL+4%22&type=issues) | missing-pr |
| systemd | 1 | 65 | github:systemd/systemd | [issues](https://github.com/search?q=repo%3Asystemd%2Fsystemd+%22OpenSSL+4%22&type=issues) | missing-pr |
| libevent | 0 | 59 | github:libevent/libevent | [issues](https://github.com/search?q=repo%3Alibevent%2Flibevent+%22OpenSSL+4%22&type=issues) | missing-pr |
| python@3.13 | 0 | 39 | python |  | missing-pr |
| libngtcp2 | closure | 21 | github:ngtcp2/ngtcp2 | [issues](https://github.com/search?q=repo%3Angtcp2%2Fngtcp2+%22OpenSSL+4%22&type=issues) | missing-pr |
| pulseaudio | 1 | 20 | gitlab:gitlab.freedesktop.org/pulseaudio/pulseaudio |  | missing-pr |
| ruby | 2 | 20 | github:ruby/ruby | [issues](https://github.com/search?q=repo%3Aruby%2Fruby+%22OpenSSL+4%22&type=issues) | missing-pr |
| cargo-c | 2 | 18 | github:lu-zero/cargo-c | [issues](https://github.com/search?q=repo%3Alu-zero%2Fcargo-c+%22OpenSSL+4%22&type=issues) | missing-pr |
| curl | 1 | 17 | github:curl/curl | [issues](https://github.com/search?q=repo%3Acurl%2Fcurl+%22OpenSSL+4%22&type=issues) | missing-pr |
| pipewire | closure | 16 | gitlab:gitlab.freedesktop.org/pipewire/pipewire |  | missing-pr |
| ffmpeg | 1 | 12 | github:FFmpeg/FFmpeg | [issues](https://github.com/search?q=repo%3AFFmpeg%2FFFmpeg+%22OpenSSL+4%22&type=issues) | missing-pr |
| apr-util | 0 | 10 | apache |  | missing-pr |
| qtbase | 1 | 9 | qt |  | missing-pr |
| folly | 1 | 7 | github:facebook/folly | [issues](https://github.com/search?q=repo%3Afacebook%2Ffolly+%22OpenSSL+4%22&type=issues) | missing-pr |
| freetds | 0 | 7 | github:FreeTDS/freetds | [issues](https://github.com/search?q=repo%3AFreeTDS%2Ffreetds+%22OpenSSL+4%22&type=issues) | missing-pr |
| httpd | 1 | 7 | apache |  | missing-pr |
| srt | 0 | 6 | github:Haivision/srt | [issues](https://github.com/search?q=repo%3AHaivision%2Fsrt+%22OpenSSL+4%22&type=issues) | missing-pr |
| aws-c-cal | closure | 5 | github:awslabs/aws-c-cal | [issues](https://github.com/search?q=repo%3Aawslabs%2Faws-c-cal+%22OpenSSL+4%22&type=issues) | missing-pr |

## Curated Upstream Issues

| Formula | Upstream | State | Status | Link | Note |
|---|---|---|---|---|---|
| cryptography | github:pyca/cryptography | closed | reference | [OpenSSL 4.0.0 support](https://github.com/pyca/cryptography/issues/14656) | Closed upstream support tracker for OpenSSL 4. |
| grpc | github:grpc/grpc | open | relevant | [Fix compile with OpenSSL 4.0](https://github.com/grpc/grpc/issues/42020) | Tracks upstream OpenSSL 4 compile compatibility. |
| libfido2 | github:Yubico/libfido2 | closed | reference | [Fix library build with OpenSSL 4.x](https://github.com/Yubico/libfido2/issues/966) | Closed upstream compatibility fix to inspect when validating Homebrew failures. |
| node | github:nodejs/node | closed | reference | [[openssl-4.0.0] - failing tests](https://github.com/nodejs/node/issues/62817) | Closed upstream test-failure tracker for OpenSSL 4. |
| rust | github:rust-lang/rust | open | relevant | [Build with OpenSSL-4.0.0 fails](https://github.com/rust-lang/rust/issues/155397) | Tracks upstream Rust build compatibility with OpenSSL 4. |
| s2n | github:aws/s2n-tls | open | relevant | [OpenSSL Support Roadmap](https://github.com/aws/s2n-tls/issues/5783) | Upstream roadmap for supported OpenSSL versions. |
