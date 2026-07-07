import json
import subprocess
from dataclasses import dataclass


@dataclass(frozen=True)
class PRMetadata:
    number: int
    title: str
    state: str
    base_ref: str
    head_ref: str
    changed_files: int


def ingest_pr_with_gh(repo: str, pr_number: int) -> PRMetadata:
    cmd = [
        "gh",
        "pr",
        "view",
        str(pr_number),
        "--repo",
        repo,
        "--json",
        "number,title,state,baseRefName,headRefName,changedFiles",
    ]
    result = subprocess.run(cmd, check=True, capture_output=True, text=True)
    payload = json.loads(result.stdout)
    return PRMetadata(
        number=int(payload["number"]),
        title=str(payload["title"]),
        state=str(payload["state"]),
        base_ref=str(payload["baseRefName"]),
        head_ref=str(payload["headRefName"]),
        changed_files=int(payload["changedFiles"]),
    )
