# blog skills — src を正とし、必要なら .cursor/skills へ展開する
.PHONY: help list install

CURSOR_SKILLS := ../.cursor/skills

help:
	@echo "Usage: make -C skills [target]"
	@echo "  list    - list skill directories under src/"
	@echo "  install - copy each src/<name>/ to $(CURSOR_SKILLS)/blog-<name>/"

list:
	@ls -d src/*/ 2>/dev/null || echo "(no skills under src/)"

install:
	@test -d src || (echo "src/ missing" && exit 1)
	@mkdir -p $(CURSOR_SKILLS)
	@for d in src/*/; do \
		[ -f "$$d/SKILL.md" ] || continue; \
		name=$$(basename $$d); \
		dest=$(CURSOR_SKILLS)/$$name; \
		echo "install $$d -> $$dest"; \
		rm -rf $$dest; \
		mkdir -p $$dest; \
		cp -R "$$d"/* "$$dest"/; \
	done
	@echo "Done."
