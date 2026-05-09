BINARY        := bin/sslmigrate
HOMEBREW_CORE ?= /opt/homebrew/Library/Taps/homebrew/homebrew-core
GO            ?= go

.PHONY: help build dep-tree status migrate migrate-dry clean

help:
	@echo "Targets:"
	@echo "  build               Compile bin/openssl4"
	@echo "  dep-tree            Rebuild data/dep_tree.json from homebrew-core"
	@echo "  status              Regenerate TRACKING.md + print dashboard"
	@echo "  migrate FORMULA=X   Migrate formula X and open a PR"
	@echo "  migrate-dry FORMULA=X   Dry-run migration (no branch/PR created)"
	@echo "  clean               Remove bin/ and generated artifacts"

build:
	$(GO) build -o $(BINARY) .

dep-tree: build
	$(BINARY) dep-tree --homebrew-core=$(HOMEBREW_CORE)

status: build
	$(BINARY) status --homebrew-core=$(HOMEBREW_CORE)

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
	rm -rf bin/ data/dep_tree.json TRACKING.md
