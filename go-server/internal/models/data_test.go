package models

import "testing"

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
	if len(Models) != 42 {
		t.Errorf("expected 42 models, got %d", len(Models))
	}
}

func TestProviderCounts(t *testing.T) {
	counts := make(map[string]int)
	for _, m := range Models {
		counts[m.Provider]++
	}

	expected := map[string]int{
		"OpenAI":    14,
		"Anthropic": 8,
		"Google":    6,
		"xAI":      4,
		"Meta":     3,
		"Mistral":  4,
		"DeepSeek": 3,
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
