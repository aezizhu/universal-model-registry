package resources

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go-server/internal/models"
)

// AllModels returns JSON of all models in the registry.
func AllModels() string {
	data, err := json.MarshalIndent(models.Models, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(data)
}

// CurrentModels returns JSON of only current-status models.
func CurrentModels() string {
	current := make(map[string]models.Model)
	for k, m := range models.Models {
		if m.Status == "current" {
			current[k] = m
		}
	}
	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(data)
}

// PricingSummary returns a markdown pricing table of current models sorted by input price.
func PricingSummary() string {
	var current []models.Model
	for _, m := range models.Models {
		if m.Status == "current" {
			current = append(current, m)
		}
	}
	sort.SliceStable(current, func(i, j int) bool {
		if current[i].PricingInput != current[j].PricingInput {
			return current[i].PricingInput < current[j].PricingInput
		}
		return current[i].ID < current[j].ID
	})

	rows := []string{
		"| Model ID | Provider | Input $/1M | Output $/1M | Context |",
		"|----------|----------|------------|-------------|---------|",
	}
	for _, m := range current {
		rows = append(rows, fmt.Sprintf(
			"| %s | %s | $%.2f | $%.2f | %s |",
			m.ID, m.Provider, m.PricingInput, m.PricingOutput, formatInt(m.ContextWindow),
		))
	}
	return strings.Join(rows, "\n")
}

// formatInt formats an integer with comma separators.
func formatInt(n int) string {
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
