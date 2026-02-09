"""Universal Model Registry — MCP server exposing curated AI model metadata."""

import json
import os
from typing import Optional

from fastmcp import FastMCP
from models_data import MODELS

mcp = FastMCP(
    "Universal Model Registry",
    instructions=(
        "Query this server to get accurate, up-to-date information about AI models. "
        "Use list_models to browse, get_model_info for details, recommend_model for "
        "task-based suggestions, and check_model_status to verify if a model ID is "
        "current, legacy, or deprecated."
    ),
)


# ── Helpers ────────────────────────────────────────────────────────────────


def _format_table(models: list[dict]) -> str:
    """Render a list of model dicts as a markdown table."""
    if not models:
        return "No models found matching the criteria."

    rows = []
    rows.append(
        "| Model ID | Display Name | Provider | Status | Context | "
        "Input $/1M | Output $/1M |"
    )
    rows.append(
        "|----------|-------------|----------|--------|---------|"
        "-----------|-------------|"
    )
    for m in models:
        ctx = f"{m['context_window']:,}"
        rows.append(
            f"| {m['id']} | {m['display_name']} | {m['provider']} | "
            f"{m['status']} | {ctx} | ${m['pricing_input']:.2f} | "
            f"${m['pricing_output']:.2f} |"
        )
    return "\n".join(rows)


def _model_detail(m: dict) -> str:
    """Render full specs for a single model as markdown."""
    caps = []
    if m["vision"]:
        caps.append("Vision")
    if m["reasoning"]:
        caps.append("Reasoning/Thinking")
    caps_str = ", ".join(caps) if caps else "None"

    return f"""## {m['display_name']} (`{m['id']}`)

| Field | Value |
|-------|-------|
| Provider | {m['provider']} |
| Status | **{m['status']}** |
| Context Window | {m['context_window']:,} tokens |
| Max Output | {m['max_output_tokens']:,} tokens |
| Capabilities | {caps_str} |
| Pricing (input) | ${m['pricing_input']:.2f} / 1M tokens |
| Pricing (output) | ${m['pricing_output']:.2f} / 1M tokens |
| Knowledge Cutoff | {m['knowledge_cutoff']} |
| Release Date | {m['release_date']} |
| Notes | {m['notes'] or '—'} |"""


# ── Tools ──────────────────────────────────────────────────────────────────


@mcp.tool()
def list_models(
    provider: Optional[str] = None,
    status: Optional[str] = None,
    capability: Optional[str] = None,
) -> str:
    """List AI models with optional filters.

    Args:
        provider: Filter by provider name (OpenAI, Anthropic, Google). Case-insensitive.
        status: Filter by status (current, legacy, deprecated).
        capability: Filter by capability (vision, reasoning).

    Returns:
        Markdown table of matching models.
    """
    results = list(MODELS.values())

    if provider:
        p = provider.lower()
        results = [m for m in results if m["provider"].lower() == p]

    if status:
        s = status.lower()
        results = [m for m in results if m["status"].lower() == s]

    if capability:
        c = capability.lower()
        if c == "vision":
            results = [m for m in results if m["vision"]]
        elif c in ("reasoning", "thinking"):
            results = [m for m in results if m["reasoning"]]

    return _format_table(results)


@mcp.tool()
def get_model_info(model_id: str) -> str:
    """Get full specifications for a specific model.

    Args:
        model_id: The API model ID string (e.g. 'claude-opus-4-6', 'gpt-5', 'gemini-2.5-pro').

    Returns:
        Detailed markdown spec sheet, or an error if the model is not found.
    """
    model = MODELS.get(model_id)
    if not model:
        # Try case-insensitive / partial match
        model_id_lower = model_id.lower()
        for key, m in MODELS.items():
            if key.lower() == model_id_lower or model_id_lower in key.lower():
                model = m
                break

    if not model:
        known = ", ".join(sorted(MODELS.keys()))
        return f"Model `{model_id}` not found in registry.\n\nKnown models: {known}"

    return _model_detail(model)


