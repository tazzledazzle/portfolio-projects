.PHONY: test demo up down
test:
	@test -f BEST_PRACTICES.md && test -f PROOF_CHECKLIST.md && test -f INTERVIEW_STORY.md
	@grep -q "Four habits" BEST_PRACTICES.md || grep -q "**Four habits**" BEST_PRACTICES.md
	@echo "evidence kit ok"
demo: test
	@echo "Evidence kit demo: open BEST_PRACTICES.md and PROOF_CHECKLIST.md"
up:
	@echo "No runtime services for evidence kits"
down:
	@echo "Nothing to tear down"
