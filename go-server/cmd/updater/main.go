package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// DocSource describes a public documentation page to scrape for model IDs.
// No API keys needed — these are all publicly accessible pages.
type DocSource struct {
	URLs           []string                // URLs to try in order (fallbacks)
	Pattern        *regexp.Regexp          // Regex to extract model IDs from page content
	ExcludePattern *regexp.Regexp          // Optional: exclude matched IDs containing this pattern
	Lowercase      bool                    // Lowercase extracted IDs before comparison
	NormalizeRe    *regexp.Regexp          // Optional: normalize extracted IDs (regex)
	NormalizeRepl  string                  // Replacement for NormalizeRe
	NormalizeFunc  func(string) string     // Optional: custom normalization function applied after NormalizeRe
}

// normalizeMistralID converts Mistral's long-form versioned API names to our
// short-form registry names. For example:
//
//	mistral-small-3-1-25-03 → mistral-small-2503
//	codestral-25-08         → codestral-2508
//	devstral-small-2-25-12  → devstral-small-2512
//
// If the ID doesn't match the long-form pattern (e.g. already short-form),
// it is returned unchanged.
func normalizeMistralID(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) < 3 {
		return id
	}
	// Long-form ends with two 2-digit segments: -{YY}-{MM}
	yy := parts[len(parts)-2]
	mm := parts[len(parts)-1]
	if len(yy) != 2 || len(mm) != 2 || !isAllDigits(yy) || !isAllDigits(mm) {
		return id
	}
	// Base name = non-numeric parts before the date, skipping version digits.
	// e.g. [ministral, 3, 14b, 25, 12] → base parts = [ministral, 14b]
	var baseParts []string
	for _, p := range parts[:len(parts)-2] {
		if !isAllDigits(p) {
			baseParts = append(baseParts, p)
		}
	}
	if len(baseParts) == 0 {
		return id
	}
	return strings.Join(baseParts, "-") + "-" + yy + mm
}

// docSources maps provider name to its public documentation source.
// Each entry contains public URLs and a regex pattern to extract model IDs.
var docSources = map[string]DocSource{
	"OpenAI": {
		URLs: []string{
			"https://raw.githubusercontent.com/openai/openai-python/main/src/openai/types/shared/chat_model.py",
			"https://cdn.jsdelivr.net/gh/openai/openai-python@main/src/openai/types/shared/chat_model.py",
			"https://raw.githubusercontent.com/openai/openai-python/main/src/openai/types/shared/all_models.py",
			"https://cdn.jsdelivr.net/gh/openai/openai-python@main/src/openai/types/shared/all_models.py",
		},
		Pattern:        regexp.MustCompile(`(?:"|')((?:gpt-[0-9][a-z0-9._-]*|o[0-9](?:-[a-z0-9-]+)*))`),
		ExcludePattern: regexp.MustCompile(`^gpt-(?:3\.|4o?(?:-|$))|^o1(?:-|$)`),
	},
	"Anthropic": {
		URLs: []string{
			"https://docs.anthropic.com/en/docs/about-claude/models",
		},
		Pattern: regexp.MustCompile(`(claude-(?:opus|sonnet|haiku)-[0-9]+(?:-[0-9]+)*(?:-[0-9]{8})?)`),
	},
	"Google": {
		URLs: []string{
			"https://ai.google.dev/gemini-api/docs/models",
		},
		Pattern: regexp.MustCompile(`(gemini-[0-9]+\.?[0-9]*-(?:pro|pro-image|flash|flash-lite)(?:-preview)?)`),
	},
	"Mistral": {
		URLs: []string{
			"https://docs.mistral.ai/getting-started/models/models_overview/",
			"https://docs.mistral.ai/getting-started/models/",
		},
		Pattern:        regexp.MustCompile(`((?:mistral|devstral|codestral|ministral|magistral)(?:-[a-z0-9]+)*-[0-9]{2,4})`),
		ExcludePattern: regexp.MustCompile(`(?:embed|moderation|ocr|nemo)`),
		NormalizeFunc:  normalizeMistralID,
	},
	"xAI": {
		URLs: []string{
			"https://docs.x.ai/docs/models",
		},
		Pattern:        regexp.MustCompile(`(grok-(?:[0-9]+(?:\.[0-9]+)?(?:-[a-z0-9-]*)?|code-(?:fast|turbo|chat)-[a-z0-9-]+))`),
		ExcludePattern: regexp.MustCompile(`(?i)(?:image|vision|imagine|video)|^grok-2(?:-|$)`),
		NormalizeRe:    regexp.MustCompile(`(\d)-(\d)([^0-9]|$)`),
		NormalizeRepl:  "${1}.${2}${3}",
	},
	"DeepSeek": {
		URLs: []string{
			"https://api-docs.deepseek.com/quick_start/pricing",
			"https://api-docs.deepseek.com/",
		},
		Pattern: regexp.MustCompile(`(deepseek-(?:chat|reasoner|r1|coder|v[0-9]+))`),
	},
	"Zhipu": {
		URLs: []string{
			"https://docs.z.ai/guides/overview/pricing",
		},
		Pattern:        regexp.MustCompile(`(?i)(GLM-[0-9]+(?:\.[0-9]+)?(?:-[A-Za-z]+)*)`),
		ExcludePattern: regexp.MustCompile(`(?i)^glm-4(?:\.[0-6])?(?:-|$)`),
		Lowercase:      true,
	},
	"MiniMax": {
		URLs: []string{
			"https://platform.minimax.io/docs/guides/models-intro",
			"https://intl.minimaxi.com/",
		},
		Pattern:        regexp.MustCompile(`(?i)(MiniMax-M[0-9](?:\.[0-9]+)?(?:-[a-z0-9]+)*)`),
		ExcludePattern: regexp.MustCompile(`(?i)^minimax-m1(?:-|$)`),
		Lowercase:      true,
	},
}

