package main

import (
	"bytes"
	"context"
	"encoding/base64"
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
	URLs           []string       // URLs to try in order (fallbacks)
	Pattern        *regexp.Regexp // Regex to extract model IDs from page content
	ExcludePattern *regexp.Regexp // Optional: exclude matched IDs containing this pattern
	Lowercase      bool           // Lowercase extracted IDs before comparison
	NormalizeRe    *regexp.Regexp // Optional: normalize extracted IDs (regex)
	NormalizeRepl  string         // Replacement for NormalizeRe
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
		ExcludePattern: regexp.MustCompile(`^gpt-(?:3\.|4(?:-|$))|^o1(?:-|$)`),
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
		Pattern: regexp.MustCompile(`((?:mistral|devstral|codestral|ministral|magistral)-[a-z]*-?[0-9]{2,4}(?:-[0-9]{4})?)`),
	},
	"xAI": {
		URLs: []string{
			"https://docs.x.ai/docs/models",
		},
		Pattern:        regexp.MustCompile(`(grok-(?:[0-9]+(?:\.[0-9]+)?(?:-[a-z0-9-]*)?|code-[a-z0-9-]+))`),
		ExcludePattern: regexp.MustCompile(`(?i)(?:image|vision|imagine|video)`),
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
		Pattern:   regexp.MustCompile(`(?i)(GLM-[0-9]+(?:\.[0-9]+)?(?:-[A-Za-z]+)*)`),
		Lowercase: true,
	},
	"MiniMax": {
		URLs: []string{
			"https://platform.minimax.io/docs/guides/models-intro",
			"https://intl.minimaxi.com/",
		},
		Pattern:   regexp.MustCompile(`(?i)(MiniMax-M[0-9](?:\.[0-9]+)?(?:-[a-z0-9]+)*)`),
		Lowercase: true,
	},
}

