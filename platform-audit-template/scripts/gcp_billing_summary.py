#!/usr/bin/env python3
"""
Read-only GCP billing summary by project and service for a single day.

Uses BigQuery billing export. Requires:
- BigQuery billing export configured for the billing account.
- Credentials with read access to the billing dataset (e.g. gcloud auth
  application-default login, or GOOGLE_APPLICATION_CREDENTIALS).
- Environment variables:
  - GCP_BILLING_DATASET: dataset name (e.g. billing_export)
  - BILLING_DATE: YYYY-MM-DD

Usage:
  export GCP_BILLING_DATASET=billing_export
  export BILLING_DATE=2025-01-15
  python3 gcp_billing_summary.py
"""

import os
import sys
from datetime import datetime

try:
    from google.cloud import bigquery
except ImportError:
    print("Install: pip install google-cloud-bigquery", file=sys.stderr)
    sys.exit(1)


def main() -> None:
    dataset = os.environ.get("GCP_BILLING_DATASET")
    date_str = os.environ.get("BILLING_DATE")
    if not dataset or not date_str:
        print(
            "Set GCP_BILLING_DATASET and BILLING_DATE (YYYY-MM-DD).",
            file=sys.stderr,
        )
        sys.exit(1)
    try:
        datetime.strptime(date_str, "%Y-%m-%d")
    except ValueError:
        print("BILLING_DATE must be YYYY-MM-DD.", file=sys.stderr)
        sys.exit(1)

    # Table pattern for exported billing data
    table_pattern = f"`{dataset}.gcp_billing_export_v1_*`"
    query = f"""
        SELECT
            project.name AS project_id,
            service.description AS service,
            SUM(cost) AS total_cost
        FROM {table_pattern}
        WHERE DATE(usage_start_time) = @date
        GROUP BY 1, 2
        ORDER BY total_cost DESC
        LIMIT 100
    """
    client = bigquery.Client()
    job_config = bigquery.QueryJobConfig(
        query_parameters=[
            bigquery.ScalarQueryParameter("date", "DATE", date_str),
        ]
    )
    try:
        rows = client.query(query, job_config=job_config).result()
    except Exception as e:
        print(f"Query failed: {e}", file=sys.stderr)
        sys.exit(1)

    print(f"Billing summary for {date_str}\n")
    print(f"{'Project':<40} {'Service':<50} {'Cost':>12}")
    print("-" * 104)
    for row in rows:
        print(f"{row.project_id or 'N/A':<40} {(row.service or 'N/A')[:50]:<50} {float(row.total_cost):>12.2f}")


if __name__ == "__main__":
    main()
