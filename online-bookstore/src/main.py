from fastapi import FastAPI, HTTPException
from models import (
    Book,
    BookWithID,
    Inventory,
    Orders,
    OrderWithID,
    OrderItems,
)
app = FastAPI()




# Root endpoint
@app.get("/")
def read_root():
    return {"message": "Welcome to the Online Bookstore API"}






## custom error handler for 404 Not Found
@app.exception_handler(HTTPException)
async def http_exception_handler(request, exc):
    return {
        "status_code": exc.status_code,
        "detail": exc.detail,
        "message": "Resource not found"
    }