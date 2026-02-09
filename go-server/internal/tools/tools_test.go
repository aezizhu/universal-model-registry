package tools

import (
	"strings"
	"testing"

	"go-server/internal/models"
)

// ── ListModels ────────────────────────────────────────────────────────────

func TestListModels_NoFilters(t *testing.T) {
	result := ListModels("", "", "")
	for id := range models.Models {
		if !strings.Contains(result, id) {
			t.Errorf("expected model %q in result", id)
		}
	}
}

func TestListModels_FilterByProvider(t *testing.T) {
	result := ListModels("Anthropic", "", "")
	if !strings.Contains(result, "Anthropic") {
		t.Error("expected 'Anthropic' in result")
	}
	if strings.Contains(result, "OpenAI") {
		t.Error("did not expect 'OpenAI' in result when filtering by Anthropic")
	}
}

func TestListModels_FilterByProviderCaseInsensitive(t *testing.T) {
	result := ListModels("anthropic", "", "")
	if !strings.Contains(result, "Anthropic") {
		t.Error("expected 'Anthropic' in result for case-insensitive filter")
	}
}

func TestListModels_FilterByStatus(t *testing.T) {
	result := ListModels("", "deprecated", "")
	lines := strings.Split(result, "\n")
	for _, line := range lines[2:] { // skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "deprecated") {
			t.Errorf("expected all rows to be deprecated, got: %s", line)
		}
	}
}

func TestListModels_FilterByVision(t *testing.T) {
	result := ListModels("", "", "vision")
	for _, m := range models.Models {
		if !m.Vision {
			if strings.Contains(result, "| "+m.ID+" |") {
				t.Errorf("non-vision model %q should not appear in vision filter", m.ID)
			}
		}
	}
}

func TestListModels_FilterByReasoning(t *testing.T) {
	result := ListModels("", "", "reasoning")
	for _, m := range models.Models {
		if !m.Reasoning {
			if strings.Contains(result, "| "+m.ID+" |") {
				t.Errorf("non-reasoning model %q should not appear in reasoning filter", m.ID)
			}
		}
	}
}

func TestListModels_NoResults(t *testing.T) {
	result := ListModels("Nonexistent", "", "")
	if !strings.Contains(result, "No models found") {
		t.Errorf("expected 'No models found' for nonexistent provider, got: %s", result)
	}
}

// ── GetModelInfo ──────────────────────────────────────────────────────────

func TestGetModelInfo_ExactMatch(t *testing.T) {
	result := GetModelInfo("gpt-5")
	if !strings.Contains(result, "GPT-5") {
		t.Error("expected 'GPT-5' in result")
	}
	if !strings.Contains(result, "OpenAI") {
		t.Error("expected 'OpenAI' in result")
	}
}

func TestGetModelInfo_CaseInsensitive(t *testing.T) {
	result := GetModelInfo("GPT-5")
	if !strings.Contains(result, "GPT-5") {
		t.Error("expected 'GPT-5' in result for case-insensitive lookup")
	}
}

func TestGetModelInfo_PartialMatch(t *testing.T) {
	result := GetModelInfo("opus-4-6")
	if !strings.Contains(result, "Claude Opus 4.6") {
		t.Error("expected 'Claude Opus 4.6' in result for partial match")
	}
}

func TestGetModelInfo_NotFound(t *testing.T) {
	result := GetModelInfo("nonexistent-model")
	if !strings.Contains(strings.ToLower(result), "not found") {
		t.Errorf("expected 'not found' in result, got: %s", result)
	}
}

// ── RecommendModel ────────────────────────────────────────────────────────

func TestRecommendModel_Coding(t *testing.T) {
	result := RecommendModel("coding", "")
	if !strings.Contains(result, "Recommendations for") {
		t.Error("expected 'Recommendations for' in result")
	}
	if !strings.Contains(result, "1.") {
		t.Error("expected numbered recommendations")
	}
}

func TestRecommendModel_Vision(t *testing.T) {
	result := RecommendModel("image analysis", "")
	if !strings.Contains(strings.ToLower(result), "vision") {
		t.Error("expected 'vision' mentioned in result")
	}
}

func TestRecommendModel_CheapBudget(t *testing.T) {
	result := RecommendModel("general tasks", "cheap")
	if !strings.Contains(result, "Budget:** cheap") {
		t.Error("expected 'Budget:** cheap' in result")
	}
}

func TestRecommendModel_Reasoning(t *testing.T) {
	result := RecommendModel("complex math reasoning", "")
	if !strings.Contains(strings.ToLower(result), "reasoning") {
		t.Error("expected 'reasoning' mentioned in result")
	}
}

