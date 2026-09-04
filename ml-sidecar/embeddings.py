from sentence_transformers import SentenceTransformer
from typing import List
import numpy as np


class EmbeddingService:
    """Generates 768D embeddings using bge-base-en-v1.5.
    
    WHY bge-base-en-v1.5:
    - 768 dimensions (good balance of quality vs storage)
    - Optimized for retrieval tasks
    - Local — zero cost per embedding
    - ~500MB model size — fits in memory
    """
    
    MODEL_NAME = "BAAI/bge-base-en-v1.5"
    EMBEDDING_DIM = 768
    
    def __init__(self):
        self.model = None
        self.is_loaded = False
        self.model_name = self.MODEL_NAME
    
    def load(self):
        """Load the model. Called once at startup."""
        self.model = SentenceTransformer(self.MODEL_NAME)
        self.is_loaded = True
    
    def embed_batch(self, texts: List[str]) -> List[List[float]]:
        """Generate embeddings for a batch of texts.
        
        Mental execution:
        Input: ["How does sharding work?", "Consistent hashing"]
        1. Normalize texts (strip whitespace)
        2. Add BGE instruction prefix (improves retrieval quality)
        3. Encode with model
        4. Normalize vectors (for cosine similarity)
        5. Convert to Python lists (JSON-serializable)
        
        Returns: List of 768D float32 vectors
        """
        if not self.is_loaded:
            raise RuntimeError("Model not loaded. Call load() first.")
        
        # Clean texts
        cleaned = [t.strip() for t in texts if t.strip()]
        if not cleaned:
            return []
        
        # BGE models benefit from instruction prefix for retrieval
        # WHY: bge-base-en-v1.5 was trained with this prefix for passage retrieval
        instruction = "Represent this sentence for searching relevant passages: "
        prefixed = [instruction + t for t in cleaned]
        
        # Generate embeddings
        # normalize_embeddings=True: L2 normalization for cosine similarity
        # WHY normalize: pgvector cosine distance requires normalized vectors
        embeddings = self.model.encode(
            prefixed,
            normalize_embeddings=True,
            batch_size=32,
            show_progress_bar=False
        )
        
        # Convert numpy arrays to Python lists for JSON serialization
        return embeddings.tolist()
