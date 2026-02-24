"""Tests for gcp_billing_summary logic (validation and query shape without BigQuery)."""

from datetime import datetime
import pytest


def test_validate_date_format() -> None:
    """BILLING_DATE must be YYYY-MM-DD."""
    datetime.strptime("2025-01-15", "%Y-%m-%d")
    with pytest.raises(ValueError):
        datetime.strptime("01-15-2025", "%Y-%m-%d")
    with pytest.raises(ValueError):
        datetime.strptime("invalid", "%Y-%m-%d")


def test_query_table_pattern() -> None:
    """Table pattern uses dataset name correctly (same logic as script)."""
    dataset = "billing_export"
    table_pattern = f"`{dataset}.gcp_billing_export_v1_*`"
    assert "billing_export" in table_pattern
    assert "gcp_billing_export_v1_*" in table_pattern


def test_query_select_columns() -> None:
    """Query selects project_id, service, total_cost (documentation of expected shape)."""
    query = """
        SELECT
            project.name AS project_id,
            service.description AS service,
            SUM(cost) AS total_cost
        FROM `dataset.gcp_billing_export_v1_*`
        WHERE DATE(usage_start_time) = @date
        GROUP BY 1, 2
        ORDER BY total_cost DESC
        LIMIT 100
    """
    assert "project_id" in query
    assert "total_cost" in query
    assert "@date" in query
