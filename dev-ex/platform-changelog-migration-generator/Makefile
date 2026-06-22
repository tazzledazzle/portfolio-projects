PYTHON ?= python3

.PHONY: install run test lint

install:
	$(PYTHON) -m pip install -e .

run:
	$(PYTHON) -m changelog_generator.cli --old api_v1.json --new api_v2.json

test:
	pytest -q

lint:
	$(PYTHON) -m py_compile src/changelog_generator/*.py src/changelog_generator/transforms/*.py
