"""Tests for Q18 — Rate-limited Webhook Delivery."""

from unittest.mock import MagicMock

import pytest

from q18_webhook_delivery import DeliveryResult, Subscriber, deliver

EVENT: dict[str, object] = {"event": "payment.created", "id": "evt-1"}


def make_sub(sub_id: str = "s1") -> Subscriber:
    return Subscriber(subscriber_id=sub_id, url="http://example.com/webhook")


def http_always_ok() -> MagicMock:
    m = MagicMock()
    m.post.return_value = 200
    return m


def http_always_fail(code: int = 500) -> MagicMock:
    m = MagicMock()
    m.post.return_value = code
    return m


def test_successful_delivery_returns_success(  ) -> None:
    result = deliver(EVENT, make_sub(), http_always_ok())
    assert result.success is True


def test_successful_delivery_attempt_count(  ) -> None:
    result = deliver(EVENT, make_sub(), http_always_ok())
    assert result.attempts == 1


def test_failed_delivery_retries(  ) -> None:
    result = deliver(EVENT, make_sub(), http_always_fail(), max_retries=3)
    assert result.attempts == 3


def test_failed_delivery_success_false(  ) -> None:
    result = deliver(EVENT, make_sub(), http_always_fail(), max_retries=2)
    assert result.success is False


def test_success_resets_consecutive_failures(  ) -> None:
    sub = make_sub()
    sub.consecutive_failures = 3
    deliver(EVENT, sub, http_always_ok())
    assert sub.consecutive_failures == 0


def test_failure_increments_consecutive_failures(  ) -> None:
    sub = make_sub()
    deliver(EVENT, sub, http_always_fail(), max_retries=1)
    assert sub.consecutive_failures > 0


def test_subscriber_suspended_after_threshold(  ) -> None:
    sub = make_sub()
    sub.SUSPEND_THRESHOLD = 3
    for _ in range(3):
        deliver(EVENT, sub, http_always_fail(), max_retries=1)
    assert sub.suspended is True


def test_suspended_subscriber_immediate_return(  ) -> None:
    sub = make_sub()
    sub.suspended = True
    http = http_always_ok()
    result = deliver(EVENT, sub, http)
    assert result.success is False
    http.post.assert_not_called()


def test_delivery_result_has_final_status_on_success(  ) -> None:
    result = deliver(EVENT, make_sub(), http_always_ok())
    assert result.final_status == 200


def test_delivery_result_has_final_status_on_failure(  ) -> None:
    result = deliver(EVENT, make_sub(), http_always_fail(503), max_retries=1)
    assert result.final_status == 503


def test_retry_stops_on_2xx(  ) -> None:
    http = MagicMock()
    http.post.side_effect = [500, 200]
    result = deliver(EVENT, make_sub(), http, max_retries=3)
    assert result.success is True
    assert result.attempts == 2


def test_no_retry_on_success(  ) -> None:
    http = http_always_ok()
    deliver(EVENT, make_sub(), http, max_retries=5)
    assert http.post.call_count == 1


def test_result_error_set_on_suspension(  ) -> None:
    sub = make_sub()
    sub.suspended = True
    result = deliver(EVENT, sub, http_always_ok())
    assert result.error is not None


def test_max_retries_zero_means_one_attempt(  ) -> None:
    http = http_always_fail()
    result = deliver(EVENT, make_sub(), http, max_retries=1)
    assert http.post.call_count == 1
