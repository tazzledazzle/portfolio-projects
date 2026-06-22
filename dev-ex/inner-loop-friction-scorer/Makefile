PYTHON ?= python3

.PHONY: install run test lint

install:
	$(PYTHON) -m pip install -e .

run:
	$(PYTHON) -m friction_scorer.main

test:
	pytest -q

lint:
	$(PYTHON) -m py_compile src/friction_scorer/*.py src/friction_scorer/core/*.py src/friction_scorer/adapters/*.py
