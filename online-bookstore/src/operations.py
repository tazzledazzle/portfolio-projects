import csv
from typing import Optional
from models import BookWithID, Order, OrderWithID

import os

# Get the directory of this file to find CSV files relative to it
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
BOOK_DATABASE_FILENAME = os.path.join(BASE_DIR, 'books.csv')
ORDER_DATABASE_FILENAME = os.path.join(BASE_DIR, 'orders.csv')
book_column_fields = [
    'id', 'title', 'authors', 'published_year', 'isbn', 'price', 'categories',
    'description', 'cover_image_url', 'rating'
]
order_column_fields = [
    'id', 'customer_name', 'customer_email', 'book_id', 'quantity'
]

def read_all_books() -> list[BookWithID]:
    """Read all books from the CSV file."""
    books = []
    try:
        with open(BOOK_DATABASE_FILENAME, mode='r', newline='', encoding='utf-8') as file:
            reader = csv.DictReader(file)  # Let it read headers automatically
            for row in reader:
                # Convert string fields to appropriate types
                processed_row = {
                    'id': int(row['id']),
                    'title': row['title'],
                    'authors': [row['authors']],  # Convert to list
                    'published_year': int(row['published_year']),
                    'isbn': row['isbn'],
                    'price': float(row['price']),
                    'categories': [row['categories']],  # Convert to list
                    'description': row['description'],
                    'cover_image_url': row['cover_image_url'],
                    'rating': float(row['rating']) if row['rating'] else None
                }
                book = BookWithID(**processed_row)
                books.append(book)
    except FileNotFoundError:
        print(f"Database file {BOOK_DATABASE_FILENAME} not found.")
    return books

def read_all_orders() -> list[OrderWithID]:
    """Read all orders from the CSV file."""
    orders = []
    try:
        with open(ORDER_DATABASE_FILENAME, mode='r', newline='', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            for row in reader:
                processed_row = {
                    'id': int(row['id']),
                    'customer_name': row['customer_name'],
                    'customer_email': row['customer_email'],
                    'book_id': int(row['book_id']),
                    'quantity': int(row['quantity'])
                }
                order = OrderWithID(**processed_row)
                orders.append(order)
    except FileNotFoundError:
        print(f"Database file {ORDER_DATABASE_FILENAME} not found.")
    return orders

def read_order(order_id: int) -> Optional[OrderWithID]:
    """Read an order by its ID from the CSV file."""
    try:
        with open(ORDER_DATABASE_FILENAME, mode='r', newline='', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            for row in reader:
                if int(row['id']) == order_id:
                    processed_row = {
                        'id': int(row['id']),
                        'customer_name': row['customer_name'],
                        'customer_email': row['customer_email'],
                        'book_id': int(row['book_id']),
                        'quantity': int(row['quantity'])
                    }
                    return OrderWithID(**processed_row)
    except FileNotFoundError:
        print(f"Database file {ORDER_DATABASE_FILENAME} not found.")
    return None

def read_book_by_id(book_id: int) -> Optional[BookWithID]:
    """Read a book by its ID from the CSV file."""
    try:
        with open(BOOK_DATABASE_FILENAME, mode='r', newline='', encoding='utf-8') as file:
            reader = csv.DictReader(file)  # Let it read headers automatically
            for row in reader:
                if int(row['id']) == book_id:
                    # Convert string fields to appropriate types
                    processed_row = {
                        'id': int(row['id']),
                        'title': row['title'],
                        'authors': [row['authors']],  # Convert to list
                        'published_year': int(row['published_year']),
                        'isbn': row['isbn'],
                        'price': float(row['price']),
                        'categories': [row['categories']],  # Convert to list
                        'description': row['description'],
                        'cover_image_url': row['cover_image_url'],
                        'rating': float(row['rating']) if row['rating'] else None
                    }
                    return BookWithID(**processed_row)
    except FileNotFoundError:
        print(f"Database file {BOOK_DATABASE_FILENAME} not found.")
    return None

def get_next_id():
    try:
        with open(ORDER_DATABASE_FILENAME, mode='r', newline='', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            rows = list(reader)
            if rows:
                return int(rows[-1]['id']) + 1
            else:
                return 1
    except FileNotFoundError:
        return 1
    
def write_order_into_csv(order: OrderWithID):
    """Write an order into the CSV file."""
    with open(ORDER_DATABASE_FILENAME, mode='a',
               newline='', encoding='utf-8') as file:
        writer = csv.DictWriter(
            file,
            fieldnames=order_column_fields
        )
        writer.writerow(order.model_dump())

def create_order(order: Order) -> OrderWithID:
    """Create a new order and write it to the CSV file."""
    next_id = get_next_id()
    order_with_id = OrderWithID(
        id=next_id,
        **order.model_dump()
    )
    write_order_into_csv(order_with_id)
    return order_with_id

def remove_order(order_id: int) -> bool:
    """Remove an order by its ID from the CSV file."""
    try:
        orders = read_all_orders()

        with open(ORDER_DATABASE_FILENAME, mode='w', newline='', encoding='utf-8') as file:
            writer = csv.DictWriter(file, fieldnames=orders[0].keys())
            writer.writeheader()
            for order in orders:
                if int(order['id']) != order_id:
                    writer.writerow(order)

        return True
    except FileNotFoundError:
        print(f"Database file {ORDER_DATABASE_FILENAME} not found.")
        return False
