# Makefile for repo
bootstrap:
	python3 -m pip install -r requirements.txt
	just install-hooks

lint:
	ruff .
	ktlintCheck

test:
	pytest
	./gradlew test

docs:
	mkdocs build