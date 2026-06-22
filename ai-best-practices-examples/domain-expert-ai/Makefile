PYTHON ?= python3
VENV_DIR ?= .venv
PYTHON_BIN := $(VENV_DIR)/bin/python
PIP_BIN := $(VENV_DIR)/bin/pip

BUILD_DIR := .build
INSTALL_STAMP := $(BUILD_DIR)/install.stamp

RAW_DATA := data/raw/tech_expert_seed.jsonl
TRAIN_DATA := data/processed/train.jsonl
VAL_DATA := data/processed/val.jsonl
QUALITY_REPORT := reports/data_quality_report.json
EVAL_REPORT := reports/baseline_vs_tuned.json
RUN_MODEL ?= domain-expert-ai
RUN_PROMPT ?= How should I choose between quicksort and mergesort for large datasets?

PREPARE_STAMP := $(BUILD_DIR)/prepare.stamp
TRAIN_STAMP := $(BUILD_DIR)/train_dry_run.stamp
EVAL_STAMP := $(BUILD_DIR)/eval.stamp

.PHONY: help venv install prepare train-dry eval run quality-summary data-improve reprepare pipeline clean-build clean-all

help:
	@echo "Targets:"
	@echo "  make venv        - Create local virtual environment"
	@echo "  make install     - Install package + dev dependencies"
	@echo "  make prepare     - Prepare curated train/val data incrementally"
	@echo "  make train-dry   - Run dry-run training incrementally"
	@echo "  make eval        - Run evaluation incrementally"
	@echo "  make run         - Run one local inference call"
	@echo "  make quality-summary - Print data quality report summary"
	@echo "  make reprepare   - Force rebuild of curated data step"
	@echo "  make data-improve - Reprepare + print quality summary"
	@echo "  make pipeline    - Run prepare -> train-dry -> eval"
	@echo "  make clean-build - Reset incremental build stamps only"
	@echo "  make clean-all   - Remove venv, reports, checkpoints, and stamps"

venv:
	@if [ ! -d "$(VENV_DIR)" ]; then $(PYTHON) -m venv "$(VENV_DIR)"; fi

install: $(INSTALL_STAMP)

$(INSTALL_STAMP): pyproject.toml README.md | $(BUILD_DIR) venv
	$(PIP_BIN) install -e ".[dev]"
	@touch "$(INSTALL_STAMP)"

$(BUILD_DIR):
	mkdir -p "$(BUILD_DIR)"

$(PREPARE_STAMP): $(RAW_DATA) src/domain_expert_ai/data/prepare_dataset.py src/domain_expert_ai/data/curation/quality_checks.py src/domain_expert_ai/data/curation/augment_cases.py | $(BUILD_DIR) $(INSTALL_STAMP)
	PYTHONPATH=src $(PYTHON_BIN) -m domain_expert_ai.cli prepare-data \
		--input "$(RAW_DATA)" \
		--train-output "$(TRAIN_DATA)" \
		--val-output "$(VAL_DATA)" \
		--quality-report-path "$(QUALITY_REPORT)" \
		--enable-augmentation
	@touch "$(PREPARE_STAMP)"

prepare: $(PREPARE_STAMP)

$(TRAIN_STAMP): $(TRAIN_DATA) $(VAL_DATA) src/domain_expert_ai/training/train_qlora.py src/domain_expert_ai/prompting/templates.py | $(BUILD_DIR) $(INSTALL_STAMP)
	PYTHONPATH=src $(PYTHON_BIN) -m domain_expert_ai.cli train \
		--train-file "$(TRAIN_DATA)" \
		--val-file "$(VAL_DATA)" \
		--output-dir checkpoints/dry-run-quality \
		--profile tiny \
		--template-strategy mixed \
		--max-steps 10 \
		--dry-run
	@touch "$(TRAIN_STAMP)"

train-dry: $(TRAIN_STAMP)

$(EVAL_STAMP): $(VAL_DATA) src/domain_expert_ai/eval/run_eval.py src/domain_expert_ai/eval/benchmarks.py | $(BUILD_DIR) $(INSTALL_STAMP)
	PYTHONPATH=src $(PYTHON_BIN) -m domain_expert_ai.cli eval \
		--eval-file "$(VAL_DATA)" \
		--report-path "$(EVAL_REPORT)"
	@touch "$(EVAL_STAMP)"

eval: $(EVAL_STAMP)

run: install
	PYTHONPATH=src $(PYTHON_BIN) -m domain_expert_ai.cli serve \
		--model "$(RUN_MODEL)" \
		--prompt "$(RUN_PROMPT)"

quality-summary: prepare
	@$(PYTHON_BIN) -c "import json; p='$(QUALITY_REPORT)'; d=json.load(open(p)); print('quality_report:', p); print('input_rows=', d.get('input_rows')); print('accepted_records=', d.get('accepted_records')); print('rejected_records=', d.get('rejected_records')); print('rejection_reasons=', d.get('rejection_reasons', {}))"

reprepare: clean-build prepare

data-improve: reprepare quality-summary

pipeline: prepare train-dry eval

clean-build:
	rm -rf "$(BUILD_DIR)"

clean-all:
	rm -rf "$(BUILD_DIR)" "$(VENV_DIR)" reports checkpoints
