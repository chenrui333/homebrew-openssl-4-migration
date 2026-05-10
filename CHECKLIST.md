# OpenSSL 4 Migration Checklist (2026-05-10)

Progress: **39/57 (68.4%)**
Tracking issue: [Homebrew/homebrew-core#278366](https://github.com/Homebrew/homebrew-core/issues/278366)

> Staging batches are depth 0 → 1 → 2 → 3 plus their computed transitive closure.
> This public checklist is scoped to formulae targeting openssl-4-migration-staging.

## Batch 0 — Roots -> openssl-4-migration-staging [20/25]

- [x] ~~apr-util~~
- [x] ~~asio~~
- [x] ~~cmake~~
- [ ] dotnet <!-- PR #280829 open -->
- [x] ~~erlang~~
- [x] ~~freetds~~
- [ ] grpc <!-- PR #280832 open -->
- [x] ~~hiredis~~
- [x] ~~krb5~~
- [x] ~~libevent~~
- [ ] libfido2 <!-- PR #280836 open -->
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

- [ ] apache-arrow <!-- PR #280851 open -->
- [x] ~~bind~~
- [x] ~~curl~~
- [x] ~~ffmpeg~~
- [x] ~~folly~~
- [x] ~~httpd~~
- [x] ~~libpq~~
- [ ] node <!-- PR #280858 open -->
- [x] ~~postgresql@17~~
- [x] ~~postgresql@18~~
- [ ] pulseaudio <!-- PR #280861 open -->
- [ ] qtbase <!-- PR #280862 open -->
- [ ] rust <!-- PR #280863 open -->
- [ ] systemd <!-- PR #280864 open -->
- [x] ~~unbound~~

## Batch 2 -> openssl-4-migration-staging [1/5]

- [ ] cargo-c <!-- PR #280867 open -->
- [ ] cryptography <!-- PR #280868 open -->
- [ ] gdal <!-- PR #280870 open -->
- [ ] php <!-- PR #280871 open -->
- [x] ~~ruby~~

## Batch 3 -> openssl-4-migration-staging [0/1]

- [ ] gstreamer <!-- PR #280873 open -->

## Staging closure -> openssl-4-migration-staging [9/11]

- [x] ~~aws-c-cal~~
- [x] ~~cgal~~
- [x] ~~libgit2~~
- [x] ~~libngtcp2~~
- [x] ~~libshout~~
- [x] ~~libzip~~
- [x] ~~net-snmp~~
- [x] ~~rtmpdump~~
- [ ] s2n <!-- PR #280912 open -->
- [x] ~~srtp~~
- [ ] thrift <!-- PR #280913 open -->
