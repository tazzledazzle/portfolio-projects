# Domain Expert AI Model Quality Improvements Design

## Goal

Improve legal-answer quality (correctness, citation quality, nuance handling) by strengthening dataset curation and prompt templating before further training optimization.

## Approved Strategy

Follow this sequence:
1. Data + prompt quality loop (primary)
2. Training quality optimization
3. Inference-time quality boosting only if still needed

## Architecture

### Quality-First Pipeline

- `data/raw`: source legal FAQ examples (including new hard cases).
- `data/curation`:
  - `augment_cases.py`: generate controlled variants.
  - `quality_checks.py`: quality gates on citations, duplicates, contradictions.
- `data/processed`: curated records that pass gates + split artifacts.
- `prompting/templates`: multiple instruction styles for stronger generalization.
- `training`: consumes curated/template-mixed records.
- `eval`: reports quality by difficulty slice and legal nuance categories.

## Approaches Considered

### A) Data + Prompt Quality Loop (selected first)
- Highest quality improvement per unit effort.
- Improves baseline and fine-tuned paths simultaneously.
- Reduces noise before expensive hyperparameter runs.

### B) Training Optimization (selected second)
- Hyperparameter/profile sweeps after data quality stabilizes.
- Better checkpoint selection by validation metrics.
- Improves convergence and stability.

### C) Inference-Time Boosting (selected third)
- Retrieval-backed grounding and stricter constrained generation.
- Useful for production trust, but adds infrastructure complexity.

## Data Flow

1. Add/collect new domain records (especially hard cases).
2. Generate controlled augmentation variants.
3. Run quality gates (citation validity, duplication, contradiction checks).
4. Produce curated train/val with quality report.
5. Train/evaluate using curated/template-mixed dataset.
6. Compare quality metrics against previous run and baseline.

## Quality Gates

- Citation presence and formatting quality.
- Duplicate and near-duplicate detection.
- Contradiction heuristics (same fact pattern, conflicting answer labels).
- Coverage checks across jurisdictions/risk levels/topic slices.

## Prompt Template Strategy

Use three templates:
- Direct FAQ template.
- Scenario analysis template.
- Ambiguity-aware template (when facts are insufficient or conflicting).

Template sampling is balanced during training data formatting to reduce overfitting to one style.

## Evaluation Plan

Add quality-focused slices:
- Easy / medium / hard.
- Federal-only vs state-specific.
- High-risk categories (termination, retaliation, wage-hour disputes).
- Citation precision and support alignment.

Primary metrics:
- Keyword coverage score.
- Citation coverage/precision.
- Format compliance.
- Ambiguity-safe behavior rate.

## Error Handling

- Any failed quality gate is logged with row-level reason.
- Training artifacts include full resolved config and data lineage references.
- Evaluation outputs include per-slice metrics to localize regressions.

## Out of Scope (for this phase)

- Full production policy engine.
- Multi-provider serving orchestration.
- Automated legal citation retrieval from external paid APIs.

