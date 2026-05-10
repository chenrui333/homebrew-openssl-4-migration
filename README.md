# OpenSSL 4 Homebrew Core Migration Harness

This repository tracks and automates mechanical openssl@3 to openssl@4 formula migrations in Homebrew/homebrew-core.

The workflow is split into two parts:

1. Build a dependency/status inventory from a local homebrew-core checkout.
2. Regenerate tracking, checklist, audit, and dataset artifacts for the migration.
3. Use the migration harness to update one formula, create the expected branch, commit, push to a fork, and open a PR.

## Prerequisites

- Go 1.26+
- git
- GitHub CLI (gh) authenticated for querying and creating Homebrew PRs
- A local homebrew-core checkout

Set HOMEBREW_CORE when your checkout is not at the Makefile default:

~~~sh
make status HOMEBREW_CORE=$HOME/path/to/homebrew-core
~~~

## Build the binary

~~~sh
make build
~~~

This compiles `bin/sslmigrate`. All other targets run it automatically.

## Build the dependency inventory

~~~sh
make dep-tree
~~~

This writes data/dep_tree.json. The inventory includes every formula that currently has a direct dependency on openssl@3 or openssl@4, source/upstream metadata for each formula, plus the staged depth from Homebrew tracking issue Homebrew/homebrew-core#278366. The scan also reads the full formula dependency graph so it can compute transitive staging blockers through bridge formulae that do not directly depend on OpenSSL.

Target branches are computed from the staged dependency closure:

- Depth 0 through depth 3: openssl-4-migration-staging
- No-depth formulae required by a staged formula's transitive dependency graph: openssl-4-migration-staging
- Main-track leaves that are not required by the staged track: main

## Refresh the status dashboard

~~~sh
make status
~~~

This refreshes the target Homebrew refs, checks formula files from each formula's computed target branch, queries open migration PRs by migration labels plus an openssl@4 title fallback, prints the dashboard, and regenerates [TRACKING.md](TRACKING.md).

## Refresh the checklist

~~~sh
make checklist
~~~

This regenerates [CHECKLIST.md](CHECKLIST.md) from the same dependency inventory and live PR/status data.

## Refresh the audit report

~~~sh
make audit
~~~

This regenerates [AUDIT.md](AUDIT.md) by combining data/dep_tree.json, live migration PR state, and the curated upstream issue dataset in data/upstream_issues.json. The report highlights staged-track blockers, main-track opportunities, readiness signals, upstream repository metadata, and known upstream OpenSSL 4 issues.

## Daily dataset sync

The GitHub Action in [.github/workflows/sync.yml](.github/workflows/sync.yml) runs daily and regenerates data/dep_tree.json, [TRACKING.md](TRACKING.md), [CHECKLIST.md](CHECKLIST.md), and [AUDIT.md](AUDIT.md). When those artifacts change, the workflow commits and pushes the updated datasets back to the repository.

## Migrate one formula

Dry-run first:

~~~sh
make migrate-dry FORMULA=wget
~~~

Then run the migration:

~~~sh
make migrate FORMULA=wget
~~~

The migration tool:

- locates the formula file (formulae live at Formula/<first_char>/<name>.rb)
- skips formulae already migrated to openssl@4
- swaps depends_on "openssl@3" to depends_on "openssl@4" (both quote styles)
- skips depends_on lines inside resource blocks
- bumps an existing revision or inserts revision 1
- adds OpenSSL 4 environment variables to Rust/cargo formulae
- creates branch rchen.openssl4.<formula> from the computed target branch; fails if the branch exists (pass --reset-existing to reset)
- commits <formula>: use openssl@4 with a DCO sign-off
- pushes to your fork remote and opens a PR unless --no-pr is used

Optional flags:

~~~sh
bin/sslmigrate migrate <formula> --dry-run
bin/sslmigrate migrate <formula> --no-pr
bin/sslmigrate migrate <formula> --push-remote=<remote>
bin/sslmigrate migrate <formula> --reset-existing
~~~

PR bodies stay short and reference Homebrew/homebrew-core#278366. Staging PRs also receive staging-branch-pr and CI-skip-recursive-dependents labels in addition to openssl-4-migration.
