import argparse
import json
from pathlib import Path

from ai_code_assistant.adapters.llm_adapter import LLMAdapter
from ai_code_assistant.audit import append_audit_event
from ai_code_assistant.automation import load_plan, run_steps
from ai_code_assistant.extensions import validate_manifest
from ai_code_assistant.github_ingest import ingest_pr_with_gh
from ai_code_assistant.policy import load_policy
from ai_code_assistant.redaction import redact_mapping
from ai_code_assistant.risk import evaluate_risk, score_write_action
from ai_code_assistant.services.repo_scanner import scan_python_files
from ai_code_assistant.services.test_generator import generate_pyramid_for_file
from ai_code_assistant.security import build_policy


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="ai-code-assistant")
    subparsers = parser.add_subparsers(dest="command", required=True)

    gen_tests = subparsers.add_parser("gen-tests", help="Generate pytest files for Python modules.")
    gen_tests.add_argument("file_path", nargs="?", help="Single Python file to generate tests for.")
    gen_tests.add_argument("--repo", help="Repository root to scan and generate tests for.")
    gen_tests.add_argument("--dry-run", action="store_true", help="Print outputs instead of writing files.")
    gen_tests.add_argument(
        "--profile",
        choices=["read-only", "workspace-write", "full-access"],
        default="workspace-write",
        help="Execution profile controlling mutation permissions.",
    )
    gen_tests.add_argument(
        "--output",
        choices=["text", "json"],
        default="text",
        help="Output format. Use json for CI/headless automation.",
    )
    gen_tests.add_argument(
        "--audit-log",
        default=".ai-code-assistant/audit.log.jsonl",
        help="Path to local JSONL audit log file.",
    )
    gen_tests.add_argument("--policy-file", help="Optional assistant-policy.toml path.")
    gen_tests.add_argument("--approve-high-risk", action="store_true", help="Allow high-risk writes.")
    gen_tests.add_argument("--headless", action="store_true", help="CI mode: implies --output json.")
    gen_tests.add_argument(
        "--pyramid",
        choices=["unit", "integration", "e2e", "all"],
        default="unit",
        help="Generate one level or full test pyramid.",
    )
    gen_tests.add_argument("--pr-repo", help="GitHub repo owner/name for PR ingestion.")
    gen_tests.add_argument("--pr-number", type=int, help="GitHub PR number for metadata enrichment.")

    ext = subparsers.add_parser("extensions", help="Extension operations.")
    ext_sub = ext.add_subparsers(dest="extensions_command", required=True)
    validate = ext_sub.add_parser("validate-manifest", help="Validate extension manifest.")
    validate.add_argument("--manifest", required=True, help="Path to extension-manifest.v1.json")
    validate.add_argument("--output", choices=["text", "json"], default="text")

    run_plan = subparsers.add_parser("run-plan", help="Run checkpointed automation plan.")
    run_plan.add_argument("--plan", required=True, help="Path to plan JSON file.")
    run_plan.add_argument("--checkpoint-dir", default=".ai-code-assistant/checkpoints")
    run_plan.add_argument("--parallel", action="store_true")
    run_plan.add_argument("--output", choices=["text", "json"], default="text")
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)

    if args.command == "gen-tests":
        return _run_gen_tests(args=args, parser=parser)
    if args.command == "extensions":
        return _run_extensions(args)
    if args.command == "run-plan":
        return _run_plan(args)
    return 1


def _run_gen_tests(args: argparse.Namespace, parser: argparse.ArgumentParser) -> int:
    if bool(args.file_path) == bool(args.repo):
        parser.error("Provide exactly one of: file_path or --repo.")

    policy = build_policy(args.profile)
    if not args.dry_run and not policy.can_write:
        parser.error("Profile 'read-only' cannot write files. Re-run with --dry-run.")

    if args.headless:
        args.output = "json"

    policy_data, policy_source = load_policy(args.policy_file)
    adapter = LLMAdapter()
    results: list[dict[str, str | bool]] = []
    audit_log_path = Path(args.audit_log).resolve()
    pr_metadata = None
    if args.pr_repo and args.pr_number:
        pr_metadata = ingest_pr_with_gh(args.pr_repo, args.pr_number)

    if args.file_path:
        source_path = Path(args.file_path).resolve()
        if not source_path.exists() or source_path.suffix != ".py":
            parser.error("file_path must be an existing .py file.")
        repo_root = Path.cwd().resolve()
        generated_tests = _generate_tests_for_levels(
            source_path=source_path,
            repo_root=repo_root,
            adapter=adapter,
            pyramid=args.pyramid,
        )
        for generated in generated_tests:
            results.append(
                _output_one(
                    generated.target_path,
                    generated.content,
                    dry_run=args.dry_run,
                    output_mode=args.output,
                    audit_log_path=audit_log_path,
                    source_path=source_path,
                    profile=args.profile,
                    policy=policy_data,
                    approve_high_risk=args.approve_high_risk,
                )
            )
        _finalize_output(
            output_mode=args.output,
            processed_count=len(generated_tests),
            dry_run=args.dry_run,
            profile=args.profile,
            results=results,
            audit_log_path=audit_log_path,
            policy_source=policy_source,
            pr_metadata=pr_metadata,
            redaction_patterns=policy_data.redaction_patterns,
            redaction_enabled=policy_data.redaction_enabled,
        )
        return 0

    repo_root = Path(args.repo).resolve()
    if not repo_root.exists() or not repo_root.is_dir():
        parser.error("--repo must be an existing directory.")

    files = scan_python_files(repo_path=repo_root)
    for source_path in files:
        generated_tests = _generate_tests_for_levels(
            source_path=source_path,
            repo_root=repo_root,
            adapter=adapter,
            pyramid=args.pyramid,
        )
        for generated in generated_tests:
            results.append(
                _output_one(
                    generated.target_path,
                    generated.content,
                    dry_run=args.dry_run,
                    output_mode=args.output,
                    audit_log_path=audit_log_path,
                    source_path=source_path,
                    profile=args.profile,
                    policy=policy_data,
                    approve_high_risk=args.approve_high_risk,
                )
            )

    _finalize_output(
        output_mode=args.output,
        processed_count=len(results),
        dry_run=args.dry_run,
        profile=args.profile,
        results=results,
        audit_log_path=audit_log_path,
        policy_source=policy_source,
        pr_metadata=pr_metadata,
        redaction_patterns=policy_data.redaction_patterns,
        redaction_enabled=policy_data.redaction_enabled,
    )
    return 0


