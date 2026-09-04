from sentence_transformers import CrossEncoder
from typing import List, Dict
import numpy as np


class RerankerService:
    """Reranks documents using bge-reranker-base (cross-encoder).
    
    WHY cross-encoder vs bi-encoder for reranking:
    - Bi-encoder (embedding model): Fast, approximate. Good for initial retrieval.
    - Cross-encoder (reranker): Slow, accurate. Good for final ranking.
    
    We use bi-encoder for semantic search (top-20 candidates),
    then cross-encoder to rerank to top-5.
    This is the China Wall Layer 1 check.
    
    WHY bge-reranker-base:
    - Local, zero cost
    - Good accuracy for technical content
    - ~280MB model size
    """
    
    MODEL_NAME = "BAAI/bge-reranker-base"
    
    def __init__(self):
        self.model = None
        self.is_loaded = False
        self.model_name = self.MODEL_NAME
    
    def load(self):
        """Load the model. Called once at startup."""
        self.model = CrossEncoder(self.MODEL_NAME)
        self.is_loaded = True
    
    def rerank(self, query: str, documents: List[str], top_k: int) -> List[Dict]:
        """Rerank documents by relevance to query.
        
        Mental execution:
        Input: query="sharding", docs=["consistent hashing...", "kafka..."], top_k=5
        1. Create (query, doc) pairs
        2. Score each pair with cross-encoder
        3. Sort by score DESC
        4. Return top-K with original index
        
        Returns: List of {index, score, text} sorted by score DESC
        """
        if not self.is_loaded:
            raise RuntimeError("Model not loaded. Call load() first.")
        
        if not documents:
            return []
        
        # Create query-document pairs
        pairs = [(query, doc) for doc in documents]
        
        # Score all pairs
        # predict() returns raw logits — apply sigmoid for 0-1 range
        scores = self.model.predict(pairs, apply_softmax=False)
        
        # Apply sigmoid to get 0-1 scores
        scores = 1 / (1 + np.exp(-scores))
        
        # Create results with original indices
        results = [
            {"index": i, "score": float(score), "text": doc}
            for i, (doc, score) in enumerate(zip(documents, scores))
        ]
        
        # Sort by score descending
        results.sort(key=lambda x: x["score"], reverse=True)
        
        # Return top-K
        return results[:top_k]