// knownModels maps provider -> set of model IDs we track in the registry.
var knownModels = map[string]map[string]bool{
	// NOTE: Only include current and legacy models here.
	// Deprecated models are already handled and should NOT be tracked
	// (otherwise they appear as false "MISSING" every run).
	"OpenAI": {
		"gpt-5.4":             true,
		"gpt-5.3-chat-latest": true,
		"gpt-5.2":             true,
		"gpt-5.2-pro":         true,
		"gpt-5.1":       true,
		"gpt-5.1-codex": true,
		"gpt-5.1-mini":  true,
		"gpt-5":         true,
		"gpt-5-mini":    true,
		"gpt-5-nano":    true,
		"gpt-4.1-mini":  true,
		"gpt-4.1-nano":  true,
		"o3":            true,
		"o4-mini":       true,
		"o3-mini":       true, // legacy
	},
	"Anthropic": {
		"claude-sonnet-4-6":          true,
		"claude-opus-4-6":            true,
		"claude-sonnet-4-5-20250929": true,
		"claude-haiku-4-5-20251001":  true,
		"claude-opus-4-5":            true, // legacy
		"claude-opus-4-1":            true, // legacy
		"claude-sonnet-4-0":          true, // legacy
		"claude-opus-4-0":            true, // legacy
	},
	"Google": {
		"gemini-3.1-pro-preview": true,
		"gemini-3.1-flash":       true,
		"gemini-3-flash-preview": true,
		"gemini-2.5-pro":         true,
		"gemini-2.5-flash":       true,
	},
	"xAI": {
		"grok-4":           true,
		"grok-4.1-alt":     true,
		"grok-4.1-fast":    true,
		"grok-4-fast":      true,
		"grok-code-fast-1": true,
		"grok-3":           true, // legacy
		"grok-3-mini":      true, // legacy
	},
	"Mistral": {
		"mistral-large-2512":          true,
		"mistral-medium-2505":         true,
		"mistral-small-2503":          true, // legacy
		"mistral-small-2506":          true,
		"mistral-small-creative-2512": true,
		"mistral-saba-2502":           true,
		"ministral-3b-2512":           true,
		"ministral-8b-2512":           true,
		"ministral-14b-2512":          true,
		"magistral-small-2509":        true,
		"magistral-medium-2509":       true,
		"devstral-medium-2507":        true,
		"devstral-2512":               true,
		"devstral-small-2512":         true,
		"codestral-2508":              true, // legacy
	},
	"DeepSeek": {
		"deepseek-reasoner": true,
		"deepseek-chat":     true,
	},
	"Meta": {
		"llama-4-maverick": true,
		"llama-4-scout":    true,
		"llama-3.3-70b":    true,
	},
	"Amazon": {
		"amazon-nova-micro":   true,
		"amazon-nova-lite":    true,
		"amazon-nova-pro":     true,
		"amazon-nova-premier": true,
		"amazon-nova-2-lite":  true,
		"amazon-nova-2-pro":   true,
	},
	"Cohere": {
		"command-a-03-2025":            true,
		"command-a-reasoning-08-2025":  true,
		"command-a-vision-07-2025":     true,
		"command-a-translate-08-2025":  true,
		"command-r7b-12-2024":          true,
	},
	"Perplexity": {
		"sonar":               true,
		"sonar-pro":           true,
		"sonar-reasoning-pro": true,
		"sonar-deep-research": true,
	},
	"AI21": {
		"jamba-large-1.7": true,
		"jamba-mini-1.7":  true,
	},
	"Moonshot": {
		"kimi-k2.5":            true,
		"kimi-k2-thinking":     true,
		"kimi-k2-0905-preview": true,
	},
	"Zhipu": {
		"glm-5":          true,
		"glm-5-code":     true,
		"glm-4.7":        true,
		"glm-4.7-flash":  true,
		"glm-4.7-flashx": true,
	},
	"NVIDIA": {
		"nvidia/nemotron-3-nano-30b-a3b":          true,
		"nvidia/llama-3.1-nemotron-ultra-253b-v1": true,
	},
	"Tencent": {
		"hunyuan-turbos": true,
		"hunyuan-t1":     true,
		"hunyuan-a13b":   true,
	},
	"Microsoft": {
		"phi-4":                     true,
		"phi-4-multimodal-instruct": true,
		"phi-4-reasoning":           true,
		"phi-4-reasoning-plus":      true,
	},
	"MiniMax": {
		"minimax-m2.5":     true,
		"minimax-m2":       true,
		"minimax-m2-her-2": true,
		"minimax-m2.1":     true, // legacy
	},
	"Xiaomi": {
		"mimo-v2-flash": true,
	},
	"Kuaishou": {
		"kat-coder-pro": true,
	},
}

