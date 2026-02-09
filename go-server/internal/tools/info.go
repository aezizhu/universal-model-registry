package tools

import (
	"fmt"
	"sort"
	"strings"

	"go-server/internal/models"
)

// GetModelInfo returns detailed specs for a specific model.
func GetModelInfo(modelID string) string {
	m, found := FindModel(modelID)
	if !found {
		keys := make([]string, 0, len(models.Models))
		for k := range models.Models {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return fmt.Sprintf("Model `%s` not found in registry.\n\nKnown models: %s",
			modelID, strings.Join(keys, ", "))
	}
	return ModelDetail(m)
}
