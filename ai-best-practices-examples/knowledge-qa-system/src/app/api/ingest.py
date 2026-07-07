from fastapi import APIRouter

from app.domain.services.ingestion_service import IngestionService
from app.schemas.ingest import IngestRequest, IngestResponse

router = APIRouter(prefix="", tags=["ingest"])
ingestion_service = IngestionService()
_JOB_STATUS: dict[str, IngestResponse] = {}


@router.post("/ingest", response_model=IngestResponse)
def ingest(request: IngestRequest) -> IngestResponse:
    result = ingestion_service.ingest(request)
    _JOB_STATUS[result.job_id] = result
    return result


@router.get("/ingest/{job_id}", response_model=IngestResponse)
def ingest_status(job_id: str) -> IngestResponse:
    if job_id not in _JOB_STATUS:
        return IngestResponse(job_id=job_id, documents_ingested=0, chunks_written=0, failures=["job not found"])
    return _JOB_STATUS[job_id]