const maxRetries = 3

func main() {
	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	hasChanges := false
	hasErrors := false
	providerOrder := []string{"OpenAI", "Anthropic", "Google", "Mistral", "xAI", "DeepSeek", "Zhipu", "MiniMax"}

	// Capture report output for GitHub issue creation.
	var report strings.Builder
	var allMissing []string
	var allNew []string

	logf := func(format string, args ...any) {
		line := fmt.Sprintf(format, args...)
		fmt.Print(line)
		report.WriteString(line)
	}

	logf("=== Model Registry Update Check ===\n")
	logf("Time: %s\n\n", time.Now().UTC().Format(time.RFC3339))

	for _, name := range providerOrder {
		src, ok := docSources[name]
		if !ok {
			logf("[%s] SKIP: no doc source configured\n", name)
			continue
		}

		ids, err := fetchModelsFromDocs(ctx, client, src)
		if err != nil {
			logf("[%s] ERROR: %v\n", name, err)
			hasErrors = true
			continue
		}

		known := knownModels[name]
		newModels, missing := diff(known, ids)

		logf("[%s] Docs returned %d model IDs, we track %d\n", name, len(ids), len(known))

		if len(newModels) > 0 {
			hasChanges = true
			sort.Strings(newModels)
			allNew = append(allNew, newModels...)
			logf("  NEW (%d):\n", len(newModels))
			for _, m := range newModels {
				logf("    + %s\n", m)
			}
		}
		if len(missing) > 0 {
			hasChanges = true
			sort.Strings(missing)
			allMissing = append(allMissing, missing...)
			logf("  MISSING from docs (%d):\n", len(missing))
			for _, m := range missing {
				logf("    - %s\n", m)
			}
		}
		if len(newModels) == 0 && len(missing) == 0 {
			logf("  OK: in sync\n")
		}
		logf("\n")
	}

	// Providers without scrapable documentation — just note them.
	logf("[Meta] SKIP: no scrapable model listing (models are provider-hosted)\n")
	logf("[Amazon] SKIP: no scrapable model listing (check AWS Bedrock console)\n")
	logf("[Cohere] SKIP: no scrapable model listing (check docs.cohere.com)\n")
	logf("[Perplexity] SKIP: no scrapable model listing (check docs.perplexity.ai)\n")
	logf("[AI21] SKIP: no scrapable model listing (check docs.ai21.com)\n")
	logf("[Moonshot] SKIP: no scrapable model listing (check platform.moonshot.cn)\n")
	logf("[NVIDIA] SKIP: no scrapable model listing (check build.nvidia.com)\n")
	logf("[Tencent] SKIP: no scrapable model listing (check cloud.tencent.com/product/hunyuan)\n")
	logf("[Microsoft] SKIP: no scrapable model listing (check ai.azure.com)\n")
	logf("[Xiaomi] SKIP: no scrapable model listing (check platform.xiaomimimo.com)\n")
	logf("[Kuaishou] SKIP: no scrapable model listing (check kwaipilot.com)\n")

	logf("\n=== Summary ===\n")
	if hasChanges {
		if hasErrors {
			logf("WARNING: Some providers failed to respond (see errors above).\n")
		}
		logf("Changes detected. Review the output above.\n")
		if len(allMissing) > 0 {
			createDeprecationIssue(ctx, client, allMissing, report.String())
		}
		if len(allNew) > 0 {
			createNewModelsIssue(ctx, client, allNew, report.String())
		}
		os.Exit(1)
	} else if hasErrors {
		logf("No model changes detected, but some providers could not be checked.\n")
		os.Exit(1)
	}
	logf("All tracked providers are in sync.\n")
	os.Exit(0)
}

