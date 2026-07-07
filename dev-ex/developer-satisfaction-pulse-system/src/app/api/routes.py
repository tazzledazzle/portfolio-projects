from fastapi import APIRouter

from app.services.scoring import rolling_nps

router = APIRouter()


@router.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@router.get("/metrics/nps")
def nps_preview() -> dict[str, float]:
    sample_scores = [9, 10, 8, 7, 6, 10]
    return {"rolling_nps": rolling_nps(sample_scores)}
