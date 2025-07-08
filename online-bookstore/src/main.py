from fastapi import FastAPI, HTTPException
from models import (
    Book,
    BookWithID,
    Inventory,
    Order,
    OrderWithID,
    OrderItems,
)
from operations import read_all_books, read_order, create_order

app = FastAPI()




# Root endpoint
@app.get("/")
def read_root():
    return {"message": "Welcome to the Online Bookstore API"}


@app.get("/v1/books", response_model=list[BookWithID])
def get_books():
    """Get all books."""
    books = read_all_books()
    if not books:
        raise HTTPException(status_code=404, detail="No books found")
    return books

@app.get("/v1/books/{book_id}", response_model=BookWithID)
def get_book(book_id: int):
    """Get a book by ID."""
    book = read_order(book_id)
    if not book:
        raise HTTPException(status_code=404, detail="Book not found")
    return book


@app.post("/v1/orders", response_model=OrderWithID)
def create_order(order: Order):
    """Create a new order."""
    return create_order(order)

@app.get("/v1/orders/{order_id}", response_model=OrderWithID)
def get_order(order_id: int):
    """Get an order by ID."""
    order = read_order(order_id)
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    return order

@app.get("/v1/search", response_model=list[BookWithID])
def search_books(query: str):
    """Search for books by title or author."""
    books = read_all_books()
    results = [book for book in books if query.lower() in book.title.lower() or query.lower() in book.authors.lower()]
    if not results:
        raise HTTPException(status_code=404, detail="No books found")
    return results


## custom error handler for 404 Not Found
@app.exception_handler(HTTPException)
async def http_exception_handler(request, exc):
    return {
        "status_code": exc.status_code,
        "detail": exc.detail,
        "message": "Resource not found"
    }