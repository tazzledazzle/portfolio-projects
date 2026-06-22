# Domain Expert AI

Prompt-first baseline and QLoRA fine-tuning scaffold for a technical expert assistant focused on computer science, software engineering, design patterns, data structures & algorithms, and machine learning & AI.

## Setup

```bash
cd domain-expert-ai
python3 -m venv .venv
source .venv/bin/activate
pip install -e ".[dev]"
```

Optional environment defaults:

```bash
set -a
source .env.example
set +a
```

CLI flags always override environment defaults.

## Run Instructions

Quick start (recommended):

```bash
cd domain-expert-ai
make install
make pipeline
```

Run individual steps:

```bash
make prepare
make train-dry
make eval
make run
```

Inspect supported build targets:

```bash
make help
```

Data quality iteration helpers:

```bash
make quality-summary      # prints accepted/rejected counts + reasons
make data-improve         # forces reprepare then prints quality summary
```

Customize one-shot run target:

```bash
make run RUN_MODEL=domain-expert-ai RUN_PROMPT="How should I choose between quicksort and mergesort for large datasets?"
```

## Incremental Build Process

This project now uses an incremental `Makefile` build:

- `make prepare` reruns only when raw data or curation/prepare code changes.
- `make train-dry` reruns only when processed data or training/template code changes.
- `make eval` reruns only when eval data or eval code changes.
- `make run` executes one inference call without changing incremental build stamps.
- `make quality-summary` reads and prints the latest quality report in `reports/`.
- `make data-improve` forces a fresh prepare pass and prints quality diagnostics.
- `make pipeline` chains all three steps in order.

Incremental state is tracked with stamp files in `.build/`.

Reset incremental state only:

```bash
make clean-build
```

Clean everything (including virtualenv/reports/checkpoints):

```bash
make clean-all
```

## Commands

### 1) Prepare data

```bash
PYTHONPATH=src domain-expert-ai prepare-data \
  --input data/raw/tech_expert_seed.jsonl \
  --train-output data/processed/train.jsonl \
  --val-output data/processed/val.jsonl \
  --quality-report-path reports/data_quality_report.json \
  --enable-augmentation
```

### 2) Train (QLoRA with Hugging Face Trainer)

```bash
PYTHONPATH=src domain-expert-ai train \
  --train-file data/processed/train.jsonl \
  --val-file data/processed/val.jsonl \
  --output-dir checkpoints/run-001
```

Profiles:
- `tiny`: fastest smoke profile for low-memory debugging
- `balanced`: default day-to-day profile
- `quality`: higher-capacity profile for stronger results

Example with explicit full configuration:

```bash
PYTHONPATH=src domain-expert-ai train \
  --train-file data/processed/train.jsonl \
  --val-file data/processed/val.jsonl \
  --output-dir checkpoints/run-configurable \
  --profile tiny \
  --max-steps 20 \
  --epochs 1 \
  --batch-size 1 \
  --grad-accum-steps 2 \
  --learning-rate 3e-4 \
  --lora-r 4 \
  --lora-alpha 8 \
  --lora-dropout 0.05 \
  --max-length 192 \
  --warmup-ratio 0.03 \
  --weight-decay 0.0 \
  --logging-steps 5 \
  --eval-strategy epoch \
  --save-strategy epoch \
  --seed 42 \
  --use-4bit auto \
  --target-modules q_proj,v_proj \
  --template-strategy mixed
```

Quick validation without model download/training:

```bash
PYTHONPATH=src domain-expert-ai train \
  --train-file data/processed/train.jsonl \
  --val-file data/processed/val.jsonl \
  --output-dir checkpoints/dry-run \
  --profile tiny \
  --dry-run
```

### 3) Evaluate baseline vs tuned

```bash
PYTHONPATH=src domain-expert-ai eval \
  --eval-file data/processed/val.jsonl \
  --report-path reports/baseline_vs_tuned.json
```

The report now includes global metrics and `slice_metrics` grouped by:
- `difficulty`
- `jurisdiction`
- `risk_category` (with compatibility fallback to `risk_level`)

### 4) Serve one response via Ollama

```bash
PYTHONPATH=src domain-expert-ai serve \
  --model domain-expert-ai \
  --prompt "How do I reduce p99 latency in a microservice?"
```

## What is implemented

- JSONL schema validation for technical Q&A records.
- Deterministic data split utility.
- Data curation quality gates and optional augmentation with quality report output.
- Baseline prompt builder and structured response parser.
- QLoRA training loop using Hugging Face `Trainer` + PEFT LoRA adapters.
- Multi-template training formatter (`direct`, `scenario`, `ambiguity`, `mixed`) with seeded determinism.
- Guardrail validators for disclaimer, citations, and confidence.
- Benchmark scoring and report generation.
- Slice-aware evaluation reporting for quality analysis.
- Minimal Ollama adapter for local demo inference.

## Notes

- Use this as a starter: swap in a production training loop in `training/train_qlora.py`.
- Keep disclaimers in every generated technical response.