// ── CheckModelStatus ──────────────────────────────────────────────────────

func TestCheckModelStatus_Current(t *testing.T) {
	result := CheckModelStatus("gpt-5")
	if !strings.Contains(strings.ToLower(result), "current") {
		t.Errorf("expected 'current' in result, got: %s", result)
	}
}

func TestCheckModelStatus_Legacy(t *testing.T) {
	result := CheckModelStatus("gpt-4.1")
	lower := strings.ToLower(result)
	if !strings.Contains(lower, "legacy") {
		t.Error("expected 'legacy' in result")
	}
	if !strings.Contains(lower, "replacement") {
		t.Error("expected 'replacement' suggestion for legacy model")
	}
}

func TestCheckModelStatus_Deprecated(t *testing.T) {
	result := CheckModelStatus("gpt-4o")
	if !strings.Contains(strings.ToLower(result), "deprecated") {
		t.Error("expected 'deprecated' in result")
	}
}

func TestCheckModelStatus_NotFound(t *testing.T) {
	result := CheckModelStatus("fake-model")
	if !strings.Contains(strings.ToLower(result), "not found") {
		t.Errorf("expected 'not found' in result, got: %s", result)
	}
}

// ── CompareModels ─────────────────────────────────────────────────────────

func TestCompareModels_Two(t *testing.T) {
	result := CompareModels([]string{"gpt-5", "claude-opus-4-6"})
	if !strings.Contains(result, "GPT-5") {
		t.Error("expected 'GPT-5' in comparison")
	}
	if !strings.Contains(result, "Claude Opus 4.6") {
		t.Error("expected 'Claude Opus 4.6' in comparison")
	}
}

func TestCompareModels_Three(t *testing.T) {
	result := CompareModels([]string{"gpt-5", "claude-opus-4-6", "gemini-2.5-pro"})
	if !strings.Contains(result, "GPT-5") {
		t.Error("expected 'GPT-5' in comparison")
	}
	if !strings.Contains(result, "Gemini 2.5 Pro") {
		t.Error("expected 'Gemini 2.5 Pro' in comparison")
	}
}

func TestCompareModels_SingleError(t *testing.T) {
	result := CompareModels([]string{"gpt-5"})
	if !strings.Contains(strings.ToLower(result), "at least 2") {
		t.Errorf("expected 'at least 2' error, got: %s", result)
	}
}

func TestCompareModels_NotFound(t *testing.T) {
	result := CompareModels([]string{"gpt-5", "nonexistent"})
	if !strings.Contains(strings.ToLower(result), "not found") {
		t.Errorf("expected 'not found' in result, got: %s", result)
	}
}

func TestCompareModels_CaseInsensitive(t *testing.T) {
	result := CompareModels([]string{"GPT-5", "CLAUDE-OPUS-4-6"})
	if !strings.Contains(result, "GPT-5") {
		t.Error("expected 'GPT-5' in case-insensitive comparison")
	}
	if !strings.Contains(result, "Claude Opus 4.6") {
		t.Error("expected 'Claude Opus 4.6' in case-insensitive comparison")
	}
}

// ── SearchModels ──────────────────────────────────────────────────────────

func TestSearchModels_ByProvider(t *testing.T) {
	result := SearchModels("OpenAI")
	if !strings.Contains(strings.ToLower(result), "gpt") {
		t.Error("expected 'gpt' models when searching for OpenAI")
	}
}

func TestSearchModels_ByName(t *testing.T) {
	result := SearchModels("Claude")
	if !strings.Contains(result, "Anthropic") {
		t.Error("expected 'Anthropic' when searching for Claude")
	}
}

func TestSearchModels_ByKeyword(t *testing.T) {
	result := SearchModels("flagship")
	if !strings.Contains(result, "|") {
		t.Error("expected table output for keyword 'flagship'")
	}
}

func TestSearchModels_CaseInsensitive(t *testing.T) {
	result := SearchModels("GEMINI")
	if !strings.Contains(result, "Google") {
		t.Error("expected 'Google' when searching for GEMINI")
	}
}

func TestSearchModels_NoResults(t *testing.T) {
	result := SearchModels("zzzznonexistent")
	if !strings.Contains(result, "No models found") {
		t.Errorf("expected 'No models found', got: %s", result)
	}
}

func TestSearchModels_PartialID(t *testing.T) {
	result := SearchModels("gpt-5")
	if !strings.Contains(result, "GPT-5") {
		t.Error("expected 'GPT-5' when searching by partial ID")
	}
}
