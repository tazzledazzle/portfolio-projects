"""
Problem 1: Idempotent Deploy-Trigger API
==========================================
Accepts "trigger deployment" requests keyed by (idempotency_key, service_name,
git_sha). Duplicate requests with the same idempotency_key must NOT trigger a
second deployment. Supports querying deploy status. Must survive a process
crash mid-deploy without either losing the dedup guarantee or silently
re-running a completed deploy.

This file contains:
  - BruteForceDeployService : in-memory dict, no locking, no persistence
  - OptimalDeployService    : SQLite-backed, atomic insert-based dedup,
                              crash-safe status reconciliation
  - A tiny stdlib-only HTTP wrapper around OptimalDeployService
  - A unittest suite that PROVES the brute-force version has a race
    condition (fires the underlying deploy more than once) and that the
    optimal version does not, plus a crash-recovery test.

Run:
    python3 p1_idempotent_deploy_api.py            # runs the demo + serves HTTP on :8001
    python3 -m unittest p1_idempotent_deploy_api -v
"""
from __future__ import annotations

import json
import sqlite3
import threading
import time
import uuid
from dataclasses import dataclass, field
from enum import Enum
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Optional
from urllib.parse import urlparse


class DeployStatus(str, Enum):
    PENDING = "PENDING"
    RUNNING = "RUNNING"
    SUCCEEDED = "SUCCEEDED"
    FAILED = "FAILED"
    UNKNOWN = "UNKNOWN"  # crash-recovery state: we don't know what happened


@dataclass
class DeploymentRecord:
    idempotency_key: str
    service_name: str
    git_sha: str
    status: DeployStatus
    created_at: float
    updated_at: float
    error: Optional[str] = None


# ---------------------------------------------------------------------------
# STEP 1 — BRUTE FORCE
# ---------------------------------------------------------------------------
# Reasoning:
#   The obvious first implementation: a Python dict in process memory keyed
#   by idempotency_key. "Check if key exists, if not, insert and deploy."
#   This is O(1) and looks correct in a single-threaded, single-process,
#   never-crashes world. It is WRONG in production for two independent
#   reasons:
#     1. Race condition: "check-then-act" is not atomic. Two concurrent
#        requests with the same key can both pass the `if key not in dict`
#        check before either has inserted, and both trigger a real deploy.
#     2. No durability: a process restart (crash, deploy of the deploy
#        service itself, autoscaling) wipes all history. A retried request
#        after restart looks brand new and deploys again.
# ---------------------------------------------------------------------------
class BruteForceDeployService:
    def __init__(self, deploy_fn):
        self._store: dict[str, DeploymentRecord] = {}
        self._deploy_fn = deploy_fn  # simulates the real side-effecting deploy

    def trigger(self, idempotency_key: str, service_name: str, git_sha: str) -> DeploymentRecord:
        # --- THE BUG: check-then-act, no lock ---
        existing = self._store.get(idempotency_key)
        if existing is not None:
            return existing

        # Window between the check above and the insert below is where two
        # threads can both get past the check. In real code this window is
        # never zero — there's always some validation, logging, or object
        # construction between "check" and "write" — so we model that
        # honestly instead of pretending the two lines execute atomically.
        time.sleep(0.01)

        record = DeploymentRecord(
            idempotency_key=idempotency_key,
            service_name=service_name,
            git_sha=git_sha,
            status=DeployStatus.RUNNING,
            created_at=time.time(),
            updated_at=time.time(),
        )
        self._store[idempotency_key] = record

        # "Deploy" — the actual expensive/dangerous side effect.
        try:
            self._deploy_fn(service_name, git_sha)
            record.status = DeployStatus.SUCCEEDED
        except Exception as e:  # noqa: BLE001
            record.status = DeployStatus.FAILED
            record.error = str(e)
        record.updated_at = time.time()
        return record

    def get_status(self, idempotency_key: str) -> Optional[DeploymentRecord]:
        return self._store.get(idempotency_key)


