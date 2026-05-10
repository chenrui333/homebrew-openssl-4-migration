# homebrew-openssl-4-migration

Migration harness and tracking repo for the Homebrew/homebrew-core openssl@3 → openssl@4 migration (tracking issue: Homebrew/homebrew-core#278366).

## What this repo does

1. **Tracks** which formulae have been migrated and which are still pending.
2. **Automates** the mechanical formula changes (dep swap + revision bump + Rust ENV wiring) and PR creation.
3. **Audits** staged blockers, main-track opportunities, upstream repositories, and curated upstream issue context.

The actual formulae live in a local homebrew-core checkout.
Set `HOMEBREW_CORE` when your checkout is not at the Makefile default.

## Binary

```sh
make build          # compiles bin/sslmigrate
```

Five subcommands:

```sh
bin/sslmigrate dep-tree   # scan homebrew-core → data/dep_tree.json
bin/sslmigrate status     # live status dashboard → TRACKING.md
bin/sslmigrate checklist  # markdown checkbox view → CHECKLIST.md
bin/sslmigrate audit      # readiness/upstream report → AUDIT.md
bin/sslmigrate migrate <formula> [--dry-run] [--no-pr] [--push-remote=REMOTE] [--reset-existing]
```

## Key Make targets

```sh
make dep-tree                    # rebuild data/dep_tree.json
make status                      # regenerate TRACKING.md + print dashboard
make checklist                   # regenerate CHECKLIST.md with checkboxes
make audit                       # regenerate AUDIT.md from datasets + live PR state
make migrate-dry FORMULA=wget    # preview diff, no changes to homebrew-core
make migrate FORMULA=wget        # migrate and open PR
```

## Migration rules

- **Depth 0–3 formulae** (cmake, curl, python@3.x, rust, etc.) target the `openssl-4-migration-staging` branch.
- **Staging closure formulae** have no explicit staged depth, but are required by a staged formula's transitive dependency graph, so they also target `openssl-4-migration-staging`.
- **Main-track leaves** are not required by the staged track and target `main`.
- Branch naming: `rchen.openssl4.<formula-name>`
- Commit message: `<formula>: use openssl@4`
- Labels: `openssl-4-migration` for all; also `staging-branch-pr` + `CI-skip-recursive-dependents` for staging PRs.
- Status/checklist tracking should find PRs through migration labels and through migration-style `openssl@4` titles, because label application is non-fatal.
- Rust formulae (detected via `system "cargo"` or `std_cargo_args`): inject `OPENSSL_DIR`, `OPENSSL_LIB_DIR`, `OPENSSL_INCLUDE_DIR`, `PKG_CONFIG_PATH` into `def install`.

## Source layout

```
main.go                        cobra CLI entry point
internal/
  formula/
    locate.go                  find formula .rb file in homebrew-core
    parse.go                   detect openssl dep, parse depends_on, detect Rust
    patch.go                   apply migration (swap dep, bump revision, Rust ENV)
    source.go                  parse formula homepage/source and upstream metadata
  deptree/
    types.go                   Formula + DepTree structs
    build.go                   scan formulae, compute staging closure, emit JSON
  tracking/tracking.go         shared live-status + PR query logic
  git/git.go                   thin wrappers around git CLI
  github/gh.go                 thin wrappers around gh CLI
  migrate/migrate.go           full migration flow (branch, commit, push, PR)
  status/status.go             TRACKING.md dashboard generator
  checklist/checklist.go       CHECKLIST.md checkbox generator
  audit/audit.go               AUDIT.md readiness/upstream report generator
data/dep_tree.json             committed snapshot; regenerate with make dep-tree
data/upstream_issues.json      curated upstream issue links for audit context
TRACKING.md                    generated dashboard; regenerate with make status
CHECKLIST.md                   generated checklist; regenerate with make checklist
AUDIT.md                       generated audit report; regenerate with make audit
.github/workflows/sync.yml     daily GitHub Action to auto-regenerate tracking data
```

## Depth levels (from tracking issue)

| Depth | Formulae |
|-------|----------|
| 0 (roots) | cmake, apr-util, asio, dotnet, erlang, freetds, grpc, hiredis, krb5, libevent, libfido2, librdkafka, libssh, libssh2, mariadb-connector-c, openldap, opusfile, python@3.11–3.14, srt, tcl-tk, tcl-tk@8, wget |
| 1 | apache-arrow, bind, curl, ffmpeg, folly, httpd, libpq, node, postgresql@17, postgresql@18, pulseaudio, qtbase, rust, systemd, unbound |
| 2 | cargo-c, cryptography, gdal, php, ruby |
| 3 | gstreamer |
| nil | no staged depth; target staging only when required by the computed staging closure |

Depth 0 must be fully merged to the staging branch before depth 1, and so on. Formulae in the computed staging closure should move with the staged track; main-track leaves that are not required by staged formulae can target `main` directly.

## Keeping data in sync

After pulling homebrew-core, regenerate and commit:

```sh
cd $HOMEBREW_CORE && git pull --rebase
cd $HOME/path/to/homebrew-openssl-4-migration
make dep-tree
make status
make checklist
make audit
git add data/dep_tree.json data/upstream_issues.json TRACKING.md CHECKLIST.md AUDIT.md
git commit -s -m "data: sync dep_tree to homebrew-core $(git -C $HOMEBREW_CORE rev-parse --short HEAD)"
git push origin main
```

The GitHub Action (`.github/workflows/sync.yml`) runs dep-tree, status, checklist, and audit automatically every day at 06:17 UTC, then commits and pushes changed datasets/reports.

## Prerequisites

- Go 1.26+
- `gh` CLI authenticated (`gh auth status`)
- A local homebrew-core checkout
- A fork of Homebrew/homebrew-core pushed as a remote in that checkout
