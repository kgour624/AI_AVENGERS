#!/usr/bin/env python3
"""End-to-end check for one workflow run (G2 / A4).

WHY this exists
---------------
Every failure this project hit looked fine to the unit tests: a workflow that
reached "completed" having produced nothing, a finished run with zero LLM cost,
a board that stayed empty while the status said done. Each part behaved, and only
the whole did not — so the check has to be made of the questions a human would
ask after watching a run:

  1. Did it actually finish?               status == completed
  2. Did it actually spend anything?       cost_spent_usd > 0
  3. Did it actually produce anything?     at least one design/code artifact

A run that answers "no" to 2 or 3 while claiming 1 is exactly the lie this script
is built to catch.

Usage
-----
    API_BASE=http://localhost:8080 TOKEN=<access token> \
        python scripts/e2e_workflow_check.py [workflow_id]

With no workflow id it checks the most recently updated workflow. Exits 0 on
PASS and 1 on FAIL, so it can be put in CI later without changes.
"""

import json
import os
import sys
import time
import urllib.error
import urllib.request

API_BASE = os.environ.get("API_BASE", "http://localhost:8080").rstrip("/")
TOKEN = os.environ.get("TOKEN", "")
POLL_SECONDS = int(os.environ.get("POLL_SECONDS", "5"))
TIMEOUT_SECONDS = int(os.environ.get("TIMEOUT_SECONDS", "1800"))

# The same event types the runner counts as produced work. Kept in sync by hand:
# a checker that asks a different question than the product does is no checker.
PRODUCED_WORK_TYPES = {
    "architecture_decision",
    "data_model_proposed",
    "api_contract_proposed",
    "module_design_proposed",
    "code_artifact_produced",
    "design_section_written",
}

TERMINAL = {"completed", "failed", "cancelled"}


def call(path):
    """GET one API path and return the `data` field of the response envelope."""
    url = f"{API_BASE}{path}"
    req = urllib.request.Request(url, headers={"Accept": "application/json"})
    if TOKEN:
        req.add_header("Authorization", f"Bearer {TOKEN}")
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            body = json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        raise SystemExit(f"FAIL: {path} -> HTTP {e.code} {e.read()[:200].decode('utf-8', 'replace')}")
    except urllib.error.URLError as e:
        raise SystemExit(f"FAIL: cannot reach {url} ({e.reason}). Is the stack up?")
    if isinstance(body, dict) and "data" in body:
        return body["data"]
    return body


def find_workflow_id():
    """The most recently updated workflow, so no id has to be copied by hand."""
    data = call("/api/v1/workflows")
    items = data if isinstance(data, list) else data.get("workflows", data.get("items", []))
    if not items:
        raise SystemExit("FAIL: no workflows found — create and start one first")
    return items[0]["id"]


def count_produced_work(workflow_id):
    """Count design/code artifacts on the blackboard."""
    data = call(f"/api/v1/workflows/{workflow_id}/blackboard")
    events = data if isinstance(data, list) else data.get("events", [])
    seen = {}
    for ev in events:
        et = ev.get("event_type") or ev.get("eventType") or ""
        if et in PRODUCED_WORK_TYPES:
            seen[et] = seen.get(et, 0) + 1
    return seen


def main():
    workflow_id = sys.argv[1] if len(sys.argv) > 1 else find_workflow_id()
    print(f"checking workflow {workflow_id} at {API_BASE}")
    if not TOKEN:
        print("note: no TOKEN set — the API will reject an authenticated read")

    deadline = time.time() + TIMEOUT_SECONDS
    wf = {}
    while True:
        wf = call(f"/api/v1/workflows/{workflow_id}")
        status = (wf.get("status") or "").lower()
        phase = wf.get("current_phase") or wf.get("currentPhase") or "?"
        cost = wf.get("cost_spent_usd", wf.get("costSpentUsd", 0)) or 0
        print(f"  status={status:<22} phase={phase:<20} cost=${float(cost):.4f}")
        if status in TERMINAL:
            break
        if time.time() > deadline:
            raise SystemExit(f"FAIL: still '{status}' after {TIMEOUT_SECONDS}s — the run is stuck, not finished")

    produced = count_produced_work(workflow_id)
    produced_total = sum(produced.values())
    status = (wf.get("status") or "").lower()
    cost = float(wf.get("cost_spent_usd", wf.get("costSpentUsd", 0)) or 0)
    reason = wf.get("failure_reason") or wf.get("failureReason") or ""

    print("")
    print(f"  status            : {status}")
    print(f"  cost spent        : ${cost:.4f}")
    print(f"  artifacts produced: {produced_total} {dict(sorted(produced.items()))}")
    if reason:
        print(f"  failure reason    : {reason}")

    problems = []
    if status != "completed":
        problems.append(f"the workflow did not complete (status={status})")
    if cost <= 0:
        problems.append("no LLM cost was recorded — nothing was actually asked of a model")
    if produced_total == 0:
        problems.append("no design or code artifact was produced — a green 'completed' over nothing")

    print("")
    if problems:
        print("FAIL")
        for p in problems:
            print(f"  - {p}")
        return 1
    print("PASS — the run finished and left evidence that it did real work")
    return 0


if __name__ == "__main__":
    sys.exit(main())
