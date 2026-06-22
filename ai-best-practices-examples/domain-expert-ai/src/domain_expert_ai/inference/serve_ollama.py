import json
import subprocess


def serve_once(prompt: str, model: str = "domain-expert-ai") -> str:
    try:
        result = subprocess.run(
            ["ollama", "run", model, prompt],
            check=True,
            capture_output=True,
            text=True,
            timeout=10,
        )
        raw_answer = result.stdout.strip()
    except (FileNotFoundError, subprocess.CalledProcessError, subprocess.TimeoutExpired):
        # Keep local demo resilient when ollama/model is unavailable.
        raw_answer = (
            "I could not reach the local model runtime. "
            "Please verify Ollama is running and model is available."
        )

    payload = {
        "answer": raw_answer,
        "citations": [],
        "confidence": 0.0,
        "disclaimer": "Educational technical guidance; validate in your environment.",
    }
    return json.dumps(payload, ensure_ascii=True)

