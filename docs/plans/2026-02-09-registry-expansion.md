# Registry Expansion Implementation Plan

> **For implementer:** REQUIRED SUB-SKILL: Use sr-executing-plans to implement this plan task-by-task.

**Goal:** Expand the model registry with 4 new providers, add a compare_models tool, add pytest test suite, and improve the recommendation engine.

**Architecture:** Direct additions to the existing flat-file architecture. New providers go into `models_data.py`, new tool into `registry.py`, tests in `tests/`. No new dependencies beyond pytest.

**Tech Stack:** Python 3.12, FastMCP, pytest

---

### Task 1: Add pytest infrastructure

**Files:**
- Modify: `pyproject.toml`
- Create: `tests/__init__.py`
- Create: `tests/test_registry.py`

**Step 1: Add pytest dev dependency to pyproject.toml**

Add after the `dependencies` list in `pyproject.toml`:

```toml
[tool.pytest.ini_options]
testpaths = ["tests"]

[dependency-groups]
dev = ["pytest>=8.0"]
```

**Step 2: Sync dependencies**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv sync`
Expected: Success, pytest now available

**Step 3: Create test scaffolding**

Create `tests/__init__.py` (empty).

Create `tests/test_registry.py`:

```python
"""Tests for the Universal Model Registry tools."""

from models_data import MODELS
from registry import (
    _format_table,
    _model_detail,
    list_models,
    get_model_info,
    recommend_model,
    check_model_status,
)


# ── Data integrity ────────────────────────────────────────────────────────


class TestModelsData:
    """Verify every model entry has the required schema."""

    REQUIRED_KEYS = {
        "id", "display_name", "provider", "context_window",
        "max_output_tokens", "vision", "reasoning", "pricing_input",
        "pricing_output", "knowledge_cutoff", "release_date", "status", "notes",
    }

    def test_all_models_have_required_keys(self):
        for model_id, model in MODELS.items():
            missing = self.REQUIRED_KEYS - set(model.keys())
            assert not missing, f"{model_id} missing keys: {missing}"

    def test_model_id_matches_dict_key(self):
        for key, model in MODELS.items():
            assert key == model["id"], f"Key {key!r} != model id {model['id']!r}"

    def test_status_values_are_valid(self):
        valid = {"current", "legacy", "deprecated"}
        for model_id, model in MODELS.items():
            assert model["status"] in valid, f"{model_id} has invalid status: {model['status']}"

    def test_pricing_is_non_negative(self):
        for model_id, model in MODELS.items():
            assert model["pricing_input"] >= 0, f"{model_id} has negative input pricing"
            assert model["pricing_output"] >= 0, f"{model_id} has negative output pricing"

    def test_context_window_is_positive(self):
        for model_id, model in MODELS.items():
            assert model["context_window"] > 0, f"{model_id} has non-positive context window"

    def test_at_least_three_providers(self):
        providers = {m["provider"] for m in MODELS.values()}
        assert len(providers) >= 3, f"Only {len(providers)} providers found"


# ── list_models ───────────────────────────────────────────────────────────


class TestListModels:
    def test_no_filters_returns_all(self):
        result = list_models()
        for model_id in MODELS:
            assert model_id in result

    def test_filter_by_provider(self):
        result = list_models(provider="Anthropic")
        assert "Anthropic" in result
        assert "OpenAI" not in result

    def test_filter_by_provider_case_insensitive(self):
        result = list_models(provider="anthropic")
        assert "Anthropic" in result

    def test_filter_by_status(self):
        result = list_models(status="deprecated")
        # All rows should be deprecated
        for line in result.split("\n")[2:]:  # skip header
            if line.strip():
                assert "deprecated" in line

    def test_filter_by_vision(self):
        result = list_models(capability="vision")
        # Should not contain models without vision
        non_vision = [m["id"] for m in MODELS.values() if not m["vision"]]
        for mid in non_vision:
            assert mid not in result

    def test_filter_by_reasoning(self):
        result = list_models(capability="reasoning")
        non_reasoning = [m["id"] for m in MODELS.values() if not m["reasoning"]]
        for mid in non_reasoning:
            assert mid not in result

    def test_no_results(self):
        result = list_models(provider="Nonexistent")
        assert "No models found" in result


# ── get_model_info ────────────────────────────────────────────────────────


