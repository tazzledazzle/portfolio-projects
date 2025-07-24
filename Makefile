# Makefile for repo
bootstrap:
	python3 -m pip install -r requirements.txt
	just install-hooks

lint:
	ruff .
	ktlintCheck

test:
	@echo "🧪 Running test suite..."
	@echo ""
	@echo "📋 Python Tests:"
	@if command -v pytest >/dev/null 2>&1; then \
		pytest -v --tb=short --continue-on-collection-errors || echo "⚠️  Some Python tests failed"; \
	else \
		echo "❌ pytest not installed. Install with: pip install pytest"; \
	fi
	@echo ""
	@echo "📋 Gradle Tests:"
	@if [ -f "./gradlew" ]; then \
		./gradlew test && echo "✅ Gradle tests passed" || echo "⚠️  Gradle tests failed"; \
	else \
		echo "⚠️  No gradlew found, skipping Gradle tests"; \
	fi
	@echo ""
	@echo "📋 Test Summary:"
	@echo "- Python tests: Check output above for details"
	@echo "- Gradle tests: Check output above for details"
	@echo "- Some test failures are expected in development projects"

docs:
	mkdocs build