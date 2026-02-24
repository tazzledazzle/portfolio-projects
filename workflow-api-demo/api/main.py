"""
API: enqueue jobs to Redis, read job status/result from PostgreSQL.
"""
import os
import uuid
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import redis.asyncio as redis
import asyncpg

app = FastAPI(title="Workflow API Demo")

REDIS_URL = os.environ.get("REDIS_URL", "redis://redis:6379")
DATABASE_URL = os.environ.get("DATABASE_URL", "postgresql://postgres:postgres@db:5432/workflow")

db_pool = None
redis_client = None

async def get_redis():
    global redis_client
    if redis_client is None:
        redis_client = redis.from_url(REDIS_URL)
    return redis_client

async def get_db():
    global db_pool
    if db_pool is None:
        db_pool = await asyncpg.create_pool(DATABASE_URL, min_size=1, max_size=5)
    return db_pool


class JobCreate(BaseModel):
    payload: str


class JobResponse(BaseModel):
    id: str
    status: str
    result: str | None = None


@app.on_event("startup")
async def startup():
    await get_db()
    pool = await get_db()
    async with pool.acquire() as conn:
        await conn.execute("""
            CREATE TABLE IF NOT EXISTS jobs (
                id TEXT PRIMARY KEY,
                status TEXT NOT NULL,
                result TEXT,
                created_at TIMESTAMPTZ DEFAULT NOW()
            )
        """)


@app.post("/jobs", response_model=JobResponse)
async def create_job(job: JobCreate):
    job_id = str(uuid.uuid4())
    r = await get_redis()
    await r.lpush("workflow:queue", job_id)
    await r.set(f"workflow:payload:{job_id}", job.payload)
    pool = await get_db()
    async with pool.acquire() as conn:
        await conn.execute(
            "INSERT INTO jobs (id, status) VALUES ($1, $2)",
            job_id, "queued"
        )
    return JobResponse(id=job_id, status="queued")


@app.get("/jobs/{job_id}", response_model=JobResponse)
async def get_job(job_id: str):
    pool = await get_db()
    async with pool.acquire() as conn:
        row = await conn.fetchrow("SELECT id, status, result FROM jobs WHERE id = $1", job_id)
    if not row:
        raise HTTPException(status_code=404, detail="Job not found")
    return JobResponse(id=row["id"], status=row["status"], result=row["result"])


@app.get("/health")
def health():
    return {"status": "ok"}