class TestGetModelInfo:
    def test_exact_match(self):
        result = get_model_info("gpt-5")
        assert "GPT-5" in result
        assert "OpenAI" in result

    def test_case_insensitive(self):
        result = get_model_info("GPT-5")
        assert "GPT-5" in result

    def test_partial_match(self):
        result = get_model_info("opus-4-6")
        assert "Claude Opus 4.6" in result

    def test_not_found(self):
        result = get_model_info("nonexistent-model")
        assert "not found" in result


# ── recommend_model ───────────────────────────────────────────────────────


class TestRecommendModel:
    def test_coding_task(self):
        result = recommend_model("coding")
        assert "Recommendations for" in result
        # Should have numbered recommendations
        assert "1." in result

    def test_vision_task(self):
        result = recommend_model("image analysis")
        assert "vision" in result.lower()

    def test_cheap_budget(self):
        result = recommend_model("general tasks", budget="cheap")
        assert "Budget:** cheap" in result

    def test_reasoning_task(self):
        result = recommend_model("complex math reasoning")
        assert "reasoning" in result.lower()


# ── check_model_status ────────────────────────────────────────────────────


class TestCheckModelStatus:
    def test_current_model(self):
        result = check_model_status("gpt-5")
        assert "current" in result.lower()

    def test_legacy_model(self):
        result = check_model_status("gpt-4o")
        assert "legacy" in result.lower()
        assert "replacement" in result.lower()

    def test_deprecated_model(self):
        result = check_model_status("gpt-4o-mini")
        assert "deprecated" in result.lower()

    def test_not_found(self):
        result = check_model_status("fake-model")
        assert "not found" in result.lower()
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv run pytest tests/ -v`
Expected: All tests PASS

**Step 5: Commit**

```bash
git add pyproject.toml tests/
git commit -m "feat: add pytest test suite for registry tools and data integrity"
```

---

### Task 2: Add xAI (Grok) models to registry

**Files:**
- Modify: `models_data.py` (append after Google section, before closing brace)

**Step 1: Add xAI models**

Append to `models_data.py` MODELS dict (before the closing `}`):

```python
    # ─── xAI: Current ────────────────────────────────────────────────
    "grok-3": {
        "id": "grok-3",
        "display_name": "Grok 3",
        "provider": "xAI",
        "context_window": 131_072,
        "max_output_tokens": 131_072,
        "vision": True,
        "reasoning": False,
        "pricing_input": 3.00,
        "pricing_output": 15.00,
        "knowledge_cutoff": "2025-03",
        "release_date": "2025-02",
        "status": "current",
        "notes": "xAI flagship model",
    },
    "grok-3-mini": {
        "id": "grok-3-mini",
        "display_name": "Grok 3 Mini",
        "provider": "xAI",
        "context_window": 131_072,
        "max_output_tokens": 131_072,
        "vision": True,
        "reasoning": True,
        "pricing_input": 0.30,
        "pricing_output": 0.50,
        "knowledge_cutoff": "2025-03",
        "release_date": "2025-02",
        "status": "current",
        "notes": "Fast thinking model from xAI",
    },
```

**Step 2: Run tests**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv run pytest tests/ -v`
Expected: All PASS (data integrity tests validate new entries automatically)

**Step 3: Commit**

```bash
git add models_data.py
git commit -m "feat: add xAI Grok 3 and Grok 3 Mini to registry"
```

---

### Task 3: Add Meta (Llama) models to registry

**Files:**
- Modify: `models_data.py`

**Step 1: Add Meta models**

Append to MODELS dict:

