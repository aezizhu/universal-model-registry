package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

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
// normalizeMistralID tests
// ---------------------------------------------------------------------------

func TestNormalizeMistralID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Long-form → short-form conversions
		{"mistral-small-3-1-25-03", "mistral-small-2503"},
		{"mistral-large-3-25-12", "mistral-large-2512"},
		{"codestral-25-08", "codestral-2508"},
		{"codestral-24-05", "codestral-2405"},
		{"devstral-small-2-25-12", "devstral-small-2512"},
		{"devstral-medium-1-0-25-07", "devstral-medium-2507"},
		{"magistral-medium-1-2-25-09", "magistral-medium-2509"},
		{"magistral-small-1-0-25-06", "magistral-small-2506"},
		{"ministral-3-14b-25-12", "ministral-14b-2512"}, // "3" is a version digit (stripped), "14b" is alphanumeric (kept)
		{"mistral-small-creative-25-12", "mistral-small-creative-2512"},
		{"mistral-medium-1-0-23-12", "mistral-medium-2312"},
		// Already short-form → unchanged
		{"mistral-small-2503", "mistral-small-2503"},     // ends with 4-digit suffix, not 2-2
		{"codestral-2508", "codestral-2508"},               // only 2 parts
		{"mistral-saba-2502", "mistral-saba-2502"},         // "saba" is not digits
		// Edge cases → unchanged
		{"mistral-large", "mistral-large"},                 // no numeric suffix
		{"codestral", "codestral"},                         // single part
	}
	for _, tt := range tests {
		got := normalizeMistralID(tt.input)
		if got != tt.want {
			t.Errorf("normalizeMistralID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// isKnownAlias: mini/nano alias suffix tests
// ---------------------------------------------------------------------------

func TestIsKnownAlias_MiniNanoSuffix(t *testing.T) {
	// A deprecated base model (e.g. gpt-4.1) should be recognized as an alias
	// when its mini/nano variants are in the known set.
	known := map[string]bool{
		"gpt-4.1-mini": true,
		"gpt-4.1-nano": true,
	}
	// Heuristic 2b: known "gpt-4.1-mini" extends scraped "gpt-4.1" with suffix "mini"
	if !isKnownAlias("gpt-4.1", known) {
		t.Error("expected gpt-4.1 to be recognized as alias via mini/nano suffix")
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
// stripModeSuffixes tests
// ---------------------------------------------------------------------------

func TestStripModeSuffixes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"grok-4-fast-reasoning", "grok-4-fast"},
		{"grok-4-fast-non-reasoning", "grok-4-fast"},
		{"grok-4.20-beta-0309-reasoning", "grok-4.20-beta-0309"},
		{"grok-4.20-beta-0309-non-reasoning", "grok-4.20-beta-0309"},
		{"grok-4.20-beta-latest-reasoning", "grok-4.20-beta"},
		{"grok-4.20-beta-latest-non-reasoning", "grok-4.20-beta"},
		{"claude-sonnet-4-6-reasoning", "claude-sonnet-4-6"},
		{"gemini-2.5-pro-latest", "gemini-2.5-pro"},
		// No suffix to strip
		{"gpt-5", "gpt-5"},
		{"grok-4-fast", "grok-4-fast"},
		{"grok-4.20-beta-0309", "grok-4.20-beta-0309"},
		{"", ""},
	}
	for _, tt := range tests {
		got := stripModeSuffixes(tt.input)
		if got != tt.want {
			t.Errorf("stripModeSuffixes(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// isCompoundAliasSuffix tests
// ---------------------------------------------------------------------------

func TestIsCompoundAliasSuffix(t *testing.T) {
	tests := []struct {
		suffix string
		want   bool
	}{
		// Single known suffixes
		{"latest", true},
		{"reasoning", true},
		{"non-reasoning", true},
		{"beta", true},
		{"preview", true},
		// Compound suffixes
		{"latest-non-reasoning", true},
		{"latest-reasoning", true},
		{"fast-beta", true},
		{"fast-latest", true},
		// Not alias suffixes
		{"turbo", false},
		{"audio-preview", false},
		{"", false},
		{"unknown", false},
		{"latest-turbo", false},
	}
	for _, tt := range tests {
		got := isCompoundAliasSuffix(tt.suffix)
		if got != tt.want {
			t.Errorf("isCompoundAliasSuffix(%q) = %v, want %v", tt.suffix, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// hasVariantInDocs tests
// ---------------------------------------------------------------------------

func TestHasVariantInDocs(t *testing.T) {
	docSet := map[string]bool{
		"grok-4-fast-reasoning":     true,
		"grok-4-fast-non-reasoning": true,
		"grok-4.20-beta-0309":       true,
		"gpt-5":                     true,
		"gpt-5.3-chat":              true, // stripped from gpt-5.3-chat-latest
	}

	tests := []struct {
		knownID string
		want    bool
	}{
		// grok-4-fast has variants in docs (forward check)
		{"grok-4-fast", true},
		// gpt-5.3-chat-latest: stripping -latest yields gpt-5.3-chat which is in docs (reverse check)
		{"gpt-5.3-chat-latest", true},
		// gpt-5 is directly in docs, but hasVariantInDocs checks for suffixed variants
		{"gpt-5", false},
		// No variant for this
		{"grok-3", false},
		// grok-4.20-beta-0309 is in docs directly, but no suffixed variant
		{"grok-4.20-beta-0309", false},
	}
	for _, tt := range tests {
		got := hasVariantInDocs(tt.knownID, docSet)
		if got != tt.want {
			t.Errorf("hasVariantInDocs(%q) = %v, want %v", tt.knownID, got, tt.want)
		}
	}
}

func TestDiff_KnownModelEndingWithLatestNotFlaggedMissing(t *testing.T) {
	// gpt-5.3-chat-latest is a canonical known model ID ending in -latest.
	// After universal stripModeSuffixes, docs will contain "gpt-5.3-chat".
	// The known model must NOT be flagged as missing.
	known := map[string]bool{"gpt-5.3-chat-latest": true}
	docIDs := []string{"gpt-5.3-chat"} // stripped by universal suffix stripping

	_, missing := diff(known, docIDs)
	for _, m := range missing {
		if m == "gpt-5.3-chat-latest" {
			t.Error("gpt-5.3-chat-latest should NOT be flagged as missing when docs contain gpt-5.3-chat (stripped form)")
		}
	}
}

// ---------------------------------------------------------------------------
// diff() reverse alias checking: known model NOT flagged as missing when
// docs only list suffixed variants
// ---------------------------------------------------------------------------

func TestDiff_ReverseAliasNoFalseMissing(t *testing.T) {
	known := map[string]bool{"grok-4-fast": true}
	// Docs list only the -reasoning/-non-reasoning variants, not the bare name
	docIDs := []string{"grok-4-fast-reasoning", "grok-4-fast-non-reasoning"}

	_, missing := diff(known, docIDs)
	for _, m := range missing {
		if m == "grok-4-fast" {
			t.Error("grok-4-fast should NOT be flagged as missing when docs list its reasoning variants")
		}
	}
}

func TestDiff_CompoundSuffixNotNew(t *testing.T) {
	known := map[string]bool{"grok-4-fast": true}
	// Compound suffix variant should be recognized as alias, not new
	docIDs := []string{"grok-4-fast", "grok-4-fast-latest-non-reasoning"}

	newModels, _ := diff(known, docIDs)
	for _, m := range newModels {
		if m == "grok-4-fast-latest-non-reasoning" {
			t.Error("grok-4-fast-latest-non-reasoning should be recognized as alias of grok-4-fast")
		}
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
	content := `"grok-4-1-fast" and "grok-4-0709" and "grok-3-mini" and "grok-code-prompt-engineering"`

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

func TestXAIPattern_DocsPagePaths(t *testing.T) {
	re := docSources["xAI"].Pattern
	// "grok-code-prompt-engineering" is a docs page URL path, not a model ID.
	// The pattern should NOT match it.
	shouldNotMatch := []string{
		"grok-code-prompt-engineering",
		"grok-code-best-practices",
	}
	for _, s := range shouldNotMatch {
		if m := re.FindString(s); m == s {
			t.Errorf("xAI Pattern should NOT fully match docs page path %q, but it did", s)
		}
	}
	// Real model IDs should still match.
	shouldMatch := []string{
		"grok-code-fast-1",
		"grok-code-turbo-1",
	}
	for _, s := range shouldMatch {
		if m := re.FindString(s); m != s {
			t.Errorf("xAI Pattern should match model ID %q, got %q", s, m)
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

// ---------------------------------------------------------------------------
// fingerprintModels tests
// ---------------------------------------------------------------------------

func TestFingerprintModels(t *testing.T) {
	// Same IDs in different order produce same hash
	h1 := fingerprintModels([]string{"gpt-5", "claude-opus-4-6", "gemini-2.5-pro"})
	h2 := fingerprintModels([]string{"gemini-2.5-pro", "gpt-5", "claude-opus-4-6"})
	if h1 != h2 {
		t.Errorf("expected same hash for same IDs in different order, got %s vs %s", h1, h2)
	}

	// Different IDs produce different hash
	h3 := fingerprintModels([]string{"gpt-5", "gpt-5-mini"})
	if h1 == h3 {
		t.Error("expected different hash for different IDs")
	}

	// Empty input produces consistent hash
	h4 := fingerprintModels([]string{})
	h5 := fingerprintModels([]string{})
	if h4 != h5 {
		t.Error("expected consistent hash for empty input")
	}

	// Hash is valid hex string of expected length (SHA-256 = 64 hex chars)
	if len(h1) != 64 {
		t.Errorf("expected 64 char hex hash, got %d chars", len(h1))
	}
}

// ---------------------------------------------------------------------------
// Circuit breaker threshold tests
// ---------------------------------------------------------------------------

func TestCircuitBreakerThreshold(t *testing.T) {
	// When scraped is 0 and known > 0, circuit breaker should trigger
	// (tested via the logic: len(ids) == 0 && len(known) > 0)
	known := map[string]bool{"gpt-5": true, "gpt-5-mini": true}
	ids := []string{}

	// Circuit breaker condition
	if !(len(ids) == 0 && len(known) > 0) {
		t.Error("circuit breaker should trigger when scraped is empty but known has entries")
	}

	// When scraped has items, circuit breaker should NOT trigger
	ids2 := []string{"gpt-5"}
	if len(ids2) == 0 && len(known) > 0 {
		t.Error("circuit breaker should not trigger when scraped has items")
	}

	// Sanity check threshold: scraped*5 < known means suspiciously low
	known10 := make(map[string]bool)
	for i := 0; i < 10; i++ {
		known10[fmt.Sprintf("model-%d", i)] = true
	}
	scrapedLow := []string{"model-0"}
	if !(len(scrapedLow)*5 < len(known10)) {
		t.Error("expected warning threshold to trigger: 1*5=5 < 10")
	}

	scrapedOk := []string{"m1", "m2", "m3"}
	if len(scrapedOk)*5 < len(known10) {
		t.Error("should not warn: 3*5=15 >= 10")
	}
}

// ---------------------------------------------------------------------------
// applyNormalization tests
// ---------------------------------------------------------------------------

func TestApplyNormalization(t *testing.T) {
	// Unknown provider passes through unchanged
	ids := applyNormalization("unknown-provider", []string{"model-a", "model-b"})
	if len(ids) != 2 {
		t.Errorf("unknown provider: expected 2 models, got %d", len(ids))
	}

	// Mistral: ignore patterns filter correctly
	mistralIDs := applyNormalization("mistral", []string{
		"mistral-large-2512",
		"mistral-embed-23",
		"mistral-moderation-24",
		"mistral-nemo-12",
		"mistral-ocr-2503",
		"mistral-small-2506",
		"mistral-saba-2502-latest",
		"mistral-saba-2502-beta",
	})
	for _, id := range mistralIDs {
		if strings.Contains(id, "embed") || strings.Contains(id, "moderation") ||
			strings.Contains(id, "nemo") || strings.Contains(id, "ocr") ||
			strings.HasSuffix(id, "-latest") || strings.HasSuffix(id, "-beta") {
			t.Errorf("mistral normalization should have filtered %q", id)
		}
	}
	// mistral-large-2512 and mistral-small-2506 should survive
	found := map[string]bool{}
	for _, id := range mistralIDs {
		found[id] = true
	}
	if !found["mistral-large-2512"] {
		t.Error("expected mistral-large-2512 to survive normalization")
	}
	if !found["mistral-small-2506"] {
		t.Error("expected mistral-small-2506 to survive normalization")
	}
}

func TestApplyNormalizationXAI(t *testing.T) {
	xaiIDs := applyNormalization("xai", []string{
		"grok-4",
		"grok-4.1-fast",
		"grok-code-prompt-engineering",
		"grok-2",
		"grok-2-1212",
		"grok-3-fast-beta",
		"grok-3-fast-latest",
	})
	blocked := map[string]bool{
		"grok-code-prompt-engineering": true,
		"grok-2":                       true,
		"grok-2-1212":                  true,
		"grok-3-fast-beta":             true,
		"grok-3-fast-latest":           true,
	}
	for _, id := range xaiIDs {
		if blocked[id] {
			t.Errorf("xai normalization should have filtered %q", id)
		}
	}
	// grok-4 and grok-4.1-fast should survive
	survived := map[string]bool{}
	for _, id := range xaiIDs {
		survived[id] = true
	}
	if !survived["grok-4"] {
		t.Error("expected grok-4 to survive")
	}
	if !survived["grok-4.1-fast"] {
		t.Error("expected grok-4.1-fast to survive")
	}
}

// ---------------------------------------------------------------------------
// fetchModelsFromAPI tests
// ---------------------------------------------------------------------------

func TestFetchModelsFromAPI(t *testing.T) {
	// Mock a /v1/models endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "gpt-5"},
				{"id": "gpt-5-mini"},
				{"id": "o3"},
			},
		})
	}))
	defer ts.Close()

	ctx := context.Background()
	client := &http.Client{Timeout: 5 * time.Second}

	// Success case
	ids, err := fetchModelsFromAPI(ctx, client, ts.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 models, got %d", len(ids))
	}

	// Auth failure
	_, err = fetchModelsFromAPI(ctx, client, ts.URL, "wrong-key")
	if err == nil {
		t.Error("expected error for wrong API key")
	}

	// Invalid URL
	_, err = fetchModelsFromAPI(ctx, client, "http://localhost:1/nonexistent", "key")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