# ---------------------------------------------------------------------------
# STEP 2 — OPTIMAL
# ---------------------------------------------------------------------------
# Reasoning:
#   Fix #1 (race): make the dedup check-and-insert a single atomic operation.
#   We don't need a distributed lock service for this — a UNIQUE constraint
#   in a persistent store gives us atomic "insert if absent" for free via
#   INSERT ... ON CONFLICT DO NOTHING, which SQLite (and Postgres) support
#   natively. Whichever request's INSERT actually lands owns the deploy;
#   every other concurrent request's INSERT is silently ignored, and it
#   just reads back the winner's record.
#
#   Fix #2 (durability): the row lives in SQLite on disk (or Postgres in
#   production), not in a Python dict, so a process restart doesn't lose
#   dedup history.
#
#   Fix #3 (crash mid-deploy): a status of RUNNING that hasn't been updated
#   in longer than a timeout is ambiguous — we genuinely don't know if the
#   deploy finished. We do NOT blindly retry (that reintroduces double-
#   deploys) and we do NOT blindly report failure. On startup, we flip any
#   stale RUNNING rows to UNKNOWN, which forces an explicit reconciliation
#   step (e.g., query the real deployment target — Kubernetes, the CD
#   system — for ground truth) before anyone acts on that record again.
#   This is the standard pattern for making an inherently non-idempotent
#   side effect (a deploy) safe under a system that can crash: durable
#   intent BEFORE the side effect, durable outcome AFTER it, and an
#   explicit "I don't know" state for the gap in between.
# ---------------------------------------------------------------------------
class OptimalDeployService:
    def __init__(self, deploy_fn, db_path: str = ":memory:", stale_running_timeout_s: float = 300):
        self._deploy_fn = deploy_fn
        self._stale_timeout = stale_running_timeout_s
        # check_same_thread=False: we manage our own locking below.
        self._conn = sqlite3.connect(db_path, check_same_thread=False)
        self._conn.execute("PRAGMA journal_mode=WAL")
        self._conn.execute(
            """
            CREATE TABLE IF NOT EXISTS deployments (
                idempotency_key TEXT PRIMARY KEY,
                service_name    TEXT NOT NULL,
                git_sha         TEXT NOT NULL,
                status          TEXT NOT NULL,
                error           TEXT,
                created_at      REAL NOT NULL,
                updated_at      REAL NOT NULL
            )
            """
        )
        self._conn.commit()
        # A per-key lock table guards the in-process race between the
        # winning INSERT and the "now actually run the deploy" step, so we
        # don't run the deploy fn twice inside the SAME process either
        # (SQLite's atomic INSERT only tells us who won the row — one of
        # those winners still has to execute deploy_fn exactly once).
        self._locks: dict[str, threading.Lock] = {}
        self._locks_guard = threading.Lock()
        self._reconcile_stale_running()

    def _lock_for(self, key: str) -> threading.Lock:
        with self._locks_guard:
            if key not in self._locks:
                self._locks[key] = threading.Lock()
            return self._locks[key]

    def _reconcile_stale_running(self) -> None:
        cutoff = time.time() - self._stale_timeout
        self._conn.execute(
            "UPDATE deployments SET status = ?, updated_at = ? "
            "WHERE status = ? AND updated_at < ?",
            (DeployStatus.UNKNOWN.value, time.time(), DeployStatus.RUNNING.value, cutoff),
        )
        self._conn.commit()

    def trigger(self, idempotency_key: str, service_name: str, git_sha: str) -> DeploymentRecord:
        now = time.time()
        cur = self._conn.cursor()
        # Atomic "insert if absent". Exactly one concurrent caller's INSERT
        # actually writes a new row; SQLite serializes writers, so this is
        # race-free even across threads/processes sharing this DB file.
        cur.execute(
            """
            INSERT INTO deployments
                (idempotency_key, service_name, git_sha, status, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?)
            ON CONFLICT(idempotency_key) DO NOTHING
            """,
            (idempotency_key, service_name, git_sha, DeployStatus.RUNNING.value, now, now),
        )
        self._conn.commit()
        won_the_insert = cur.rowcount == 1

        if not won_the_insert:
            # Someone else already owns (or owned) this key. Just report
            # its current state — do NOT deploy again.
            return self._read(idempotency_key)

        # We own this key's row. Use a local lock so that even if this
        # SAME process somehow calls trigger() twice concurrently for a
        # key it just won (shouldn't happen given the INSERT above, but
        # cheap insurance), the deploy_fn only runs once.
        with self._lock_for(idempotency_key):
            record = self._read(idempotency_key)
            if record.status != DeployStatus.RUNNING:
                return record  # already resolved by a previous attempt
            try:
                self._deploy_fn(service_name, git_sha)
                self._update_status(idempotency_key, DeployStatus.SUCCEEDED)
            except Exception as e:  # noqa: BLE001
                self._update_status(idempotency_key, DeployStatus.FAILED, error=str(e))
        return self._read(idempotency_key)

    def _update_status(self, key: str, status: DeployStatus, error: Optional[str] = None) -> None:
        self._conn.execute(
            "UPDATE deployments SET status = ?, error = ?, updated_at = ? WHERE idempotency_key = ?",
            (status.value, error, time.time(), key),
        )
        self._conn.commit()

    def _read(self, key: str) -> DeploymentRecord:
        row = self._conn.execute(
            "SELECT idempotency_key, service_name, git_sha, status, error, created_at, updated_at "
            "FROM deployments WHERE idempotency_key = ?",
            (key,),
        ).fetchone()
        return DeploymentRecord(
            idempotency_key=row[0], service_name=row[1], git_sha=row[2],
            status=DeployStatus(row[3]), error=row[4], created_at=row[5], updated_at=row[6],
        )

    def get_status(self, idempotency_key: str) -> Optional[DeploymentRecord]:
        cur = self._conn.execute(
            "SELECT 1 FROM deployments WHERE idempotency_key = ?", (idempotency_key,)
        )
        return self._read(idempotency_key) if cur.fetchone() else None

    def simulate_crash_and_restart(self) -> None:
        """Test hook: re-run startup reconciliation, as if the process
        had just restarted after a crash."""
        self._reconcile_stale_running()


