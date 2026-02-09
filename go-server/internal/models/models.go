package models

// Model represents an AI model entry in the registry.
type Model struct {
	ID              string  `json:"id"`
	DisplayName     string  `json:"display_name"`
	Provider        string  `json:"provider"`
	ContextWindow   int     `json:"context_window"`
	MaxOutputTokens int     `json:"max_output_tokens"`
	Vision          bool    `json:"vision"`
	Reasoning       bool    `json:"reasoning"`
	PricingInput    float64 `json:"pricing_input"`
	PricingOutput   float64 `json:"pricing_output"`
	KnowledgeCutoff string  `json:"knowledge_cutoff"`
	ReleaseDate     string  `json:"release_date"`
	Status          string  `json:"status"`
	Notes           string  `json:"notes"`
}
