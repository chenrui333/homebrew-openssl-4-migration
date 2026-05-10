# OpenSSL 4 Migration Audit (2026-05-10)

Tracking issue: Homebrew/homebrew-core#278366

## Summary

- Formulae tracked: 759
- Live pending: 679
- Live done: 80 (10.5%)
- Open migration PRs: 25
- Draft migration PRs: 25
- PRs with merge/check blockers: 25
- Pending formulae without open migration PRs: 653

## Staging Priority

Pending staged formulae are sorted by transitive dependent count.

| Formula | Target | Depth | Impact | Status | PR | Readiness | Upstream | Issues |
|---|---|---:|---:|---|---|---|---|---|
| python@3.14 | openssl-4-migration-staging | 0 | 481 | PENDING | #280846 | draft, checks-blocked, merge-unstable | python |  |
| rust | openssl-4-migration-staging | 1 | 279 | PENDING | #280863 | draft, checks-blocked, merge-unstable | github:rust-lang/rust | [issues#155397](https://github.com/rust-lang/rust/issues/155397) open |
| python@3.13 | openssl-4-migration-staging | 0 | 80 | PENDING | #280845 | draft, checks-blocked, merge-unstable | python |  |
| systemd | openssl-4-migration-staging | 1 | 72 | PENDING | #280864 | draft, checks-blocked, merge-unstable | github:systemd/systemd |  |
| cargo-c | openssl-4-migration-staging | 2 | 25 | PENDING | #280867 | draft, checks-blocked, merge-unstable | github:lu-zero/cargo-c |  |
| pulseaudio | openssl-4-migration-staging | 1 | 21 | PENDING | #280861 | draft, checks-blocked, merge-unstable | gitlab:gitlab.freedesktop.org/pulseaudio/pulseaudio |  |
| cryptography | openssl-4-migration-staging | 2 | 15 | PENDING | #280868 | draft, checks-blocked, merge-unstable | github:pyca/cryptography | [issues#14656](https://github.com/pyca/cryptography/issues/14656) closed |
| grpc | openssl-4-migration-staging | 0 | 11 | PENDING | #280832 | draft, checks-blocked, merge-unstable | github:grpc/grpc | [issues#42020](https://github.com/grpc/grpc/issues/42020) open |
| libfido2 | openssl-4-migration-staging | 0 | 10 | PENDING | #280836 | draft, checks-blocked, merge-unstable | github:Yubico/libfido2 | [issues#966](https://github.com/Yubico/libfido2/issues/966) closed |
| qtbase | openssl-4-migration-staging | 1 | 10 | PENDING | #280862 | draft, checks-blocked, merge-unstable | qt |  |
| node | openssl-4-migration-staging | 1 | 8 | PENDING | #280858 | draft, checks-blocked, merge-unstable | github:nodejs/node | [issues#62817](https://github.com/nodejs/node/issues/62817) closed |
| s2n | openssl-4-migration-staging | closure | 7 | PENDING | #280912 | draft, checks-blocked, merge-unstable | github:aws/s2n-tls | [issues#5783](https://github.com/aws/s2n-tls/issues/5783) open |
| thrift | openssl-4-migration-staging | closure | 5 | PENDING | #280913 | draft, checks-blocked, merge-unstable | apache |  |
| apache-arrow | openssl-4-migration-staging | 1 | 4 | PENDING | #280851 | draft, checks-blocked, merge-unstable | github:apache/arrow |  |
| gstreamer | openssl-4-migration-staging | 3 | 4 | PENDING | #280873 | draft, checks-blocked, merge-unstable | gitlab:gitlab.freedesktop.org/gstreamer/gstreamer |  |
| dotnet | openssl-4-migration-staging | 0 | 2 | PENDING | #280829 | draft, checks-blocked, merge-unstable | github:dotnet/dotnet |  |
| gdal | openssl-4-migration-staging | 2 | 2 | PENDING | #280870 | draft, checks-blocked, merge-unstable | github:OSGeo/gdal |  |
| php | openssl-4-migration-staging | 2 | 0 | PENDING | #280871 | draft, checks-blocked, merge-unstable | php |  |

## Main-Track Opportunities

Top 30 pending main-track formulae sorted by transitive dependent count.

