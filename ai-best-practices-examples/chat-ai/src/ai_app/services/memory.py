import json
import math
import re
from dataclasses import dataclass
from pathlib import Path


WORD_RE = re.compile(r"[a-zA-Z0-9']+")


@dataclass
class MemoryItem:
    user_id: str
    text: str
    vector: dict[str, float]


class MemoryService:
    def __init__(self, store_path: Path) -> None:
        self.store_path = store_path
        self._items: list[MemoryItem] = []
        self._load()

    def add_turn(self, user_id: str, text: str) -> None:
        item = MemoryItem(user_id=user_id, text=text, vector=self._embed(text))
        self._items.append(item)
        self._save()

    def retrieve(self, user_id: str, query: str, top_k: int = 3) -> list[MemoryItem]:
        query_vector = self._embed(query)
        scored: list[tuple[float, MemoryItem]] = []
        for item in self._items:
            if item.user_id != user_id:
                continue
            score = self._cosine_similarity(query_vector, item.vector)
            if score > 0:
                scored.append((score, item))
        scored.sort(key=lambda pair: pair[0], reverse=True)
        return [item for _, item in scored[:top_k]]

    def _embed(self, text: str) -> dict[str, float]:
        tokens = [self._normalize_token(token.lower()) for token in WORD_RE.findall(text)]
        if not tokens:
            return {}
        counts: dict[str, float] = {}
        for token in tokens:
            counts[token] = counts.get(token, 0.0) + 1.0
        norm = math.sqrt(sum(value * value for value in counts.values()))
        if norm == 0:
            return {}
        return {token: value / norm for token, value in counts.items()}

    def _normalize_token(self, token: str) -> str:
        if token.endswith("s") and len(token) > 3:
            return token[:-1]
        return token

    def _cosine_similarity(self, left: dict[str, float], right: dict[str, float]) -> float:
        if not left or not right:
            return 0.0
        if len(left) > len(right):
            left, right = right, left
        return sum(value * right.get(token, 0.0) for token, value in left.items())

    def _save(self) -> None:
        self.store_path.parent.mkdir(parents=True, exist_ok=True)
        payload = [
            {"user_id": item.user_id, "text": item.text, "vector": item.vector}
            for item in self._items
        ]
        self.store_path.write_text(json.dumps(payload), encoding="utf-8")

    def _load(self) -> None:
        if not self.store_path.exists():
            return
        raw = self.store_path.read_text(encoding="utf-8").strip()
        if not raw:
            return
        for item in json.loads(raw):
            self._items.append(
                MemoryItem(
                    user_id=item["user_id"],
                    text=item["text"],
                    vector={k: float(v) for k, v in item["vector"].items()},
                )
            )
