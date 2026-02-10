package tools

import (
	"fmt"
	"strings"

	"go-server/internal/models"
)

// CompareModelsInput holds parameters for the compare_models tool.
type CompareModelsInput struct {
	ModelIDs []string `json:"model_ids" jsonschema:"List of 2-5 model IDs to compare"`
}

// CompareModels returns a side-by-side markdown comparison table for 2-5 models.
func CompareModels(modelIDs []string) string {
	if len(modelIDs) < 2 {
		return "Please provide at least 2 model IDs to compare."
	}

	if len(modelIDs) > 5 {
		modelIDs = modelIDs[:5]
	}

	var found []models.Model
	var notFound []string
	for _, mid := range modelIDs {
		m, ok := FindModel(mid)
		if ok {
			found = append(found, m)
		} else {
			notFound = append(notFound, mid)
		}
	}

	if len(notFound) > 0 {
		var parts []string
		for _, nf := range notFound {
			suggestions := SuggestModels(nf, 3)
			parts = append(parts, fmt.Sprintf("`%s` (did you mean: %s)", nf, strings.Join(suggestions, ", ")))
		}
		return fmt.Sprintf("Model(s) not found: %s", strings.Join(parts, "; "))
	}

	// Build comparison table — fields as rows, models as columns
	names := make([]string, len(found))
	for i, m := range found {
		names[i] = m.DisplayName
	}

	header := "| Field | " + strings.Join(names, " | ") + " |"
	sep := "|-------|" + strings.Repeat("------|", len(found))

	providers := make([]string, len(found))
	statuses := make([]string, len(found))
	contexts := make([]string, len(found))
	maxOutputs := make([]string, len(found))
	capabilities := make([]string, len(found))
	inputPrices := make([]string, len(found))
	outputPrices := make([]string, len(found))
	cutoffs := make([]string, len(found))
	releases := make([]string, len(found))

	for i, m := range found {
		providers[i] = m.Provider
		statuses[i] = m.Status
		contexts[i] = models.FormatInt(m.ContextWindow)
		maxOutputs[i] = models.FormatInt(m.MaxOutputTokens)
		capabilities[i] = caps(m)
		inputPrices[i] = fmt.Sprintf("$%.2f", m.PricingInput)
		outputPrices[i] = fmt.Sprintf("$%.2f", m.PricingOutput)
		cutoffs[i] = m.KnowledgeCutoff
		releases[i] = m.ReleaseDate
	}

	rows := []string{
		header,
		sep,
		"| Provider | " + strings.Join(providers, " | ") + " |",
		"| Status | " + strings.Join(statuses, " | ") + " |",
		"| Context | " + strings.Join(contexts, " | ") + " |",
		"| Max Output | " + strings.Join(maxOutputs, " | ") + " |",
		"| Capabilities | " + strings.Join(capabilities, " | ") + " |",
		"| Input $/1M | " + strings.Join(inputPrices, " | ") + " |",
		"| Output $/1M | " + strings.Join(outputPrices, " | ") + " |",
		"| Knowledge Cutoff | " + strings.Join(cutoffs, " | ") + " |",
		"| Release Date | " + strings.Join(releases, " | ") + " |",
	}

	return strings.Join(rows, "\n")
}

// caps returns a comma-separated capability string for a model.
func caps(m models.Model) string {
	var c []string
	if m.Vision {
		c = append(c, "Vision")
	}
	if m.Reasoning {
		c = append(c, "Reasoning")
	}
	if len(c) == 0 {
		return "None"
	}
	return strings.Join(c, ", ")
}