| Formula | Target | Depth | Impact | Status | PR | Readiness | Upstream | Issues |
|---|---|---:|---:|---|---|---|---|---|
| crystal | main | - | 5 | PENDING | none | missing-pr | github:crystal-lang/crystal |  |
| fizz | main | - | 5 | PENDING | none | missing-pr | github:facebookincubator/fizz |  |
| ldns | main | - | 4 | PENDING | #281770 | draft, base-mismatch, checks-blocked, merge-unstable | other |  |
| pkcs11-helper | main | - | 4 | PENDING | none | missing-pr | github:OpenSC/pkcs11-helper |  |
| cpprestsdk | main | - | 3 | PENDING | none | missing-pr | github:microsoft/cpprestsdk |  |
| libwebsockets | main | - | 3 | PENDING | none | missing-pr | github:warmcat/libwebsockets |  |
| libxmlsec1 | main | - | 3 | PENDING | none | missing-pr | github:lsh123/xmlsec |  |
| mvfst | main | - | 3 | PENDING | none | missing-pr | github:facebook/mvfst |  |
| nim | main | - | 3 | PENDING | none | missing-pr | github:nim-lang/Nim |  |
| wangle | main | - | 3 | PENDING | none | missing-pr | github:facebook/wangle |  |
| xml-security-c | main | - | 3 | PENDING | none | missing-pr | apache |  |
| capnp | main | - | 2 | PENDING | none | missing-pr | github:capnproto/capnproto |  |
| davix | main | - | 2 | PENDING | none | missing-pr | github:cern-fts/davix |  |
| libks | main | - | 2 | PENDING | none | missing-pr | github:signalwire/libks |  |
| libretls | main | - | 2 | PENDING | none | missing-pr | other |  |
| mongo-c-driver | main | - | 2 | PENDING | none | missing-pr | github:mongodb/mongo-c-driver |  |
| pypy | main | - | 2 | PENDING | none | missing-pr | github:pypy/pypy |  |
| riemann-client | main | - | 2 | PENDING | none | missing-pr | other |  |
| samba | main | - | 2 | PENDING | none | missing-pr | other |  |
| w3m | main | - | 2 | PENDING | none | missing-pr | other |  |
| xml-tooling-c | main | - | 2 | PENDING | none | missing-pr | other |  |
| afflib | main | - | 1 | PENDING | none | missing-pr | github:sshock/AFFLIBv3 |  |
| azure-core-cpp | main | - | 1 | PENDING | #281235 | draft, base-mismatch, checks-blocked, merge-unstable | github:Azure/azure-sdk-for-cpp |  |
| cyrus-sasl | main | - | 1 | PENDING | none | missing-pr | github:cyrusimap/cyrus-sasl |  |
| erlang@26 | main | - | 1 | PENDING | none | missing-pr | github:erlang/otp |  |
| etcd-cpp-apiv3 | main | - | 1 | PENDING | none | missing-pr | github:etcd-cpp-apiv3/etcd-cpp-apiv3 |  |
| fbthrift | main | - | 1 | PENDING | none | missing-pr | github:facebook/fbthrift |  |
| getdns | main | - | 1 | PENDING | none | missing-pr | github:getdnsapi/getdns |  |
| gwenhywfar | main | - | 1 | PENDING | none | missing-pr | other |  |
| libewf | main | - | 1 | PENDING | none | missing-pr | github:libyal/libewf-legacy |  |

## Curated Upstream Issues

| Formula | Upstream | State | Status | Link | Note |
|---|---|---|---|---|---|
| cryptography | github:pyca/cryptography | closed | reference | [OpenSSL 4.0.0 support](https://github.com/pyca/cryptography/issues/14656) | Closed upstream support tracker for OpenSSL 4. |
| grpc | github:grpc/grpc | open | relevant | [Fix compile with OpenSSL 4.0](https://github.com/grpc/grpc/issues/42020) | Tracks upstream OpenSSL 4 compile compatibility. |
| libfido2 | github:Yubico/libfido2 | closed | reference | [Fix library build with OpenSSL 4.x](https://github.com/Yubico/libfido2/issues/966) | Closed upstream compatibility fix to inspect when validating Homebrew failures. |
| node | github:nodejs/node | closed | reference | [[openssl-4.0.0] - failing tests](https://github.com/nodejs/node/issues/62817) | Closed upstream test-failure tracker for OpenSSL 4. |
| rust | github:rust-lang/rust | open | relevant | [Build with OpenSSL-4.0.0 fails](https://github.com/rust-lang/rust/issues/155397) | Tracks upstream Rust build compatibility with OpenSSL 4. |
| s2n | github:aws/s2n-tls | open | relevant | [OpenSSL Support Roadmap](https://github.com/aws/s2n-tls/issues/5783) | Upstream roadmap for supported OpenSSL versions. |
