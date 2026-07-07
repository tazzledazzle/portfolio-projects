from fastapi import FastAPI

from ws_chat.chat import router as chat_router
from ws_chat.exclusive_chatroom import router as exclusive_router
from ws_chat.security import router as security_router

app = FastAPI(title="ws-chat-fast")
app.include_router(chat_router)
app.include_router(security_router)
app.include_router(exclusive_router)