def _generate_tests_for_levels(source_path: Path, repo_root: Path, adapter: LLMAdapter, pyramid: str):
    if pyramid == "all":
        return generate_pyramid_for_file(source_path, repo_root, adapter, ("unit", "integration", "e2e"))
    return generate_pyramid_for_file(source_path, repo_root, adapter, (pyramid,))


def _output_one(
    target_path: Path,
    content: str,
    dry_run: bool,
    output_mode: str,
    audit_log_path: Path,
    source_path: Path,
    profile: str,
    policy,
    approve_high_risk: bool,
) -> dict[str, str | bool]:
    score, reasons = score_write_action(str(target_path), profile=profile, dry_run=dry_run)
    risk = evaluate_risk(
        score=score,
        reasons=reasons,
        auto_allow_max=policy.auto_allow_max,
        approval_required_min=policy.approval_required_min,
        hard_block_min=policy.hard_block_min,
    )
    if risk.blocked and not approve_high_risk:
        raise PermissionError("Blocked high-risk write. Re-run with --approve-high-risk to override.")

    action = "previewed" if dry_run else "wrote"
    append_audit_event(
        audit_log_path,
        {
            "event_type": "file_result",
            "action": action,
            "source_path": str(source_path),
            "target_path": str(target_path),
            "output_mode": output_mode,
            "risk_score": risk.score,
            "risk_reasons": risk.reasons,
            "approval_required": risk.approval_required,
        },
    )
    if dry_run:
        if output_mode == "text":
            print(f"--- {target_path} ---")
            print(content)
        return {"target_path": str(target_path), "action": action, "content": content}

    target_path.parent.mkdir(parents=True, exist_ok=True)
    target_path.write_text(content, encoding="utf-8")
    if output_mode == "text":
        print(f"Wrote {target_path}")
    return {"target_path": str(target_path), "action": action, "content": content}


def _finalize_output(
    output_mode: str,
    processed_count: int,
    dry_run: bool,
    profile: str,
    results: list[dict[str, str | bool]],
    audit_log_path: Path,
    policy_source: str,
    pr_metadata,
    redaction_patterns: list[str],
    redaction_enabled: bool,
) -> None:
    append_audit_event(
        audit_log_path,
        {
            "event_type": "run_summary",
            "profile": profile,
            "dry_run": dry_run,
            "processed_count": processed_count,
            "output_mode": output_mode,
            "policy_source": policy_source,
        },
    )
    safe_results = results
    if redaction_enabled:
        safe_results = [redact_mapping(result, redaction_patterns) for result in results]
    if output_mode == "json":
        payload = {
            "status": "ok",
            "profile": profile,
            "dry_run": dry_run,
            "processed_count": processed_count,
            "policy_source": policy_source,
            "results": safe_results,
        }
        if pr_metadata is not None:
            payload["pr_metadata"] = {
                "number": pr_metadata.number,
                "title": pr_metadata.title,
                "state": pr_metadata.state,
                "base_ref": pr_metadata.base_ref,
                "head_ref": pr_metadata.head_ref,
                "changed_files": pr_metadata.changed_files,
            }
        print(json.dumps(payload))
        return

    print(f"Processed {processed_count} file(s).")


def _run_extensions(args: argparse.Namespace) -> int:
    if args.extensions_command == "validate-manifest":
        payload = validate_manifest(Path(args.manifest).resolve())
        if args.output == "json":
            print(json.dumps({"status": "ok", "manifest": payload}))
        else:
            print(f"Manifest valid: {payload['name']}@{payload['version']}")
        return 0
    return 1


def _run_plan(args: argparse.Namespace) -> int:
    steps = load_plan(Path(args.plan).resolve())
    results = run_steps(steps=steps, checkpoint_dir=Path(args.checkpoint_dir).resolve(), parallel=args.parallel)
    if args.output == "json":
        print(json.dumps({"status": "ok", "steps": results}))
    else:
        print(f"Executed {len(results)} step(s).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
