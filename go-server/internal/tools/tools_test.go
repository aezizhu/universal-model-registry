package tools

import (
	"fmt"
	"strings"
	"testing"
	"time"

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
		if line == "" || !strings.HasPrefix(line, "|") {
			continue // skip empty lines and footer instruction
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
	result := CheckModelStatus("o3-mini")
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

// ── Helper function tests ────────────────────────────────────────────────

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{128000, "128,000"},
		{1048576, "1,048,576"},
		{10000000, "10,000,000"},
	}
	for _, tc := range tests {
		got := models.FormatInt(tc.input)
		if got != tc.want {
			t.Errorf("FormatInt(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFormatInt_Negative(t *testing.T) {
	got := models.FormatInt(-1000)
	if got != "-1,000" {
		t.Errorf("FormatInt(-1000) = %q, want %q", got, "-1,000")
	}
}

func TestFormatTable_Empty(t *testing.T) {
	result := FormatTable(nil)
	if !strings.Contains(result, "No models found") {
		t.Errorf("expected 'No models found' for empty slice, got: %s", result)
	}
}

func TestFormatTable_SingleModel(t *testing.T) {
	ms := []models.Model{{
		ID:          "test-model",
		DisplayName: "Test Model",
		Provider:    "TestProvider",
		Status:      "current",
	}}
	result := FormatTable(ms)
	if !strings.Contains(result, "| test-model |") {
		t.Error("expected model ID in table")
	}
	if !strings.Contains(result, "Model ID") {
		t.Error("expected table header")
	}
}

func TestModelDetail_WithAllCapabilities(t *testing.T) {
	m := models.Model{
		ID:              "test-model",
		DisplayName:     "Test Model",
		Provider:        "TestProvider",
		ContextWindow:   200000,
		MaxOutputTokens: 4096,
		Vision:          true,
		Reasoning:       true,
		PricingInput:    1.0,
		PricingOutput:   5.0,
		KnowledgeCutoff: "2025-01",
		ReleaseDate:     "2025-01",
		Status:          "current",
		Notes:           "Test note",
	}
	result := ModelDetail(m)
	if !strings.Contains(result, "Vision") {
		t.Error("expected 'Vision' in detail")
	}
	if !strings.Contains(result, "Reasoning/Thinking") {
		t.Error("expected 'Reasoning/Thinking' in detail")
	}
	if !strings.Contains(result, "Test note") {
		t.Error("expected notes in detail")
	}
}

func TestModelDetail_NoCapabilities(t *testing.T) {
	m := models.Model{
		ID:          "test-model",
		DisplayName: "Test Model",
		Provider:    "TestProvider",
		Vision:      false,
		Reasoning:   false,
	}
	result := ModelDetail(m)
	if !strings.Contains(result, "None") {
		t.Error("expected 'None' for capabilities when neither vision nor reasoning")
	}
}

func TestFindModel_ExactMatch(t *testing.T) {
	m, found := FindModel("gpt-5")
	if !found {
		t.Fatal("expected to find gpt-5")
	}
	if m.ID != "gpt-5" {
		t.Errorf("expected ID 'gpt-5', got %q", m.ID)
	}
}

func TestFindModel_CaseInsensitive(t *testing.T) {
	m, found := FindModel("GPT-5")
	if !found {
		t.Fatal("expected to find GPT-5 via case-insensitive match")
	}
	if m.Provider != "OpenAI" {
		t.Errorf("expected provider 'OpenAI', got %q", m.Provider)
	}
}

func TestFindModel_NotFound(t *testing.T) {
	_, found := FindModel("nonexistent-model-xyz")
	if found {
		t.Error("did not expect to find nonexistent model")
	}
}

func TestFilterModels_CombinedFilters(t *testing.T) {
	results := FilterModels("OpenAI", "current", "vision")
	for _, m := range results {
		if m.Provider != "OpenAI" {
			t.Errorf("expected provider OpenAI, got %s", m.Provider)
		}
		if m.Status != "current" {
			t.Errorf("expected status current, got %s", m.Status)
		}
		if !m.Vision {
			t.Errorf("expected vision=true for model %s", m.ID)
		}
	}
	if len(results) == 0 {
		t.Error("expected at least one OpenAI current vision model")
	}
}

func TestFilterModels_UnknownCapability(t *testing.T) {
	all := FilterModels("", "", "")
	unknown := FilterModels("", "", "teleportation")
	// Unknown capability falls through to default (returns all)
	if len(unknown) != len(all) {
		t.Errorf("unknown capability should return all models, got %d vs %d", len(unknown), len(all))
	}
}

func TestFilterModels_ThinkingCapability(t *testing.T) {
	results := FilterModels("", "", "thinking")
	for _, m := range results {
		if !m.Reasoning {
			t.Errorf("model %s should have reasoning=true when filtering by thinking", m.ID)
		}
	}
	if len(results) == 0 {
		t.Error("expected at least one model with thinking/reasoning")
	}
}

func TestCaps_VisionOnly(t *testing.T) {
	m := models.Model{Vision: true, Reasoning: false}
	result := caps(m)
	if result != "Vision" {
		t.Errorf("expected 'Vision', got %q", result)
	}
}

func TestCaps_ReasoningOnly(t *testing.T) {
	m := models.Model{Vision: false, Reasoning: true}
	result := caps(m)
	if result != "Reasoning" {
		t.Errorf("expected 'Reasoning', got %q", result)
	}
}

func TestCaps_Both(t *testing.T) {
	m := models.Model{Vision: true, Reasoning: true}
	result := caps(m)
	if result != "Vision, Reasoning" {
		t.Errorf("expected 'Vision, Reasoning', got %q", result)
	}
}

func TestCaps_None(t *testing.T) {
	m := models.Model{Vision: false, Reasoning: false}
	result := caps(m)
	if result != "None" {
		t.Errorf("expected 'None', got %q", result)
	}
}

// ── Additional edge case tests ───────────────────────────────────────────

func TestGetModelInfo_EmptyString(t *testing.T) {
	result := GetModelInfo("")
	// Empty string may partial-match everything; just ensure no panic
	if result == "" {
		t.Error("expected non-empty result for empty input")
	}
}

func TestSearchModels_EmptyString(t *testing.T) {
	result := SearchModels("")
	// Empty query should return an error message prompting for a search term
	if !strings.Contains(result, "Please provide a search term") {
		t.Errorf("expected 'Please provide a search term' for empty query, got: %s", result)
	}
}

func TestSearchModels_SpecialCharacters(t *testing.T) {
	result := SearchModels("!@#$%^&*()")
	if !strings.Contains(result, "No models found") {
		t.Errorf("expected 'No models found' for special characters, got: %s", result)
	}
}

func TestCompareModels_EmptySlice(t *testing.T) {
	result := CompareModels([]string{})
	if !strings.Contains(strings.ToLower(result), "at least 2") {
		t.Errorf("expected 'at least 2' for empty slice, got: %s", result)
	}
}

func TestCompareModels_MoreThanFive(t *testing.T) {
	ids := []string{"gpt-5", "claude-opus-4-6", "gemini-2.5-pro", "grok-4", "deepseek-chat", "o3"}
	result := CompareModels(ids)
	// Should truncate to 5, so "o3" (6th) may or may not appear depending on ordering
	// but should not error
	if strings.Contains(strings.ToLower(result), "not found") {
		t.Error("should not report not found when truncating to 5")
	}
	if !strings.Contains(result, "Field") {
		t.Error("expected comparison table header")
	}
}

func TestCompareModels_DuplicateIDs(t *testing.T) {
	result := CompareModels([]string{"gpt-5", "gpt-5"})
	// Should work without error - comparing a model with itself
	if !strings.Contains(result, "GPT-5") {
		t.Error("expected 'GPT-5' in duplicate comparison")
	}
}

func TestListModels_CombinedProviderAndStatus(t *testing.T) {
	result := ListModels("OpenAI", "current", "")
	if strings.Contains(result, "deprecated") {
		t.Error("should not contain deprecated models when filtering for current")
	}
	if strings.Contains(result, "Anthropic") {
		t.Error("should not contain Anthropic models when filtering for OpenAI")
	}
}

func TestListModels_InvalidStatus(t *testing.T) {
	result := ListModels("", "invalid_status", "")
	if !strings.Contains(result, "No models found") {
		t.Errorf("expected 'No models found' for invalid status, got: %s", result)
	}
}

func TestRecommendModel_EmptyTask(t *testing.T) {
	result := RecommendModel("", "")
	// Should still return recommendations even with empty task
	if !strings.Contains(result, "Recommendations for") {
		t.Error("expected recommendations even for empty task")
	}
	if !strings.Contains(result, "1.") {
		t.Error("expected at least one recommendation")
	}
}

func TestRecommendModel_UnlimitedBudget(t *testing.T) {
	result := RecommendModel("general tasks", "unlimited")
	// "unlimited" normalizes to "expensive"
	if !strings.Contains(result, "Budget:** expensive") {
		t.Error("expected 'Budget:** expensive' in result (unlimited normalizes to expensive)")
	}
}

func TestRecommendModel_LongContext(t *testing.T) {
	result := RecommendModel("long context document analysis", "")
	if !strings.Contains(result, "1.") {
		t.Error("expected recommendations for long context task")
	}
}

func TestRecommendModel_OpenWeight(t *testing.T) {
	result := RecommendModel("open weight model for self-hosting", "")
	if !strings.Contains(result, "1.") {
		t.Error("expected recommendations for open weight task")
	}
}

func TestRecommendModel_LowBudgetAvoidsExpensive(t *testing.T) {
	result := RecommendModel("code generation", "low")
	// "low" should be treated as "cheap" — the top recommendations
	// must NOT include models costing > $5/M input.
	if strings.Contains(result, "gpt-5.2-pro") {
		t.Error("low budget should NOT recommend gpt-5.2-pro ($21/M)")
	}
	if strings.Contains(result, "o3-pro") {
		t.Error("low budget should NOT recommend o3-pro ($20/M)")
	}
	if strings.Contains(result, "o3-deep-research") {
		t.Error("low budget should NOT recommend o3-deep-research ($10/M)")
	}
}

func TestRecommendModel_BudgetNormalization(t *testing.T) {
	// "low" and "cheap" should produce the same results
	low := RecommendModel("general tasks", "low")
	cheap := RecommendModel("general tasks", "cheap")
	if low != cheap {
		t.Error("expected 'low' and 'cheap' budgets to produce identical results")
	}
	// "high" and "expensive" should produce the same results
	high := RecommendModel("general tasks", "high")
	expensive := RecommendModel("general tasks", "expensive")
	if high != expensive {
		t.Error("expected 'high' and 'expensive' budgets to produce identical results")
	}
}

func TestRecommendModel_CodingPrefersCodingModels(t *testing.T) {
	result := RecommendModel("coding tasks", "moderate")
	// At least one coding-specialized model should appear
	hasCodingModel := strings.Contains(result, "codex") ||
		strings.Contains(result, "devstral") ||
		strings.Contains(result, "kat-coder") ||
		strings.Contains(strings.ToLower(result), "code")
	if !hasCodingModel {
		t.Error("coding task should recommend at least one coding-specialized model")
	}
}

func TestCheckModelStatus_CaseInsensitive(t *testing.T) {
	result := CheckModelStatus("GPT-5")
	if !strings.Contains(strings.ToLower(result), "current") {
		t.Errorf("expected 'current' for case-insensitive GPT-5 lookup, got: %s", result)
	}
}

func TestSearchModels_SearchByNotes(t *testing.T) {
	result := SearchModels("flagship")
	if strings.Contains(result, "No models found") {
		t.Error("expected to find models with 'flagship' in notes")
	}
	if !strings.Contains(result, "|") {
		t.Error("expected table format in results")
	}
}

func TestSearchModels_SearchByStatus(t *testing.T) {
	// SearchModels searches ID, DisplayName, Provider, Status, and Notes
	result := SearchModels("deprecated")
	if strings.Contains(result, "No models found") {
		t.Error("expected to find deprecated models when searching by status")
	}
	if !strings.Contains(result, "deprecated") {
		t.Error("expected 'deprecated' in result when searching by status")
	}
}

func TestSearchModels_MultiWord(t *testing.T) {
	// Multi-word queries should match across different fields
	result := SearchModels("zhipu glm")
	if strings.Contains(result, "No models found") {
		t.Error("expected 'zhipu glm' to find Zhipu GLM models (provider + ID)")
	}
	if !strings.Contains(result, "glm-4.7") {
		t.Error("expected glm-4.7 in results for 'zhipu glm'")
	}
}

func TestSearchModels_VisionCapability(t *testing.T) {
	// "google vision" should find Google vision models via capability keyword injection
	result := SearchModels("google vision")
	if strings.Contains(result, "No models found") {
		t.Error("expected 'google vision' to find Google vision models")
	}
}

func TestSearchModels_ReasoningCapability(t *testing.T) {
	result := SearchModels("openai reasoning")
	if strings.Contains(result, "No models found") {
		t.Error("expected 'openai reasoning' to find OpenAI reasoning models")
	}
}

func TestSearchModels_ProviderAlternateNames(t *testing.T) {
	// z.ai should find Zhipu models via Notes field
	result := SearchModels("z.ai")
	if strings.Contains(result, "No models found") {
		t.Error("expected 'z.ai' to find Zhipu models")
	}
	// nim should find NVIDIA models via Notes field
	result = SearchModels("nim")
	if strings.Contains(result, "No models found") {
		t.Error("expected 'nim' to find NVIDIA models")
	}
}

func TestListModels_ProviderAlias(t *testing.T) {
	// "kimi" should resolve to Moonshot provider
	result := ListModels("kimi", "", "")
	if strings.Contains(result, "No models found") {
		t.Error("expected list_models(provider='kimi') to find Moonshot models")
	}
	if !strings.Contains(result, "kimi-k2.5") {
		t.Error("expected kimi-k2.5 in results for provider 'kimi'")
	}
}

func TestListModels_ProviderAliasZhipu(t *testing.T) {
	result := ListModels("z.ai", "", "")
	if strings.Contains(result, "No models found") {
		t.Error("expected list_models(provider='z.ai') to find Zhipu models")
	}
	if !strings.Contains(result, "glm-4.7") {
		t.Error("expected glm-4.7 in results for provider 'z.ai'")
	}
}

func TestListModels_ProviderAliasPhi(t *testing.T) {
	result := ListModels("phi", "", "")
	if strings.Contains(result, "No models found") {
		t.Error("expected list_models(provider='phi') to find Microsoft models")
	}
	if !strings.Contains(result, "phi-4") {
		t.Error("expected phi-4 in results for provider 'phi'")
	}
}

// ── Alias resolution tests ───────────────────────────────────────────

func TestFindModel_AliasResolution(t *testing.T) {
	m, found := FindModel("claude-sonnet-4-5")
	if !found {
		t.Fatal("expected to find model via alias 'claude-sonnet-4-5'")
	}
	if m.ID != "claude-sonnet-4-5-20250929" {
		t.Errorf("expected alias to resolve to 'claude-sonnet-4-5-20250929', got %q", m.ID)
	}
}

func TestFindModel_AliasResolution_Haiku(t *testing.T) {
	m, found := FindModel("claude-haiku-4-5")
	if !found {
		t.Fatal("expected to find model via alias 'claude-haiku-4-5'")
	}
	if m.ID != "claude-haiku-4-5-20251001" {
		t.Errorf("expected alias to resolve to 'claude-haiku-4-5-20251001', got %q", m.ID)
	}
}

func TestFindModel_AliasResolution_DateAlias(t *testing.T) {
	m, found := FindModel("gpt-4o-2024-05-13")
	if !found {
		t.Fatal("expected to find model via alias 'gpt-4o-2024-05-13'")
	}
	if m.ID != "gpt-4o" {
		t.Errorf("expected alias to resolve to 'gpt-4o', got %q", m.ID)
	}
}

// ── SuggestModels tests ──────────────────────────────────────────────

func TestSuggestModels_ClosestMatch(t *testing.T) {
	suggestions := SuggestModels("gpt-55", 3)
	if len(suggestions) == 0 {
		t.Fatal("expected at least one suggestion")
	}
	// "gpt-5" should be the closest match (edit distance 1)
	if suggestions[0] != "gpt-5" {
		t.Errorf("expected first suggestion to be 'gpt-5', got %q", suggestions[0])
	}
}

func TestSuggestModels_ReturnsRequestedCount(t *testing.T) {
	suggestions := SuggestModels("nonexistent", 5)
	if len(suggestions) != 5 {
		t.Errorf("expected 5 suggestions, got %d", len(suggestions))
	}
}

func TestSuggestModels_CaseInsensitive(t *testing.T) {
	lower := SuggestModels("GPT-55", 3)
	upper := SuggestModels("gpt-55", 3)
	if len(lower) != len(upper) {
		t.Fatal("case should not affect suggestion count")
	}
	for i := range lower {
		if lower[i] != upper[i] {
			t.Errorf("suggestion %d differs: %q vs %q", i, lower[i], upper[i])
		}
	}
}

// ── levenshteinDistance tests ─────────────────────────────────────────

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"kitten", "sitting", 3},
		{"abc", "xyz", 3},
		{"", "hello", 5},
		{"hello", "", 5},
	}
	for _, tc := range tests {
		got := levenshteinDistance(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

// ── FormatTable ★ marking tests ──────────────────────────────────────

func TestFormatTable_StarMarking(t *testing.T) {
	ms := []models.Model{
		{
			ID:          "old-model",
			DisplayName: "Old Model",
			Provider:    "TestProvider",
			Status:      "legacy",
			ReleaseDate: "2024-01",
		},
		{
			ID:          "new-model",
			DisplayName: "New Model",
			Provider:    "TestProvider",
			Status:      "current",
			ReleaseDate: "2025-06",
		},
	}
	result := FormatTable(ms)
	// The newest model should have ★
	if !strings.Contains(result, "★ new-model") {
		t.Error("expected ★ before newest model 'new-model'")
	}
	// The older model should NOT have ★
	if strings.Contains(result, "★ old-model") {
		t.Error("did not expect ★ before older model 'old-model'")
	}
}

func TestFormatTable_UseInCodeFooter(t *testing.T) {
	ms := []models.Model{
		{
			ID:          "model-a",
			DisplayName: "Model A",
			Provider:    "ProviderA",
			Status:      "current",
			ReleaseDate: "2025-06",
		},
		{
			ID:          "model-b",
			DisplayName: "Model B",
			Provider:    "ProviderB",
			Status:      "current",
			ReleaseDate: "2025-03",
		},
	}
	result := FormatTable(ms)
	if !strings.Contains(result, "USE IN CODE:") {
		t.Error("expected 'USE IN CODE:' in FormatTable footer")
	}
	// Both are newest for their respective providers, so both should appear
	if !strings.Contains(result, "model-a") {
		t.Error("expected 'model-a' in USE IN CODE footer")
	}
	if !strings.Contains(result, "model-b") {
		t.Error("expected 'model-b' in USE IN CODE footer")
	}
}

// ── CompareModels field completeness test ────────────────────────────

func TestCompareModels_FieldCompleteness(t *testing.T) {
	result := CompareModels([]string{"gpt-5", "claude-opus-4-6"})
	requiredFields := []string{
		"Provider",
		"Status",
		"Context",
		"Max Output",
		"Capabilities",
		"Input $/1M",
		"Output $/1M",
		"Knowledge Cutoff",
		"Release Date",
	}
	for _, field := range requiredFields {
		if !strings.Contains(result, "| "+field+" |") {
			t.Errorf("CompareModels output missing field row %q", field)
		}
	}
}

// ── recencyBonus tests ───────────────────────────────────────────────

func TestRecencyBonus(t *testing.T) {
	now := time.Now()

	fmtDate := func(t time.Time) string {
		return fmt.Sprintf("%d-%02d", t.Year(), t.Month())
	}

	tests := []struct {
		name        string
		releaseDate string
		want        float64
	}{
		{"this month", fmtDate(now), 1.5},
		{"3 months ago (within ≤6 window)", fmtDate(now.AddDate(0, -3, 0)), 1.5},
		{"12 months ago (mid-decay)", fmtDate(now.AddDate(0, -12, 0)), 0.75},
		{"24 months ago (beyond 18-month cutoff)", fmtDate(now.AddDate(0, -24, 0)), 0},
		{"invalid date", "abc", 0},
		{"empty date", "", 0},
	}
	for _, tc := range tests {
		got := recencyBonus(tc.releaseDate)
		if got != tc.want {
			t.Errorf("recencyBonus(%q) [%s] = %.4f, want %.4f", tc.releaseDate, tc.name, got, tc.want)
		}
	}
}
