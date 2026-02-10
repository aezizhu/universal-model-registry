package models

import (
	"fmt"
	"strings"
)

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

// Aliases maps common shorthand model IDs to their canonical registry key.
var Aliases = map[string]string{
	"claude-sonnet-4-5":        "claude-sonnet-4-5-20250929",
	"claude-haiku-4-5":         "claude-haiku-4-5-20251001",
	"claude-3-7-sonnet-latest": "claude-3-7-sonnet-20250219",
	"claude-opus-4-5-20251101": "claude-opus-4-5",
	"claude-sonnet-4-20250514": "claude-sonnet-4-0",
	"claude-opus-4-20250514":   "claude-opus-4-0",
	"claude-opus-4-1-20250805": "claude-opus-4-1",
	"gpt-4o-2024-05-13":        "gpt-4o",
}

// FormatInt formats an integer with comma separators.
func FormatInt(n int) string {
	if n < 0 {
		if -n < 0 {
			// math.MinInt: -n overflows back to negative.
			// Format the string representation directly.
			s := fmt.Sprintf("%d", n)
			digits := s[1:]
			var result strings.Builder
			result.WriteByte('-')
			rem := len(digits) % 3
			if rem > 0 {
				result.WriteString(digits[:rem])
				if len(digits) > rem {
					result.WriteString(",")
				}
			}
			for i := rem; i < len(digits); i += 3 {
				if i > rem {
					result.WriteString(",")
				}
				result.WriteString(digits[i : i+3])
			}
			return result.String()
		}
		return "-" + FormatInt(-n)
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
