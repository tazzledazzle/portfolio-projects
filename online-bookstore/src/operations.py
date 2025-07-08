import csv
from typing import Optional
from models import Book, BookWithID, Order, OrderWithID

BOOK_DATABASE_FILENAME = 'books.csv'
ORDER_DATABASE_FILENAME = 'orders.csv'
column_fields = [
    'id', 'title', 'authors', 'published_year', 'isbn', 'price', 'categories',
    'description', 'cover_image_url', 'rating'
]

def read_all_books() -> list[BookWithID]:
    """Read all books from the CSV file."""
    books = []
    try:
        with open(BOOK_DATABASE_FILENAME, mode='r', newline='', encoding='utf-8') as file:
            reader = csv.DictReader(file, fieldnames=column_fields)
            for row in reader:
                book = BookWithID(**row)
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
                order = OrderWithID(**row)
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
                    return OrderWithID(**row)
    except FileNotFoundError:
        print(f"Database file {ORDER_DATABASE_FILENAME} not found.")
    return None

def read_book_by_id(book_id: int) -> Optional[BookWithID]:
    """Read a book by its ID from the CSV file."""
    try:
        with open(BOOK_DATABASE_FILENAME, mode='r', newline='', encoding='utf-8') as file:
            reader = csv.DictReader(file, fieldnames=column_fields)
            for row in reader:
                if int(row['id']) == book_id:
                    return BookWithID(**row)
    except FileNotFoundError:
        print(f"Database file {BOOK_DATABASE_FILENAME} not found.")
    return None

def get_next_id():
    try:
        with open(ORDER_DATABASE_FILENAME, mode='r', newline='', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            last_row = list(reader)[-1] if list(reader) else None
            return int(last_row['id']) + 1 if last_row else 1
    except FileNotFoundError:
        return 1
    
def write_order_into_csv(order: OrderWithID):
    """Write an order into the CSV file."""
    with open(ORDER_DATABASE_FILENAME, mode='a',
               newline='', encoding='utf-8') as file:
        writer = csv.DictWriter(
            file,
            fieldnames=column_fields
        )
        writer.writerow(order.model_dump())

def create_order(order: Order) -> OrderWithID:
    """Create a new order and write it to the CSV file."""
    order.id = get_next_id()
    order_with_id = OrderWithID(
        id=order.id,
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
