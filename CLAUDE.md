# homebrew-openssl-4-migration

Migration harness and tracking repo for the Homebrew/homebrew-core openssl@3 → openssl@4 migration (tracking issue: Homebrew/homebrew-core#278366).

## What this repo does

1. **Tracks** which formulae have been migrated and which are still pending.
2. **Automates** the mechanical formula changes (dep swap + revision bump + Rust ENV wiring) and PR creation.

The actual formulae live in homebrew-core at `/opt/homebrew/Library/Taps/homebrew/homebrew-core`.
This repo holds scripts/tooling that operate on that checkout.

## Binary

```sh
make build          # compiles bin/sslmigrate
```

Three subcommands:

```sh
bin/sslmigrate dep-tree   # scan homebrew-core → data/dep_tree.json
bin/sslmigrate status     # live status dashboard → TRACKING.md
bin/sslmigrate migrate <formula> [--dry-run] [--no-pr] [--push-remote=REMOTE]
```

## Key Make targets

```sh
make dep-tree                    # rebuild data/dep_tree.json
make status                      # regenerate TRACKING.md + print dashboard
make migrate-dry FORMULA=wget    # preview diff, no changes to homebrew-core
make migrate FORMULA=wget        # migrate and open PR
```

## Migration rules

- **Depth 0–3 formulae** (cmake, curl, python@3.x, rust, etc.) target the `openssl-4-migration-staging` branch.
- **Leaf formulae** (no depth in dep_tree.json) target `main`.
- Branch naming: `rchen.openssl4.<formula-name>`
- Commit message: `<formula>: use openssl@4`
- Labels: `openssl-4-migration` for all; also `staging-branch-pr` + `CI-skip-recursive-dependents` for staging PRs.
- Rust formulae (detected via `system "cargo"` or `std_cargo_args`): inject `OPENSSL_DIR`, `OPENSSL_LIB_DIR`, `OPENSSL_INCLUDE_DIR`, `PKG_CONFIG_PATH` into `def install`.

## Source layout

```
main.go                        cobra CLI entry point
internal/
  formula/
    locate.go                  find formula .rb file in homebrew-core
    parse.go                   detect openssl dep, parse depends_on, detect Rust
    patch.go                   apply migration (swap dep, bump revision, Rust ENV)
  deptree/
    types.go                   Formula + DepTree structs
    build.go                   scan formulae, assign depths, emit JSON
  git/git.go                   thin wrappers around git CLI
  github/gh.go                 thin wrappers around gh CLI
  migrate/migrate.go           full migration flow (branch, commit, push, PR)
  status/status.go             dashboard generator
data/dep_tree.json             committed snapshot; regenerate with make dep-tree
TRACKING.md                    generated dashboard; regenerate with make status
```

## Depth levels (from tracking issue)

| Depth | Formulae |
|-------|----------|
| 0 (roots) | cmake, apr-util, asio, dotnet, erlang, freetds, grpc, hiredis, krb5, libevent, libfido2, librdkafka, libssh, libssh2, mariadb-connector-c, openldap, opusfile, python@3.11–3.14, srt, tcl-tk, tcl-tk@8, wget |
| 1 | apache-arrow, bind, curl, ffmpeg, folly, httpd, libpq, node, postgresql@17, postgresql@18, pulseaudio, qtbase, rust, systemd, unbound |
| 2 | cargo-c, cryptography, gdal, php, ruby |
| 3 | gstreamer |
| nil (leaves) | ~677 formulae with no staged depth |

Depth 0 must be fully merged to the staging branch before depth 1, and so on.

## Keeping data in sync

After pulling homebrew-core, regenerate and commit:

```sh
cd /opt/homebrew/Library/Taps/homebrew/homebrew-core && git pull --rebase
cd /path/to/this/repo
make dep-tree
make status
git add data/dep_tree.json TRACKING.md
git commit -m "data: sync dep_tree to homebrew-core $(git -C /opt/homebrew/Library/Taps/homebrew/homebrew-core rev-parse --short HEAD)"
git push origin main
```

## Prerequisites

- Go 1.21+
- `gh` CLI authenticated (`gh auth status`)
- A local homebrew-core checkout (default: `/opt/homebrew/Library/Taps/homebrew/homebrew-core`)
- A fork of Homebrew/homebrew-core pushed as a remote in that checkout
