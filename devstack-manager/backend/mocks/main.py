from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from .services import docker_control

app = FastAPI()

origins = ["http://localhost:5173"]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/api/health")
def health_check():
    return {"status": "ok"}

@app.get("/api/services")
def list_services():
    return docker_control.list_containers()

@app.post("/api/services/{name}/start")
def start_service(name: str):
    return docker_control.start_container(name)

@app.post("/api/services/{name}/stop")
def stop_service(name: str):
    return docker_control.stop_container(name)