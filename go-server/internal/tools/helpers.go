package tools

import (
	"fmt"
	"sort"
	"strings"

	"go-server/internal/models"
)

// newestPerProvider returns a set of model IDs that are the newest (by ReleaseDate) for each provider.
func newestPerProvider(ms []models.Model) map[string]bool {
	best := make(map[string]string)   // provider -> best release date
	bestID := make(map[string]string) // provider -> model ID with best date
	for _, m := range ms {
		if m.ReleaseDate > best[m.Provider] ||
			(m.ReleaseDate == best[m.Provider] && m.ID < bestID[m.Provider]) {
			best[m.Provider] = m.ReleaseDate
			bestID[m.Provider] = m.ID
		}
	}
	result := make(map[string]bool)
	for _, id := range bestID {
		result[id] = true
	}
	return result
}

// FormatTable renders a list of models as a markdown table.
// Models are grouped by provider and sorted newest-first within each group.
// The newest model per provider is marked with ★.
func FormatTable(ms []models.Model) string {
	if len(ms) == 0 {
		return "No models found matching the criteria."
	}

	newest := newestPerProvider(ms)

	// Sort: by provider name ascending, then by release date descending within provider
	sorted := make([]models.Model, len(ms))
	copy(sorted, ms)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Provider != sorted[j].Provider {
			return sorted[i].Provider < sorted[j].Provider
		}
		return sorted[i].ReleaseDate > sorted[j].ReleaseDate
	})

	rows := []string{
		"| Model ID | Display Name | Provider | Status | Context | Input $/1M | Output $/1M |",
		"|----------|-------------|----------|--------|---------|-----------|-------------|",
	}
	for _, m := range sorted {
		displayName := m.DisplayName
		if newest[m.ID] {
			displayName = "★ " + displayName
		}
		rows = append(rows, fmt.Sprintf(
			"| %s | %s | %s | %s | %s | $%.2f | $%.2f |",
			m.ID, displayName, m.Provider, m.Status,
			models.FormatInt(m.ContextWindow),
			m.PricingInput, m.PricingOutput,
		))
	}
	return strings.Join(rows, "\n")
}

// ModelDetail renders full specs for a single model as markdown.
func ModelDetail(m models.Model) string {
	var caps []string
	if m.Vision {
		caps = append(caps, "Vision")
	}
	if m.Reasoning {
		caps = append(caps, "Reasoning/Thinking")
	}
	capsStr := "None"
	if len(caps) > 0 {
		capsStr = strings.Join(caps, ", ")
	}

	notes := m.Notes
	if notes == "" {
		notes = "—"
	}

	return fmt.Sprintf(`## %s (`+"`%s`"+`)

| Field | Value |
|-------|-------|
| Provider | %s |
| Status | **%s** |
| Context Window | %s tokens |
| Max Output | %s tokens |
| Capabilities | %s |
| Pricing (input) | $%.2f / 1M tokens |
| Pricing (output) | $%.2f / 1M tokens |
| Knowledge Cutoff | %s |
| Release Date | %s |
| Notes | %s |`,
		m.DisplayName, m.ID,
		m.Provider,
		m.Status,
		models.FormatInt(m.ContextWindow),
		models.FormatInt(m.MaxOutputTokens),
		capsStr,
		m.PricingInput,
		m.PricingOutput,
		m.KnowledgeCutoff,
		m.ReleaseDate,
		notes,
	)
}

// levenshteinDistance computes the Levenshtein edit distance between two strings.
func levenshteinDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			min := del
			if ins < min {
				min = ins
			}
			if sub < min {
				min = sub
			}
			curr[j] = min
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// SuggestModels returns the n closest model IDs to the input by Levenshtein distance.
func SuggestModels(input string, n int) []string {
	type candidate struct {
		id   string
		dist int
	}
	lower := strings.ToLower(input)
	var candidates []candidate
	for key := range models.Models {
		dist := levenshteinDistance(lower, strings.ToLower(key))
		candidates = append(candidates, candidate{id: key, dist: dist})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].dist != candidates[j].dist {
			return candidates[i].dist < candidates[j].dist
		}
		return candidates[i].id < candidates[j].id
	})
	var result []string
	for i := 0; i < n && i < len(candidates); i++ {
		result = append(result, candidates[i].id)
	}
	return result
}

// FindModel finds a model by exact match, alias, case-insensitive, or partial match.
// Partial matching is deterministic: shortest ID first, then alphabetically.
func FindModel(modelID string) (models.Model, bool) {
	if modelID == "" {
		return models.Model{}, false
	}

	// Exact match
	if m, ok := models.Models[modelID]; ok {
		return m, true
	}

	// Alias resolution
	if canonical, ok := models.Aliases[modelID]; ok {
		if m, ok := models.Models[canonical]; ok {
			return m, true
		}
	}

	// Case-insensitive / partial match — collect all candidates, then sort deterministically
	lower := strings.ToLower(modelID)
	var candidates []models.Model
	for key, m := range models.Models {
		if strings.ToLower(key) == lower {
			return m, true // Exact case-insensitive — return immediately
		}
		if strings.Contains(strings.ToLower(key), lower) {
			candidates = append(candidates, m)
		}
	}

	// Sort: shortest ID first, then alphabetically
	sort.SliceStable(candidates, func(i, j int) bool {
		if len(candidates[i].ID) != len(candidates[j].ID) {
			return len(candidates[i].ID) < len(candidates[j].ID)
		}
		return candidates[i].ID < candidates[j].ID
	})
	if len(candidates) > 0 {
		return candidates[0], true
	}

	return models.Model{}, false
}

// FilterModels returns models matching the given provider, status, and capability filters.
// Empty string means no filter for that field.
func FilterModels(provider, status, capability string) []models.Model {
	var results []models.Model
	for _, m := range models.Models {
		results = append(results, m)
	}

	if provider != "" {
		p := strings.ToLower(provider)
		var filtered []models.Model
		for _, m := range results {
			if strings.ToLower(m.Provider) == p {
				filtered = append(filtered, m)
			}
		}
		results = filtered
	}

	if status != "" {
		s := strings.ToLower(status)
		var filtered []models.Model
		for _, m := range results {
			if strings.ToLower(m.Status) == s {
				filtered = append(filtered, m)
			}
		}
		results = filtered
	}

	if capability != "" {
		c := strings.ToLower(capability)
		var filtered []models.Model
		switch c {
		case "vision":
			for _, m := range results {
				if m.Vision {
					filtered = append(filtered, m)
				}
			}
		case "reasoning", "thinking":
			for _, m := range results {
				if m.Reasoning {
					filtered = append(filtered, m)
				}
			}
		default:
			filtered = results
		}
		results = filtered
	}

	return results
}
