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
	// ─── OpenAI Aliases ────────────────────────────────────────────
	// GPT-5.2 series
	"gpt52":         "gpt-5.2",
	"gpt5.2":        "gpt-5.2",
	"gpt-5-2":       "gpt-5.2",
	"gpt52codex":    "gpt-5.2-codex",
	"gpt-5-2-codex": "gpt-5.2-codex",
	"gpt52pro":      "gpt-5.2-pro",
	"gpt-5-2-pro":   "gpt-5.2-pro",

	// GPT-5.1 series
	"gpt51":              "gpt-5.1",
	"gpt5.1":             "gpt-5.1",
	"gpt-5-1":            "gpt-5.1",
	"gpt51codex":         "gpt-5.1-codex",
	"gpt-5-1-codex":      "gpt-5.1-codex",
	"gpt51codexmini":     "gpt-5.1-codex-mini",
	"gpt-5-1-codex-mini": "gpt-5.1-codex-mini",
	"gpt51mini":          "gpt-5.1-mini",
	"gpt5.1mini":         "gpt-5.1-mini",
	"gpt-5-1-mini":       "gpt-5.1-mini",

	// GPT-5 series
	"gpt5":      "gpt-5",
	"gpt5mini":  "gpt-5-mini",
	"gpt-5mini": "gpt-5-mini",
	"gpt5nano":  "gpt-5-nano",
	"gpt-5nano": "gpt-5-nano",

	// GPT-4.1 series
	"gpt41":        "gpt-4.1",
	"gpt4.1":       "gpt-4.1",
	"gpt-4-1":      "gpt-4.1",
	"gpt41mini":    "gpt-4.1-mini",
	"gpt4.1mini":   "gpt-4.1-mini",
	"gpt-4-1-mini": "gpt-4.1-mini",
	"gpt41nano":    "gpt-4.1-nano",
	"gpt4.1nano":   "gpt-4.1-nano",
	"gpt-4-1-nano": "gpt-4.1-nano",

	// GPT-4o series
	"gpt4o":            "gpt-4o",
	"gpt-4-o":          "gpt-4o",
	"gpt4omini":        "gpt-4o-mini",
	"gpt-4-o-mini":     "gpt-4o-mini",
	"gpt-4o-2024-05-13": "gpt-4o",

	// o-series reasoning models
	"o-3":            "o3",
	"o3pro":          "o3-pro",
	"o-3-pro":        "o3-pro",
	"o4mini":         "o4-mini",
	"o-4-mini":       "o4-mini",
	"o3mini":         "o3-mini",
	"o-3-mini":       "o3-mini",
	"o3deepresearch": "o3-deep-research",
	"o3-research":    "o3-deep-research",

	// ─── Anthropic Aliases ─────────────────────────────────────────
	// Claude Opus 4.6
	"claude-opus":        "claude-opus-4-6",
	"opus":               "claude-opus-4-6",
	"claude-opus-latest": "claude-opus-4-6",
	"opus-4-6":           "claude-opus-4-6",
	"claude-opus-4.6":    "claude-opus-4-6",
	"claude-4.6-opus":    "claude-opus-4-6",

	// Claude Sonnet 4.6
	"claude-sonnet-4.6":          "claude-sonnet-4-6",
	"sonnet-4-6":                 "claude-sonnet-4-6",
	"claude-4.6-sonnet":          "claude-sonnet-4-6",
	"claude-sonnet-4-6-20260217": "claude-sonnet-4-6",
	"sonnet":                     "claude-sonnet-4-6",
	"claude-sonnet":              "claude-sonnet-4-6",
	"claude-sonnet-latest":       "claude-sonnet-4-6",

	// Claude Sonnet 4.5
	"claude-sonnet-4-5":  "claude-sonnet-4-5-20250929",
	"sonnet-4-5":         "claude-sonnet-4-5-20250929",
	"claude-4.5-sonnet":  "claude-sonnet-4-5-20250929",
	"claude-sonnet-4.5":  "claude-sonnet-4-5-20250929",

	// Claude Haiku 4.5
	"claude-haiku-4-5":  "claude-haiku-4-5-20251001",
	"haiku":             "claude-haiku-4-5-20251001",
	"claude-haiku":      "claude-haiku-4-5-20251001",
	"haiku-4-5":         "claude-haiku-4-5-20251001",
	"claude-4.5-haiku":  "claude-haiku-4-5-20251001",
	"claude-haiku-4.5":  "claude-haiku-4-5-20251001",

	// Claude Opus 4.5
	"claude-opus-4-5-20251101": "claude-opus-4-5",
	"claude-opus-4.5":          "claude-opus-4-5",
	"claude-4.5-opus":          "claude-opus-4-5",

	// Claude Opus 4.1
	"claude-opus-4-1-20250805": "claude-opus-4-1",
	"claude-opus-4.1":          "claude-opus-4-1",
	"claude-4.1-opus":          "claude-opus-4-1",

	// Claude Sonnet 4.0
	"claude-sonnet-4-20250514": "claude-sonnet-4-0",
	"claude-sonnet-4.0":        "claude-sonnet-4-0",
	"claude-4.0-sonnet":        "claude-sonnet-4-0",
	"claude-4-sonnet":          "claude-sonnet-4-0",

	// Claude Opus 4.0
	"claude-opus-4-20250514": "claude-opus-4-0",
	"claude-opus-4.0":        "claude-opus-4-0",
	"claude-4.0-opus":        "claude-opus-4-0",
	"claude-4-opus":          "claude-opus-4-0",

	// Claude 3.7
	"claude-3-7-sonnet-latest": "claude-3-7-sonnet-20250219",
	"claude-3.7-sonnet":        "claude-3-7-sonnet-20250219",
	"claude-3-7-sonnet":        "claude-3-7-sonnet-20250219",

	// ─── Google Aliases ────────────────────────────────────────────
	// Gemini 3 series
	"gemini-3-pro":                  "gemini-3-pro-preview",
	"gemini3pro":                    "gemini-3-pro-preview",
	"gemini-pro":                    "gemini-3-pro-preview",
	"google/gemini-3-pro-preview":   "gemini-3-pro-preview",
	"models/gemini-3-pro-preview":   "gemini-3-pro-preview",
	"gemini-3-pro-image":                "gemini-3-pro-image-preview",
	"gemini3proimage":                   "gemini-3-pro-image-preview",
	"google/gemini-3-pro-image-preview": "gemini-3-pro-image-preview",
	"gemini-3-flash":                    "gemini-3-flash-preview",
	"gemini3flash":                      "gemini-3-flash-preview",
	"google/gemini-3-flash-preview":     "gemini-3-flash-preview",
	"models/gemini-3-flash-preview":     "gemini-3-flash-preview",
	"gemini-flash":                      "gemini-3-flash-preview",

	// Gemini 2.5 series
	"gemini-2-5-pro":          "gemini-2.5-pro",
	"gemini25pro":             "gemini-2.5-pro",
	"gemini2.5pro":            "gemini-2.5-pro",
	"google/gemini-2.5-pro":   "gemini-2.5-pro",
	"gemini-2-5-flash":        "gemini-2.5-flash",
	"gemini25flash":           "gemini-2.5-flash",
	"gemini2.5flash":          "gemini-2.5-flash",
	"google/gemini-2.5-flash": "gemini-2.5-flash",
	"gemini-2-5-flash-lite":        "gemini-2.5-flash-lite",
	"gemini25flashlite":            "gemini-2.5-flash-lite",
	"gemini2.5flashlite":           "gemini-2.5-flash-lite",
	"google/gemini-2.5-flash-lite": "gemini-2.5-flash-lite",

	// Gemini 2.0 series
	"gemini-2-0-flash-lite": "gemini-2.0-flash-lite",
	"gemini20flashlite":     "gemini-2.0-flash-lite",
	"gemini2.0flashlite":    "gemini-2.0-flash-lite",
	"gemini-2-0-flash":      "gemini-2.0-flash",
	"gemini20flash":         "gemini-2.0-flash",
	"gemini2.0flash":        "gemini-2.0-flash",

	// ─── xAI Aliases ───────────────────────────────────────────────
	"grok4":          "grok-4",
	"grok41":         "grok-4.1",
	"grok4.1":        "grok-4.1",
	"grok-4-1":       "grok-4.1",
	"grok41fast":     "grok-4.1-fast",
	"grok4.1fast":    "grok-4.1-fast",
	"grok-4-1-fast":  "grok-4.1-fast",
	"grok4fast":      "grok-4-fast",
	"grok-code-fast": "grok-code-fast-1",
	"grok-code":      "grok-code-fast-1",
	"grokcode":       "grok-code-fast-1",
	"grok3":          "grok-3",
	"grok3mini":      "grok-3-mini",

	// ─── Meta Aliases ──────────────────────────────────────────────
	"llama4maverick":  "llama-4-maverick",
	"llama-4maverick": "llama-4-maverick",
	"llama4-maverick": "llama-4-maverick",
	"llama4scout":     "llama-4-scout",
	"llama-4scout":    "llama-4-scout",
	"llama4-scout":    "llama-4-scout",
	"llama-3-3-70b":   "llama-3.3-70b",
	"llama3370b":      "llama-3.3-70b",
	"llama3.370b":     "llama-3.3-70b",

	// ─── Mistral Aliases ───────────────────────────────────────────
	"mistral-large":        "mistral-large-2512",
	"mistral-large-3":      "mistral-large-2512",
	"ministral-3b":         "ministral-3b-2512",
	"ministral-8b":         "ministral-8b-2512",
	"ministral-14b":        "ministral-14b-2512",
	"magistral-small":      "magistral-small-2509",
	"magistral-small-1.2":  "magistral-small-2509",
	"magistral-medium":     "magistral-medium-2509",
	"magistral-medium-1.2": "magistral-medium-2509",
	"mistral-small":        "mistral-small-2506",
	"mistral-small-3.2":    "mistral-small-2506",
	"devstral":             "devstral-2512",
	"devstral-2":           "devstral-2512",
	"mistral-medium":       "mistral-medium-2505",
	"mistral-medium-3":     "mistral-medium-2505",
	"devstral-small":       "devstral-small-2512",
	"devstral-small-2":     "devstral-small-2512",
	"codestral":            "codestral-2508",

	// ─── DeepSeek Aliases ──────────────────────────────────────────
	"deepseek":           "deepseek-chat",
	"deepseek-v3.2":      "deepseek-reasoner",
	"deepseek-v3-2":      "deepseek-reasoner",
	"deepseek-thinking":  "deepseek-reasoner",
	"deepseek-v3.2-chat": "deepseek-chat",
	"deepseek-r-1":       "deepseek-r1",
	"r1":                 "deepseek-r1",
	"deepseek-v-3":       "deepseek-v3",

	// ─── Amazon Aliases ────────────────────────────────────────────
	"nova-micro":       "amazon-nova-micro",
	"aws/nova-micro":   "amazon-nova-micro",
	"nova-lite":        "amazon-nova-lite",
	"aws/nova-lite":    "amazon-nova-lite",
	"nova-pro":         "amazon-nova-pro",
	"aws/nova-pro":     "amazon-nova-pro",
	"nova-premier":     "amazon-nova-premier",
	"aws/nova-premier": "amazon-nova-premier",
	"nova-2-lite":      "amazon-nova-2-lite",
	"aws/nova-2-lite":  "amazon-nova-2-lite",
	"nova-2-pro":       "amazon-nova-2-pro",
	"aws/nova-2-pro":   "amazon-nova-2-pro",

	// ─── Cohere Aliases ────────────────────────────────────────────
	"command-a":           "command-a-03-2025",
	"cohere-command-a":    "command-a-03-2025",
	"command-a-reasoning": "command-a-reasoning-08-2025",
	"command-a-vision":    "command-a-vision-07-2025",
	"command-r7b":         "command-r7b-12-2024",

	// ─── Perplexity Aliases ────────────────────────────────────────
	"perplexity-sonar":                 "sonar",
	"perplexity-sonar-pro":             "sonar-pro",
	"perplexity-sonar-reasoning-pro":   "sonar-reasoning-pro",
	"sonar-reasoning":                  "sonar-reasoning-pro",
	"perplexity-sonar-deep-research":   "sonar-deep-research",

	// ─── AI21 Aliases ──────────────────────────────────────────────
	"jamba-large": "jamba-large-1.7",
	"jamba-1.7":   "jamba-large-1.7",
	"jamba-mini":  "jamba-mini-1.7",

	// ─── Moonshot/Kimi Aliases (NEW PROVIDER) ─────────────────────
	"kimi":                          "kimi-k2.5",
	"kimi-latest":                   "kimi-k2.5",
	"kimi-k2-5":                     "kimi-k2.5",
	"kimi-k25":                      "kimi-k2.5",
	"kimik2.5":                      "kimi-k2.5",
	"k2.5":                          "kimi-k2.5",
	"k25":                           "kimi-k2.5",
	"moonshot-kimi":                 "kimi-k2.5",
	"moonshot/kimi-k2.5":            "kimi-k2.5",
	"kimi-k2-think":                 "kimi-k2-thinking",
	"kimi-thinking":                 "kimi-k2-thinking",
	"kimi-reasoner":                 "kimi-k2-thinking",
	"kimi-k2thinking":               "kimi-k2-thinking",
	"k2-thinking":                   "kimi-k2-thinking",
	"k2thinking":                    "kimi-k2-thinking",
	"moonshot/kimi-k2-thinking":     "kimi-k2-thinking",
	"kimi-k2":                       "kimi-k2-0905-preview",
	"kimi-k2-preview":               "kimi-k2-0905-preview",
	"kimi-0905":                     "kimi-k2-0905-preview",
	"kimi-k20905":                   "kimi-k2-0905-preview",
	"k2-0905":                       "kimi-k2-0905-preview",
	"moonshot/kimi-k2":              "kimi-k2-0905-preview",

	// ─── Zhipu/GLM Aliases (NEW PROVIDER) ─────────────────────────
	"glm":                      "glm-5",
	"glm-latest":               "glm-5",
	"glm5":                     "glm-5",
	"glm-5-0":                  "glm-5",
	"zhipu-glm-5":              "glm-5",
	"chatglm-5":                "glm-5",
	"zhipu/glm-5":              "glm-5",
	"glm-4-7":                  "glm-4.7",
	"glm47":                    "glm-4.7",
	"glm4.7":                   "glm-4.7",
	"zhipu-glm-4.7":            "glm-4.7",
	"chatglm-4.7":              "glm-4.7",
	"zhipu/glm-4.7":            "glm-4.7",
	"glm-flashx":               "glm-4.7-flashx",
	"glm-flash":                "glm-4.7-flashx",
	"glm-4-7-flashx":           "glm-4.7-flashx",
	"glm47flashx":              "glm-4.7-flashx",
	"glm4.7flashx":             "glm-4.7-flashx",
	"zhipu/glm-4.7-flashx":     "glm-4.7-flashx",
	"glm-vision":               "glm-4.6v",
	"glm-v":                    "glm-4.6v",
	"glm-4-6v":                 "glm-4.6v",
	"glm46v":                   "glm-4.6v",
	"glm4.6v":                  "glm-4.6v",
	"zhipu/glm-4.6v":           "glm-4.6v",

	// ─── NVIDIA Aliases (NEW PROVIDER) ────────────────────────────
	"nemotron-3-nano":                         "nvidia/nemotron-3-nano-30b-a3b",
	"nemotron-nano":                           "nvidia/nemotron-3-nano-30b-a3b",
	"nvidia-nemotron-nano":                    "nvidia/nemotron-3-nano-30b-a3b",
	"nemotron-3-nano-30b":                     "nvidia/nemotron-3-nano-30b-a3b",
	"nemotron-nano-30b":                       "nvidia/nemotron-3-nano-30b-a3b",
	"nvidia/nemotron-nano":                    "nvidia/nemotron-3-nano-30b-a3b",
	"nemotron-ultra":                          "nvidia/llama-3.1-nemotron-ultra-253b-v1",
	"nemotron-ultra-253b":                     "nvidia/llama-3.1-nemotron-ultra-253b-v1",
	"nvidia-nemotron-ultra":                   "nvidia/llama-3.1-nemotron-ultra-253b-v1",
	"llama-nemotron-ultra":                    "nvidia/llama-3.1-nemotron-ultra-253b-v1",
	"nemotron-253b":                           "nvidia/llama-3.1-nemotron-ultra-253b-v1",
	"nvidia/nemotron-ultra":                   "nvidia/llama-3.1-nemotron-ultra-253b-v1",

	// ─── Tencent/Hunyuan Aliases (NEW PROVIDER) ───────────────────
	"hunyuan":                  "hunyuan-turbos",
	"hunyuan-latest":           "hunyuan-turbos",
	"hunyuan-turbo-s":          "hunyuan-turbos",
	"hunyuan-turbo":            "hunyuan-turbos",
	"tencent-hunyuan-turbos":   "hunyuan-turbos",
	"tencent/hunyuan-turbos":   "hunyuan-turbos",
	"hunyuan-t-1":              "hunyuan-t1",
	"tencent-hunyuan-t1":       "hunyuan-t1",
	"hunyuan-thinking":         "hunyuan-t1",
	"tencent/hunyuan-t1":       "hunyuan-t1",
	"hunyuan-a-13b":            "hunyuan-a13b",
	"hunyuan-13b":              "hunyuan-a13b",
	"tencent-hunyuan-a13b":     "hunyuan-a13b",
	"tencent/hunyuan-a13b":     "hunyuan-a13b",

	// ─── Microsoft/Phi Aliases (NEW PROVIDER) ─────────────────────
	"phi4":                                    "phi-4",
	"microsoft-phi-4":                         "phi-4",
	"microsoft/phi-4":                         "phi-4",
	"phi4multimodal":                          "phi-4-multimodal-instruct",
	"phi-4-multimodal":                        "phi-4-multimodal-instruct",
	"phi-4-mm":                                "phi-4-multimodal-instruct",
	"microsoft/phi-4-multimodal":              "phi-4-multimodal-instruct",
	"microsoft/phi-4-multimodal-instruct":     "phi-4-multimodal-instruct",
	"phi4reasoningplus":                       "phi-4-reasoning-plus",
	"phi-4-reasoning":                         "phi-4-reasoning-plus",
	"phi-4-rp":                                "phi-4-reasoning-plus",
	"microsoft/phi-4-reasoning":               "phi-4-reasoning-plus",
	"microsoft/phi-4-reasoning-plus":          "phi-4-reasoning-plus",

	// ─── MiniMax Aliases (NEW PROVIDER) ───────────────────────────
	"minimax":        "minimax-m2.1",
	"minimax-latest": "minimax-m2.1",
	"minimax-m2-1":   "minimax-m2.1",
	"minimax-m21":    "minimax-m2.1",
	"minimaxm2.1":    "minimax-m2.1",
	"MiniMax-M2.1":   "minimax-m2.1",
	"minimax-m2":     "minimax-m2.1",
	"minimax-0-1":    "minimax-01",
	"MiniMax-01":     "minimax-01",
	"minimax-vl-01":  "minimax-01",
	"MiniMax-VL-01":  "minimax-01",

	// ─── Xiaomi/MiMo Aliases (NEW PROVIDER) ───────────────────────
	"mimo":                 "mimo-v2-flash",
	"mimo-latest":          "mimo-v2-flash",
	"mimo-v2":              "mimo-v2-flash",
	"mimo-flash":           "mimo-v2-flash",
	"mimov2flash":          "mimo-v2-flash",
	"xiaomi-mimo":          "mimo-v2-flash",
	"xiaomi/mimo-v2-flash": "mimo-v2-flash",

	// ─── Kuaishou/KAT Aliases (NEW PROVIDER) ──────────────────────
	"kat":                     "kat-coder-pro",
	"kat-pro":                 "kat-coder-pro",
	"kat-coder":               "kat-coder-pro",
	"katcoder":                "kat-coder-pro",
	"katcoderpro":             "kat-coder-pro",
	"kwaipilot/kat-coder-pro": "kat-coder-pro",
	"kuaishou/kat-coder-pro":  "kat-coder-pro",
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
