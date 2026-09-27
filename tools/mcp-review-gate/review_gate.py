#!/usr/bin/env python3
"""Ask a domain expert to review a change, and gate on the verdict.

WHAT IT DOES
    Sends a diff (or a set of files, or stdin) to the AI AVENGERS MCP server's
    `review_change` tool and reads back the expert's verdict. A
    "VERDICT: CHANGES_REQUESTED" reply is a failure, so it can be used as a
    Claude Code hook or a git pre-commit hook and actually stop the change.

WHY A STANDALONE SCRIPT AND NOT PRODUCT CODE
    This is a developer tool. It lives outside the backend on purpose: it may
    change as often as a team's workflow changes, it must run on a laptop with no
    Go toolchain, and nothing in it should ever be able to affect the running
    product. It only speaks HTTP to the MCP server.

WHY IT FAILS OPEN BY DEFAULT
    If the expert cannot be reached (no network, server down, expired token) the
    honest answer is "no review happened" — not "your commit is bad". Blocking a
    developer because a review service was unreachable teaches them to bypass the
    gate. So a transport failure prints a loud warning and exits 0. Use
    --fail-closed where a gate must never be silently skipped (release branches,
    CI), and the default stays out of the way.

CONFIGURATION (environment)
    MCP_URL     base URL of the MCP server, e.g. http://127.0.0.1:8090
    MCP_TOKEN   bearer token minted with cmd/mcp-token
    MCP_DOMAIN  which expert domain reviews, e.g. "frontend" (required)

USAGE
    # everything staged for commit
    python review_gate.py

    # an explicit diff
    git diff | python review_gate.py --diff-from - --domain frontend

    # specific files, warn only
    python review_gate.py --files src/app.ts src/api.ts --warn-only

    # verify the parsing logic without touching the network
    python review_gate.py --selftest

EXIT CODES
    0  approved, or warn-only, or the expert was unreachable (fail-open)
    1  changes requested  (the change is blocked)
    2  misconfiguration or a rejected token
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import urllib.error
import urllib.request

# One review carries a diff, not a repository. The cap protects the provider and
# the developer's latency budget; a change larger than this is reviewed in parts
# (or by a human), not silently truncated.
MAX_CHANGE_BYTES = 200_000

# Timeouts: connecting/uploading should be quick, but the expert may think for
# minutes on a large change.
CONNECT_TIMEOUT_SECONDS = 15
REVIEW_TIMEOUT_SECONDS = 600

# The verdict contract. The prompt asks for the verdict on its own first line;
# matching it case-insensitively keeps the gate stable if an expert writes
# "verdict: changes_requested".
APPROVED = "VERDICT: APPROVED"
CHANGES_REQUESTED = "VERDICT: CHANGES_REQUESTED"


class ConfigError(Exception):
    """The gate cannot run at all (missing config, rejected token)."""


class TransportError(Exception):
    """The expert could not be reached — fail-open territory."""


def read_config(args: argparse.Namespace) -> tuple[str, str, str]:
    """Resolve URL, token and domain from flags, falling back to the environment."""
    url = (args.url or os.environ.get("MCP_URL", "")).strip().rstrip("/")
    token = (args.token or os.environ.get("MCP_TOKEN", "")).strip()
    domain = (args.domain or os.environ.get("MCP_DOMAIN", "")).strip()

    missing = [name for name, value in (("MCP_URL", url), ("MCP_TOKEN", token), ("domain", domain)) if not value]
    if missing:
        raise ConfigError("missing configuration: " + ", ".join(missing))
    return url, token, domain


def gather_change(args: argparse.Namespace) -> str:
    """Collect the text to review from stdin, files, or the staged diff."""
    if args.diff_from == "-":
        return sys.stdin.read()
    if args.diff_from:
        return subprocess.run(
            ["git", "diff", args.diff_from],
            capture_output=True, text=True, check=False,
        ).stdout
    if args.files:
        chunks = []
        for path in args.files:
            try:
                with open(path, "r", encoding="utf-8", errors="replace") as handle:
                    chunks.append(f"### {path}\n{handle.read()}")
            except OSError as exc:
                # A file the agent deleted or cannot read is not a review
                # failure; note it and review what exists.
                chunks.append(f"### {path}\n(could not read: {exc})")
        return "\n\n".join(chunks)
    # Default: what is staged for the next commit. That is what a pre-commit
    # hook must judge, and it is also the most useful default for an agent.
    staged = subprocess.run(
        ["git", "diff", "--cached"],
        capture_output=True, text=True, check=False,
    ).stdout
    if staged.strip():
        return staged
    # Nothing staged: fall back to unstaged work so a hook is still useful while
    # a developer iterates.
    return subprocess.run(
        ["git", "diff"],
        capture_output=True, text=True, check=False,
    ).stdout


def call_review(url: str, token: str, domain: str, change: str, context: str) -> str:
    """Call the MCP review_change tool and return its text."""
    if len(change.encode("utf-8")) > MAX_CHANGE_BYTES:
        raise ConfigError(
            f"the change is larger than {MAX_CHANGE_BYTES} bytes; review it in smaller parts"
        )

    payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "review_change",
            "arguments": {"domain": domain, "change": change, "context": context},
        },
    }
    request = urllib.request.Request(
        url + "/",
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": "Bearer " + token,
        },
        method="POST",
    )

    try:
        with urllib.request.urlopen(request, timeout=REVIEW_TIMEOUT_SECONDS) as response:
            body = response.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")[:400]
        if exc.code in (401, 403):
            raise ConfigError(f"the MCP server rejected the token ({exc.code}): {detail}") from exc
        raise TransportError(f"the MCP server returned {exc.code}: {detail}") from exc
    except (urllib.error.URLError, TimeoutError, OSError) as exc:
        raise TransportError(f"could not reach the MCP server at {url}: {exc}") from exc

    return extract_text(body)


def extract_text(body: str) -> str:
    """Pull the tool's text out of a JSON-RPC reply."""
    try:
        envelope = json.loads(body)
    except json.JSONDecodeError as exc:
        raise TransportError(f"the MCP server sent a reply that is not JSON: {exc}") from exc

    if isinstance(envelope.get("error"), dict):
        message = envelope["error"].get("message", "unknown protocol error")
        raise TransportError(f"protocol error from the MCP server: {message}")

    result = envelope.get("result") or {}
    if result.get("isError"):
        # A tool-level failure (unknown domain, forbidden token scope) is a
        # configuration problem the developer must fix, not a silent pass.
        texts = [part.get("text", "") for part in result.get("content", []) if isinstance(part, dict)]
        raise ConfigError("the expert refused the review: " + " ".join(texts).strip())

    texts = [part.get("text", "") for part in result.get("content", []) if isinstance(part, dict)]
    text = "\n".join(t for t in texts if t).strip()
    if not text:
        raise TransportError("the MCP server returned an empty review")
    return text


