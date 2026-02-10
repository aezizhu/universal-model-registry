package tools

import (
	"fmt"
	"strings"
)

// GetModelInfo returns detailed specs for a specific model.
func GetModelInfo(modelID string) string {
	m, found := FindModel(modelID)
	if !found {
		suggestions := SuggestModels(modelID, 3)
		return fmt.Sprintf("Model `%s` not found in registry. Did you mean: %s",
			modelID, strings.Join(suggestions, ", "))
	}
	return ModelDetail(m)
}
