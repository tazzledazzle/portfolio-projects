from pydantic import BaseModel, Field

## todo: add field validations and constraints as needed
class Book(BaseModel):
    id: int
    title: str = Field(..., min_length=1, max_length=200)
    authors: list[str] = Field(..., min_items=1)
    published_year: int = Field(..., ge=0)
    isbn: str = Field(..., regex=r'^\d{3}-\d{1,5}-\d{1,7}-\d{1,7}-\d{1}$')
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

class Order(BaseModel):
    order_id: int
    user_id: int
    order_date: str
    total_amount: float

class OrderWithID(Order):
    id: int

class OrderItems(BaseModel):
    order_id: int
    book_id: int
    quantity: int