def verdict_of(review: str) -> str:
    """Return 'approved', 'changes_requested' or 'unclear'.

    WHY 'unclear' instead of guessing: a review whose verdict line is missing is
    not an approval. Treating it as one would let a malformed answer silently
    pass a gate, which is exactly the failure this tool exists to prevent.
    """
    head = review[:600].upper()
    if CHANGES_REQUESTED in head:
        return "changes_requested"
    if APPROVED in head:
        return "approved"
    return "unclear"


def run_gate(args: argparse.Namespace) -> int:
    url, token, domain = read_config(args)
    change = gather_change(args)
    if not change.strip():
        print("review-gate: there is no change to review — nothing to do.")
        return 0

    context = args.context or os.environ.get("MCP_REVIEW_CONTEXT", "")

    try:
        review = call_review(url, token, domain, change, context)
    except ConfigError as exc:
        print(f"review-gate: configuration error: {exc}", file=sys.stderr)
        return 2
    except TransportError as exc:
        message = f"review-gate: no review happened — {exc}"
        if args.fail_closed:
            print(message + " (fail-closed: blocking)", file=sys.stderr)
            return 1
        # Fail open, loudly: a developer must never be blocked by a dead service,
        # and must never believe a review happened when it did not.
        print(message + " (fail-open: continuing)", file=sys.stderr)
        return 0

    verdict = verdict_of(review)

    if verdict == "changes_requested":
        print(review)
        if args.warn_only:
            print("\nreview-gate: the expert requested changes (warn-only: not blocking).", file=sys.stderr)
            return 0
        print("\nreview-gate: the expert requested changes — the change is blocked.", file=sys.stderr)
        return 1

    if verdict == "unclear":
        message = "review-gate: the expert's reply carried no verdict line; not treating it as approval"
        if args.fail_closed:
            print(review)
            print(message + " (fail-closed: blocking)", file=sys.stderr)
            return 1
        print(message + " (fail-open: continuing)", file=sys.stderr)
        return 0

    print("review-gate: the expert approved this change.")
    if args.verbose:
        print(review)
    return 0


