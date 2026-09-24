from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from typing import List
import time
import asyncio
from concurrent.futures import ThreadPoolExecutor

import extractor
from embeddings import EmbeddingService
from reranker import RerankerService

app = FastAPI(
    title="AI Avengers ML Sidecar",
    description=(
        "Local ML inference: embeddings (bge-base-en-v1.5) + reranking "
        "(bge-reranker-base) + document text extraction for ingestion"
    ),
    version="1.0.0"
)

# Initialize services at startup
# WHY module-level: Models are large (~500MB each). Load once, reuse forever.
# Loading per-request would be catastrophically slow.
embedding_service = EmbeddingService()
reranker_service = RerankerService()

# Thread pool for CPU-bound ML operations
# WHY ThreadPoolExecutor: asyncio is single-threaded. ML inference is CPU-bound.
# Running in thread pool prevents blocking the event loop.
# WHY max_workers=2: Two models, each can run in parallel.
# More workers = more memory, not more speed (CPU-bound).
executor = ThreadPoolExecutor(max_workers=2)

# Separate pool for document extraction.
# WHY separate from the ML executor: extraction is a burst of CPU + memory
# (parsing a 50MB PDF), and sharing the pool would let one big upload delay
# embedding calls for an in-flight ingestion. Two workers = two concurrent
# ingestions, which matches the admin workflow and bounds peak memory.
extract_executor = ThreadPoolExecutor(max_workers=2)

# Hard cap on uploaded bytes for /extract. The Go API enforces 50MB and gives a
# friendlier error; this is the defense-in-depth backstop.
MAX_UPLOAD_BYTES = 60 * 1024 * 1024


class EmbedRequest(BaseModel):
    texts: List[str]


class EmbedResponse(BaseModel):
    embeddings: List[List[float]]
    model: str
    duration_ms: float


class RerankRequest(BaseModel):
    query: str
    documents: List[str]
    top_k: int = 5


class RerankResult(BaseModel):
    index: int
    score: float
    text: str


class RerankResponse(BaseModel):
    results: List[RerankResult]
    model: str
    duration_ms: float


class ExtractResponse(BaseModel):
    text: str
    format: str
    chars: int
    pages: int = 0
    warnings: List[str] = []
    duration_ms: float


@app.on_event("startup")
async def startup_event():
    """Load ML models on startup.
    
    WHY load at startup:
    - First request would be very slow if models loaded lazily
    - Better to fail fast if models can't load
    - Health check will fail until models are ready
    """
    loop = asyncio.get_event_loop()
    
    # Load embedding model
    await loop.run_in_executor(executor, embedding_service.load)
    print(f"Embedding model loaded: {embedding_service.model_name}")
    
    # Load reranker model
    await loop.run_in_executor(executor, reranker_service.load)
    print(f"Reranker model loaded: {reranker_service.model_name}")
    
    print("ML Sidecar ready")


@app.get("/health")
async def health():
    """Health check endpoint.
    
    Returns 200 only if both models are loaded.
    Go backend checks this before marking ML sidecar as available.
    """
    if not embedding_service.is_loaded or not reranker_service.is_loaded:
        raise HTTPException(status_code=503, detail="Models not loaded yet")
    
    return {
        "status": "ok",
        "embedding_model": embedding_service.model_name,
        "reranker_model": reranker_service.model_name,
        # Surfaced so ops can see at a glance whether document extraction is
        # available in this deployment (it needs no model, only the parsers).
        "extract_formats": len(extractor.SUPPORTED_EXTENSIONS),
    }


