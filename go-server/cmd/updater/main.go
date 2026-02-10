package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// Provider describes how to query a provider's model listing API.
type Provider struct {
	URL        string
	AuthEnv    string
	AuthHeader string // empty means use query param auth (Google)
}

var providers = map[string]Provider{
	"OpenAI":    {URL: "https://api.openai.com/v1/models", AuthEnv: "OPENAI_API_KEY", AuthHeader: "Authorization"},
	"Anthropic": {URL: "https://api.anthropic.com/v1/models", AuthEnv: "ANTHROPIC_API_KEY", AuthHeader: "x-api-key"},
	"Google":    {URL: "https://generativelanguage.googleapis.com/v1beta/models", AuthEnv: "GEMINI_API_KEY", AuthHeader: ""},
	"Mistral":   {URL: "https://api.mistral.ai/v1/models", AuthEnv: "MISTRAL_API_KEY", AuthHeader: "Authorization"},
	"xAI":       {URL: "https://api.x.ai/v1/models", AuthEnv: "XAI_API_KEY", AuthHeader: "Authorization"},
	"DeepSeek":  {URL: "https://api.deepseek.com/models", AuthEnv: "DEEPSEEK_API_KEY", AuthHeader: "Authorization"},
}

// knownModels maps provider -> set of model IDs we track in the registry.
var knownModels = map[string]map[string]bool{
	"OpenAI": {
		"gpt-5.2":        true,
		"gpt-5.2-codex":  true,
		"gpt-5.2-pro":    true,
		"gpt-5.1":        true,
		"gpt-5":          true,
		"gpt-5-mini":     true,
		"gpt-5-nano":     true,
		"gpt-4.1-mini":   true,
		"gpt-4.1-nano":   true,
		"o3":             true,
		"o3-pro":         true,
		"o4-mini":        true,
		"o3-mini":        true,
		"gpt-4.1":        true,
		"gpt-4o":         true,
		"gpt-4o-mini":    true,
	},
	"Anthropic": {
		"claude-opus-4-6":              true,
		"claude-sonnet-4-5-20250929":   true,
		"claude-haiku-4-5-20251001":    true,
		"claude-opus-4-5":              true,
		"claude-opus-4-1":              true,
		"claude-sonnet-4-0":            true,
		"claude-3-7-sonnet-20250219":   true,
		"claude-opus-4-0":              true,
	},
	"Google": {
		"gemini-3-pro-preview":   true,
		"gemini-3-flash-preview": true,
		"gemini-2.5-pro":         true,
		"gemini-2.5-flash":       true,
		"gemini-2.5-flash-lite":  true,
		"gemini-2.0-flash":       true,
	},
	"xAI": {
		"grok-4":           true,
		"grok-4.1-fast":    true,
		"grok-4-fast":      true,
		"grok-code-fast-1": true,
		"grok-3":           true,
		"grok-3-mini":      true,
	},
	"Mistral": {
		"mistral-large-2512":  true,
		"mistral-medium-2505": true,
		"mistral-small-2506":  true,
		"devstral-2512":       true,
		"devstral-small-2512": true,
		"codestral-2508":      true,
	},
	"DeepSeek": {
		"deepseek-reasoner": true,
		"deepseek-chat":     true,
		"deepseek-r1":       true,
		"deepseek-v3":       true,
	},
	"Meta": {
		"llama-4-maverick": true,
		"llama-4-scout":    true,
		"llama-3.3-70b":    true,
	},
	"Amazon": {
		"amazon-nova-micro":     true,
		"amazon-nova-lite":      true,
		"amazon-nova-pro":       true,
		"amazon-nova-premier":   true,
		"amazon-nova-2-lite":    true,
		"amazon-nova-2-pro":     true,
	},
	"Cohere": {
		"command-a-03-2025":            true,
		"command-a-reasoning-08-2025":  true,
		"command-a-vision-07-2025":     true,
		"command-r7b-12-2024":          true,
	},
	"Perplexity": {
		"sonar":                true,
		"sonar-pro":            true,
		"sonar-reasoning-pro":  true,
	},
	"AI21": {
		"jamba-large-1.7": true,
		"jamba-mini-1.7":  true,
	},
}

// apiResponse is the common shape returned by OpenAI-compatible model list APIs.
type apiResponse struct {
	Data   []apiModel `json:"data"`
	Models []apiModel `json:"models"` // Google uses top-level "models" array
}

