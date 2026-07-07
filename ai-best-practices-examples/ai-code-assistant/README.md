# AI Code Assistant

MVP CLI for generating Python unit tests.

## MVP Scope

- In scope: unit-test generation for Python files (single file and repo-scan modes)
- Out of scope: PR review, patch suggestions, CI automation

## Setup

```bash
cd ai-code-assistant
python3 -m venv .venv
source .venv/bin/activate
pip install -e ".[dev]"
```

## Commands

Generate tests for one file:

```bash
PYTHONPATH=src ai-code-assistant gen-tests path/to/file.py
```

Generate a full test pyramid for one file:

```bash
PYTHONPATH=src ai-code-assistant gen-tests path/to/file.py --pyramid all
```

Generate tests for a repository:

```bash
PYTHONPATH=src ai-code-assistant gen-tests --repo .
```

Dry run (print generated output, do not write files):

```bash
PYTHONPATH=src ai-code-assistant gen-tests --repo . --dry-run
```

JSON mode for CI/headless workflows:

```bash
PYTHONPATH=src ai-code-assistant gen-tests --repo . --dry-run --output json
```

Audit logging (JSONL):

```bash
PYTHONPATH=src ai-code-assistant gen-tests --repo . --dry-run --audit-log .ai-code-assistant/audit.log.jsonl
```

Policy-driven run with headless CI output:

```bash
PYTHONPATH=src ai-code-assistant gen-tests --repo . --headless --policy-file assistant-policy.toml
```

Validate extension manifest:

```bash
PYTHONPATH=src ai-code-assistant extensions validate-manifest --manifest extension-manifest.v1.json --output json
```

Run checkpointed plan:

```bash
PYTHONPATH=src ai-code-assistant run-plan --plan plan.json --output json
```

## Optional OpenAI Integration

If `OPENAI_API_KEY` is set and the `openai` package is installed, the adapter will
attempt model-based generation. Otherwise, it uses a deterministic fallback template.
