from pydantic import BaseModel


class Book(BaseModel):
    title: str
    authors: list[str]
    published_year: int
    isbn: str
    price: float
    categories: list[str]
    description: str | None = None
    cover_image_url: str | None = None
    rating: float | None = None

class BookWithID(Book):
    id: int
class Inventory(BaseModel):
    book_id: int
    quantity: int

class Orders(BaseModel):
    order_id: int
    user_id: int
    order_date: str
    total_amount: float

class OrderWithID(Orders):
    id: int

class OrderItems(BaseModel):
    order_id: int
    book_id: int
    quantity: int