import json
from argparse import Namespace

from domain_expert_ai.training.train_qlora import _format_instruction, run_training
from domain_expert_ai.training.train_qlora import _resolve_training_profile
from domain_expert_ai.training.train_qlora import _resolve_template_strategy


def test_format_instruction_contains_required_sections():
    row = {
        "question": "When should I prefer BFS over DFS?",
        "context": "Need shortest path in an unweighted graph.",
        "answer": "Use BFS for shortest path levels in unweighted graphs.",
        "citations": ["CLRS Chapter 22"],
    }
    text = _format_instruction(row)
    assert "Question:" in text
    assert "Context:" in text
    assert "Answer:" in text
    assert "References:" in text


def test_run_training_dry_run_writes_metadata(tmp_path):
    train_file = tmp_path / "train.jsonl"
    val_file = tmp_path / "val.jsonl"
    output_dir = tmp_path / "output"
    row = {
        "question": "How do I reduce overfitting in my model?",
        "context": "Validation performance drops while train accuracy rises.",
        "answer": "Use regularization, augmentation, and early stopping.",
        "jurisdiction": "ML-AI",
        "risk_level": "medium",
        "citations": ["Deep Learning (Goodfellow et al.)"],
    }
    train_file.write_text(json.dumps(row) + "\n", encoding="utf-8")
    val_file.write_text(json.dumps(row) + "\n", encoding="utf-8")

    args = Namespace(
        train_file=str(train_file),
        val_file=str(val_file),
        output_dir=str(output_dir),
        base_model="TinyLlama/TinyLlama-1.1B-Chat-v1.0",
        epochs=1,
        batch_size=1,
        learning_rate=2e-4,
        lora_r=8,
        lora_alpha=16,
        lora_dropout=0.05,
        max_length=256,
        profile="tiny",
        max_steps=5,
        grad_accum_steps=1,
        warmup_ratio=0.03,
        weight_decay=0.0,
        logging_steps=10,
        eval_strategy="epoch",
        save_strategy="epoch",
        seed=42,
        use_4bit=None,
        target_modules=None,
        dry_run=True,
    )

    metadata = run_training(args)
    assert metadata["status"] == "dry_run"
    assert metadata["profile"] == "tiny"
    assert metadata["max_steps"] == 5
    assert (output_dir / "training_config.json").exists()


def test_resolve_training_profile_applies_profile_defaults():
    args = Namespace(
        profile="tiny",
        epochs=None,
        batch_size=None,
        learning_rate=None,
        lora_r=None,
        lora_alpha=None,
        lora_dropout=None,
        max_length=None,
    )
    resolved = _resolve_training_profile(args)
    assert resolved["epochs"] == 1
    assert resolved["batch_size"] == 1
    assert resolved["max_length"] <= 256


def test_format_instruction_backward_compatible_default():
    row = {
        "question": "How do I implement an LRU cache?",
        "context": "Need O(1) operations for get and put.",
        "answer": "Use hash map with doubly linked list.",
        "citations": ["CLRS Chapter 10"],
    }
    text = _format_instruction(row)
    assert text.startswith("You are a senior computer science")
    assert "Question:\nHow do I implement an LRU cache?" in text


def test_resolve_template_strategy_defaults_to_direct():
    args = Namespace(template_strategy=None, seed=123)
    strategy, rng = _resolve_template_strategy(args)
    assert strategy == "direct"
    assert rng is not None


def test_format_instruction_supports_all_template_variants():
    row = {
        "question": "How should I partition data for distributed training?",
        "context": "GPU cluster with uneven batch processing times.",
        "answer": "Balance shards, monitor skew, and tune communication strategy.",
        "citations": ["PyTorch Docs"],
    }
    assert "Question:" in _format_instruction(row, template_strategy="direct")
    assert "Scenario:" in _format_instruction(row, template_strategy="scenario")
    assert "Potential ambiguity:" in _format_instruction(row, template_strategy="ambiguity")


def test_format_instruction_mixed_strategy_is_seeded_deterministic():
    row = {
        "question": "How should I partition data for distributed training?",
        "context": "GPU cluster with uneven batch processing times.",
        "answer": "Balance shards, monitor skew, and tune communication strategy.",
        "citations": ["PyTorch Docs"],
    }
    args_a = Namespace(template_strategy="mixed", seed=7)
    args_b = Namespace(template_strategy="mixed", seed=7)
    strategy_a, rng_a = _resolve_template_strategy(args_a)
    strategy_b, rng_b = _resolve_template_strategy(args_b)

    rendered_a = [_format_instruction(row, template_strategy=strategy_a, rng=rng_a) for _ in range(5)]
    rendered_b = [_format_instruction(row, template_strategy=strategy_b, rng=rng_b) for _ in range(5)]

    assert rendered_a == rendered_b

