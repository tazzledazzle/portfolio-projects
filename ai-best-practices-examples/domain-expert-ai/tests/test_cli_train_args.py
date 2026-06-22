from domain_expert_ai.cli import build_parser


def test_train_cli_supports_max_steps_and_profile():
    parser = build_parser()
    args = parser.parse_args(
        [
            "train",
            "--train-file",
            "train.jsonl",
            "--val-file",
            "val.jsonl",
            "--output-dir",
            "out",
            "--profile",
            "tiny",
            "--max-steps",
            "8",
            "--grad-accum-steps",
            "2",
            "--template-strategy",
            "mixed",
        ]
    )
    assert args.profile == "tiny"
    assert args.max_steps == 8
    assert args.grad_accum_steps == 2
    assert args.template_strategy == "mixed"

