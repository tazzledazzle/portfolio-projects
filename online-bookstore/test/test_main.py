import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))
from main import app
from fastapi.testclient import TestClient

client = TestClient(app)
def test_read_root():
    response = client.get("/")
    assert response.status_code == 200
    assert response.json() == {"message": "Welcome to the Online Bookstore API"}

def test_get_books():
    response = client.get("/v1/books")
    assert response.status_code == 200
    books = response.json()
    assert isinstance(books, list)
    assert len(books) > 0
    for book in books:
        assert "id" in book
        assert "title" in book
        assert "authors" in book
        assert "published_year" in book
        assert "isbn" in book
        assert "price" in book
        assert "categories" in book
        assert "description" in book
        assert "cover_image_url" in book
        assert "rating" in book

def test_get_book():
    response = client.get("/v1/books/1")
    assert response.status_code == 200
    book = response.json()
    assert "id" in book
    assert "title" in book
    assert "authors" in book
    assert "published_year" in book
    assert "isbn" in book
    assert "price" in book
    assert "categories" in book
    assert "description" in book
    assert "cover_image_url" in book
    assert "rating" in book

def test_get_non_existent_book():
    response = client.get("/v1/books/9999")
    assert response.status_code == 404
    response_data = response.json()
    assert response_data["detail"] == "Book not found"

def test_create_order():
    order_data = {
        "customer_name": "John Doe",
        "customer_email": "john.doe@example.com",
        "book_id": 1,
        "quantity": 2
    }
    response = client.post("/v1/orders", json=order_data)
    assert response.status_code == 200
    order = response.json()
    assert "id" in order
    assert order["customer_name"] == "John Doe"
    assert order["customer_email"] == "john.doe@example.com"
    assert order["book_id"] == 1
    assert order["quantity"] == 2

def test_get_order():
    response = client.get("/v1/orders/1")
    assert response.status_code == 200
    order = response.json()
    assert "id" in order
    assert "customer_name" in order
    assert "customer_email" in order
    assert "book_id" in order
    assert "quantity" in order