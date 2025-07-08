
import csv
import os
from pathlib import Path
from unittest.mock import patch
import pytest

TEST_DATABASE_FILENAME = 'test_books.csv'
TEST_ORDER_DATABASE_FILENAME = 'test_orders.csv'

TEST_BOOKS_CSV = [
    {
        "id": 1,
        "title": "Test Book 1",
        "authors": "Author A",
        "published_year": "2021",
        "isbn": "1234567890",
        "price": 19.99,
        "categories": "Fiction",
        "description": "A test book for unit testing.",
        "cover_image_url": "http://example.com/test_book_1.jpg",
        "rating": 4.5
    },
    {
        "id": 2,
        "title": "Test Book 2",
        "authors": "Author B",
        "published_year": "2022",
        "isbn": "0987654321",
        "price": 29.99,
        "categories": "Non-Fiction",
        "description": "Another test book for unit testing.",
        "cover_image_url": "http://example.com/test_book_2.jpg",
        "rating": 4.0
    }
]

TEST_BOOKS = [
    { **book_json, "id": int(book_json["id"]) }
    for book_json in TEST_BOOKS_CSV
]


@pytest.fixture(autouse=True)
def create_test_database():
    database_file_location = str(Path(__file__).parent / TEST_DATABASE_FILENAME)
    with patch(
        "operations.BOOK_DATABASE_FILENAME",
        database_file_location
    ) as csv_test:
        with open(
            database_file_location,
            mode='w',
            newline='',
            encoding='utf-8'
        ) as csvfile:
            writer = csv.DictWriter(
                csvfile,
                fieldnames=[
                    "id",
                    "title",
                    "authors",
                    "published_year",
                    "isbn",
                    "price",
                    "categories",
                    "description",
                    "cover_image_url",
                    "rating"
                ],
            )
            writer.writeheader()
            writer.writerows(TEST_BOOKS_CSV)
            print("")
        yield csv_test
        os.remove(database_file_location)