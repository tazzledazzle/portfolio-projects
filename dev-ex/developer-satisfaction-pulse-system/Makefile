PYTHON ?= python3

.PHONY: install run test lint

install:
	$(PYTHON) -m pip install -e .

run:
	uvicorn app.main:app --reload --app-dir src

test:
	pytest -q

lint:
	$(PYTHON) -m py_compile src/app/*.py src/app/api/*.py src/app/domain/*.py src/app/services/*.py
