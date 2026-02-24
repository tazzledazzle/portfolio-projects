"""
Worker: pop job IDs from Redis, load payload, process, write result to PostgreSQL.
"""
import os
import time
import redis.asyncio as redis
import asyncpg

REDIS_URL = os.environ.get("REDIS_URL", "redis://redis:6379")
DATABASE_URL = os.environ.get("DATABASE_URL", "postgresql://postgres:postgres@db:5432/workflow")
QUEUE = "workflow:queue"


async def main():
    r = redis.from_url(REDIS_URL)
    pool = await asyncpg.create_pool(DATABASE_URL, min_size=1, max_size=2)
    while True:
        result = await r.brpop(QUEUE, timeout=5)
        if result is None:
            continue
        _, job_id = result
        job_id = job_id.decode() if isinstance(job_id, bytes) else job_id
        payload = await r.get(f"workflow:payload:{job_id}")
        payload = (payload or b"").decode() if isinstance(payload, bytes) else (payload or "")
        # "Process": sleep briefly and set result
        time.sleep(0.5)
        result = f"Processed: {payload}"
        async with pool.acquire() as conn:
            await conn.execute(
                "UPDATE jobs SET status = $1, result = $2 WHERE id = $3",
                "completed", result, job_id
            )
        print(f"Job {job_id} completed: {result}")


if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