@mcp.tool()
def recommend_model(
    task: str,
    budget: Optional[str] = None,
) -> str:
    """Recommend the best model for a given task.

    Args:
        task: Description of the task (e.g. 'coding', 'vision', 'cheap batch processing',
              'long context analysis', 'complex reasoning').
        budget: Optional budget constraint ('cheap', 'moderate', 'unlimited'). Default: 'moderate'.

    Returns:
        Markdown recommendation with reasoning.
    """
    budget = (budget or "moderate").lower()
    task_lower = task.lower()

    current = [m for m in MODELS.values() if m["status"] == "current"]

    # Score each model
    scored: list[tuple[float, dict]] = []
    for m in current:
        score = 0.0

        # Task relevance
        if "coding" in task_lower or "code" in task_lower:
            if m["reasoning"]:
                score += 3
            if m["context_window"] >= 200_000:
                score += 1
        if "vision" in task_lower or "image" in task_lower:
            if m["vision"]:
                score += 4
            else:
                score -= 10
        if "reason" in task_lower or "think" in task_lower or "math" in task_lower:
            if m["reasoning"]:
                score += 5
        if "long context" in task_lower or "large document" in task_lower:
            if m["context_window"] >= 1_000_000:
                score += 4
            elif m["context_window"] >= 200_000:
                score += 2
        if "cheap" in task_lower or "batch" in task_lower or "cost" in task_lower:
            score += max(0, 5 - m["pricing_input"])

        # Budget modifier
        if budget == "cheap":
            score += max(0, 3 - m["pricing_input"])
            if m["pricing_input"] > 5:
                score -= 5
        elif budget == "unlimited":
            # Prefer the most capable
            score += min(m["pricing_input"], 5)

        # General quality signal: higher-priced models are generally more capable
        score += min(m["pricing_input"] * 0.3, 2)

        scored.append((score, m))

    scored.sort(key=lambda x: x[0], reverse=True)
    top = scored[:3]

    lines = [f"## Recommendations for: *{task}*", f"**Budget:** {budget}", ""]
    for i, (score, m) in enumerate(top, 1):
        caps = []
        if m["vision"]:
            caps.append("vision")
        if m["reasoning"]:
            caps.append("reasoning")
        cap_str = ", ".join(caps) if caps else "standard"
        lines.append(
            f"{i}. **{m['display_name']}** (`{m['id']}`)\n"
            f"   - Provider: {m['provider']} | Capabilities: {cap_str}\n"
            f"   - Pricing: ${m['pricing_input']:.2f} / ${m['pricing_output']:.2f} per 1M tokens\n"
            f"   - Context: {m['context_window']:,} tokens\n"
        )

    return "\n".join(lines)


@mcp.tool()
def check_model_status(model_id: str) -> str:
    """Check whether a model ID is current, legacy, or deprecated.

    Args:
        model_id: The API model ID string to check.

    Returns:
        Status information and recommended replacement if applicable.
    """
    model = MODELS.get(model_id)
    if not model:
        model_id_lower = model_id.lower()
        for key, m in MODELS.items():
            if key.lower() == model_id_lower or model_id_lower in key.lower():
                model = m
                break

    if not model:
        return (
            f"`{model_id}` is **not found** in the registry. "
            "It may be misspelled, very old, or not yet tracked."
        )

    status = model["status"]
    result = f"**{model['display_name']}** (`{model['id']}`): status = **{status}**"

    if status in ("legacy", "deprecated"):
        # Find current replacement from same provider
        replacements = [
            m
            for m in MODELS.values()
            if m["provider"] == model["provider"] and m["status"] == "current"
        ]
        replacements.sort(
            key=lambda m: abs(m["pricing_input"] - model["pricing_input"])
        )
        if replacements:
            r = replacements[0]
            result += (
                f"\n\nRecommended replacement: **{r['display_name']}** (`{r['id']}`)"
            )

    if model["notes"]:
        result += f"\n\nNote: {model['notes']}"

    return result


# ── Resources ──────────────────────────────────────────────────────────────


@mcp.resource("model://registry/all")
def get_all_models() -> str:
    """Full JSON dump of the entire model registry."""
    return json.dumps(MODELS, indent=2)


@mcp.resource("model://registry/current")
def get_current_models() -> str:
    """JSON dump of only current (non-legacy, non-deprecated) models."""
    current = {k: v for k, v in MODELS.items() if v["status"] == "current"}
    return json.dumps(current, indent=2)


# ── Entry point ────────────────────────────────────────────────────────────

if __name__ == "__main__":
    transport = os.getenv("MCP_TRANSPORT", "stdio")
    if transport == "sse":
        mcp.run(
            transport="sse",
            host="0.0.0.0",
            port=int(os.getenv("PORT", "8000")),
        )
    else:
        mcp.run()
