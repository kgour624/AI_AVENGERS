"""FastAPI wrapper that runs one Aider iteration inside a workspace.

Every LLM call Aider makes is routed to the Go backend's /llm/proxy endpoint,
which is OpenAI chat-completions compatible. The proxy decides which real
model runs (the admin's configured provider and "strong" tier) — this service
never talks to a model vendor directly, so cost tracking, retries and provider
switching stay in one place.
"""

import os
import subprocess

from fastapi import FastAPI
from pydantic import BaseModel

# Aider source is in aider-main/ (installed via pip install -e /aider-main)
from aider import models as aider_models
from aider.coders import Coder
from aider.io import InputOutput
from aider.models import Model
from aider.repo import GitRepo

app = FastAPI()

# The base URL Aider's HTTP client is pointed at. The OpenAI client library
# inside litellm appends "/chat/completions" to it, so the backend registers
# both /llm/proxy and /llm/proxy/chat/completions.
MODEL_GATEWAY_URL = os.getenv("MODEL_GATEWAY_URL", "http://localhost:8080/api/v1/llm/proxy")

# Shared secret for the proxy. It is sent as the OpenAI api_key, which means it
# arrives as "Authorization: Bearer <token>" — the only header the OpenAI
# client library lets us control on that path. Must match the backend's
# AIDER_PROXY_TOKEN.
PROXY_TOKEN = os.getenv("AIDER_PROXY_TOKEN", "")

# Model name handed to litellm.
#
# The "openai/" prefix is load-bearing: it selects litellm's OpenAI-compatible
# transport, which is the one that honours OPENAI_API_BASE. Without the prefix
# litellm picks a vendor transport with a hardcoded endpoint and the proxy is
# bypassed. The part after the prefix is only a label — the proxy ignores the
# requested model and uses the admin-configured provider.
AIDER_MODEL = os.getenv("AIDER_MODEL", "openai/deepseek-chat")

# "diff" (search/replace blocks) instead of Aider's "whole" default. "whole"
# makes the model re-emit every edited file in full, which on a real repo burns
# output tokens and truncates against the output cap.
AIDER_EDIT_FORMAT = os.getenv("AIDER_EDIT_FORMAT", "diff")

# Context window Aider should assume. Aider normally reads this from litellm's
# model database, but "openai/<label>" is not in that database, so it would
# assume 0 and mis-size the repo map and history summarisation.
AIDER_MAX_INPUT_TOKENS = int(os.getenv("AIDER_MAX_INPUT_TOKENS", "64000"))
AIDER_MAX_OUTPUT_TOKENS = int(os.getenv("AIDER_MAX_OUTPUT_TOKENS", "8000"))

# litellm reads credentials from the environment, not from constructor
# arguments. Setting them here covers every Model this process builds —
# including the weak model Aider uses for commit messages and history
# summarisation, which we never construct by hand.
os.environ["OPENAI_API_BASE"] = MODEL_GATEWAY_URL
os.environ["OPENAI_API_KEY"] = PROXY_TOKEN or "unset"

# Teach Aider the model's limits and that it costs nothing to call. Cost is
# recorded server-side against the workflow, so a second figure here would
# double-count. This is the same dict Aider's --model-metadata-file option
# populates.
aider_models.model_info_manager.local_model_metadata[AIDER_MODEL] = {
    "max_input_tokens": AIDER_MAX_INPUT_TOKENS,
    "max_output_tokens": AIDER_MAX_OUTPUT_TOKENS,
    "max_tokens": AIDER_MAX_INPUT_TOKENS,
    "input_cost_per_token": 0.0,
    "output_cost_per_token": 0.0,
    "litellm_provider": "openai",
    "mode": "chat",
}


class IterateRequest(BaseModel):
    workspace_path: str
    message: str
    expert_id: str
    workflow_id: str
    # Files added to Aider's chat as editable, relative to workspace_path.
    # Empty is normal on a greenfield task — the edit format lets the model
    # create new files without asking.
    edit_files: list[str] = []
    # Files added as reference material the model may read but must not change
    # (the approved design documents).
    #
    # WHY they must be listed: Aider only reads files that are in the chat.
    # Writing them into the workspace leaves them closed, and the repo map shows
    # little more than an .md file's headings — so the model was asked to
    # implement a design it could not actually see.
    read_only_files: list[str] = []


class IterateResponse(BaseModel):
    commit_sha: str
    files_changed: list[str]
    success: bool
    error: str = ""


@app.get("/health")
def health():
    return {"status": "ok"}


def _existing_paths(workspace_path: str, names: list[str]) -> list[str]:
    """Turn workspace-relative names into absolute paths, skipping missing ones.

    Missing entries are skipped rather than passed through: for a name that does
    not exist, Aider's Coder creates an empty file and adds it to the chat, which
    would litter the repository with empty files whenever the caller's list drifts
    from what is actually on disk.
    """
    resolved = []
    for name in names:
        path = os.path.join(workspace_path, name)
        if os.path.isfile(path):
            resolved.append(path)
    return resolved


