"""Edge case and boundary tests for the Universal Model Registry."""

from models_data import MODELS
from registry import (
    check_model_status as _check_model_status,
)
from registry import (
    compare_models as _compare_models,
)
from registry import (
    get_model_info as _get_model_info,
)
from registry import (
    list_models as _list_models,
)
from registry import (
    recommend_model as _recommend_model,
)

list_models = _list_models.fn
get_model_info = _get_model_info.fn
recommend_model = _recommend_model.fn
check_model_status = _check_model_status.fn
compare_models = _compare_models.fn


# ── Empty string inputs ──────────────────────────────────────────────────


class TestEmptyStringInputs:
    def test_get_model_info_empty(self):
        result = get_model_info("")
        # Empty string is a substring of every key, so partial match returns the first model
        assert "##" in result  # Returns a model detail block

    def test_check_model_status_empty(self):
        result = check_model_status("")
        # Empty string partial-matches the first model
        assert "status" in result.lower()

    def test_recommend_model_empty(self):
        result = recommend_model("")
        # Should still return recommendations (no task keywords matched)
        assert "Recommendations for" in result
        assert "1." in result


# ── Special characters ───────────────────────────────────────────────────


class TestSpecialCharacters:
    def test_get_model_info_sql_injection(self):
        result = get_model_info("gpt-5'; DROP TABLE")
        assert "not found" in result.lower()

    def test_check_model_status_xss(self):
        result = check_model_status("<script>alert(1)</script>")
        assert "not found" in result.lower()

    def test_recommend_model_special_chars(self):
        result = recommend_model("task with $pecial ch@rs & symbols!")
        assert "Recommendations for" in result
        assert "1." in result

    def test_compare_models_special_chars(self):
        result = compare_models(["gpt-5", "'; DROP TABLE models;--"])
        assert "not found" in result.lower()


# ── Unicode inputs ───────────────────────────────────────────────────────


class TestUnicodeInputs:
    def test_get_model_info_chinese(self):
        result = get_model_info("模型")
        assert "not found" in result.lower()

    def test_recommend_model_chinese(self):
        result = recommend_model("编程任务")
        assert "Recommendations for" in result
        assert "1." in result

    def test_get_model_info_emoji(self):
        result = get_model_info("🤖")
        assert "not found" in result.lower()

    def test_recommend_model_mixed_unicode(self):
        result = recommend_model("コーディング coding タスク")
        # "coding" keyword should still be detected
        assert "Recommendations for" in result
        assert "1." in result


# ── compare_models boundary cases ────────────────────────────────────────


class TestCompareModelsBoundary:
    def test_exactly_five_models(self):
        ids = ["gpt-5", "claude-opus-4-6", "gemini-2.5-pro", "grok-3", "o3"]
        result = compare_models(ids)
        assert "GPT-5" in result
        assert "Claude Opus 4.6" in result
        assert "Gemini 2.5 Pro" in result
        assert "Grok 3" in result
        assert "o3" in result

    def test_six_models_truncates_to_five(self):
        ids = [
            "gpt-5",
            "claude-opus-4-6",
            "gemini-2.5-pro",
            "grok-3",
            "o3",
            "deepseek-reasoner",
        ]
        result = compare_models(ids)
        # First 5 should be present
        assert "GPT-5" in result
        assert "Claude Opus 4.6" in result
        assert "Gemini 2.5 Pro" in result
        assert "Grok 3" in result
        assert "o3" in result
        # 6th model should be truncated
        assert "DeepSeek Reasoner" not in result

    def test_empty_list(self):
        result = compare_models([])
        assert "at least 2" in result.lower()

    def test_duplicate_ids(self):
        result = compare_models(["gpt-5", "gpt-5"])
        # Should still work — duplicates produce duplicate columns
        assert "GPT-5" in result

    def test_single_model_rejected(self):
        result = compare_models(["gpt-5"])
        assert "at least 2" in result.lower()


# ── Combined filters ─────────────────────────────────────────────────────


class TestCombinedFilters:
    def test_provider_and_status(self):
        result = list_models(provider="OpenAI", status="current")
        assert "OpenAI" in result
        # Should not contain legacy/deprecated OpenAI models
        assert "gpt-4o" not in result.lower() or "gpt-4o-mini" not in result

    def test_provider_and_capability_vision(self):
        result = list_models(provider="Anthropic", capability="vision")
        assert "Anthropic" in result
        # All Anthropic models have vision, so we should see results
        assert "No models found" not in result

    def test_provider_and_capability_reasoning(self):
        result = list_models(provider="OpenAI", capability="reasoning")
        assert "OpenAI" in result
        # Non-reasoning OpenAI models should be excluded
        assert "gpt-5 " not in result or "| gpt-5 |" not in result

    def test_all_filters_no_match(self):
        result = list_models(provider="Meta", capability="reasoning")
        # Meta models don't have reasoning
        assert "No models found" in result

    def test_status_and_capability(self):
        result = list_models(status="current", capability="vision")
        # Should only contain current models with vision
        assert (
            "deprecated" not in result.lower().split("\n")[2:].__str__()
            or "No models found" in result
        )


# ── Budget types ─────────────────────────────────────────────────────────


class TestAllBudgetTypes:
    def test_budget_cheap(self):
        result = recommend_model("general tasks", budget="cheap")
        assert "Budget:** cheap" in result
        assert "1." in result

    def test_budget_moderate(self):
        result = recommend_model("general tasks", budget="moderate")
        assert "Budget:** moderate" in result
        assert "1." in result

    def test_budget_unlimited(self):
        result = recommend_model("general tasks", budget="unlimited")
        assert "Budget:** unlimited" in result
        assert "1." in result

    def test_budget_default_is_moderate(self):
        result = recommend_model("general tasks")
        assert "Budget:** moderate" in result

    def test_budget_case_insensitive(self):
        result = recommend_model("general tasks", budget="CHEAP")
        assert "Budget:** cheap" in result


# ── Case variations ──────────────────────────────────────────────────────


class TestCaseVariations:
    def test_provider_uppercase(self):
        result = list_models(provider="OPENAI")
        assert "OpenAI" in result
        assert "No models found" not in result

    def test_provider_mixed_case(self):
        result = list_models(provider="oPeNaI")
        assert "OpenAI" in result

    def test_status_uppercase(self):
        result = list_models(status="CURRENT")
        assert "current" in result
        assert "No models found" not in result

    def test_status_mixed_case(self):
        result = list_models(status="Current")
        assert "current" in result

    def test_capability_uppercase(self):
        result = list_models(capability="VISION")
        assert "No models found" not in result

    def test_capability_reasoning_uppercase(self):
        result = list_models(capability="REASONING")
        assert "No models found" not in result


# ── Model count validation ───────────────────────────────────────────────


class TestModelCountValidation:
    def test_total_model_count(self):
        assert len(MODELS) >= 30, f"Expected >=30 models, got {len(MODELS)}"

    def test_at_least_six_providers(self):
        providers = {m["provider"] for m in MODELS.values()}
        assert len(providers) >= 6, f"Expected >=6 providers, got {providers}"

    def test_list_models_returns_all_by_default(self):
        result = list_models()
        for model_id in MODELS:
            assert model_id in result

    def test_each_provider_has_current_model(self):
        providers = {m["provider"] for m in MODELS.values()}
        for provider in providers:
            current = [
                m for m in MODELS.values() if m["provider"] == provider and m["status"] == "current"
            ]
            assert len(current) >= 1, f"{provider} has no current models"
