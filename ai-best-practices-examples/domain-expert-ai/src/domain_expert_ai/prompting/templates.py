from typing import Callable


TemplateRenderer = Callable[[dict], str]


def render_direct_template(row: dict) -> str:
    citations = ", ".join(row.get("citations", []))
    return (
        "You are a senior computer science and software engineering assistant. "
        "Provide technically accurate, practical guidance.\n\n"
        f"Question:\n{row.get('question', '')}\n\n"
        f"Context:\n{row.get('context', '')}\n\n"
        f"Answer:\n{row.get('answer', '')}\n\n"
        f"References:\n{citations}\n"
        "Disclaimer: This is educational technical guidance; validate in your environment."
    )


def render_scenario_template(row: dict) -> str:
    citations = ", ".join(row.get("citations", []))
    return (
        "You are a senior computer science and software engineering assistant. "
        "Analyze the engineering scenario and respond concisely.\n\n"
        f"Scenario:\n{row.get('context', '')}\n\n"
        f"User question:\n{row.get('question', '')}\n\n"
        f"Recommended answer:\n{row.get('answer', '')}\n\n"
        f"Supporting references:\n{citations}\n"
        "Disclaimer: This is educational technical guidance; validate in your environment."
    )


def render_ambiguity_template(row: dict) -> str:
    citations = ", ".join(row.get("citations", []))
    return (
        "You are a senior computer science and software engineering assistant. "
        "Highlight uncertainty and provide safe implementation guidance.\n\n"
        f"Question:\n{row.get('question', '')}\n\n"
        f"Known facts:\n{row.get('context', '')}\n\n"
        "Potential ambiguity:\nIdentify missing technical details that may change the recommendation.\n\n"
        f"Best available answer:\n{row.get('answer', '')}\n\n"
        f"References:\n{citations}\n"
        "Disclaimer: This is educational technical guidance; validate in your environment."
    )


TEMPLATE_RENDERERS: dict[str, TemplateRenderer] = {
    "direct": render_direct_template,
    "scenario": render_scenario_template,
    "ambiguity": render_ambiguity_template,
}
