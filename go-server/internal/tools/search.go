package tools

import (
	"fmt"
	"strings"

	"go-server/internal/models"
)

// SearchModels searches for models by keyword across names, providers, and notes.
// Multi-word queries require ALL words to match across any combination of fields.
func SearchModels(query string) string {
	if query == "" {
		return "Please provide a search term."
	}
	words := strings.Fields(strings.ToLower(query))
	var matches []models.Model
	for _, m := range models.Models {
		// Combine all searchable fields into one string for multi-word matching
		combined := strings.ToLower(m.ID + " " + m.DisplayName + " " + m.Provider + " " + m.Status + " " + m.Notes)
		allMatch := true
		for _, w := range words {
			if !strings.Contains(combined, w) {
				allMatch = false
				break
			}
		}
		if allMatch {
			matches = append(matches, m)
		}
	}
	if len(matches) == 0 {
		return fmt.Sprintf("No models found matching '%s'.", query)
	}
	return FormatTable(matches)
}
