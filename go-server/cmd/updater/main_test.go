package main

import (
	"sort"
	"testing"

	"go-server/internal/models"
)

// ---------------------------------------------------------------------------
// diff() tests
// ---------------------------------------------------------------------------

func TestDiff_NewModels(t *testing.T) {
	known := map[string]bool{"a": true, "b": true}
	apiIDs := []string{"a", "b", "c", "d"}

	newModels, missing := diff(known, apiIDs)

	sort.Strings(newModels)
	if len(newModels) != 2 || newModels[0] != "c" || newModels[1] != "d" {
		t.Errorf("expected new=[c d], got %v", newModels)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing, got %v", missing)
	}
}

func TestDiff_RemovedModels(t *testing.T) {
	known := map[string]bool{"a": true, "b": true, "c": true}
	apiIDs := []string{"a"}

	newModels, missing := diff(known, apiIDs)

	sort.Strings(missing)
	if len(newModels) != 0 {
		t.Errorf("expected no new models, got %v", newModels)
	}
	if len(missing) != 2 || missing[0] != "b" || missing[1] != "c" {
		t.Errorf("expected missing=[b c], got %v", missing)
	}
}

func TestDiff_NoChanges(t *testing.T) {
	known := map[string]bool{"x": true, "y": true, "z": true}
	apiIDs := []string{"x", "y", "z"}

	newModels, missing := diff(known, apiIDs)

	if len(newModels) != 0 {
		t.Errorf("expected no new models, got %v", newModels)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing, got %v", missing)
	}
}

func TestDiff_EmptyAPIResponse(t *testing.T) {
	known := map[string]bool{"a": true, "b": true}
	var apiIDs []string

	newModels, missing := diff(known, apiIDs)

	sort.Strings(missing)
	if len(newModels) != 0 {
		t.Errorf("expected no new models, got %v", newModels)
	}
	if len(missing) != 2 || missing[0] != "a" || missing[1] != "b" {
		t.Errorf("expected missing=[a b], got %v", missing)
	}
}

func TestDiff_EmptyKnownModels(t *testing.T) {
	known := map[string]bool{}
	apiIDs := []string{"m1", "m2", "m3"}

	newModels, missing := diff(known, apiIDs)

	if len(newModels) != 3 {
		t.Errorf("expected 3 new models, got %d: %v", len(newModels), newModels)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing, got %v", missing)
	}
}

// ---------------------------------------------------------------------------
// knownModels ↔ models.Models cross-reference tests
// ---------------------------------------------------------------------------

func TestKnownModels_MatchDataGo(t *testing.T) {
	for provider, ids := range knownModels {
		for id := range ids {
			m, ok := models.Models[id]
			if !ok {
				t.Errorf("knownModels[%s] has %q but it is missing from models.Models", provider, id)
				continue
			}
			if m.Provider != provider {
				t.Errorf("knownModels[%s] has %q but models.Models[%q].Provider = %q", provider, id, id, m.Provider)
			}
		}
	}
}

func TestKnownModels_CompleteCount(t *testing.T) {
	total := 0
	for _, ids := range knownModels {
		total += len(ids)
	}
	want := len(models.Models)
	if total != want {
		t.Errorf("knownModels total entries = %d, models.Models has %d entries", total, want)
	}
}

func TestKnownModels_AllProvidersPresent(t *testing.T) {
	expected := []string{
		"OpenAI", "Anthropic", "Google", "xAI", "Mistral", "DeepSeek",
		"Meta", "Amazon", "Cohere", "Perplexity", "AI21",
	}
	for _, p := range expected {
		ids, ok := knownModels[p]
		if !ok {
			t.Errorf("provider %q missing from knownModels", p)
			continue
		}
		if len(ids) == 0 {
			t.Errorf("provider %q has 0 entries in knownModels", p)
		}
	}
	if len(knownModels) != len(expected) {
		t.Errorf("knownModels has %d providers, expected %d", len(knownModels), len(expected))
	}
}