// fetchModelsFromDocs fetches a public documentation page and extracts model IDs
// using the provider's regex pattern. No API keys needed.
func fetchModelsFromDocs(ctx context.Context, client *http.Client, src DocSource) ([]string, error) {
	var lastErr error
	for _, url := range src.URLs {
		ids, err := fetchAndExtract(ctx, client, url, src.Pattern)
		if err != nil {
			lastErr = err
			continue
		}
		if len(ids) > 0 {
			if src.ExcludePattern != nil {
				filtered := make([]string, 0, len(ids))
				for _, id := range ids {
					if !src.ExcludePattern.MatchString(id) {
						filtered = append(filtered, id)
					}
				}
				ids = filtered
			}
			if src.NormalizeRe != nil {
				for i, id := range ids {
					ids[i] = src.NormalizeRe.ReplaceAllString(id, src.NormalizeRepl)
				}
			}
			if src.NormalizeFunc != nil {
				for i, id := range ids {
					ids[i] = src.NormalizeFunc(id)
				}
			}
			if src.Lowercase {
				for i, id := range ids {
					ids[i] = strings.ToLower(id)
				}
			}
			seen := make(map[string]bool, len(ids))
			deduped := make([]string, 0, len(ids))
			for _, id := range ids {
				if !seen[id] {
					seen[id] = true
					deduped = append(deduped, id)
				}
			}
			ids = deduped
			if len(ids) > 0 {
				return ids, nil
			}
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("all URLs failed: %w", lastErr)
	}
	return nil, fmt.Errorf("no model IDs found in any URL")
}

// fetchAndExtract fetches a URL and extracts model IDs using a regex pattern.
func fetchAndExtract(ctx context.Context, client *http.Client, url string, pattern *regexp.Regexp) ([]string, error) {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "ModelRegistryUpdater/1.0")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024)) // 2MB max
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}
			continue
		}
		if err != nil {
			lastErr = err
			continue
		}

		// Extract unique model IDs using the regex pattern.
		matches := pattern.FindAllStringSubmatch(string(body), -1)
		seen := make(map[string]bool)
		var ids []string
		for _, m := range matches {
			if len(m) >= 2 {
				id := m[1]
				if !seen[id] {
					seen[id] = true
					ids = append(ids, id)
				}
			}
		}
		return ids, nil
	}
	return nil, fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}

