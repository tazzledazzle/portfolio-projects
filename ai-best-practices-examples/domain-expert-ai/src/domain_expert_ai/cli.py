import argparse
import os

from domain_expert_ai.data.prepare_dataset import prepare_dataset
from domain_expert_ai.eval.run_eval import run_eval
from domain_expert_ai.inference.serve_ollama import serve_once
from domain_expert_ai.training.train_qlora import run_training


def _env_str(name: str, default: str) -> str:
    return os.getenv(name, default)


def _env_int(name: str, default: int) -> int:
    try:
        return int(os.getenv(name, str(default)))
    except ValueError:
        return default


def _env_float(name: str, default: float) -> float:
    try:
        return float(os.getenv(name, str(default)))
    except ValueError:
        return default


def _env_bool(name: str, default: bool) -> bool:
    value = os.getenv(name)
    if value is None:
        return default
    return value.strip().lower() in {"1", "true", "yes", "on"}


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="domain-expert-ai")
    subparsers = parser.add_subparsers(dest="command", required=True)

    prepare = subparsers.add_parser("prepare-data", help="Prepare train/val dataset files.")
    prepare.add_argument("--input", required=True, help="Path to raw JSONL file.")
    prepare.add_argument("--train-output", required=True, help="Path to train JSONL file.")
    prepare.add_argument("--val-output", required=True, help="Path to val JSONL file.")
    prepare.add_argument("--val-ratio", type=float, default=_env_float("VAL_RATIO", 0.2), help="Validation split ratio.")
    prepare.add_argument("--seed", type=int, default=_env_int("PREPARE_SEED", 42), help="Random seed for split.")
    prepare.add_argument(
        "--quality-report-path",
        default=os.getenv("QUALITY_REPORT_PATH"),
        help="Optional path to write quality filtering report JSON.",
    )
    prepare.add_argument(
        "--enable-augmentation",
        action="store_true",
        default=_env_bool("ENABLE_AUGMENTATION", False),
        help="Enable controlled data augmentation before quality filtering and splitting.",
    )

    train = subparsers.add_parser("train", help="Run QLoRA fine-tuning.")
    train.add_argument("--train-file", required=True, help="Path to processed train JSONL.")
    train.add_argument("--val-file", required=True, help="Path to processed val JSONL.")
    train.add_argument("--output-dir", required=True, help="Directory to write adapters/checkpoints.")
    train.add_argument("--base-model", default=_env_str("HF_MODEL_ID", "TinyLlama/TinyLlama-1.1B-Chat-v1.0"))
    train.add_argument("--profile", choices=["tiny", "balanced", "quality"], default=_env_str("TRAIN_PROFILE", "balanced"))
    train.add_argument("--epochs", type=int, default=_env_int("TRAIN_EPOCHS", None))
    train.add_argument("--batch-size", type=int, default=_env_int("BATCH_SIZE", None))
    train.add_argument("--learning-rate", type=float, default=_env_float("LEARNING_RATE", None))
    train.add_argument("--lora-r", type=int, default=_env_int("LORA_R", None))
    train.add_argument("--lora-alpha", type=int, default=_env_int("LORA_ALPHA", None))
    train.add_argument("--lora-dropout", type=float, default=_env_float("LORA_DROPOUT", None))
    train.add_argument("--max-length", type=int, default=_env_int("MAX_LENGTH", None))
    train.add_argument("--max-steps", type=int, default=_env_int("MAX_STEPS", -1))
    train.add_argument("--grad-accum-steps", type=int, default=_env_int("GRAD_ACCUM_STEPS", 1))
    train.add_argument("--warmup-ratio", type=float, default=_env_float("WARMUP_RATIO", 0.03))
    train.add_argument("--weight-decay", type=float, default=_env_float("WEIGHT_DECAY", 0.0))
    train.add_argument("--logging-steps", type=int, default=_env_int("LOGGING_STEPS", 10))
    train.add_argument("--eval-strategy", choices=["no", "steps", "epoch"], default=_env_str("EVAL_STRATEGY", "epoch"))
    train.add_argument("--save-strategy", choices=["no", "steps", "epoch"], default=_env_str("SAVE_STRATEGY", "epoch"))
    train.add_argument("--seed", type=int, default=_env_int("TRAIN_SEED", 42))
    train.add_argument(
        "--use-4bit",
        choices=["auto", "on", "off"],
        default=_env_str("USE_4BIT", "auto"),
        help="4-bit quantization mode. auto enables on CUDA only.",
    )
    train.add_argument(
        "--target-modules",
        default=_env_str("TARGET_MODULES", ""),
        help="Comma-separated LoRA target module names (e.g. q_proj,v_proj).",
    )
    train.add_argument(
        "--template-strategy",
        choices=["direct", "scenario", "ambiguity", "mixed"],
        default=_env_str("TEMPLATE_STRATEGY", "direct"),
        help="Instruction template strategy for training record formatting.",
    )
    train.add_argument("--dry-run", action="store_true", help="Validate config without running trainer.")

    evaluate = subparsers.add_parser("eval", help="Run baseline vs tuned evaluation.")
    evaluate.add_argument("--eval-file", required=True, help="Path to eval JSONL.")
    evaluate.add_argument("--report-path", required=True, help="Path to write report JSON.")
    evaluate.add_argument("--baseline-model", default=_env_str("BASELINE_MODEL", "llama3.1"))
    evaluate.add_argument("--tuned-model", default=_env_str("TUNED_MODEL", "domain-expert-ai"))

    serve = subparsers.add_parser("serve", help="Run one-shot Ollama inference.")
    serve.add_argument("--prompt", required=True, help="User prompt.")
    serve.add_argument("--model", default="domain-expert-ai")

    return parser


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()

    if args.command == "prepare-data":
        if args.val_ratio <= 0 or args.val_ratio >= 1:
            raise ValueError("--val-ratio must be between 0 and 1 (exclusive).")
        prepare_dataset(
            args.input,
            args.train_output,
            args.val_output,
            args.val_ratio,
            args.seed,
            quality_report_path=args.quality_report_path,
            enable_augmentation=args.enable_augmentation,
        )
    elif args.command == "train":
        run_training(args)
    elif args.command == "eval":
        run_eval(args.eval_file, args.report_path, args.baseline_model, args.tuned_model)
    elif args.command == "serve":
        response = serve_once(args.prompt, args.model)
        print(response)


if __name__ == "__main__":
    main()

