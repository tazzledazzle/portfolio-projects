# Domain Expert AI Design

## Goal

Build a legal FAQ domain expert for US employment topics with two tracks:
prompt-engineered baseline and QLoRA adapter fine-tuning.

## Architecture

- Data layer validates JSONL records and builds deterministic train/val splits.
- Prompting layer creates structured baseline prompts and parses structured outputs.
- Training layer stores reproducible QLoRA configuration metadata for local runs.
- Guardrails enforce minimum output quality (citations, disclaimer, confidence range).
- Evaluation compares baseline vs tuned outputs with domain-specific metrics.
- Inference adapter runs local Ollama model for one-shot demo responses.

## Data Schema

- `question`: user-facing legal question.
- `context`: scenario detail such as state or worker role.
- `answer`: gold target response.
- `jurisdiction`: state/federal jurisdiction tag.
- `risk_level`: low/medium/high legal risk category.
- `citations`: legal authority list.

## Guardrails

- Reject responses without disclaimer.
- Reject responses without at least one citation.
- Enforce confidence in `[0, 1]`.

## Evaluation

- `keyword_score`: expected concept coverage in response text.
- `citation_score`: expected legal authority presence.
- `format_ok`: disclaimer + list-shaped citations present.

