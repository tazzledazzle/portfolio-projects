from ai_app.services.memory import MemoryService


def test_memory_retrieval_prefers_semantic_overlap(tmp_path) -> None:
    store_path = tmp_path / "memory.json"
    memory = MemoryService(store_path=store_path)

    memory.add_turn("u1", "I enjoy hiking in mountains")
    memory.add_turn("u1", "I prefer tea over coffee")

    matches = memory.retrieve("u1", "What mountain trails should I hike?", top_k=1)
    assert len(matches) == 1
    assert "hiking" in matches[0].text


def test_memory_persists_to_disk(tmp_path) -> None:
    store_path = tmp_path / "memory.json"
    memory = MemoryService(store_path=store_path)
    memory.add_turn("u1", "remember this fact")

    reloaded = MemoryService(store_path=store_path)
    matches = reloaded.retrieve("u1", "fact", top_k=1)
    assert len(matches) == 1
    assert matches[0].text == "remember this fact"