```python
    # ─── Meta: Current ────────────────────────────────────────────────
    "llama-4-maverick": {
        "id": "llama-4-maverick",
        "display_name": "Llama 4 Maverick",
        "provider": "Meta",
        "context_window": 1_048_576,
        "max_output_tokens": 32_768,
        "vision": True,
        "reasoning": False,
        "pricing_input": 0.20,
        "pricing_output": 0.60,
        "knowledge_cutoff": "2025-03",
        "release_date": "2025-04",
        "status": "current",
        "notes": "Open-weight MoE model, 1M context, available via API providers",
    },
    "llama-4-scout": {
        "id": "llama-4-scout",
        "display_name": "Llama 4 Scout",
        "provider": "Meta",
        "context_window": 10_000_000,
        "max_output_tokens": 32_768,
        "vision": True,
        "reasoning": False,
        "pricing_input": 0.15,
        "pricing_output": 0.40,
        "knowledge_cutoff": "2025-03",
        "release_date": "2025-04",
        "status": "current",
        "notes": "Open-weight, 10M context window, natively multimodal",
    },
    # ─── Meta: Legacy ─────────────────────────────────────────────────
    "llama-3.3-70b": {
        "id": "llama-3.3-70b",
        "display_name": "Llama 3.3 70B",
        "provider": "Meta",
        "context_window": 128_000,
        "max_output_tokens": 4_096,
        "vision": False,
        "reasoning": False,
        "pricing_input": 0.10,
        "pricing_output": 0.30,
        "knowledge_cutoff": "2024-12",
        "release_date": "2024-12",
        "status": "legacy",
        "notes": "Superseded by Llama 4 series",
    },
```

**Step 2: Run tests**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv run pytest tests/ -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add models_data.py
git commit -m "feat: add Meta Llama 4 Maverick, Scout, and 3.3 70B to registry"
```

---

### Task 4: Add Mistral models to registry

**Files:**
- Modify: `models_data.py`

**Step 1: Add Mistral models**

Append to MODELS dict:

```python
    # ─── Mistral: Current ─────────────────────────────────────────────
    "mistral-large-latest": {
        "id": "mistral-large-latest",
        "display_name": "Mistral Large",
        "provider": "Mistral",
        "context_window": 128_000,
        "max_output_tokens": 8_192,
        "vision": True,
        "reasoning": False,
        "pricing_input": 2.00,
        "pricing_output": 6.00,
        "knowledge_cutoff": "2025-03",
        "release_date": "2025-03",
        "status": "current",
        "notes": "Mistral flagship, strong multilingual",
    },
    "mistral-small-latest": {
        "id": "mistral-small-latest",
        "display_name": "Mistral Small",
        "provider": "Mistral",
        "context_window": 128_000,
        "max_output_tokens": 8_192,
        "vision": True,
        "reasoning": False,
        "pricing_input": 0.10,
        "pricing_output": 0.30,
        "knowledge_cutoff": "2025-03",
        "release_date": "2025-03",
        "status": "current",
        "notes": "Fast and cost-efficient",
    },
    "codestral-latest": {
        "id": "codestral-latest",
        "display_name": "Codestral",
        "provider": "Mistral",
        "context_window": 256_000,
        "max_output_tokens": 8_192,
        "vision": False,
        "reasoning": False,
        "pricing_input": 0.30,
        "pricing_output": 0.90,
        "knowledge_cutoff": "2025-03",
        "release_date": "2025-05",
        "status": "current",
        "notes": "Specialized coding model, 256K context",
    },
```

**Step 2: Run tests**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv run pytest tests/ -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add models_data.py
git commit -m "feat: add Mistral Large, Small, and Codestral to registry"
```

---

### Task 5: Add DeepSeek models to registry

**Files:**
- Modify: `models_data.py`

**Step 1: Add DeepSeek models**

Append to MODELS dict:

```python
    # ─── DeepSeek: Current ────────────────────────────────────────────
    "deepseek-r1": {
        "id": "deepseek-r1",
        "display_name": "DeepSeek R1",
        "provider": "DeepSeek",
        "context_window": 128_000,
        "max_output_tokens": 16_384,
        "vision": False,
        "reasoning": True,
        "pricing_input": 0.55,
        "pricing_output": 2.19,
        "knowledge_cutoff": "2025-01",
        "release_date": "2025-01",
        "status": "current",
        "notes": "Open-weight reasoning model, chain-of-thought",
    },
    "deepseek-v3": {
        "id": "deepseek-v3",
        "display_name": "DeepSeek V3",
        "provider": "DeepSeek",
        "context_window": 128_000,
        "max_output_tokens": 16_384,
        "vision": False,
        "reasoning": False,
        "pricing_input": 0.27,
        "pricing_output": 1.10,
        "knowledge_cutoff": "2025-01",
        "release_date": "2025-01",
        "status": "current",
        "notes": "Open-weight MoE general model",
    },
```