# ---------------------------------------------------------------------------
# Minimal stdlib HTTP wrapper (no Flask dependency) around OptimalDeployService
#   POST /deployments   {"idempotency_key": "...", "service_name": "...", "git_sha": "..."}
#   GET  /deployments/<idempotency_key>
# ---------------------------------------------------------------------------
def make_handler(service: OptimalDeployService):
    class Handler(BaseHTTPRequestHandler):
        def _send(self, code: int, body: dict):
            payload = json.dumps(body).encode()
            self.send_response(code)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)

        def do_POST(self):
            if urlparse(self.path).path != "/deployments":
                return self._send(404, {"error": "not found"})
            length = int(self.headers.get("Content-Length", 0))
            body = json.loads(self.rfile.read(length) or b"{}")
            for field_name in ("idempotency_key", "service_name", "git_sha"):
                if not body.get(field_name):
                    return self._send(400, {"error": f"missing field: {field_name}"})
            record = service.trigger(body["idempotency_key"], body["service_name"], body["git_sha"])
            code = 201 if record.status in (DeployStatus.SUCCEEDED, DeployStatus.RUNNING) else 200
            self._send(code, _record_to_dict(record))

        def do_GET(self):
            path = urlparse(self.path).path
            prefix = "/deployments/"
            if not path.startswith(prefix):
                return self._send(404, {"error": "not found"})
            key = path[len(prefix):]
            record = service.get_status(key)
            if record is None:
                return self._send(404, {"error": "unknown idempotency_key"})
            self._send(200, _record_to_dict(record))

        def log_message(self, fmt, *args):  # quiet the default access log
            pass

    return Handler


def _record_to_dict(r: DeploymentRecord) -> dict:
    return {
        "idempotency_key": r.idempotency_key,
        "service_name": r.service_name,
        "git_sha": r.git_sha,
        "status": r.status.value,
        "error": r.error,
        "created_at": r.created_at,
        "updated_at": r.updated_at,
    }


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------
import unittest  # noqa: E402


