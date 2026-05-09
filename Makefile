HOMEBREW_CORE ?= /opt/homebrew/Library/Taps/homebrew/homebrew-core
RUBY ?= ruby

.PHONY: dep-tree status migrate migrate-dry

dep-tree:
	$(RUBY) scripts/build_dep_tree.rb --homebrew-core=$(HOMEBREW_CORE)

status:
	$(RUBY) scripts/status.rb --homebrew-core=$(HOMEBREW_CORE)

migrate:
	$(RUBY) scripts/migrate.rb $(FORMULA) --homebrew-core=$(HOMEBREW_CORE)

migrate-dry:
	$(RUBY) scripts/migrate.rb $(FORMULA) --homebrew-core=$(HOMEBREW_CORE) --dry-run
