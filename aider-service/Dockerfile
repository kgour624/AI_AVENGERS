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
COPY aider-main/ /aider-main/
RUN pip install --no-cache-dir "setuptools-scm<8" && \
    pip install --no-cache-dir -e /aider-main

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
