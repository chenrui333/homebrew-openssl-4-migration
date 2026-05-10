BINARY        := bin/sslmigrate
HOMEBREW_CORE ?= /opt/homebrew/Library/Taps/homebrew/homebrew-core
GO            ?= go

.PHONY: help build dep-tree status checklist audit site migrate migrate-dry clean

help:
	@echo "Targets:"
	@echo "  build               Compile bin/sslmigrate"
	@echo "  dep-tree            Rebuild data/dep_tree.json from homebrew-core"
	@echo "  status              Regenerate TRACKING.md + print dashboard"
	@echo "  checklist           Regenerate CHECKLIST.md with markdown checkboxes"
	@echo "  audit               Regenerate AUDIT.md with readiness and upstream context"
	@echo "  site                Regenerate MkDocs pages and run mkdocs build --strict"
	@echo "  migrate FORMULA=X   Migrate formula X and open a PR"
	@echo "  migrate-dry FORMULA=X   Dry-run migration (no branch/PR created)"
	@echo "  clean               Remove bin/"

build:
	$(GO) build -o $(BINARY) .

dep-tree: build
	$(BINARY) dep-tree --homebrew-core=$(HOMEBREW_CORE)

status: build
	$(BINARY) status --homebrew-core=$(HOMEBREW_CORE)

checklist: build
	$(BINARY) checklist --homebrew-core=$(HOMEBREW_CORE)

audit: build
	$(BINARY) audit --homebrew-core=$(HOMEBREW_CORE)

site: build
	$(BINARY) site --homebrew-core=$(HOMEBREW_CORE)
	mkdocs build --strict

migrate: build
ifndef FORMULA
	$(error FORMULA is required. Usage: make migrate FORMULA=wget)
endif
	$(BINARY) migrate $(FORMULA) --homebrew-core=$(HOMEBREW_CORE)

migrate-dry: build
ifndef FORMULA
	$(error FORMULA is required. Usage: make migrate-dry FORMULA=wget)
endif
	$(BINARY) migrate $(FORMULA) --homebrew-core=$(HOMEBREW_CORE) --dry-run

clean:
	rm -rf bin/