def _git(workspace_path: str, *args: str) -> str:
    return (
        subprocess.check_output(["git", *args], cwd=workspace_path)
        .decode(errors="replace")
        .strip()
    )


def _build_model(workflow_id: str, expert_id: str) -> Model:
    """Build a Model whose requests carry workflow/expert attribution headers.

    extra_params is the only injection point Aider offers: Model.send_completion
    merges it straight into the litellm.completion(**kwargs) call. Assigning
    model.extra_headers instead just sets an attribute nobody reads, so the
    backend saw no attribution headers at all.
    """
    model = Model(AIDER_MODEL, weak_model=AIDER_MODEL, editor_model=AIDER_MODEL)

    model.extra_params = dict(model.extra_params or {})
    model.extra_params["extra_headers"] = {
        "X-Workflow-ID": workflow_id,
        "X-Expert-ID": expert_id,
    }

    # The proxy answers with a single complete body, never an SSE stream.
    model.streaming = False
    model.use_repo_map = True
    return model


@app.post("/iterate", response_model=IterateResponse)
def iterate(req: IterateRequest) -> IterateResponse:
    try:
        model = _build_model(req.workflow_id, req.expert_id)

        # yes=True auto-confirms every prompt (adding files, creating files,
        # committing). pretty/fancy_input off because there is no terminal:
        # prompt_toolkit's interactive input needs a tty and would raise here.
        io = InputOutput(
            pretty=False,
            yes=True,
            fancy_input=False,
            root=req.workspace_path,
        )

        # The repo has to be built explicitly with git_dname. Left to itself,
        # Coder builds a GitRepo with git_dname=None, which resolves to the
        # process working directory (/app) — not this workflow's workspace.
        # chdir is not an option: FastAPI serves sync endpoints on a thread
        # pool, so a process-global cwd would race between concurrent tasks.
        repo = GitRepo(
            io,
            fnames=[],
            git_dname=req.workspace_path,
            models=model.commit_message_models(),
        )

        head_before = _git(req.workspace_path, "rev-parse", "HEAD")

        # Resolve against the workspace: Aider stores absolute paths, and a
        # relative name would resolve against this process's cwd (/app).
        edit_paths = _existing_paths(req.workspace_path, req.edit_files)
        read_only_paths = _existing_paths(req.workspace_path, req.read_only_files)

        coder = Coder.create(
            main_model=model,
            edit_format=AIDER_EDIT_FORMAT,
            io=io,
            repo=repo,
            fnames=edit_paths,
            read_only_fnames=read_only_paths,
            auto_commits=True,
            dirty_commits=False,
            # stream=False: Coder defaults to streaming, and the proxy returns
            # one complete JSON body. A streaming client would wait for SSE
            # frames that never come and fail as an unexplained timeout.
            stream=False,
            # Off because nothing here can service them: there is no user to
            # approve a shell command, no linter configured in the workspace,
            # and no reason to fetch URLs mentioned in a task description.
            suggest_shell_commands=False,
            detect_urls=False,
            auto_lint=False,
            auto_test=False,
        )

        coder.run(req.message)

        head_after = _git(req.workspace_path, "rev-parse", "HEAD")

        if head_after == head_before:
            # Aider reports LLM and parsing failures through its console and
            # then returns normally. Without this check the caller was told the
            # iteration succeeded while nothing had changed. Report the reason
            # Aider recorded instead of a bare "no changes".
            reasons = []
            if coder.num_exhausted_context_windows:
                reasons.append(
                    f"context window exhausted x{coder.num_exhausted_context_windows}"
                )
            if coder.num_malformed_responses:
                reasons.append(
                    f"malformed edit blocks x{coder.num_malformed_responses}"
                )
            detail = "; ".join(reasons) if reasons else "model proposed no edits"
            return IterateResponse(
                commit_sha=head_before,
                files_changed=[],
                success=False,
                error=f"aider made no commit ({detail})",
            )

        # Diff against the pre-run HEAD, not HEAD~1. Aider may produce more
        # than one commit in an iteration, and HEAD~1 does not exist when the
        # workspace's seed commit is the only one.
        files_output = _git(
            req.workspace_path, "diff", "--name-only", head_before, head_after
        )
        files_changed = [f for f in files_output.split("\n") if f]

        return IterateResponse(
            commit_sha=coder.last_aider_commit_hash or head_after,
            files_changed=files_changed,
            success=True,
        )

    except subprocess.CalledProcessError as e:
        output = (e.output or b"").decode(errors="replace").strip()
        return IterateResponse(
            commit_sha="",
            files_changed=[],
            success=False,
            error=f"git {' '.join(e.cmd[1:])} failed (exit {e.returncode}): {output}",
        )
    except Exception as e:
        return IterateResponse(
            commit_sha="",
            files_changed=[],
            success=False,
            error=f"{type(e).__name__}: {e}",
        )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8082)
