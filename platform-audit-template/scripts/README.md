# Diagnostic Scripts Pattern

This folder holds **read-only diagnostic** scripts used for platform and cost audits. They are intended to be safe to run in sandbox or audit contexts: no resource creation, modification, or deletion.

## Principles

- **Read-only:** Only list, describe, or query; never mutate.
- **No credentials in repo:** Require env vars or CLI login (e.g. `gcloud auth application-default login`). Document required permissions in this README or in the script docstring.
- **Scoped:** Prefer date ranges and project filters to limit output and cost.

## How to adapt

1. Copy a script and change the project ID, dataset, or date range via env vars or arguments.
2. Ensure your identity has only the minimal read-only permissions needed (e.g. `billing.resourceCosts.list`, `logging.logEntries.list` for GCP).
3. Run in a sandbox or audit project first; never point at production billing or logs without review.

## Scripts

| Script | Purpose | Requirements |
|--------|---------|--------------|
| `gcp_billing_summary.py` | Summarize GCP billing by project/service for a given day (uses BigQuery billing export). | BigQuery billing export configured; `GOOGLE_APPLICATION_CREDENTIALS` or `gcloud auth`; read access to billing dataset. |

## Example usage (GCP billing summary)

```bash
export GCP_BILLING_DATASET=my_billing_dataset   # e.g. billing_export
export BILLING_DATE=2025-01-15
python3 gcp_billing_summary.py
```

Output is printed to stdout (project, service, cost). Redirect to a file or pipe to your audit doc as needed.
