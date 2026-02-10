package models

import (
	"regexp"
	"testing"
)

func TestAllModelsHaveRequiredFields(t *testing.T) {
	for key, m := range Models {
		if m.ID == "" {
			t.Errorf("%s: missing ID", key)
		}
		if m.DisplayName == "" {
			t.Errorf("%s: missing DisplayName", key)
		}
		if m.Provider == "" {
			t.Errorf("%s: missing Provider", key)
		}
		if m.ContextWindow == 0 {
			t.Errorf("%s: ContextWindow is zero", key)
		}
		if m.MaxOutputTokens == 0 {
			t.Errorf("%s: MaxOutputTokens is zero", key)
		}
		if m.KnowledgeCutoff == "" {
			t.Errorf("%s: missing KnowledgeCutoff", key)
		}
		if m.ReleaseDate == "" {
			t.Errorf("%s: missing ReleaseDate", key)
		}
		if m.Status == "" {
			t.Errorf("%s: missing Status", key)
		}
		if m.Notes == "" {
			t.Errorf("%s: missing Notes", key)
		}
	}
}

func TestModelIDMatchesMapKey(t *testing.T) {
	for key, m := range Models {
		if key != m.ID {
			t.Errorf("map key %q != model ID %q", key, m.ID)
		}
	}
}

func TestStatusValuesAreValid(t *testing.T) {
	valid := map[string]bool{
		"current":    true,
		"legacy":     true,
		"deprecated": true,
	}
	for key, m := range Models {
		if !valid[m.Status] {
			t.Errorf("%s: invalid status %q", key, m.Status)
		}
	}
}

func TestPricingIsNonNegative(t *testing.T) {
	for key, m := range Models {
		if m.PricingInput < 0 {
			t.Errorf("%s: negative input pricing %f", key, m.PricingInput)
		}
		if m.PricingOutput < 0 {
			t.Errorf("%s: negative output pricing %f", key, m.PricingOutput)
		}
	}
}

func TestContextWindowIsPositive(t *testing.T) {
	for key, m := range Models {
		if m.ContextWindow <= 0 {
			t.Errorf("%s: non-positive context window %d", key, m.ContextWindow)
		}
	}
}

func TestAtLeastThreeProviders(t *testing.T) {
	providers := make(map[string]bool)
	for _, m := range Models {
		providers[m.Provider] = true
	}
	if len(providers) < 3 {
		t.Errorf("expected at least 3 providers, got %d", len(providers))
	}
}

func TestTotalModelCount(t *testing.T) {
	if len(Models) != 64 {
		t.Errorf("expected 64 models, got %d", len(Models))
	}
}

func TestProviderCounts(t *testing.T) {
	counts := make(map[string]int)
	for _, m := range Models {
		counts[m.Provider]++
	}

	expected := map[string]int{
		"OpenAI":     16,
		"Anthropic":  8,
		"Google":     6,
		"xAI":       6,
		"Meta":      3,
		"Mistral":   6,
		"DeepSeek":  4,
		"Amazon":    6,
		"Cohere":    4,
		"Perplexity": 3,
		"AI21":      2,
	}

	for provider, want := range expected {
		got := counts[provider]
		if got != want {
			t.Errorf("provider %s: expected %d models, got %d", provider, want, got)
		}
	}

	// Also verify no unexpected providers
	for provider := range counts {
		if _, ok := expected[provider]; !ok {
			t.Errorf("unexpected provider %q with %d models", provider, counts[provider])
		}
	}
}

func TestMaxOutputTokensIsPositive(t *testing.T) {
	for key, m := range Models {
		if m.MaxOutputTokens <= 0 {
			t.Errorf("%s: non-positive MaxOutputTokens %d", key, m.MaxOutputTokens)
		}
	}
}

func TestMaxOutputDoesNotExceedContext(t *testing.T) {
	for key, m := range Models {
		if m.MaxOutputTokens > m.ContextWindow {
			t.Errorf("%s: MaxOutputTokens (%d) > ContextWindow (%d)", key, m.MaxOutputTokens, m.ContextWindow)
		}
	}
}

func TestOutputPricingAtLeastInputPricing(t *testing.T) {
	for key, m := range Models {
		if m.PricingOutput < m.PricingInput {
			t.Errorf("%s: output pricing $%.2f < input pricing $%.2f", key, m.PricingOutput, m.PricingInput)
		}
	}
}

func TestDateFormats(t *testing.T) {
	dateRe := regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)
	for key, m := range Models {
		if !dateRe.MatchString(m.KnowledgeCutoff) {
			t.Errorf("%s: KnowledgeCutoff %q does not match YYYY-MM format", key, m.KnowledgeCutoff)
		}
		if !dateRe.MatchString(m.ReleaseDate) {
			t.Errorf("%s: ReleaseDate %q does not match YYYY-MM format", key, m.ReleaseDate)
		}
	}
}

func TestNoDuplicateDisplayNames(t *testing.T) {
	seen := make(map[string]string) // displayName -> first model ID
	for key, m := range Models {
		if prev, ok := seen[m.DisplayName]; ok {
			t.Errorf("duplicate DisplayName %q: used by both %s and %s", m.DisplayName, prev, key)
		}
		seen[m.DisplayName] = key
	}
}

func TestAliasesPointToValidModels(t *testing.T) {
	for alias, target := range Aliases {
		if _, ok := Models[target]; !ok {
			t.Errorf("alias %q points to non-existent model %q", alias, target)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{1234, "1,234"},
		{100000, "100,000"},
		{1000000, "1,000,000"},
		{10000000, "10,000,000"},
		{-1, "-1"},
		{-1000, "-1,000"},
		{-1234567, "-1,234,567"},
	}
	for _, tt := range tests {
		got := FormatInt(tt.input)
		if got != tt.want {
			t.Errorf("FormatInt(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEveryProviderHasAtLeastOneCurrentModel(t *testing.T) {
	providers := make(map[string]bool)
	currentProviders := make(map[string]bool)
	for _, m := range Models {
		providers[m.Provider] = true
		if m.Status == "current" {
			currentProviders[m.Provider] = true
		}
	}
	for p := range providers {
		if !currentProviders[p] {
			t.Errorf("provider %s has no current models", p)
		}
	}
}
