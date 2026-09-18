FROM python:3.11-slim

# Install git (required for workspace git operations)
RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash aider

# Install Aider from source (aider-main/ is already in the repo).
# Pin setuptools-scm<8 first: setuptools_scm>=8 changed _version_missing()
# signature, breaking the editable install of aider-main which calls it
# internally. Installing the pinned version before the editable install
# ensures pip resolves a compatible setuptools-scm for the entire env.
#
# --no-build-isolation is required on the second install: by default pip
# builds each package in a fresh, isolated temp env and re-resolves build
# dependencies there, which pulls the latest (incompatible) setuptools-scm
# right back in — silently undoing the pin above. --no-build-isolation
# makes the editable install reuse THIS image's site-packages (where the
# pinned version now lives) instead of resolving its own.
#
# SETUPTOOLS_SCM_PRETEND_VERSION avoids a second, unrelated failure: aider
# is installed from a COPY'd source tree with no .git directory, so
# setuptools-scm has no VCS metadata to derive a version from and errors
# out. Pretending a version skips that lookup entirely.
COPY aider-main/ /aider-main/
RUN pip install --no-cache-dir "setuptools-scm<8" && \
    SETUPTOOLS_SCM_PRETEND_VERSION="1.0.0" pip install --no-cache-dir --no-build-isolation -e /aider-main

# Install service dependencies
COPY aider-service/requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r /app/requirements.txt

# Copy service code
COPY aider-service/main.py /app/main.py

WORKDIR /app

# Workspaces directory (mounted as volume)
RUN mkdir -p /workspaces && chown aider:aider /workspaces

USER aider

EXPOSE 8082

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8082", "--workers", "1"]
