# WebSocket Chat Application using FastAPI and React

----
This project demonstrates a simple WebSocket chat application built with FastAPI for the backend and React for the frontend.
 It allows users to send and receive messages in real-time using WebSockets.
 The application is designed to be lightweight and easy to understand, making it suitable for learning purposes.
----
## Project Structure

```shell
ws-chat-fast/
 ├── app
 │   ├── chat.py
 │   ├── exclusive_chatroom.py
 │   ├── security.py
 │   ├── templating.py
 │   ├── ws_manager.py
 │   ├── ws_password_bearer.py
 ├── templates
 │   ├── chatroom.html
 │   ├── exclusivechatroom.html
 │   ├── login.html
 ├── requirements.txt
``` 


## Features
- **WebSocket Communication**: Real-time messaging between clients.
- **Chat Rooms**: Users can join public or exclusive chat rooms.
- **Authentication**: Basic authentication for exclusive chat rooms.
- **HTML Templating**: Render HTML pages using Jinja2 templates.
- **Static Files**: Serve static files like CSS and JavaScript.
- **Dependency Management**: Use `requirements.txt` for Python dependencies.

