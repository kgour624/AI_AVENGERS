FROM python:3.11-slim

# LibreOffice is optional and OFF by default.
# WHY: it is ~500MB and only needed for the long tail of formats that have no
# Python parser (.wpd, .pages, .key, ...). The legacy formats admins actually
# upload (.doc, .ppt, .xls) are covered by catdoc (~2MB), installed below.
# Enable with: docker build --build-arg WITH_LIBREOFFICE=true
ARG WITH_LIBREOFFICE=false

WORKDIR /app

# System packages for document extraction:
#   catdoc  -> catdoc (.doc), catppt (.ppt), xls2csv (.xls) as a last resort
#   Others  -> runtime libs lxml/odfpy need (installed as wheels where possible)
# --no-install-recommends keeps the image small; --no-install-suggests too.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        catdoc \
    && if [ "$WITH_LIBREOFFICE" = "true" ]; then \
         apt-get install -y --no-install-recommends \
             libreoffice-writer libreoffice-calc libreoffice-impress; \
       fi \
    && rm -rf /var/lib/apt/lists/*

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
# One worker is sufficient — async handles concurrent requests, and CPU-bound
# work (embeddings, document extraction) runs in bounded ThreadPoolExecutors.
# If more throughput needed, scale horizontally (multiple containers).
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8001", "--workers", "1"]
