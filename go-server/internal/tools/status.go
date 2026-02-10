package tools

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"go-server/internal/models"
)

// CheckModelStatusInput holds parameters for the check_model_status tool.
type CheckModelStatusInput struct {
	ModelID string `json:"model_id" jsonschema:"The model ID to check"`
}

// CheckModelStatus returns status information for a model, including
// replacement suggestions for legacy/deprecated models.
func CheckModelStatus(modelID string) string {
	m, found := FindModel(modelID)
	if !found {
		suggestions := SuggestModels(modelID, 3)
		return fmt.Sprintf("`%s` is **not found** in the registry. "+
			"Did you mean: %s", modelID, strings.Join(suggestions, ", "))
	}

	result := fmt.Sprintf("**%s** (`%s`): status = **%s**",
		m.DisplayName, m.ID, m.Status)

	if m.Status == "legacy" || m.Status == "deprecated" {
		// Find current replacements from the same provider
		var replacements []models.Model
		for _, r := range models.Models {
			if r.Provider == m.Provider && r.Status == "current" {
				replacements = append(replacements, r)
			}
		}
		// Sort by closest pricing to the original model
		sort.SliceStable(replacements, func(i, j int) bool {
			return math.Abs(replacements[i].PricingInput-m.PricingInput) <
				math.Abs(replacements[j].PricingInput-m.PricingInput)
		})
		if len(replacements) > 0 {
			r := replacements[0]
			result += fmt.Sprintf("\n\nRecommended replacement: **%s** (`%s`)",
				r.DisplayName, r.ID)
		}
	}

	if m.Notes != "" {
		result += fmt.Sprintf("\n\nNote: %s", m.Notes)
	}

	return result
}
