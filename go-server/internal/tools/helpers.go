package tools

import (
	"fmt"
	"strings"

	"go-server/internal/models"
)

// FormatTable renders a list of models as a markdown table.
func FormatTable(ms []models.Model) string {
	if len(ms) == 0 {
		return "No models found matching the criteria."
	}

	rows := []string{
		"| Model ID | Display Name | Provider | Status | Context | Input $/1M | Output $/1M |",
		"|----------|-------------|----------|--------|---------|-----------|-------------|",
	}
	for _, m := range ms {
		rows = append(rows, fmt.Sprintf(
			"| %s | %s | %s | %s | %s | $%.2f | $%.2f |",
			m.ID, m.DisplayName, m.Provider, m.Status,
			formatInt(m.ContextWindow),
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
		formatInt(m.ContextWindow),
		formatInt(m.MaxOutputTokens),
		capsStr,
		m.PricingInput,
		m.PricingOutput,
		m.KnowledgeCutoff,
		m.ReleaseDate,
		notes,
	)
}

// formatInt formats an integer with comma separators.
func formatInt(n int) string {
	if n < 0 {
		return "-" + formatInt(-n)
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	s := fmt.Sprintf("%d", n)
	var result strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		result.WriteString(s[:remainder])
		if len(s) > remainder {
			result.WriteString(",")
		}
	}
	for i := remainder; i < len(s); i += 3 {
		if i > remainder {
			result.WriteString(",")
		}
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

// FindModel finds a model by exact match, case-insensitive, or partial match.
func FindModel(modelID string) (models.Model, bool) {
	// Exact match
	if m, ok := models.Models[modelID]; ok {
		return m, true
	}

	// Case-insensitive / partial match
	lower := strings.ToLower(modelID)
	for key, m := range models.Models {
		if strings.ToLower(key) == lower || strings.Contains(strings.ToLower(key), lower) {
			return m, true
		}
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
