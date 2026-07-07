# Security Hotspot Annotator

## Overview
This project converts Semgrep SAST findings into rich pull request annotations with severity, CWE mapping, and remediation playbook links.

## Architecture
- `src/input`: Semgrep JSON ingestion and validation.
- `src/enrich`: severity normalization, CWE extraction, and taxonomy mapping.
- `src/format`: PR comment payload formatting.
- `src/publish`: VCS adapter for posting review comments.
- `config`: severity policy and remediation URL catalog.

## Use Cases
- Turn static analysis into developer-actionable review comments.
- Standardize vulnerability communication across teams.
- Improve remediation speed by linking directly to playbooks.

## Usage
1. Run Semgrep and export JSON results.
2. Run annotator with repo and PR metadata.
3. Publish line-level comments to pull request.

## Control Flow
1. Parse Semgrep findings.
2. Enrich findings with CWE + severity taxonomy.
3. Build plain-English risk explanation strings.
4. Format inline PR annotations.
5. Post comments and publish summary status.

## Project Structure
```text
security-hotspot-annotator/
  .github/workflows/ci.yml
  config/severity-policy.yml
  scripts/annotate-pr.sh
  src/input/semgrep_parser.py
  src/enrich/cwe_mapper.py
  src/format/comment_formatter.py
  src/publish/pr_comment_client.py
  tests/test_comment_formatter.py
```