**Step 2: Run tests**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv run pytest tests/ -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add models_data.py
git commit -m "feat: add DeepSeek R1 and V3 to registry"
```

---

### Task 6: Add compare_models tool

**Files:**
- Modify: `registry.py` (add new tool after `check_model_status`)
- Modify: `tests/test_registry.py` (add test class)

**Step 1: Write failing tests**

Add to `tests/test_registry.py` imports:

```python
from registry import (
    _format_table,
    _model_detail,
    list_models,
    get_model_info,
    recommend_model,
    check_model_status,
    compare_models,
)
```

Add new test class:

```python
# ── compare_models ────────────────────────────────────────────────────────


class TestCompareModels:
    def test_two_models(self):
        result = compare_models(["gpt-5", "claude-opus-4-6"])
        assert "GPT-5" in result
        assert "Claude Opus 4.6" in result

    def test_three_models(self):
        result = compare_models(["gpt-5", "claude-opus-4-6", "gemini-2.5-pro"])
        assert "GPT-5" in result
        assert "Gemini 2.5 Pro" in result

    def test_single_model_error(self):
        result = compare_models(["gpt-5"])
        assert "at least 2" in result.lower()

    def test_not_found_model(self):
        result = compare_models(["gpt-5", "nonexistent"])
        assert "not found" in result.lower()

    def test_case_insensitive(self):
        result = compare_models(["GPT-5", "CLAUDE-OPUS-4-6"])
        assert "GPT-5" in result
        assert "Claude Opus 4.6" in result
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv run pytest tests/test_registry.py::TestCompareModels -v`
Expected: FAIL — `compare_models` not importable

**Step 3: Implement compare_models in registry.py**

Add after `check_model_status` function (before Resources section, ~line 267):

```python
@mcp.tool()
def compare_models(model_ids: list[str]) -> str:
    """Compare 2-5 models side by side.

    Args:
        model_ids: List of model ID strings to compare (2-5 models).

    Returns:
        Markdown comparison table, or an error if fewer than 2 valid models found.
    """
    if len(model_ids) < 2:
        return "Please provide at least 2 model IDs to compare."

    if len(model_ids) > 5:
        model_ids = model_ids[:5]

    models = []
    not_found = []
    for mid in model_ids:
        m = MODELS.get(mid)
        if not m:
            mid_lower = mid.lower()
            for key, val in MODELS.items():
                if key.lower() == mid_lower or mid_lower in key.lower():
                    m = val
                    break
        if m:
            models.append(m)
        else:
            not_found.append(mid)

    if not_found:
        return f"Model(s) not found: {', '.join(not_found)}"

    if len(models) < 2:
        return "Need at least 2 valid models to compare."

    # Build comparison table — fields as rows, models as columns
    header = "| Field | " + " | ".join(m["display_name"] for m in models) + " |"
    sep = "|-------|" + "|".join("------" for _ in models) + "|"

    def caps(m):
        c = []
        if m["vision"]:
            c.append("Vision")
        if m["reasoning"]:
            c.append("Reasoning")
        return ", ".join(c) if c else "None"

    rows = [
        header,
        sep,
        "| Provider | " + " | ".join(m["provider"] for m in models) + " |",
        "| Status | " + " | ".join(m["status"] for m in models) + " |",
        "| Context | " + " | ".join(f"{m['context_window']:,}" for m in models) + " |",
        "| Max Output | " + " | ".join(f"{m['max_output_tokens']:,}" for m in models) + " |",
        "| Capabilities | " + " | ".join(caps(m) for m in models) + " |",
        "| Input $/1M | " + " | ".join(f"${m['pricing_input']:.2f}" for m in models) + " |",
        "| Output $/1M | " + " | ".join(f"${m['pricing_output']:.2f}" for m in models) + " |",
        "| Knowledge Cutoff | " + " | ".join(m["knowledge_cutoff"] for m in models) + " |",
        "| Release Date | " + " | ".join(m["release_date"] for m in models) + " |",
    ]
    return "\n".join(rows)
```

**Step 4: Run tests**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv run pytest tests/ -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add registry.py tests/test_registry.py
git commit -m "feat: add compare_models tool for side-by-side model comparison"
```

---

### Task 7: Improve recommendation engine

**Files:**
- Modify: `registry.py` (rewrite `recommend_model` scoring logic)
- Modify: `tests/test_registry.py` (add edge case tests)

**Step 1: Write additional tests**

Add to `TestRecommendModel`:

```python
    def test_long_context_prefers_large_context(self):
        result = recommend_model("long context analysis of documents")
        # Top recommendation should be a 1M+ context model
        lines = result.split("\n")
        first_rec = next(l for l in lines if l.startswith("1."))
        assert any(
            mid in first_rec
            for mid in ["gemini", "gpt-4.1", "llama-4"]
        ), f"Expected large-context model in first rec: {first_rec}"

    def test_cheap_budget_excludes_expensive(self):
        result = recommend_model("general assistant", budget="cheap")
        lines = result.split("\n")
        first_rec = next(l for l in lines if l.startswith("1."))
        # Should not recommend $10+ models first
        assert "o3" not in first_rec or "o3-mini" in first_rec
        assert "claude-opus" not in first_rec

    def test_coding_specialist(self):
        result = recommend_model("code generation and review")
        assert "1." in result

    def test_multilingual(self):
        result = recommend_model("multilingual translation")
        assert "1." in result
```

**Step 2: Implement improved scoring**

Replace the scoring section in `recommend_model` (lines 158-197) with:

```python
    # Score each model
    scored: list[tuple[float, dict]] = []
    for m in current:
        score = 0.0

        # ── Task relevance signals ──
        if "coding" in task_lower or "code" in task_lower or "programming" in task_lower:
            if m["reasoning"]:
                score += 3
            if m["context_window"] >= 200_000:
                score += 1
            if "codestral" in m["id"]:
                score += 2

        if "vision" in task_lower or "image" in task_lower or "screenshot" in task_lower:
            if m["vision"]:
                score += 4
            else:
                score -= 10

        if "reason" in task_lower or "think" in task_lower or "math" in task_lower or "logic" in task_lower:
            if m["reasoning"]:
                score += 5

        if "long context" in task_lower or "large document" in task_lower or "summariz" in task_lower:
            if m["context_window"] >= 1_000_000:
                score += 4
            elif m["context_window"] >= 200_000:
                score += 2

        if "cheap" in task_lower or "batch" in task_lower or "cost" in task_lower:
            score += max(0, 5 - m["pricing_input"])

        if "multilingual" in task_lower or "translat" in task_lower:
            if m["provider"] == "Mistral":
                score += 2
            if m["context_window"] >= 128_000:
                score += 1

        if "open" in task_lower and ("weight" in task_lower or "source" in task_lower):
            if m["provider"] in ("Meta", "DeepSeek", "Mistral"):
                score += 3

        # ── Budget modifier ──
        if budget == "cheap":
            score += max(0, 3 - m["pricing_input"])
            if m["pricing_input"] > 5:
                score -= 5
        elif budget == "unlimited":
            score += min(m["pricing_input"], 5)

        # General quality signal
        score += min(m["pricing_input"] * 0.3, 2)

        scored.append((score, m))
```

**Step 3: Run tests**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv run pytest tests/ -v`
Expected: All PASS

**Step 4: Commit**

```bash
git add registry.py tests/test_registry.py
git commit -m "feat: improve recommendation engine with broader task matching"
```

---

### Task 8: Update README and Dockerfile

**Files:**
- Modify: `README.md` (update provider list, add compare_models to tools table)
- Verify: `Dockerfile` (no changes needed — it copies `models_data.py` already)

**Step 1: Update README**

Update the tools table to include `compare_models`:

```markdown
| `compare_models(model_ids)` | Side-by-side comparison of 2-5 models |
```

Update the providers section:

```markdown
## Covered Providers & Models

- **OpenAI:** GPT-5.2, GPT-5, GPT-5 Mini, GPT-4.1 series, o3, o4-mini, o3-mini + legacy GPT-4o
- **Anthropic:** Claude Opus 4.6, Sonnet 4.5, Haiku 4.5 + legacy Opus 4.5/4.0, Sonnet 4.0, 3.7
- **Google:** Gemini 3 Pro/Flash (preview), 2.5 Pro/Flash/Flash Lite + deprecated 2.0 Flash
- **xAI:** Grok 3, Grok 3 Mini
- **Meta:** Llama 4 Maverick, Llama 4 Scout + legacy Llama 3.3 70B
- **Mistral:** Mistral Large, Mistral Small, Codestral
- **DeepSeek:** DeepSeek R1, DeepSeek V3
```

**Step 2: Run full test suite one final time**

Run: `cd /Users/aezi/Desktop/universal-model-registry && uv run pytest tests/ -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: update README with new providers and compare_models tool"
```
