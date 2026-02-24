# Runbook: Cost / Billing Diagnostic (read-only)

**Purpose:** Safely run read-only checks to understand where cloud costs are coming from (e.g. by project, service, or resource type). Use for sandbox or billing alerts—never modify resources or expose credentials.

**When to use:** After a cost spike, for periodic cost reviews, or when setting up budget alerts.

**Prerequisites:** Read-only billing access (e.g. `billing.accounts.get`, `billing.resourceCosts.list` or equivalent) and authenticated CLI or SDK.

---

## Principles

- **Read-only only:** No resource creation, deletion, or modification.
- **No credentials in repo:** Use environment variables or secret manager; document required permissions in the runbook, not keys.
- **Scoped queries:** Prefer date range and project filters to avoid large exports.

---

## GCP example (BigQuery billing export)

If your org exports GCP billing to BigQuery:

1. **Identify dataset and table** (e.g. `billing_dataset.gcp_billing_export_v1_*`).
2. **Run a summarized query** (replace dataset and date as needed):

```sql
SELECT
  project.name,
  service.description,
  sku.description,
  SUM(cost) AS total_cost
FROM `billing_dataset.gcp_billing_export_v1_*`
WHERE DATE(usage_start_time) = 'YYYY-MM-DD'
  AND project.name = 'your-project-id'
GROUP BY 1, 2, 3
ORDER BY total_cost DESC
LIMIT 50;
```

3. **Interpret:** High cost from a single `service.description` or `sku.description` suggests where to optimize or add budgets.

---

## GCP example (gcloud logging volume)

To diagnose **log volume** (often a major cost driver):

```bash
# Set project and time range
PROJECT="your-project-id"
START_TIME="2025-01-01T00:00:00Z"
END_TIME="2025-01-02T00:00:00Z"

# Log volume by resource type
gcloud logging read "timestamp>=\"$START_TIME\" timestamp<\"$END_TIME\"" \
  --project="$PROJECT" \
  --format=json \
  --limit=10000 \
  | jq -r '.[] | .resource.type' | sort | uniq -c | sort -rn
```

Use results to target log exclusion or sampling.

---

## Budget alerts (recommended)

- Create budget alerts at 80%, 100%, and 120% of expected spend.
- Document the budget ID and alert channels in your runbook so the team can adjust thresholds without code changes.

---

## Escalation

- Share query results (with PII/sensitive project names redacted if needed).
- Include date range and any filters used so others can reproduce.
