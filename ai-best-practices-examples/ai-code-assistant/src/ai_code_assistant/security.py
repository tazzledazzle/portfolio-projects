from dataclasses import dataclass


@dataclass(frozen=True)
class ExecutionPolicy:
    profile: str

    @property
    def can_write(self) -> bool:
        return self.profile in {"workspace-write", "full-access"}


def build_policy(profile: str) -> ExecutionPolicy:
    return ExecutionPolicy(profile=profile)
