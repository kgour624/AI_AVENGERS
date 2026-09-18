import os
import subprocess
from fastapi import FastAPI
from pydantic import BaseModel

# Aider source is in aider-main/ (installed via pip install -e /aider-main)
from aider.coders import Coder
from aider.models import Model

app = FastAPI()

MODEL_GATEWAY_URL = os.getenv("MODEL_GATEWAY_URL", "http://localhost:8080/api/v1/llm/proxy")


class IterateRequest(BaseModel):
    workspace_path: str
    message: str
    expert_id: str
    workflow_id: str


class IterateResponse(BaseModel):
    commit_sha: str
    files_changed: list[str]
    success: bool
    error: str = ""


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/iterate", response_model=IterateResponse)
def iterate(req: IterateRequest) -> IterateResponse:
    try:
        # Configure Aider to route LLM calls through ModelGateway proxy
        # WHY: Centralized cost tracking, retry, provider switching
        model = Model(
            name="gpt-4",
            api_base=MODEL_GATEWAY_URL,
            api_key="proxy",  # Not used — ModelGateway handles auth
        )

        coder = Coder.create(
            main_model=model,
            fnames=[],  # Aider auto-detects files in workspace
            auto_commits=True,
            dirty_commits=False,
            git_dname=req.workspace_path,
        )

        # Set cost attribution headers on every LLM call
        if hasattr(coder.main_model, "extra_headers"):
            coder.main_model.extra_headers = {
                "X-Workflow-ID": req.workflow_id,
                "X-Expert-ID": req.expert_id,
            }

        # Run one Aider iteration
        coder.run(req.message)

        # Extract latest commit SHA
        commit_sha = subprocess.check_output(
            ["git", "rev-parse", "HEAD"],
            cwd=req.workspace_path,
        ).decode().strip()

        # Get files changed in last commit
        files_output = subprocess.check_output(
            ["git", "diff", "--name-only", "HEAD~1", "HEAD"],
            cwd=req.workspace_path,
        ).decode().strip()
        files_changed = [f for f in files_output.split("\n") if f]

        return IterateResponse(
            commit_sha=commit_sha,
            files_changed=files_changed,
            success=True,
        )

    except Exception as e:
        return IterateResponse(
            commit_sha="",
            files_changed=[],
            success=False,
            error=str(e),
        )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8082)