// existingIssueCoversModels checks if any open issue with the auto-update label
// already mentions all of the given model IDs in its body. Returns true if a
// covering issue exists (meaning we should skip creating a new one).
func existingIssueCoversModels(ctx context.Context, client *http.Client, token, repo string, modelIDs []string) bool {
	searchURL := fmt.Sprintf("https://api.github.com/search/issues?q=repo:%s+state:open+label:auto-update",
		repo)
	searchReq, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return false
	}
	searchReq.Header.Set("Authorization", "Bearer "+token)
	searchReq.Header.Set("Accept", "application/vnd.github+json")

	searchResp, err := client.Do(searchReq)
	if err != nil {
		return false
	}
	defer searchResp.Body.Close()

	if searchResp.StatusCode != http.StatusOK {
		return false
	}

	var searchResult struct {
		TotalCount int `json:"total_count"`
		Items      []struct {
			Body string `json:"body"`
		} `json:"items"`
	}
	if err := json.NewDecoder(searchResp.Body).Decode(&searchResult); err != nil || searchResult.TotalCount == 0 {
		return false
	}

	// Check if any existing issue body already mentions ALL the model IDs.
	// Use word-boundary matching to avoid substring false positives
	// (e.g., "gpt-5" matching inside "gpt-5-mini").
	for _, issue := range searchResult.Items {
		allFound := true
		for _, id := range modelIDs {
			// Match the model ID as a whole token: preceded by start/non-alnum,
			// followed by end/non-alnum. This prevents "gpt-5" matching "gpt-5-mini".
			pat := `(?:^|[^a-zA-Z0-9_-])` + regexp.QuoteMeta(id) + `(?:$|[^a-zA-Z0-9_.-])`
			matched, _ := regexp.MatchString(pat, issue.Body)
			if !matched {
				allFound = false
				break
			}
		}
		if allFound {
			return true
		}
	}
	return false
}

// createGitHubIssue creates a GitHub issue with the given title, body, and
// the "auto-update" label. Returns silently if GITHUB_TOKEN or GITHUB_REPO
// environment variables are not set.
func createGitHubIssue(ctx context.Context, client *http.Client, title, body string) {
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPO")
	if token == "" || repo == "" {
		return
	}

	issueURL := fmt.Sprintf("https://api.github.com/repos/%s/issues", repo)
	payload := map[string]any{
		"title":  title,
		"body":   body,
		"labels": []string{"auto-update"},
	}
	bodyJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("[GitHub] failed to marshal issue body: %v\n", err)
		return
	}

	issueReq, err := http.NewRequestWithContext(ctx, http.MethodPost, issueURL, bytes.NewReader(bodyJSON))
	if err != nil {
		fmt.Printf("[GitHub] failed to create issue request: %v\n", err)
		return
	}
	issueReq.Header.Set("Authorization", "Bearer "+token)
	issueReq.Header.Set("Accept", "application/vnd.github+json")
	issueReq.Header.Set("Content-Type", "application/json")

	issueResp, err := client.Do(issueReq)
	if err != nil {
		fmt.Printf("[GitHub] failed to create issue: %v\n", err)
		return
	}
	defer issueResp.Body.Close()

	if issueResp.StatusCode == http.StatusCreated {
		var created struct {
			HTMLURL string `json:"html_url"`
		}
		_ = json.NewDecoder(issueResp.Body).Decode(&created)
		fmt.Printf("[GitHub] Issue created: %s\n", created.HTMLURL)
	} else {
		respBody, _ := io.ReadAll(io.LimitReader(issueResp.Body, 512))
		fmt.Printf("[GitHub] Failed to create issue (HTTP %d): %s\n", issueResp.StatusCode, string(respBody))
	}
}

