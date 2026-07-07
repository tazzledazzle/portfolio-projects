from pathlib import Path

from pydantic import BaseModel, Field


class ProjectPaths(BaseModel):
    root: Path = Field(default_factory=lambda: Path(__file__).resolve().parents[2])

    @property
    def raw_data_dir(self) -> Path:
        return self.root / "data" / "raw"

    @property
    def processed_data_dir(self) -> Path:
        return self.root / "data" / "processed"

    @property
    def outputs_dir(self) -> Path:
        return self.root / "outputs"

    @property
    def reports_dir(self) -> Path:
        return self.root / "reports"

    @property
    def checkpoints_dir(self) -> Path:
        return self.root / "checkpoints"

