.PHONY: test demo demo-local up down
ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/../../..)
test:
	python3 $(ROOT)/scripts/check_evidence_kit.py .
demo: test
	@echo "Evidence kit demo: $(notdir $(CURDIR))"
	@wc -w BEST_PRACTICES.md INTERVIEW_STORY.md PROOF_CHECKLIST.md
demo-local: demo
up:
	@echo "No runtime services for evidence kits"
down:
	@echo "Nothing to tear down"
