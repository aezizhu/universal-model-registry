package tools

// ListModelsInput defines the input parameters for the list_models tool.
type ListModelsInput struct {
	Provider   string `json:"provider,omitempty" jsonschema:"Filter by provider name (case-insensitive)"`
	Status     string `json:"status,omitempty" jsonschema:"Filter by status: current, legacy, or deprecated"`
	Capability string `json:"capability,omitempty" jsonschema:"Filter by capability: vision or reasoning"`
}

// ListModels returns a markdown table of models with optional filters.
func ListModels(provider, status, capability string) string {
	results := FilterModels(provider, status, capability)
	return FormatTable(results)
}