// createNewModelsIssue creates a GitHub issue reporting newly detected model IDs.
// It checks for existing open issues that already cover the same model IDs to
// avoid duplicates.
func createNewModelsIssue(ctx context.Context, client *http.Client, newModelIDs []string, reportBody string) {
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPO")
	if token == "" || repo == "" {
		return
	}

	if existingIssueCoversModels(ctx, client, token, repo, newModelIDs) {
		fmt.Printf("[GitHub] Existing open issue already covers these new models, skipping.\n")
		return
	}

	today := time.Now().Format("2006-01-02")
	title := "New models detected - " + today

	sort.Strings(newModelIDs)
	var body strings.Builder
	body.WriteString("## New Models Detected\n\n")
	body.WriteString("The following model IDs were found in provider documentation but are not in the registry:\n\n")
	for _, id := range newModelIDs {
		body.WriteString(fmt.Sprintf("- `%s`\n", id))
	}
	body.WriteString("\n### Action Items\n\n")
	body.WriteString("- [ ] Add each model to `go-server/internal/models/data.go` (all 13 fields)\n")
	body.WriteString("- [ ] Add model IDs to `knownModels` in `go-server/cmd/updater/main.go`\n")
	body.WriteString("- [ ] Update `TestTotalModelCount` and `TestProviderCounts` in `data_test.go`\n")
	body.WriteString("- [ ] Run `go test ./... -v` to verify\n")
	body.WriteString("\n<details>\n<summary>Full update report</summary>\n\n```\n")
	body.WriteString(reportBody)
	body.WriteString("\n```\n</details>\n")

	createGitHubIssue(ctx, client, title, body.String())
}

// createDeprecationIssue creates a GitHub issue reporting models that were
// removed from provider documentation and may need deprecation. This replaces
// the old createDeprecationPR approach which was fundamentally broken: it could
// only update data.go's Status field via regex, but deprecation also requires
// updating data_test.go counts, knownModels in main.go, and the model's Notes
// field. Creating an issue lets a human (or CI-aware tool) handle all the
// required changes properly.
func createDeprecationIssue(ctx context.Context, client *http.Client, missingIDs []string, reportBody string) {
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPO")
	if token == "" || repo == "" {
		return
	}

	if existingIssueCoversModels(ctx, client, token, repo, missingIDs) {
		fmt.Printf("[GitHub] Existing open issue already covers these missing models, skipping.\n")
		return
	}

	today := time.Now().Format("2006-01-02")
	title := "Models removed from provider docs - " + today

	sort.Strings(missingIDs)
	var body strings.Builder
	body.WriteString("## Models Missing From Provider Documentation\n\n")
	body.WriteString("The following models are in our registry but were NOT found in their provider's public documentation.\n")
	body.WriteString("They may have been deprecated, renamed, or the documentation page may have changed.\n\n")
	for _, id := range missingIDs {
		body.WriteString(fmt.Sprintf("- `%s`\n", id))
	}
	body.WriteString("\n### Action Items\n\n")
	body.WriteString("For each model, verify whether it has actually been deprecated, then:\n\n")
	body.WriteString("- [ ] Set `Status: \"deprecated\"` in `go-server/internal/models/data.go`\n")
	body.WriteString("- [ ] Update the `Notes` field with deprecation context\n")
	body.WriteString("- [ ] Remove the model from `knownModels` in `go-server/cmd/updater/main.go`\n")
	body.WriteString("- [ ] Update `TestTotalModelCount` and `TestProviderCounts` in `data_test.go`\n")
	body.WriteString("- [ ] Run `go test ./... -v` to verify\n")
	body.WriteString("\n<details>\n<summary>Full update report</summary>\n\n```\n")
	body.WriteString(reportBody)
	body.WriteString("\n```\n</details>\n")

	createGitHubIssue(ctx, client, title, body.String())
}

// dateStampRe matches model IDs ending with a date stamp in YYYYMMDD or
// YYYY-MM-DD format (e.g. "gpt-5-2025-08-07" or "gpt-4.1-20250414").
var dateStampRe = regexp.MustCompile(`-(?:\d{8}|\d{4}-\d{2}-\d{2})$`)

