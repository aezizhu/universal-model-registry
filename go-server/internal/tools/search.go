package tools

import (
	"fmt"
	"strings"

	"go-server/internal/models"
)

// SearchModels searches for models by keyword across names, providers, and notes.
func SearchModels(query string) string {
	q := strings.ToLower(query)
	var matches []models.Model
	for _, m := range models.Models {
		if strings.Contains(strings.ToLower(m.ID), q) ||
			strings.Contains(strings.ToLower(m.DisplayName), q) ||
			strings.Contains(strings.ToLower(m.Provider), q) ||
			strings.Contains(strings.ToLower(m.Notes), q) {
			matches = append(matches, m)
		}
	}
	if len(matches) == 0 {
		return fmt.Sprintf("No models found matching '%s'.", query)
	}
	return FormatTable(matches)
}
