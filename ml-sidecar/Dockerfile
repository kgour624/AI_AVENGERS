FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --default-timeout=1000 --no-cache-dir -r requirements.txt

# Copy source
COPY . .

# Non-root user
RUN useradd -m avengers
USER avengers

EXPOSE 8001

# WHY --workers 1:
# ML models are loaded in memory. Multiple workers = multiple model copies.
# One worker is sufficient — async handles concurrent requests.
# If more throughput needed, scale horizontally (multiple containers).
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8001", "--workers", "1"]