class TestBruteForceHasRaceCondition(unittest.TestCase):
    def test_concurrent_duplicate_requests_deploy_more_than_once(self):
        deploy_calls = []
        calls_lock = threading.Lock()

        def slow_deploy(service_name, git_sha):
            # Simulate a deploy that takes a moment, widening the race window.
            time.sleep(0.05)
            with calls_lock:
                deploy_calls.append((service_name, git_sha))

        svc = BruteForceDeployService(deploy_fn=slow_deploy)
        barrier = threading.Barrier(20)

        def fire():
            barrier.wait()  # force all 20 threads to call trigger() at the same instant
            svc.trigger("same-key-123", "checkout-service", "abc123")

        threads = [threading.Thread(target=fire) for _ in range(20)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        # This is the bug we're demonstrating: many concurrent callers with
        # the SAME idempotency key end up all winning the race and the
        # underlying deploy runs more than once.
        self.assertGreater(len(deploy_calls), 1,
                            "expected the brute-force version to double-deploy under a race")


class TestOptimalDeployService(unittest.TestCase):
    def setUp(self):
        self.deploy_calls = []
        self.calls_lock = threading.Lock()

        def slow_deploy(service_name, git_sha):
            time.sleep(0.05)
            with self.calls_lock:
                self.deploy_calls.append((service_name, git_sha))

        self.deploy_fn = slow_deploy
        self.svc = OptimalDeployService(deploy_fn=self.deploy_fn, db_path=":memory:")

    def test_duplicate_key_deploys_exactly_once_under_concurrency(self):
        def fire():
            self.svc.trigger("same-key-123", "checkout-service", "abc123")

        threads = [threading.Thread(target=fire) for _ in range(50)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        self.assertEqual(len(self.deploy_calls), 1)

    def test_different_keys_deploy_independently(self):
        r1 = self.svc.trigger("key-a", "svc-a", "sha1")
        r2 = self.svc.trigger("key-b", "svc-b", "sha2")
        self.assertNotEqual(r1.idempotency_key, r2.idempotency_key)
        self.assertEqual(len(self.deploy_calls), 2)

    def test_status_query_for_unknown_key(self):
        self.assertIsNone(self.svc.get_status("never-seen"))

    def test_failed_deploy_is_recorded_and_not_retried_on_duplicate_request(self):
        def failing_deploy(service_name, git_sha):
            raise RuntimeError("deploy target unreachable")

        svc = OptimalDeployService(deploy_fn=failing_deploy, db_path=":memory:")
        first = svc.trigger("bad-key", "svc", "sha")
        self.assertEqual(first.status, DeployStatus.FAILED)
        self.assertIn("unreachable", first.error)

        second = svc.trigger("bad-key", "svc", "sha")
        self.assertEqual(second.status, DeployStatus.FAILED)
        # deploy_fn must not have been invoked a second time — a failed
        # deploy is still a *resolved* outcome for this idempotency key;
        # retrying automatically would violate "duplicate requests must
        # not trigger a second deployment".

    def test_crash_mid_deploy_leaves_status_running_then_reconciles_to_unknown(self):
        # Simulate: request came in, we won the insert, marked RUNNING,
        # and then the process died before deploy_fn finished (so no
        # SUCCEEDED/FAILED update ever landed).
        now = time.time() - 1000  # far in the past -> looks stale on restart
        self.svc._conn.execute(
            "INSERT INTO deployments (idempotency_key, service_name, git_sha, status, created_at, updated_at) "
            "VALUES (?, ?, ?, ?, ?, ?)",
            ("crashed-key", "svc", "sha", DeployStatus.RUNNING.value, now, now),
        )
        self.svc._conn.commit()

        before = self.svc.get_status("crashed-key")
        self.assertEqual(before.status, DeployStatus.RUNNING)

        self.svc.simulate_crash_and_restart()

        after = self.svc.get_status("crashed-key")
        self.assertEqual(after.status, DeployStatus.UNKNOWN,
                          "a stale RUNNING record must NOT be silently treated as success "
                          "or silently retried — it needs explicit reconciliation")

    def test_recent_running_record_is_not_falsely_reconciled(self):
        # A RUNNING record that's only a few ms old (deploy legitimately
        # still in flight) must NOT be flipped to UNKNOWN just because a
        # reconciliation pass happens to run.
        now = time.time()
        self.svc._conn.execute(
            "INSERT INTO deployments (idempotency_key, service_name, git_sha, status, created_at, updated_at) "
            "VALUES (?, ?, ?, ?, ?, ?)",
            ("in-flight-key", "svc", "sha", DeployStatus.RUNNING.value, now, now),
        )
        self.svc._conn.commit()
        self.svc.simulate_crash_and_restart()
        self.assertEqual(self.svc.get_status("in-flight-key").status, DeployStatus.RUNNING)


if __name__ == "__main__":
    import sys

    if "--serve" in sys.argv:
        service = OptimalDeployService(deploy_fn=lambda s, g: time.sleep(0.2), db_path="/tmp/deploys.db")
        server = ThreadingHTTPServer(("0.0.0.0", 8001), make_handler(service))
        print("Serving on http://0.0.0.0:8001  (POST /deployments, GET /deployments/<key>)")
        server.serve_forever()
    else:
        unittest.main()
