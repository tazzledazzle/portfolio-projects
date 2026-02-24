"""FastAPI app with OpenAPI."""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(
    title="Items API",
    description="Minimal REST API for showcase (OpenAPI, tests, Docker).",
    version="1.0.0",
)


class ItemCreate(BaseModel):
    name: str
    description: str | None = None


class Item(BaseModel):
    id: int
    name: str
    description: str | None = None


# In-memory store for demo; replace with DB in real use.
_items: dict[int, Item] = {}
_next_id = 1


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.get("/items", response_model=list[Item])
def list_items() -> list[Item]:
    return list(_items.values())


@app.get("/items/{item_id}", response_model=Item)
def get_item(item_id: int) -> Item:
    if item_id not in _items:
        raise HTTPException(status_code=404, detail="Item not found")
    return _items[item_id]


@app.post("/items", response_model=Item, status_code=201)
def create_item(item: ItemCreate) -> Item:
    global _next_id
    obj = Item(id=_next_id, name=item.name, description=item.description)
    _items[_next_id] = obj
    _next_id += 1
    return obj


@app.delete("/items/{item_id}", status_code=204)
def delete_item(item_id: int) -> None:
    if item_id not in _items:
        raise HTTPException(status_code=404, detail="Item not found")
    del _items[item_id]
