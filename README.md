# OpenSSL 4 Homebrew Core Migration Harness

This repository tracks and automates mechanical openssl@3 to openssl@4 formula migrations in Homebrew/homebrew-core.

The workflow is split into two parts:

1. Build a dependency/status inventory from a local homebrew-core checkout.
2. Use the migration harness to update one formula, create the expected branch, commit, push to a fork, and open a PR.

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

This writes data/dep_tree.json. The inventory includes every formula that currently has a direct dependency on openssl@3 or openssl@4, plus the staged depth from Homebrew tracking issue Homebrew/homebrew-core#278366.

Staged formulae use the openssl-4-migration-staging base branch:

- Depth 0 through depth 3: staging branch
- Leaves with no staged depth: main

## Refresh the status dashboard

~~~sh
make status
~~~

This checks the live formula files in the homebrew-core checkout, queries open PRs mentioning openssl@4, prints the dashboard, and regenerates [TRACKING.md](TRACKING.md).

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
- creates branch rchen.openssl4.<formula> from the correct base; fails if the branch exists (pass --reset-existing to reset)
- commits <formula>: use openssl@4 with a DCO sign-off
- pushes to your fork remote and opens a PR unless --no-pr is used

Optional flags:

~~~sh
bin/sslmigrate migrate <formula> --dry-run
bin/sslmigrate migrate <formula> --no-pr
bin/sslmigrate migrate <formula> --push-remote=<remote>
~~~

PR bodies stay short and reference Homebrew/homebrew-core#278366. Staging PRs also receive staging-branch-pr and CI-skip-recursive-dependents labels in addition to openssl-4-migration.
