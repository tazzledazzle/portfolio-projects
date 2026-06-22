import json
import random
from argparse import Namespace
from pathlib import Path

from domain_expert_ai.prompting.templates import TEMPLATE_RENDERERS

TRAINING_PROFILES = {
    "tiny": {
        "epochs": 1,
        "batch_size": 1,
        "learning_rate": 3e-4,
        "lora_r": 4,
        "lora_alpha": 8,
        "lora_dropout": 0.05,
        "max_length": 192,
    },
    "balanced": {
        "epochs": 2,
        "batch_size": 2,
        "learning_rate": 2e-4,
        "lora_r": 8,
        "lora_alpha": 16,
        "lora_dropout": 0.05,
        "max_length": 512,
    },
    "quality": {
        "epochs": 3,
        "batch_size": 2,
        "learning_rate": 1e-4,
        "lora_r": 16,
        "lora_alpha": 32,
        "lora_dropout": 0.1,
        "max_length": 768,
    },
}


def _load_jsonl(path: str) -> list[dict]:
    with Path(path).open("r", encoding="utf-8") as file:
        return [json.loads(line) for line in file if line.strip()]


def _format_instruction(row: dict, template_strategy: str = "direct", rng: random.Random | None = None) -> str:
    if template_strategy == "mixed":
        if rng is None:
            rng = random.Random(0)
        chosen = rng.choice(["direct", "scenario", "ambiguity"])
    else:
        chosen = template_strategy
    if chosen not in TEMPLATE_RENDERERS:
        raise ValueError(f"Unsupported template strategy: {template_strategy}")
    return TEMPLATE_RENDERERS[chosen](row)


def _select_target_modules(model) -> list[str]:
    preferred = {"q_proj", "k_proj", "v_proj", "o_proj", "gate_proj", "up_proj", "down_proj"}
    present = {name.split(".")[-1] for name, _ in model.named_modules()}
    selected = sorted(preferred.intersection(present))
    return selected if selected else ["q_proj", "v_proj"]


def _resolve_training_profile(args: Namespace) -> dict:
    profile_name = getattr(args, "profile", "balanced")
    profile = TRAINING_PROFILES[profile_name].copy()
    for field in ["epochs", "batch_size", "learning_rate", "lora_r", "lora_alpha", "lora_dropout", "max_length"]:
        value = getattr(args, field, None)
        if value is not None:
            profile[field] = value
    return profile


def _resolve_4bit_mode(mode: str, has_cuda: bool) -> bool:
    if mode == "on":
        return True
    if mode == "off":
        return False
    return has_cuda


def _resolve_template_strategy(args: Namespace) -> tuple[str, random.Random]:
    strategy = getattr(args, "template_strategy", None) or "direct"
    if strategy not in {"direct", "scenario", "ambiguity", "mixed"}:
        raise ValueError("template_strategy must be one of: direct, scenario, ambiguity, mixed.")
    seed = getattr(args, "seed", 42)
    return strategy, random.Random(seed)


