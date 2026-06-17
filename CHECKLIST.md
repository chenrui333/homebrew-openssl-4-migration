# OpenSSL 4 Migration Checklist (2026-06-17)

Progress: **22/58 (37.9%)**
Tracking issue: [Homebrew/homebrew-core#278366](https://github.com/Homebrew/homebrew-core/issues/278366)

> Staging batches are depth 0 → 1 → 2 → 3 plus their computed transitive closure.
> This public checklist is scoped to formulae targeting openssl-4-migration-staging.

## Batch 0 — Roots -> openssl-4-migration-staging [14/25]

- [x] ~~apr-util~~
- [x] ~~asio~~
- [ ] cmake
- [ ] dotnet
- [ ] erlang
- [ ] freetds <!-- PR #280846 open -->
- [ ] grpc
- [ ] hiredis <!-- PR #280846 open -->
- [x] ~~krb5~~
- [x] ~~libevent~~
- [ ] libfido2
- [ ] librdkafka <!-- PR #280846 open -->
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
- [ ] tcl-tk@8
- [x] ~~wget~~

## Batch 1 -> openssl-4-migration-staging [3/15]

- [ ] apache-arrow
- [ ] bind
- [x] ~~curl~~
- [x] ~~ffmpeg~~
- [ ] folly
- [x] ~~httpd~~
- [ ] libpq
- [ ] node
- [ ] postgresql@17
- [ ] postgresql@18
- [ ] pulseaudio
- [ ] qtbase
- [ ] rust
- [ ] systemd <!-- PR #280864 open -->
- [ ] unbound

## Batch 2 -> openssl-4-migration-staging [0/5]

- [ ] cargo-c
- [ ] cryptography
- [ ] gdal
- [ ] php
- [ ] ruby

## Batch 3 -> openssl-4-migration-staging [0/1]

- [ ] gstreamer

## Staging closure -> openssl-4-migration-staging [5/12]

- [ ] aws-c-cal
- [ ] aws-c-io
- [ ] cgal
- [ ] libgit2
- [ ] libngtcp2
- [x] ~~libshout~~
- [x] ~~libzip~~
- [x] ~~net-snmp~~
- [x] ~~rtmpdump~~
- [ ] s2n
- [x] ~~srtp~~
- [ ] thrift
