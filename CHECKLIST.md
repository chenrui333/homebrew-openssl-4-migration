# OpenSSL 4 Migration Checklist (2026-06-04)

Progress: **40/57 (70.2%)**
Tracking issue: [Homebrew/homebrew-core#278366](https://github.com/Homebrew/homebrew-core/issues/278366)

> Staging batches are depth 0 → 1 → 2 → 3 plus their computed transitive closure.
> This public checklist is scoped to formulae targeting openssl-4-migration-staging.

## Batch 0 — Roots -> openssl-4-migration-staging [20/25]

- [x] ~~apr-util~~
- [x] ~~asio~~
- [x] ~~cmake~~
- [ ] dotnet
- [x] ~~erlang~~
- [x] ~~freetds~~
- [ ] grpc
- [x] ~~hiredis~~
- [x] ~~krb5~~
- [x] ~~libevent~~
- [ ] libfido2
- [x] ~~librdkafka~~
- [x] ~~libssh~~
- [x] ~~libssh2~~
- [x] ~~mariadb-connector-c~~
- [x] ~~openldap~~
- [x] ~~opusfile~~
- [x] ~~python@3.11~~
- [x] ~~python@3.12~~
- [ ] python@3.13 <!-- PR #280845 open -->
- [ ] python@3.14 <!-- PR #280846 open -->
- [x] ~~srt~~
- [x] ~~tcl-tk~~
- [x] ~~tcl-tk@8~~
- [x] ~~wget~~

## Batch 1 -> openssl-4-migration-staging [9/15]

- [ ] apache-arrow
- [x] ~~bind~~
- [x] ~~curl~~
- [x] ~~ffmpeg~~
- [x] ~~folly~~
- [x] ~~httpd~~
- [x] ~~libpq~~
- [ ] node
- [x] ~~postgresql@17~~
- [x] ~~postgresql@18~~
- [ ] pulseaudio
- [ ] qtbase
- [ ] rust
- [ ] systemd <!-- PR #280864 open -->
- [x] ~~unbound~~

## Batch 2 -> openssl-4-migration-staging [2/5]

- [ ] cargo-c <!-- PR #280867 open -->
- [ ] cryptography
- [ ] gdal
- [x] ~~php~~
- [x] ~~ruby~~

## Batch 3 -> openssl-4-migration-staging [0/1]

- [ ] gstreamer

## Staging closure -> openssl-4-migration-staging [9/12]

- [x] ~~aws-c-cal~~
- [ ] aws-c-io
- [x] ~~cgal~~
- [x] ~~libgit2~~
- [x] ~~libngtcp2~~
- [x] ~~libshout~~
- [x] ~~libzip~~
- [x] ~~net-snmp~~
- [x] ~~rtmpdump~~
- [ ] s2n
- [x] ~~srtp~~
- [ ] thrift
