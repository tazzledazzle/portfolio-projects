# Phase 3 Verification: AI Workflows

**Verified:** 2026-05-20  
**Environment:** macOS, Temporal + WireMock LLM stub via Docker Compose

## Requirements

| ID | Status | Evidence |
|----|--------|----------|
| WF-01 | **PASS** | `starter rag` → answer + citation from WireMock stub |
| WF-02 | **PASS** | `starter agent` → summary with tool_calls=1; retry policy configured (max 3) |
| WF-03 | **PASS** | `starter batch demo-eval 5` → item_count=5, mean_score computed |
| WF-04 | **PASS** | Starter commands: `rag`, `agent`, `batch` (+ legacy `ping`) |

## Unit tests

```bash
./gradlew :workflows:test
```

- `RagQaWorkflowTest`
- `AgentToolsWorkflowTest`
- `BatchEvalWorkflowTest`
- `PingWorkflowTest`

## E2E commands (sample output)

```text
workflow_id=rag-...
answer=Stub LLM answer grounded in the provided architecture and operations documentation.
citation=fixtures/chunks.json#chunk-0

workflow_id=agent-...
summary=Synthesis for 'test goal' using tool-result:search_docs:test goal
tool_calls=1

workflow_id=batch-...
item_count=5
mean_score=0.3764705882352941
```

## Agent retry simulation

Set `SIMULATE_TOOL_FAILURE=true` on the **worker** process, then run `starter agent`. Temporal UI shows `callTool` retry attempts (max 3 per workflow retry policy).

## Artifacts

| Component | Path |
|-----------|------|
| RAG workflow | `workflows/.../RagQaWorkflow*.kt` |
| Agent workflow | `workflows/.../AgentToolsWorkflow*.kt` |
| Batch workflow | `workflows/.../BatchEvalWorkflow*.kt` |
| Activity impls | `worker/.../RagActivitiesImpl.kt`, etc. |
| LLM stub | `deploy/llm-stub/`, Compose service `llm-stub:8090` |
| Fixtures | `worker/src/main/resources/fixtures/chunks.json` |
| E2E script | `scripts/ai-workflows-e2e.sh` |
| Kotlin Jackson converter | `workflows/.../TemporalDataConverter.kt` |

**Phase 3 complete.** Next: `/gsd-execute-phase 4` (OpenTelemetry instrumentation).