// isDateStampVariant reports whether id ends with a date-stamp suffix,
// which indicates a pinned snapshot rather than a distinct new model.
func isDateStampVariant(id string) bool {
	return dateStampRe.MatchString(id)
}

// aliasSuffixes lists well-known suffixes that providers append to a base
// model ID to create convenience aliases (e.g. "gpt-5-chat-latest").
// IDs whose suffix (after the last dash relative to a known model) appears
// here are treated as aliases rather than new models.
var aliasSuffixes = map[string]bool{
	"latest": true, "beta": true, "preview": true,
	"chat-latest": true, "non-reasoning": true, "reasoning": true,
	"non-reasoning-latest": true, "reasoning-latest": true,
	"fast": true, "fast-beta": true, "fast-latest": true,
	"alt": true, "highspeed": true, "lightning": true,
	"mini": true, "nano": true,
	"mini-fast": true, "mini-fast-beta": true, "mini-fast-latest": true,
}

// isAllDigits reports whether s is a non-empty string composed entirely of
// ASCII digits.
func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isKnownAlias reports whether id is a variant of an already-known model.
// It checks three heuristics:
//  1. id is a prefix of a known ID whose remaining suffix is all-digits
//     (e.g. known "gpt-5-mini-2025" when id is "gpt-5-mini").
//  2. id extends a known ID with a well-known alias suffix
//     (e.g. "gpt-5-chat-latest" when "gpt-5" is known).
//  3. id shares a base name with a known ID and both have ≥2-digit numeric
//     suffixes (e.g. "codestral-25" or "codestral-2405" when "codestral-2508" is known).
func isKnownAlias(id string, known map[string]bool) bool {
	for knownID := range known {
		// Heuristic 1: known ID extends scraped ID with an all-digit suffix
		// e.g. known "gpt-5-mini-2025" when scraped ID is "gpt-5-mini"
		if knownID != id && strings.HasPrefix(knownID, id+"-") {
			suffix := knownID[len(id)+1:]
			if isAllDigits(suffix) {
				return true
			}
		}
		// Heuristic 2: scraped ID extends known ID with a well-known suffix
		// e.g. "gpt-5-chat-latest" when "gpt-5" is known
		if id != knownID && strings.HasPrefix(id, knownID+"-") {
			suffix := id[len(knownID)+1:]
			if aliasSuffixes[suffix] {
				return true
			}
		}
		// Heuristic 2b (reverse): known ID extends scraped ID with a well-known suffix
		// e.g. scraped "gemini-3-flash" when known is "gemini-3-flash-preview"
		if knownID != id && strings.HasPrefix(knownID, id+"-") {
			suffix := knownID[len(id)+1:]
			if aliasSuffixes[suffix] {
				return true
			}
		}
	}
	// Heuristic 3: shared base name with numeric suffixes (≥2 digits)
	// e.g. "codestral-25" matches "codestral-2508"
	if lastDash := strings.LastIndex(id, "-"); lastDash > 0 {
		idBase := id[:lastDash]
		idSuffix := id[lastDash+1:]
		if isAllDigits(idSuffix) && len(idSuffix) >= 2 {
			if known[idBase] {
				return true
			}
			for knownID := range known {
				if kd := strings.LastIndex(knownID, "-"); kd > 0 {
					if idBase == knownID[:kd] && isAllDigits(knownID[kd+1:]) {
						return true
					}
				}
			}
		}
	}
	return false
}

// diff compares the set of known model IDs against those scraped from
// documentation. It returns IDs found in docs but not in known (newModels)
// and IDs in known but absent from docs (missing), filtering out date-stamp
// variants and known aliases from the "new" list.
func diff(known map[string]bool, docIDs []string) (newModels, missing []string) {
	docSet := make(map[string]bool, len(docIDs))
	for _, id := range docIDs {
		docSet[id] = true
	}

	for _, id := range docIDs {
		if known[id] {
			continue
		}
		if isDateStampVariant(id) {
			continue
		}
		if isKnownAlias(id, known) {
			continue
		}
		newModels = append(newModels, id)
	}

	for id := range known {
		if !docSet[id] {
			missing = append(missing, id)
		}
	}

	return newModels, missing
}