@app.post("/embed", response_model=EmbedResponse)
async def embed(request: EmbedRequest):
    """Generate 768D embeddings for a batch of texts.
    
    Uses bge-base-en-v1.5 (local, zero cost).
    
    Mental execution:
    Input: ["How does sharding work?"]
    1. Validate: texts not empty
    2. Run in thread pool (CPU-bound)
    3. Return 768D float32 vectors
    """
    if not request.texts:
        return EmbedResponse(embeddings=[], model=embedding_service.model_name, duration_ms=0)
    
    if len(request.texts) > 100:
        raise HTTPException(status_code=400, detail="Maximum 100 texts per request")
    
    start = time.time()
    
    loop = asyncio.get_event_loop()
    embeddings = await loop.run_in_executor(
        executor,
        embedding_service.embed_batch,
        request.texts
    )
    
    duration_ms = (time.time() - start) * 1000
    
    return EmbedResponse(
        embeddings=embeddings,
        model=embedding_service.model_name,
        duration_ms=round(duration_ms, 2)
    )


@app.post("/rerank", response_model=RerankResponse)
async def rerank(request: RerankRequest):
    """Rerank documents by relevance to query.
    
    Uses bge-reranker-base (cross-encoder, local, zero cost).
    Returns top-K results sorted by score descending.
    
    Mental execution:
    Input: query="sharding", docs=["consistent hashing...", "kafka..."], top_k=5
    1. Validate: query not empty, docs not empty
    2. Create (query, doc) pairs
    3. Score with cross-encoder
    4. Sort by score DESC
    5. Return top-K
    """
    if not request.query:
        raise HTTPException(status_code=400, detail="Query cannot be empty")
    
    if not request.documents:
        return RerankResponse(results=[], model=reranker_service.model_name, duration_ms=0)
    
    if len(request.documents) > 50:
        raise HTTPException(status_code=400, detail="Maximum 50 documents per request")
    
    top_k = min(request.top_k, len(request.documents))
    
    start = time.time()
    
    loop = asyncio.get_event_loop()
    results = await loop.run_in_executor(
        executor,
        reranker_service.rerank,
        request.query,
        request.documents,
        top_k
    )
    
    duration_ms = (time.time() - start) * 1000
    
    return RerankResponse(
        results=[RerankResult(**r) for r in results],
        model=reranker_service.model_name,
        duration_ms=round(duration_ms, 2)
    )


@app.post("/extract", response_model=ExtractResponse)
async def extract(file: UploadFile = File(...)):
    """Turn an uploaded document into plain text for ingestion.

    WHY the sidecar owns this: parsing untrusted documents is a hostile-input
    job, and the mature parsers are Python. Doing it here keeps the file out of
    the Go API process (which holds DB credentials) and reuses the existing
    sidecar deployment instead of adding a service.

    Mental execution:
    Input:  multipart upload, e.g. lecture.pdf (12MB)
    1. Read the upload, enforce the size backstop.
    2. Extract in a worker thread (CPU-bound; must not block the event loop, or
       embeddings for an in-flight ingestion would stall).
    3. Return plain text + structure markers + metadata.

    Error cases:
      * 400 — no filename (cannot pick a parser)
      * 413 — upload above MAX_UPLOAD_BYTES
      * 422 — ExtractionError: the body carries a stable `reason` code
              (unsupported_format, pdf_encrypted, pdf_no_text, too_large, ...)
              and an admin-readable `message`. The Go layer shows the message
              verbatim, so it must stay actionable.
    """
    data = await file.read()
    if len(data) > MAX_UPLOAD_BYTES:
        raise HTTPException(
            status_code=413,
            detail=f"File is {len(data)} bytes, above the {MAX_UPLOAD_BYTES} limit",
        )

    filename = file.filename or ""
    if not filename:
        raise HTTPException(status_code=400, detail="filename is required to pick a parser")

    start = time.time()
    loop = asyncio.get_event_loop()
    try:
        result = await loop.run_in_executor(
            extract_executor, extractor.extract_document, filename, data
        )
    except extractor.ExtractionError as exc:
        # Stable machine-readable reason + human message. 422 (not 500): the
        # request was well-formed, the document's content is the problem.
        raise HTTPException(
            status_code=422, detail={"reason": exc.reason, "message": exc.message}
        )

    duration_ms = (time.time() - start) * 1000
    return ExtractResponse(
        text=result.text,
        format=result.format,
        chars=result.chars,
        pages=result.pages,
        warnings=result.warnings,
        duration_ms=round(duration_ms, 2),
    )