type apiModel struct {
	ID   string `json:"id"`
	Name string `json:"name"` // Google uses "name" (e.g. "models/gemini-2.5-pro")
}

func main() {
	client := &http.Client{Timeout: 15 * time.Second}
	ctx := context.Background()

	hasChanges := false
	providerOrder := []string{"OpenAI", "Anthropic", "Google", "Mistral", "xAI", "DeepSeek"}

	fmt.Println("=== Model Registry Update Check ===")
	fmt.Printf("Time: %s\n\n", time.Now().UTC().Format(time.RFC3339))

	for _, name := range providerOrder {
		p := providers[name]
		key := os.Getenv(p.AuthEnv)
		if key == "" {
			fmt.Printf("[%s] SKIP: %s not set\n", name, p.AuthEnv)
			continue
		}

		ids, err := fetchModels(ctx, client, name, p, key)
		if err != nil {
			fmt.Printf("[%s] ERROR: %v\n", name, err)
			continue
		}

		known := knownModels[name]
		newModels, missing := diff(known, ids)

		fmt.Printf("[%s] API returned %d models, we track %d\n", name, len(ids), len(known))

		if len(newModels) > 0 {
			hasChanges = true
			sort.Strings(newModels)
			fmt.Printf("  NEW (%d):\n", len(newModels))
			for _, m := range newModels {
				fmt.Printf("    + %s\n", m)
			}
		}
		if len(missing) > 0 {
			hasChanges = true
			sort.Strings(missing)
			fmt.Printf("  MISSING from API (%d):\n", len(missing))
			for _, m := range missing {
				fmt.Printf("    - %s\n", m)
			}
		}
		if len(newModels) == 0 && len(missing) == 0 {
			fmt.Printf("  OK: in sync\n")
		}
		fmt.Println()
	}

	// Providers without direct model-listing APIs — just note them.
	fmt.Println("[Meta] SKIP: no direct API (models are provider-hosted)")
	fmt.Println("[Amazon] SKIP: no public model-listing API (check AWS Bedrock console)")
	fmt.Println("[Cohere] SKIP: no public model-listing API (check docs.cohere.com)")
	fmt.Println("[Perplexity] SKIP: no public model-listing API (check docs.perplexity.ai)")
	fmt.Println("[AI21] SKIP: no public model-listing API (check docs.ai21.com)")

	fmt.Println("\n=== Summary ===")
	if hasChanges {
		fmt.Println("Changes detected. Review the output above.")
		os.Exit(1)
	}
	fmt.Println("All tracked providers are in sync.")
	os.Exit(0)
}

// fetchModels queries a provider's model listing endpoint and returns model IDs.
func fetchModels(ctx context.Context, client *http.Client, name string, p Provider, key string) ([]string, error) {
	url := p.URL
	if p.AuthHeader == "" {
		// Google: API key as query parameter
		url += "?key=" + key
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if p.AuthHeader != "" {
		if p.AuthHeader == "x-api-key" {
			// Anthropic uses x-api-key header + version header
			req.Header.Set("x-api-key", key)
			req.Header.Set("anthropic-version", "2023-06-01")
		} else {
			// OpenAI-style Bearer auth
			req.Header.Set(p.AuthHeader, "Bearer "+key)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	// Collect model IDs from whichever field is populated.
	var ids []string
	models := result.Data
	if len(models) == 0 {
		models = result.Models
	}
	for _, m := range models {
		id := m.ID
		if id == "" && m.Name != "" {
			// Google returns "models/gemini-2.5-pro" — strip prefix.
			id = strings.TrimPrefix(m.Name, "models/")
		}
		if id != "" {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// diff compares our known models against API results.
// Returns new models (in API but not known) and missing models (known but not in API).
func diff(known map[string]bool, apiIDs []string) (newModels, missing []string) {
	apiSet := make(map[string]bool, len(apiIDs))
	for _, id := range apiIDs {
		apiSet[id] = true
	}

	// New: in API but not in our registry
	for _, id := range apiIDs {
		if !known[id] {
			newModels = append(newModels, id)
		}
	}

	// Missing: in our registry but not in API
	for id := range known {
		if !apiSet[id] {
			missing = append(missing, id)
		}
	}

	return newModels, missing
}