def selftest() -> int:
    """Verify parsing and verdict handling without a server.

    A gate that silently passes is worse than no gate, so the failure modes are
    tested as carefully as the happy path.
    """
    failures = []

    def check(name: str, got, want) -> None:
        if got != want:
            failures.append(f"{name}: got {got!r}, want {want!r}")

    def reply(text: str, is_error: bool = False) -> str:
        return json.dumps({
            "jsonrpc": "2.0", "id": 1,
            "result": {"content": [{"type": "text", "text": text}], "isError": is_error},
        })

    check("approved", verdict_of("VERDICT: APPROVED\n- all good"), "approved")
    check("changes", verdict_of("VERDICT: CHANGES_REQUESTED\n- x"), "changes_requested")
    check("case", verdict_of("verdict: changes_requested\n- x"), "changes_requested")
    check("missing verdict", verdict_of("Looks fine to me."), "unclear")
    check("empty", verdict_of(""), "unclear")
    check("extract", extract_text(reply("VERDICT: APPROVED")), "VERDICT: APPROVED")

    try:
        extract_text(reply("no expert for domain", is_error=True))
        failures.append("tool error: expected ConfigError")
    except ConfigError:
        pass

    try:
        extract_text(json.dumps({"jsonrpc": "2.0", "error": {"code": -32601, "message": "no method"}}))
        failures.append("protocol error: expected TransportError")
    except TransportError:
        pass

    if failures:
        for line in failures:
            print("selftest FAILED —", line, file=sys.stderr)
        return 1
    print("review-gate selftest: all checks passed.")
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Gate a change on a domain expert's review (via the AI AVENGERS MCP server).",
    )
    parser.add_argument("--url", help="MCP server base URL (default: $MCP_URL)")
    parser.add_argument("--token", help="bearer token (default: $MCP_TOKEN)")
    parser.add_argument("--domain", help="expert domain that reviews (default: $MCP_DOMAIN)")
    parser.add_argument("--files", nargs="*", help="review these files instead of a diff")
    parser.add_argument("--diff-from", help="git ref to diff against, or '-' to read stdin")
    parser.add_argument("--context", help="what the change is meant to do")
    parser.add_argument("--warn-only", action="store_true", help="never block, only report")
    parser.add_argument("--fail-closed", action="store_true",
                        help="block when the review cannot be obtained (default: fail open)")
    parser.add_argument("--verbose", action="store_true", help="print the full review on approval")
    parser.add_argument("--selftest", action="store_true", help="verify parsing logic offline")
    return parser


def main() -> int:
    args = build_parser().parse_args()
    if args.selftest:
        return selftest()
    try:
        return run_gate(args)
    except ConfigError as exc:
        print(f"review-gate: configuration error: {exc}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    sys.exit(main())
