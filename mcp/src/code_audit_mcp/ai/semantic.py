from typing import List, Tuple

class SemanticIndexer:
    """
    Lightweight semantic indexer using txtai if available. Falls back to keyword search.
    """

    def __init__(self):
        self._embeddings = None
        self._items: List[str] = []
        self._idmap: List[int] = []
        try:
            from txtai.embeddings import Embeddings
            self._embeddings = Embeddings({
                "method": "sentence-transformers",
                "path": "sentence-transformers/all-MiniLM-L6-v2"
            })
        except Exception:
            self._embeddings = None

    def build(self, texts: List[str]) -> bool:
        self._items = texts
        self._idmap = list(range(len(texts)))
        if self._embeddings and texts:
            self._embeddings.index([(i, t, None) for i, t in zip(self._idmap, texts)])
            return True
        return False

    def search(self, query: str, k: int = 5) -> List[Tuple[float, int]]:
        if self._embeddings and self._items:
            return [(float(score), int(idx)) for score, idx in self._embeddings.search(query, k)]
        # naive fallback: keyword occurrence score
        results = []
        q = query.lower()
        for idx, text in enumerate(self._items):
            score = text.lower().count(q)
            if score > 0:
                results.append((float(score), idx))
        results.sort(key=lambda x: x[0], reverse=True)
        return results[:k]