// knownModels maps provider -> set of model IDs we track in the registry.
var knownModels = map[string]map[string]bool{
	"OpenAI": {
		"gpt-5.2":            true,
		"gpt-5.2-codex":      true,
		"gpt-5.2-pro":        true,
		"gpt-5.1":            true,
		"gpt-5.1-codex":      true,
		"gpt-5.1-codex-mini": true,
		"gpt-5.1-mini":       true,
		"gpt-5":              true,
		"gpt-5-mini":         true,
		"gpt-5-nano":         true,
		"gpt-4.1-mini":       true,
		"gpt-4.1-nano":       true,
		"o3":                 true,
		"o3-pro":             true,
		"o3-deep-research":   true,
		"o4-mini":            true,
		"o3-mini":            true,
		"gpt-4.1":            true,
		"gpt-4o":             true,
		"gpt-4o-mini":        true,
	},
	"Anthropic": {
		"claude-sonnet-4-6":          true,
		"claude-opus-4-6":            true,
		"claude-sonnet-4-5-20250929": true,
		"claude-haiku-4-5-20251001":  true,
		"claude-opus-4-5":            true,
		"claude-opus-4-1":            true,
		"claude-sonnet-4-0":          true,
		"claude-3-7-sonnet-20250219": true,
		"claude-opus-4-0":            true,
	},
	"Google": {
		"gemini-3-pro-preview":       true,
		"gemini-3-pro-image-preview": true,
		"gemini-3-flash-preview":     true,
		"gemini-2.5-pro":             true,
		"gemini-2.5-flash":           true,
		"gemini-2.5-flash-lite":      true,
		"gemini-2.0-flash-lite":      true,
		"gemini-2.0-flash":           true,
	},
	"xAI": {
		"grok-4":           true,
		"grok-4.1":         true,
		"grok-4.1-fast":    true,
		"grok-4-fast":      true,
		"grok-code-fast-1": true,
		"grok-3":           true,
		"grok-3-mini":      true,
	},
	"Mistral": {
		"mistral-large-2512":   true,
		"mistral-medium-2505":  true,
		"mistral-small-2506":   true,
		"ministral-3b-2512":    true,
		"ministral-8b-2512":    true,
		"ministral-14b-2512":   true,
		"magistral-small-2509": true,
		"magistral-medium-2509": true,
		"devstral-2512":        true,
		"devstral-small-2512":  true,
		"codestral-2508":       true,
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
		"amazon-nova-micro":   true,
		"amazon-nova-lite":    true,
		"amazon-nova-pro":     true,
		"amazon-nova-premier": true,
		"amazon-nova-2-lite":  true,
		"amazon-nova-2-pro":   true,
	},
	"Cohere": {
		"command-a-03-2025":           true,
		"command-a-reasoning-08-2025": true,
		"command-a-vision-07-2025":    true,
		"command-r7b-12-2024":         true,
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
		"glm-4.7":        true,
		"glm-4.7-flashx": true,
		"glm-4.6v":       true,
	},
	"NVIDIA": {
		"nvidia/nemotron-3-nano-30b-a3b":            true,
		"nvidia/llama-3.1-nemotron-ultra-253b-v1": true,
	},
	"Tencent": {
		"hunyuan-turbos": true,
		"hunyuan-t1":     true,
		"hunyuan-a13b":   true,
	},
	"Microsoft": {
		"phi-4":                      true,
		"phi-4-multimodal-instruct":  true,
		"phi-4-reasoning-plus":       true,
	},
	"MiniMax": {
		"minimax-m2.1": true,
		"minimax-01":   true,
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

	// Capture report output for GitHub issue/PR creation.
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
			createDeprecationPR(ctx, client, allMissing, report.String())
		}
		if len(allNew) > 0 {
			createGitHubIssue(ctx, client, report.String())
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
				filtered := ids[:0]
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

// createGitHubIssue creates a GitHub issue with the update report.
// Requires GITHUB_TOKEN and GITHUB_REPO (e.g. "owner/repo") environment variables.
func createGitHubIssue(ctx context.Context, client *http.Client, reportBody string) {
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPO")
	if token == "" || repo == "" {
		return
	}

	today := time.Now().Format("2006-01-02")
	title := "Model Update Detected - " + today

	// Check for existing open issue with the same title to avoid duplicates.
	searchURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s+repo:%s+state:open+label:auto-update",
		strings.ReplaceAll(title, " ", "+"), repo)
	searchReq, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		fmt.Printf("[GitHub] failed to create search request: %v\n", err)
		return
	}
	searchReq.Header.Set("Authorization", "Bearer "+token)
	searchReq.Header.Set("Accept", "application/vnd.github+json")

	searchResp, err := client.Do(searchReq)
	if err != nil {
		fmt.Printf("[GitHub] failed to search issues: %v\n", err)
		return
	}
	defer searchResp.Body.Close()

	if searchResp.StatusCode == http.StatusOK {
		var searchResult struct {
			TotalCount int `json:"total_count"`
		}
		if err := json.NewDecoder(searchResp.Body).Decode(&searchResult); err == nil && searchResult.TotalCount > 0 {
			fmt.Printf("[GitHub] Issue already exists for today, skipping.\n")
			return
		}
	}

	issueURL := fmt.Sprintf("https://api.github.com/repos/%s/issues", repo)
	body := map[string]any{
		"title":  title,
		"body":   "```\n" + reportBody + "\n```",
		"labels": []string{"auto-update"},
	}
	bodyJSON, err := json.Marshal(body)
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

// createDeprecationPR creates a GitHub PR that changes the status of missing models
// to "deprecated" in data.go. Uses the GitHub Contents API — no git clone needed.
func createDeprecationPR(ctx context.Context, client *http.Client, missingIDs []string, reportBody string) {
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPO")
	if token == "" || repo == "" {
		return
	}

	apiBase := "https://api.github.com"
	filePath := "go-server/internal/models/data.go"
	today := time.Now().Format("2006-01-02")
	branchName := "auto-deprecate-" + today

	doReq := func(method, url string, body any) (*http.Response, error) {
		var reader io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			reader = bytes.NewReader(b)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, reader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return client.Do(req)
	}

	// Step 1: Get current data.go content and blob SHA.
	fileURL := fmt.Sprintf("%s/repos/%s/contents/%s", apiBase, repo, filePath)
	fileResp, err := doReq(http.MethodGet, fileURL, nil)
	if err != nil {
		fmt.Printf("[GitHub PR] failed to get file: %v\n", err)
		return
	}
	defer fileResp.Body.Close()
	if fileResp.StatusCode != http.StatusOK {
		fmt.Printf("[GitHub PR] failed to get file: HTTP %d\n", fileResp.StatusCode)
		return
	}

	var fileInfo struct {
		SHA     string `json:"sha"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(fileResp.Body).Decode(&fileInfo); err != nil {
		fmt.Printf("[GitHub PR] failed to decode file info: %v\n", err)
		return
	}

	rawContent, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(fileInfo.Content, "\n", ""))
	if err != nil {
		fmt.Printf("[GitHub PR] failed to decode file content: %v\n", err)
		return
	}

	// Step 2: Apply deprecation changes.
	content := string(rawContent)
	changed := false
	for _, id := range missingIDs {
		pattern := fmt.Sprintf(`("%s":\s*\{[^}]*Status:\s*)"(?:current|legacy)"`, regexp.QuoteMeta(id))
		re := regexp.MustCompile(pattern)
		if re.MatchString(content) {
			content = re.ReplaceAllString(content, `${1}"deprecated"`)
			changed = true
			fmt.Printf("[GitHub PR] Marking %s as deprecated\n", id)
		}
	}

	if !changed {
		fmt.Printf("[GitHub PR] No status changes needed in data.go\n")
		return
	}

	// Step 3: Get main branch SHA.
	refURL := fmt.Sprintf("%s/repos/%s/git/ref/heads/main", apiBase, repo)
	refResp, err := doReq(http.MethodGet, refURL, nil)
	if err != nil {
		fmt.Printf("[GitHub PR] failed to get main ref: %v\n", err)
		return
	}
	var refInfo struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if refResp.StatusCode != http.StatusOK {
		refResp.Body.Close()
		fmt.Printf("[GitHub PR] failed to get main ref: HTTP %d\n", refResp.StatusCode)
		return
	}
	err = json.NewDecoder(refResp.Body).Decode(&refInfo)
	refResp.Body.Close()
	if err != nil {
		fmt.Printf("[GitHub PR] failed to decode ref info: %v\n", err)
		return
	}

	// Step 4: Create new branch.
	createRefURL := fmt.Sprintf("%s/repos/%s/git/refs", apiBase, repo)
	branchResp, err := doReq(http.MethodPost, createRefURL, map[string]string{
		"ref": "refs/heads/" + branchName,
		"sha": refInfo.Object.SHA,
	})
	if err != nil {
		fmt.Printf("[GitHub PR] failed to create branch: %v\n", err)
		return
	}
	if branchResp.StatusCode != http.StatusCreated {
		errBody, _ := io.ReadAll(io.LimitReader(branchResp.Body, 512))
		branchResp.Body.Close()
		fmt.Printf("[GitHub PR] failed to create branch (HTTP %d): %s\n", branchResp.StatusCode, string(errBody))
		return
	}
	branchResp.Body.Close()

	// Step 5: Update file on new branch.
	sort.Strings(missingIDs)
	commitMsg := fmt.Sprintf("auto: deprecate %s (removed from provider docs)", strings.Join(missingIDs, ", "))
	updateResp, err := doReq(http.MethodPut, fileURL, map[string]string{
		"message": commitMsg,
		"content": base64.StdEncoding.EncodeToString([]byte(content)),
		"sha":     fileInfo.SHA,
		"branch":  branchName,
	})
	if err != nil {
		fmt.Printf("[GitHub PR] failed to update file: %v\n", err)
		return
	}
	if updateResp.StatusCode != http.StatusOK && updateResp.StatusCode != http.StatusCreated {
		errBody, _ := io.ReadAll(io.LimitReader(updateResp.Body, 512))
		updateResp.Body.Close()
		fmt.Printf("[GitHub PR] failed to update file (HTTP %d): %s\n", updateResp.StatusCode, string(errBody))
		return
	}
	updateResp.Body.Close()

	// Step 6: Create pull request.
	prURL := fmt.Sprintf("%s/repos/%s/pulls", apiBase, repo)
	prBody := "## Auto-Deprecation\n\nModels removed from provider docs:\n"
	for _, id := range missingIDs {
		prBody += fmt.Sprintf("- `%s`\n", id)
	}
	prBody += fmt.Sprintf("\n<details>\n<summary>Full update report</summary>\n\n```\n%s\n```\n</details>", reportBody)

	prResp, err := doReq(http.MethodPost, prURL, map[string]any{
		"title": "auto: deprecate models removed from provider docs — " + today,
		"body":  prBody,
		"head":  branchName,
		"base":  "main",
	})
	if err != nil {
		fmt.Printf("[GitHub PR] failed to create PR: %v\n", err)
		return
	}
	defer prResp.Body.Close()

	if prResp.StatusCode == http.StatusCreated {
		var pr struct {
			HTMLURL string `json:"html_url"`
			Number  int    `json:"number"`
		}
		_ = json.NewDecoder(prResp.Body).Decode(&pr)
		fmt.Printf("[GitHub PR] Created: %s\n", pr.HTMLURL)
		labelURL := fmt.Sprintf("%s/repos/%s/issues/%d/labels", apiBase, repo, pr.Number)
		labelResp, labelErr := doReq(http.MethodPost, labelURL, []string{"auto-update"})
		if labelErr == nil {
			labelResp.Body.Close()
		}
	} else {
		errBody, _ := io.ReadAll(io.LimitReader(prResp.Body, 512))
		fmt.Printf("[GitHub PR] Failed to create PR (HTTP %d): %s\n", prResp.StatusCode, string(errBody))
	}
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
//  3. id shares a base name with a known ID and both have ≥4-digit numeric
//     suffixes (e.g. "codestral-2405" when "codestral-2508" is known).
func isKnownAlias(id string, known map[string]bool) bool {
	for knownID := range known {
		if knownID != id && strings.HasPrefix(knownID, id+"-") {
			suffix := knownID[len(id)+1:]
			if isAllDigits(suffix) {
				return true
			}
		}
		if id != knownID && strings.HasPrefix(id, knownID+"-") {
			suffix := id[len(knownID)+1:]
			if aliasSuffixes[suffix] {
				return true
			}
		}
	}
	if lastDash := strings.LastIndex(id, "-"); lastDash > 0 {
		idBase := id[:lastDash]
		idSuffix := id[lastDash+1:]
		if isAllDigits(idSuffix) && len(idSuffix) >= 4 {
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