def run_training(args: Namespace) -> dict:
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    dry_run = bool(getattr(args, "dry_run", False))
    profile = _resolve_training_profile(args)
    template_strategy, formatter_rng = _resolve_template_strategy(args)

    metadata = {
        "profile": getattr(args, "profile", "balanced"),
        "base_model": args.base_model,
        "train_file": args.train_file,
        "val_file": args.val_file,
        "epochs": profile["epochs"],
        "batch_size": profile["batch_size"],
        "learning_rate": profile["learning_rate"],
        "lora": {
            "r": profile["lora_r"],
            "alpha": profile["lora_alpha"],
            "dropout": profile["lora_dropout"],
        },
        "max_length": profile["max_length"],
        "max_steps": args.max_steps,
        "grad_accum_steps": args.grad_accum_steps,
        "warmup_ratio": args.warmup_ratio,
        "weight_decay": args.weight_decay,
        "logging_steps": args.logging_steps,
        "eval_strategy": args.eval_strategy,
        "save_strategy": args.save_strategy,
        "seed": args.seed,
        "use_4bit": args.use_4bit,
        "target_modules": args.target_modules,
        "dry_run": dry_run,
        "template_strategy": template_strategy,
    }

    if dry_run:
        metadata["status"] = "dry_run"
        (output_dir / "training_config.json").write_text(json.dumps(metadata, indent=2), encoding="utf-8")
        return metadata

    import torch
    from datasets import Dataset
    from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
    from transformers import (
        AutoModelForCausalLM,
        AutoTokenizer,
        BitsAndBytesConfig,
        DataCollatorForLanguageModeling,
        Trainer,
        TrainingArguments,
    )

    train_rows = _load_jsonl(args.train_file)
    val_rows = _load_jsonl(args.val_file)
    if not train_rows or not val_rows:
        raise ValueError("Both train and val files must include at least one row.")

    tokenizer = AutoTokenizer.from_pretrained(args.base_model, use_fast=True)
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token

    has_cuda = torch.cuda.is_available()
    use_4bit = _resolve_4bit_mode(args.use_4bit, has_cuda)
    quantization_config = None
    if use_4bit:
        quantization_config = BitsAndBytesConfig(
            load_in_4bit=True,
            bnb_4bit_quant_type="nf4",
            bnb_4bit_compute_dtype=torch.float16,
            bnb_4bit_use_double_quant=True,
        )

    model = AutoModelForCausalLM.from_pretrained(
        args.base_model,
        quantization_config=quantization_config,
        device_map="auto" if use_4bit else None,
    )
    if use_4bit:
        model = prepare_model_for_kbit_training(model)

    explicit_targets = None
    if args.target_modules:
        explicit_targets = [name.strip() for name in args.target_modules.split(",") if name.strip()]
        if not explicit_targets:
            raise ValueError("target_modules must include at least one module name when provided.")
    resolved_targets = explicit_targets or _select_target_modules(model)

    lora_config = LoraConfig(
        r=profile["lora_r"],
        lora_alpha=profile["lora_alpha"],
        lora_dropout=profile["lora_dropout"],
        bias="none",
        task_type="CAUSAL_LM",
        target_modules=resolved_targets,
    )
    model = get_peft_model(model, lora_config)

    def to_features(rows: list[dict]) -> Dataset:
        texts = [_format_instruction(row, template_strategy=template_strategy, rng=formatter_rng) for row in rows]
        dataset = Dataset.from_dict({"text": texts})
        return dataset.map(
            lambda batch: tokenizer(batch["text"], truncation=True, max_length=profile["max_length"]),
            batched=True,
            remove_columns=["text"],
        )

    train_dataset = to_features(train_rows)
    val_dataset = to_features(val_rows)
    collator = DataCollatorForLanguageModeling(tokenizer=tokenizer, mlm=False)

    train_args = TrainingArguments(
        output_dir=str(output_dir / "trainer_state"),
        per_device_train_batch_size=profile["batch_size"],
        per_device_eval_batch_size=max(1, profile["batch_size"]),
        gradient_accumulation_steps=args.grad_accum_steps,
        num_train_epochs=profile["epochs"],
        max_steps=args.max_steps,
        learning_rate=profile["learning_rate"],
        warmup_ratio=args.warmup_ratio,
        weight_decay=args.weight_decay,
        evaluation_strategy=args.eval_strategy,
        save_strategy=args.save_strategy,
        logging_steps=args.logging_steps,
        report_to=[],
        fp16=has_cuda,
        bf16=False,
        seed=args.seed,
    )

    trainer = Trainer(
        model=model,
        args=train_args,
        train_dataset=train_dataset,
        eval_dataset=val_dataset,
        tokenizer=tokenizer,
        data_collator=collator,
    )
    train_result = trainer.train()
    eval_metrics = trainer.evaluate()

    adapter_dir = output_dir / "adapter"
    model.save_pretrained(adapter_dir)
    tokenizer.save_pretrained(adapter_dir)

    metadata["status"] = "trained"
    metadata["resolved_target_modules"] = resolved_targets
    metadata["resolved_use_4bit"] = use_4bit
    metadata["train_loss"] = getattr(train_result, "training_loss", None)
    metadata["eval_metrics"] = eval_metrics
    metadata["adapter_dir"] = str(adapter_dir)
    (output_dir / "training_config.json").write_text(json.dumps(metadata, indent=2), encoding="utf-8")
    return metadata

