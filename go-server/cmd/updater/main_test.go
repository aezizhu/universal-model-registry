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
	// Count only non-deprecated models in models.Models.
	// Deprecated models are intentionally excluded from knownModels
	// so the updater doesn't flag them as "MISSING" every run.
	want := 0
	for _, m := range models.Models {
		if m.Status != "deprecated" {
			want++
		}
	}
	if total != want {
		t.Errorf("knownModels total entries = %d, non-deprecated models.Models has %d entries", total, want)
	}
}

func TestKnownModels_AllProvidersPresent(t *testing.T) {
	expected := []string{
		"OpenAI", "Anthropic", "Google", "xAI", "Mistral", "DeepSeek",
		"Meta", "Amazon", "Cohere", "Perplexity", "AI21",
		"Moonshot", "Zhipu", "NVIDIA", "Tencent", "Microsoft", "MiniMax", "Xiaomi", "Kuaishou",
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

// ---------------------------------------------------------------------------
// isDateStampVariant tests
// ---------------------------------------------------------------------------

func TestIsDateStampVariant(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		// YYYYMMDD format
		{"gpt-4.1-20250414", true},
		{"claude-sonnet-4-5-20250929", true},
		{"o3-mini-20250131", true},
		// YYYY-MM-DD format
		{"gpt-5-2025-08-07", true},
		{"gpt-5-mini-2025-08-07", true},
		// Not date stamps
		{"gpt-5.2", false},
		{"o3-mini", false},
		{"gpt-5-nano", false},
		{"mistral-large-2512", false}, // 4 digits, not 8
		{"codestral-2508", false},     // 4 digits
		{"gpt-4o", false},
		{"", false},
	}
	for _, tt := range tests {
		got := isDateStampVariant(tt.id)
		if got != tt.want {
			t.Errorf("isDateStampVariant(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// isAllDigits tests
// ---------------------------------------------------------------------------

func TestIsAllDigits(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"0", true},
		{"12345", true},
		{"2508", true},
		{"20250414", true},
		{"", false},
		{"12a3", false},
		{"abc", false},
		{"-1", false},
		{"12.3", false},
	}
	for _, tt := range tests {
		got := isAllDigits(tt.s)
		if got != tt.want {
			t.Errorf("isAllDigits(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// isKnownAlias tests
// ---------------------------------------------------------------------------

func TestIsKnownAlias_PrefixDigitSuffix(t *testing.T) {
	// Heuristic 1: id is a prefix of a known ID whose remaining suffix is all-digits.
	known := map[string]bool{
		"gpt-5-mini-2025": true,
		"gpt-5":           true,
	}
	if !isKnownAlias("gpt-5-mini", known) {
		t.Error("expected gpt-5-mini to be recognized as prefix of gpt-5-mini-2025")
	}
}

func TestIsKnownAlias_AliasSuffix(t *testing.T) {
	// Heuristic 2: id extends a known ID with a well-known alias suffix.
	known := map[string]bool{"gpt-5": true}
	cases := []struct {
		id   string
		want bool
	}{
		{"gpt-5-latest", true},
		{"gpt-5-beta", true},
		{"gpt-5-preview", true},
		{"gpt-5-chat-latest", true},
		{"gpt-5-reasoning", true},
		{"gpt-5-non-reasoning", true},
		{"gpt-5-audio-preview", false}, // not in aliasSuffixes
		{"gpt-5-turbo", false},
	}
	for _, tt := range cases {
		got := isKnownAlias(tt.id, known)
		if got != tt.want {
			t.Errorf("isKnownAlias(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

func TestIsKnownAlias_NumericVariant(t *testing.T) {
	// Heuristic 3: shared base name with ≥2-digit numeric suffixes.
	known := map[string]bool{
		"codestral-2508":       true,
		"mistral-large-2512":   true,
		"magistral-small-2509": true,
	}
	cases := []struct {
		id   string
		want bool
	}{
		{"codestral-2405", true},
		{"codestral-2501", true},
		{"codestral-25", true},           // 2-digit suffix still matches base "codestral"
		{"mistral-large-2407", true},
		{"magistral-small-2506", true},
		{"mistral-small-2402", false},    // base "mistral-small" ≠ "mistral-large"
		{"devstral-2507", false},         // no known model with base "devstral"
		{"codestral-2", false},           // 1-digit suffix too short
	}
	for _, tt := range cases {
		got := isKnownAlias(tt.id, known)
		if got != tt.want {
			t.Errorf("isKnownAlias(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

func TestIsKnownAlias_ExactMatchIsNotAlias(t *testing.T) {
	known := map[string]bool{"gpt-5": true}
	if isKnownAlias("gpt-5", known) {
		t.Error("exact match should not be treated as alias")
	}
}

func TestIsKnownAlias_EmptyKnown(t *testing.T) {
	if isKnownAlias("gpt-5-latest", map[string]bool{}) {
		t.Error("should return false with empty known set")
	}
}

// ---------------------------------------------------------------------------
// diff() filtering integration tests
// ---------------------------------------------------------------------------

func TestDiff_FiltersDateStamps(t *testing.T) {
	known := map[string]bool{"gpt-5": true}
	docIDs := []string{"gpt-5", "gpt-5-20250807"}

	newModels, _ := diff(known, docIDs)
	if len(newModels) != 0 {
		t.Errorf("date-stamped variant should be filtered, got newModels = %v", newModels)
	}
}

func TestDiff_FiltersAliases(t *testing.T) {
	known := map[string]bool{"gpt-5": true}
	docIDs := []string{"gpt-5", "gpt-5-chat-latest", "gpt-5-latest"}

	newModels, _ := diff(known, docIDs)
	if len(newModels) != 0 {
		t.Errorf("alias variants should be filtered, got newModels = %v", newModels)
	}
}

func TestDiff_FiltersNumericVariants(t *testing.T) {
	known := map[string]bool{"codestral-2508": true}
	docIDs := []string{"codestral-2405", "codestral-2501"}

	newModels, _ := diff(known, docIDs)
	if len(newModels) != 0 {
		t.Errorf("numeric variants should be filtered, got newModels = %v", newModels)
	}
}

func TestDiff_KeepsGenuineNewModels(t *testing.T) {
	known := map[string]bool{"gpt-5": true}
	docIDs := []string{"gpt-5", "gpt-6"}

	newModels, _ := diff(known, docIDs)
	if len(newModels) != 1 || newModels[0] != "gpt-6" {
		t.Errorf("genuinely new model should appear, got newModels = %v", newModels)
	}
}

// ---------------------------------------------------------------------------
// OpenAI ExcludePattern verification (PR #4 review checklist)
// ---------------------------------------------------------------------------

func TestOpenAIExcludePattern(t *testing.T) {
	re := docSources["OpenAI"].ExcludePattern

	// These should be EXCLUDED (regex matches → filtered out as legacy).
	shouldExclude := []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4-turbo",
		"o1",
		"o1-mini",
		"o1-preview",
	}
	for _, id := range shouldExclude {
		if !re.MatchString(id) {
			t.Errorf("ExcludePattern should match %q (legacy model), but it did not", id)
		}
	}

	// gpt-4o and gpt-4o-mini are now deprecated and should be excluded.
	shouldAlsoExclude := []string{
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4o-2024-08-06",
	}
	for _, id := range shouldAlsoExclude {
		if !re.MatchString(id) {
			t.Errorf("ExcludePattern should match %q (deprecated model), but it did not", id)
		}
	}

	// These should NOT be excluded (regex must not match → kept as current).
	shouldKeep := []string{
		"gpt-4.1",
		"gpt-4.1-mini",
		"gpt-4.1-nano",
		"gpt-5",
		"gpt-5.1",
		"gpt-5.2",
		"o3",
		"o3-pro",
		"o3-mini",
		"o4-mini",
	}
	for _, id := range shouldKeep {
		if re.MatchString(id) {
			t.Errorf("ExcludePattern should NOT match %q (current model), but it did", id)
		}
	}
}

// ---------------------------------------------------------------------------
// xAI NormalizeRe verification (PR #4 review checklist)
// ---------------------------------------------------------------------------

func TestXAINormalizeRe(t *testing.T) {
	src := docSources["xAI"]
	re := src.NormalizeRe
	repl := src.NormalizeRepl

	tests := []struct {
		input string
		want  string
	}{
		// Single-digit-dash-single-digit followed by non-digit → should normalize.
		{"grok-4-1-fast", "grok-4.1-fast"},
		{"grok-4-1", "grok-4.1"},
		// 4-digit date suffix: 4-0 is followed by digit 7, so no match → unchanged.
		{"grok-4-0709", "grok-4-0709"},
		// No digit-dash-digit pattern → unchanged.
		{"grok-3-mini", "grok-3-mini"},
		{"grok-code-fast-1", "grok-code-fast-1"},
	}
	for _, tt := range tests {
		got := re.ReplaceAllString(tt.input, repl)
		if got != tt.want {
			t.Errorf("NormalizeRe(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// isKnownAlias: additional coverage for bare base model (PR #4 review checklist)
// ---------------------------------------------------------------------------

func TestIsKnownAlias_NumericVariant_BareBase(t *testing.T) {
	// Heuristic 3, first branch: known set contains the bare base (no numeric suffix).
	known := map[string]bool{"devstral": true}
	if !isKnownAlias("devstral-2507", known) {
		t.Error("expected devstral-2507 to be recognized as numeric variant of bare base devstral")
	}
}

// ---------------------------------------------------------------------------
// docSources ExcludePattern / NormalizeRe field presence sanity checks
// ---------------------------------------------------------------------------

func TestDocSources_OpenAIHasExcludePattern(t *testing.T) {
	src, ok := docSources["OpenAI"]
	if !ok {
		t.Fatal("OpenAI missing from docSources")
	}
	if src.ExcludePattern == nil {
		t.Fatal("OpenAI ExcludePattern is nil")
	}
}

func TestDocSources_XAIHasNormalizeRe(t *testing.T) {
	src, ok := docSources["xAI"]
	if !ok {
		t.Fatal("xAI missing from docSources")
	}
	if src.NormalizeRe == nil {
		t.Fatal("xAI NormalizeRe is nil")
	}
}

// ---------------------------------------------------------------------------
// xAI pattern + normalization end-to-end: ensure extracting from raw HTML-like
// content and normalizing produces the expected model IDs.
// ---------------------------------------------------------------------------

func TestXAIPatternAndNormalize_EndToEnd(t *testing.T) {
	src := docSources["xAI"]
	// Simulate HTML snippets containing model IDs in various forms.
	content := `"grok-4-1-fast" and "grok-4-0709" and "grok-3-mini"`

	matches := src.Pattern.FindAllStringSubmatch(content, -1)
	var ids []string
	for _, m := range matches {
		if len(m) >= 2 {
			ids = append(ids, m[1])
		}
	}
	// Apply normalization.
	for i, id := range ids {
		ids[i] = src.NormalizeRe.ReplaceAllString(id, src.NormalizeRepl)
	}

	// Verify expected output.
	expected := map[string]bool{
		"grok-4.1-fast": true, // normalized from grok-4-1-fast
		"grok-4-0709":   true, // unchanged (date suffix)
		"grok-3-mini":   true, // unchanged (non-digit after 3)
	}
	for _, id := range ids {
		if !expected[id] {
			t.Errorf("unexpected normalized ID %q", id)
		}
		delete(expected, id)
	}
	for id := range expected {
		t.Errorf("expected ID %q not found in output", id)
	}
}

// ---------------------------------------------------------------------------
// Verify xAI ExcludePattern filters image/vision models
// ---------------------------------------------------------------------------

func TestXAIExcludePattern(t *testing.T) {
	re := docSources["xAI"].ExcludePattern

	shouldExclude := []string{
		"grok-2-vision-1212",
		"grok-image-gen",
		"grok-imagine-1",
		"grok-2-video-gen",
	}
	for _, id := range shouldExclude {
		if !re.MatchString(id) {
			t.Errorf("xAI ExcludePattern should match %q, but it did not", id)
		}
	}

	shouldKeep := []string{
		"grok-4",
		"grok-4-1-fast",
		"grok-3-mini",
		"grok-code-fast-1",
	}
	for _, id := range shouldKeep {
		if re.MatchString(id) {
			t.Errorf("xAI ExcludePattern should NOT match %q, but it did", id)
		}
	}
}

// ---------------------------------------------------------------------------
// Verify OpenAI pattern extracts expected model IDs
// ---------------------------------------------------------------------------

func TestOpenAIPattern(t *testing.T) {
	re := docSources["OpenAI"].Pattern

	// Simulated content from OpenAI's chat_model.py
	content := `"gpt-4o", "gpt-4o-mini", "gpt-5", "gpt-3.5-turbo", "o1-mini", "o3-pro", "o4-mini"`

	matches := re.FindAllStringSubmatch(content, -1)
	found := make(map[string]bool)
	for _, m := range matches {
		if len(m) >= 2 {
			found[m[1]] = true
		}
	}

	expected := []string{"gpt-4o", "gpt-4o-mini", "gpt-5", "gpt-3.5-turbo", "o1-mini", "o3-pro", "o4-mini"}
	for _, id := range expected {
		if !found[id] {
			t.Errorf("OpenAI Pattern should extract %q, but did not", id)
		}
	}
}

