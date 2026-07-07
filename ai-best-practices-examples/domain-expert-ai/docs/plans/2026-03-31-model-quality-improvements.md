# Model Quality Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Improve legal-answer quality by adding data curation quality gates, template diversification, and quality-sliced evaluation.

**Architecture:** Introduce a new curation stage between raw and processed data, add reusable prompt templates, and extend evaluation with slice-aware metrics. Training then consumes cleaner, more diverse instruction-formatted data. This keeps quality gains reproducible and measurable.

**Tech Stack:** Python, Pydantic, pytest, Hugging Face datasets/transformers, PEFT LoRA.

---

### Task 1: Add data curation module and quality gates

**Files:**
- Create: `domain-expert-ai/src/domain_expert_ai/data/curation/__init__.py`
- Create: `domain-expert-ai/src/domain_expert_ai/data/curation/quality_checks.py`
- Create: `domain-expert-ai/src/domain_expert_ai/data/curation/augment_cases.py`
- Test: `domain-expert-ai/tests/test_data_curation.py`

**Step 1: Write the failing test**

```python
def test_quality_checks_reject_missing_citations():
    ...
```

**Step 2: Run test to verify it fails**

Run: `PYTHONPATH=src python3 -m pytest tests/test_data_curation.py -q`  
Expected: FAIL because curation module/functions do not exist.

**Step 3: Write minimal implementation**

- Implement citation format and presence checks.
- Implement duplicate/near-duplicate detection helper.
- Implement simple augmentation function for state/fact variants.

**Step 4: Run test to verify it passes**

Run: `PYTHONPATH=src python3 -m pytest tests/test_data_curation.py -q`  
Expected: PASS.

**Step 5: Commit**

```bash
git add domain-expert-ai/src/domain_expert_ai/data/curation domain-expert-ai/tests/test_data_curation.py
git commit -m "feat: add dataset curation quality gates and augmentation helpers"
```

### Task 2: Integrate curation into data preparation pipeline

**Files:**
- Modify: `domain-expert-ai/src/domain_expert_ai/data/prepare_dataset.py`
- Modify: `domain-expert-ai/src/domain_expert_ai/cli.py`
- Test: `domain-expert-ai/tests/test_dataset_schema.py`

**Step 1: Write the failing test**

```python
def test_prepare_dataset_filters_failed_quality_rows():
    ...
```

**Step 2: Run test to verify it fails**

Run: `PYTHONPATH=src python3 -m pytest tests/test_dataset_schema.py -q`  
Expected: FAIL because quality filtering/reporting is not wired in.

**Step 3: Write minimal implementation**

- Pipe rows through quality checks before splitting.
- Emit quality report (counts + rejection reasons) to output directory.
- Add CLI options for quality-report path and optional augmentation toggle.

**Step 4: Run test to verify it passes**

Run: `PYTHONPATH=src python3 -m pytest tests/test_dataset_schema.py -q`  
Expected: PASS.

**Step 5: Commit**

```bash
git add domain-expert-ai/src/domain_expert_ai/data/prepare_dataset.py domain-expert-ai/src/domain_expert_ai/cli.py domain-expert-ai/tests/test_dataset_schema.py
git commit -m "feat: integrate curation quality gates into data preparation"
```

### Task 3: Add prompt template library and formatter selection

**Files:**
- Create: `domain-expert-ai/src/domain_expert_ai/prompting/templates.py`
- Modify: `domain-expert-ai/src/domain_expert_ai/training/train_qlora.py`
- Test: `domain-expert-ai/tests/test_training_pipeline.py`

**Step 1: Write the failing test**

```python
def test_template_selection_supports_direct_scenario_ambiguity():
    ...
```

**Step 2: Run test to verify it fails**

Run: `PYTHONPATH=src python3 -m pytest tests/test_training_pipeline.py -q`  
Expected: FAIL due to missing template selection.

**Step 3: Write minimal implementation**

- Add direct/scenario/ambiguity templates.
- Add configurable template strategy in training formatter.
- Ensure deterministic selection with seed control.

**Step 4: Run test to verify it passes**

Run: `PYTHONPATH=src python3 -m pytest tests/test_training_pipeline.py -q`  
Expected: PASS.

**Step 5: Commit**

```bash
git add domain-expert-ai/src/domain_expert_ai/prompting/templates.py domain-expert-ai/src/domain_expert_ai/training/train_qlora.py domain-expert-ai/tests/test_training_pipeline.py
git commit -m "feat: add multi-template instruction formatting for training"
```

### Task 4: Extend evaluation with quality slices

**Files:**
- Modify: `domain-expert-ai/src/domain_expert_ai/eval/benchmarks.py`
- Modify: `domain-expert-ai/src/domain_expert_ai/eval/run_eval.py`
- Test: `domain-expert-ai/tests/test_eval_pipeline.py`

**Step 1: Write the failing test**

```python
def test_run_eval_reports_slice_metrics():
    ...
```

**Step 2: Run test to verify it fails**

Run: `PYTHONPATH=src python3 -m pytest tests/test_eval_pipeline.py -q`  
Expected: FAIL because slice metrics are not present.

**Step 3: Write minimal implementation**

- Compute per-slice metrics (difficulty/jurisdiction/risk category).
- Include slice table in report JSON.
- Keep global metrics backward-compatible.

**Step 4: Run test to verify it passes**

Run: `PYTHONPATH=src python3 -m pytest tests/test_eval_pipeline.py -q`  
Expected: PASS.

**Step 5: Commit**

```bash
git add domain-expert-ai/src/domain_expert_ai/eval/benchmarks.py domain-expert-ai/src/domain_expert_ai/eval/run_eval.py domain-expert-ai/tests/test_eval_pipeline.py
git commit -m "feat: add quality-sliced evaluation reporting"
```

### Task 5: Documentation and end-to-end verification

**Files:**
- Modify: `domain-expert-ai/README.md`
- Modify: `domain-expert-ai/.env.example`
- Create/Modify: `domain-expert-ai/docs/plans/2026-03-31-model-quality-improvements-design.md`

**Step 1: Write/update failing expectation tests (if docs tests exist)**

If no docs tests, skip and proceed to verification commands.

**Step 2: Run full verification**

Run:
- `PYTHONPATH=src python3 -m pytest tests -q`
- `PYTHONPATH=src python3 -m domain_expert_ai.cli prepare-data ...`
- `PYTHONPATH=src python3 -m domain_expert_ai.cli train ... --dry-run`
- `PYTHONPATH=src python3 -m domain_expert_ai.cli eval ...`

Expected: all tests pass; pipeline commands succeed; reports generated.

**Step 3: Commit**

```bash
git add domain-expert-ai/README.md domain-expert-ai/.env.example domain-expert-ai/docs/plans/2026-03-31-model-quality-improvements-design.md
git commit -m "docs: document model quality workflow and verification path"
```

