# Feature Improvements Report

> Codebase audit performed on 2026-02-10 against the `universal-model-registry` Go MCP server.

---

## Architecture Summary

| Component | Location | Description |
|-----------|----------|-------------|
| Entry point | `go-server/cmd/server/main.go` | Registers 6 tools + 3 resources, supports stdio/SSE/streamable-HTTP transports |
| Model data | `go-server/internal/models/data.go` | ~40 models across 7 providers as a Go map literal |
| Tool handlers | `go-server/internal/tools/` | `list_models`, `get_model_info`, `search_models`, `recommend_model`, `check_model_status`, `compare_models` |
| Resources | `go-server/internal/resources/resources.go` | `model://registry/all`, `model://registry/current`, `model://registry/pricing` |
| Middleware | `go-server/internal/middleware/ratelimit.go` | Per-IP rate limiting + connection caps |
| Auto-updater | `go-server/cmd/updater/main.go` | Weekly GitHub Action that diffs provider APIs against tracked models |
| CI | `.github/workflows/go-ci.yml` | `go test`, `go build`, `golangci-lint` |
| Docker | `Dockerfile` | Multi-stage Alpine build with health check |

---

## P0 — Quick Wins (< 1 hour each)

### 1. Enhanced Health Check Endpoint

**Current state:** `/health` returns a bare `"ok"` string (main.go:210-213). No structured data for monitoring.

**What it does:** Return JSON with model count, server version, uptime, and transport type. This enables monitoring dashboards and load balancer health checks that need more than a 200 OK.

**Effort:** ~20 minutes

**Implementation:**
- File: `go-server/cmd/server/main.go`
- Add a package-level `var startTime = time.Now()` at the top of `main.go`
- Replace the health handler with:
```go
mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "status":      "ok",
        "models":      len(models.Models),
        "version":     "1.0.1",
        "uptime_secs": int(time.Since(startTime).Seconds()),
        "transport":   transport,
    })
})
```

**Priority:** P0 — trivial change, immediately useful for ops.

---

### 2. Better Error Messages with "Did You Mean?" Suggestions

**Current state:** Tool handlers return inconsistent error messages:
- `get_model_info` lists ALL known model IDs on not-found (info.go:15-21) — noisy for 40+ models.
- `check_model_status` gives a generic "may be misspelled" message (status.go:21-22).
- `compare_models` just lists the not-found IDs (compare.go:37).
- `search_models` says "No models found matching 'X'" (search.go:23).

**What it does:** When a model isn't found, suggest the 3 closest matches using edit-distance. This helps LLM clients self-correct without a round-trip, and gives human users actionable hints.

**Effort:** ~45 minutes

**Implementation:**
- File: `go-server/internal/tools/helpers.go`
- Add a `SuggestModels(input string, n int) []string` function that computes Levenshtein distance against all model IDs and returns the top `n` closest.
- A minimal Levenshtein implementation is ~20 lines in Go (no external dependency needed).
- Update `FindModel` to return suggestions when no match is found (change signature to `FindModel(id string) (Model, bool, []string)`), or add a separate `Suggestions(id string) string` helper.
- Update `GetModelInfo`, `CheckModelStatus`, and `CompareModels` to include suggestions in their error output:
  ```
  Model `cladue-opus-4-6` not found. Did you mean: claude-opus-4-6, claude-opus-4-5, claude-opus-4-1?
  ```

**Priority:** P0 — high-impact UX improvement, especially for LLM tool-calling where typos are common.

---

### 3. Model Aliases Support

**Current state:** Several models have known aliases mentioned in their Notes field (e.g., `claude-sonnet-4-5` -> `claude-sonnet-4-5-20250929`, `claude-haiku-4-5` -> `claude-haiku-4-5-20251001`), but `FindModel` doesn't resolve these. Users must know the exact date-stamped ID.

**What it does:** Add an aliases map so common shorthand IDs resolve to their canonical model entry. This is the single most common failure mode when LLMs call `get_model_info` or `check_model_status`.

**Effort:** ~30 minutes

**Implementation:**
- File: `go-server/internal/models/models.go` — add a new exported map:
```go
// Aliases maps common shorthand model IDs to their canonical registry key.
var Aliases = map[string]string{
    "claude-sonnet-4-5":           "claude-sonnet-4-5-20250929",
    "claude-haiku-4-5":            "claude-haiku-4-5-20251001",
    "claude-3-7-sonnet-latest":    "claude-3-7-sonnet-20250219",
    "claude-opus-4-5-20251101":    "claude-opus-4-5",
    "claude-sonnet-4-20250514":    "claude-sonnet-4-0",
    "claude-opus-4-20250514":      "claude-opus-4-0",
    "claude-opus-4-1-20250805":    "claude-opus-4-1",
    "gpt-4o-2024-05-13":           "gpt-4o",
    "gemini-2.5-pro-preview-03-25":"gemini-2.5-pro",
}
```
- File: `go-server/internal/tools/helpers.go` — update `FindModel`:
```go
func FindModel(modelID string) (models.Model, bool) {
    // Exact match
    if m, ok := models.Models[modelID]; ok {
        return m, true
    }
    // Alias resolution
    if canonical, ok := models.Aliases[modelID]; ok {
        if m, ok := models.Models[canonical]; ok {
            return m, true
        }
    }
    // Case-insensitive / partial match (existing logic)
    ...
}
```

**Priority:** P0 — directly fixes the most common lookup failures.

---

### 4. Fix Duplicated `formatInt` Helper

**Current state:** `formatInt` is defined in both `tools/helpers.go:115-139` and `resources/resources.go:65-85` with slightly different implementations (the tools version handles negatives, the resources version doesn't).

**What it does:** Eliminate code duplication by extracting the helper into a shared location.

**Effort:** ~10 minutes

**Implementation:**
- Move the more complete `formatInt` (from `tools/helpers.go`, which handles negatives) to `go-server/internal/models/models.go` as `FormatInt` (exported).
- Replace both usages in `tools/helpers.go` and `resources/resources.go` with `models.FormatInt()`.
- Delete both local copies.

**Priority:** P0 — prevents divergent behavior and reduces maintenance burden.

---

### 5. Fix Updater `knownModels` Mismatch

**Current state:** The updater's `knownModels["Mistral"]` map contains `"mistral-large-3-25-12"` (cmd/updater/main.go:78), but the actual model ID in `data.go` is `"mistral-large-2512"`. This means the updater will always report a false diff for Mistral.

**What it does:** Fix the updater so weekly checks are accurate.

**Effort:** ~5 minutes

**Implementation:**
- File: `go-server/cmd/updater/main.go:78`
- Change `"mistral-large-3-25-12": true` to `"mistral-large-2512": true`

**Priority:** P0 — bug fix, the auto-update workflow is currently broken for Mistral.

---

## P1 — Medium Effort (1-4 hours each)

### 6. Fuzzy Search in `search_models`

**Current state:** `SearchModels` (search.go) uses `strings.Contains` for substring matching. This works for partial matches like "claud" but fails on typos like "cladue" or "gemin".

**What it does:** Add fuzzy matching so typos still return relevant results. This dramatically improves the tool's usefulness for LLM clients which occasionally misspell model names.

**Effort:** ~2 hours

**Implementation:**
- File: `go-server/internal/tools/search.go`
- Add a scoring function that combines:
  1. **Exact substring match** (current behavior) — score 100
  2. **Levenshtein distance** — score inversely proportional to edit distance (reuse the function from improvement #2)
  3. **Trigram overlap** — split query and target into 3-char chunks, score = shared trigrams / total trigrams
- Return results sorted by score descending, with a minimum threshold to avoid garbage matches.
- Keep zero-dependency: both Levenshtein and trigram matching are straightforward to implement in Go.
- Update the "no results" message to suggest the closest model names even when no fuzzy match exceeds the threshold:
  ```
  No models found matching 'cladue'. Closest matches: claude-opus-4-6, claude-sonnet-4-5-20250929
  ```

**Priority:** P1 — meaningful UX improvement, moderate implementation effort.

---

### 7. Provider Summary Tool

**Current state:** There's no way to get a high-level overview of what each provider offers without calling `list_models` for every provider individually. LLMs frequently want to compare providers before recommending a model.

**What it does:** New `provider_summary` tool that returns a table with each provider's model count, newest model, cheapest model, and price range.

**Effort:** ~1.5 hours

**Implementation:**
- New file: `go-server/internal/tools/provider_summary.go`
- New input type (empty — no parameters needed):
```go
type ProviderSummaryInput struct{}
```
- Function `ProviderSummary() string` that:
  1. Groups `models.Models` by provider
  2. For each provider, computes: total count, current count, newest (by ReleaseDate), cheapest (by PricingInput), price range
  3. Returns a markdown table:
  ```
  | Provider | Models | Current | Newest Model | Cheapest | Input $/1M Range |
  ```
- Register in `main.go` alongside the other tools.
- Add tests in `tools_test.go`.

**Priority:** P1 — fills a real information gap, clean implementation.

---

### 8. Model Changelog Resource

**Current state:** No way to see which models were recently added, deprecated, or updated. The auto-updater detects changes but doesn't expose them to MCP clients.

**What it does:** New `model://registry/changelog` resource that returns models sorted by release date, grouped into "recent" (last 3 months), "legacy" (recently deprecated), etc.

**Effort:** ~2 hours

**Implementation:**
- File: `go-server/internal/resources/resources.go` — add `Changelog() string` function:
  1. Sort all models by `ReleaseDate` descending
  2. Group into sections: "Recently Released" (last 3 months), "Recently Deprecated/Legacy" (status != current, sorted by date)
  3. Format as markdown with release dates
- File: `go-server/cmd/server/main.go` — register the new resource:
```go
server.AddResource(&mcp.Resource{
    URI:         "model://registry/changelog",
    Name:        "model-changelog",
    Description: "Recently added, updated, and deprecated models.",
    MIMEType:    "text/markdown",
}, ...)
```
- To track actual changes over time (not just release dates), optionally add a `LastUpdated` field to the Model struct, or maintain a separate changelog data file.

**Priority:** P1 — very useful for keeping LLM applications up to date.

---

### 9. Fix Nondeterministic `FindModel` Partial Match

**Current state:** `FindModel` in helpers.go:148-155 iterates over a Go map for partial matching. Map iteration order in Go is randomized, so if multiple model IDs contain the search string, the result is nondeterministic. For example, searching for "opus-4" could match any of `claude-opus-4-6`, `claude-opus-4-5`, `claude-opus-4-1`, `claude-opus-4-0`.

**What it does:** Make partial matching deterministic by preferring: (1) exact case-insensitive match, (2) shortest ID containing the substring, (3) alphabetically first on tie.

**Effort:** ~30 minutes

**Implementation:**
- File: `go-server/internal/tools/helpers.go`
- Replace the single-pass map iteration with a collect-and-sort approach:
```go
// Collect all partial matches
var candidates []models.Model
for key, m := range models.Models {
    if strings.ToLower(key) == lower {
        return m, true // Exact case-insensitive — return immediately
    }
    if strings.Contains(strings.ToLower(key), lower) {
        candidates = append(candidates, m)
    }
}
// Sort: shortest ID first, then alphabetically
sort.SliceStable(candidates, func(i, j int) bool {
    if len(candidates[i].ID) != len(candidates[j].ID) {
        return len(candidates[i].ID) < len(candidates[j].ID)
    }
    return candidates[i].ID < candidates[j].ID
})
if len(candidates) > 0 {
    return candidates[0], true
}
```
- Add test cases for ambiguous partial matches.

**Priority:** P1 — fixes a subtle correctness bug.

---

### 10. Fix `recencyBonus` Staleness

**Current state:** `recencyBonus` in recommend.go:184-204 uses a hardcoded `2024` base year and divides by `16.0` to normalize. As time passes, all models will max out at 1.5 bonus and the function becomes useless for differentiation.

**What it does:** Make the recency bonus relative to the current date so it naturally scales over time.

**Effort:** ~20 minutes

**Implementation:**
- File: `go-server/internal/tools/recommend.go`
- Replace the static calculation with a relative one:
```go
func recencyBonus(releaseDate string) float64 {
    parts := strings.Split(releaseDate, "-")
    if len(parts) < 2 { return 0 }
    year, err1 := strconv.Atoi(parts[0])
    month, err2 := strconv.Atoi(parts[1])
    if err1 != nil || err2 != nil { return 0 }

    now := time.Now()
    releaseMonths := year*12 + month
    currentMonths := now.Year()*12 + int(now.Month())
    monthsAgo := float64(currentMonths - releaseMonths)

    // Models released in the last 6 months get full bonus, decaying to 0 at 18 months
    bonus := 1.5 * (1.0 - monthsAgo/18.0)
    if bonus < 0 { bonus = 0 }
    if bonus > 1.5 { bonus = 1.5 }
    return bonus
}
```

**Priority:** P1 — the current implementation will become ineffective by mid-2026.

---

### 11. Search by Status Field

**Current state:** `SearchModels` searches ID, DisplayName, Provider, and Notes — but not the `Status` field. Searching for "deprecated" only works if "deprecated" happens to appear in a model's Notes.

**What it does:** Add Status to the searched fields so users can search for "deprecated" or "legacy" models directly.

**Effort:** ~10 minutes

**Implementation:**
- File: `go-server/internal/tools/search.go:16` — add one line:
```go
strings.Contains(strings.ToLower(m.Status), q) ||
```
- Add a test: `TestSearchModels_ByStatus` that verifies searching "deprecated" returns deprecated models.

**Priority:** P1 — trivial change, completes search coverage.

---

## P2 — Larger Features (4+ hours)

### 12. Batch `check_model_status` Tool

**Current state:** `check_model_status` accepts a single model ID. LLM clients frequently need to validate multiple model IDs (e.g., checking a codebase's model references), requiring N separate tool calls.

**What it does:** New `batch_check_status` tool that accepts an array of model IDs and returns status for all in one call.

**Effort:** ~2-3 hours

**Implementation:**
- File: `go-server/internal/tools/status.go` — add:
```go
type BatchCheckStatusInput struct {
    ModelIDs []string `json:"model_ids" jsonschema:"List of model IDs to check (max 20)"`
}

func BatchCheckStatus(modelIDs []string) string {
    if len(modelIDs) > 20 { modelIDs = modelIDs[:20] }
    var sections []string
    for _, id := range modelIDs {
        sections = append(sections, CheckModelStatus(id))
    }
    return strings.Join(sections, "\n\n---\n\n")
}
```
- File: `go-server/cmd/server/main.go` — register the new tool.
- Add tests for batch with mixed found/not-found, empty list, over-limit.

**Priority:** P2 — nice efficiency gain, straightforward extension of existing code.

---

### 13. Capability Field Expansion

**Current state:** The Model struct only has two boolean capability fields: `Vision` and `Reasoning`. This doesn't capture other important capabilities like function/tool calling, JSON mode, streaming, file uploads, audio, etc.

**What it does:** Replace boolean fields with a `Capabilities []string` slice for extensible capability tracking.

**Effort:** ~4-6 hours (data migration + all tool/resource/test updates)

**Implementation:**
- File: `go-server/internal/models/models.go` — change Model struct:
```go
type Model struct {
    // ... existing fields ...
    Capabilities []string `json:"capabilities"` // e.g., ["vision", "reasoning", "tool_calling", "json_mode"]
}
```
- Update `data.go` — migrate all model entries from `Vision: true, Reasoning: true` to `Capabilities: []string{"vision", "reasoning", "tool_calling"}`.
- Update all tool handlers that reference `m.Vision` or `m.Reasoning` to use a helper `HasCapability(m, "vision")`.
- Update `FilterModels` capability filter to check the slice.
- Update all tests.

**Priority:** P2 — significant structural improvement but high blast radius.

---

### 14. Webhook Notifications for Model Updates

**Current state:** The auto-updater runs weekly and creates a GitHub issue when changes are detected. There's no way for downstream consumers to be notified programmatically.

**What it does:** Add an optional webhook system where the updater can POST a JSON payload to configured endpoints when model changes are detected.

**Effort:** ~6-8 hours

**Implementation:**
- File: `go-server/cmd/updater/main.go` — add webhook dispatch after change detection:
  1. Read webhook URLs from environment variable `WEBHOOK_URLS` (comma-separated)
  2. Construct a JSON payload: `{ "timestamp": "...", "changes": [{ "provider": "...", "new": [...], "removed": [...] }] }`
  3. POST to each URL with a timeout and retry
- File: `.github/workflows/auto-update.yml` — add `WEBHOOK_URLS` secret
- Consider adding HMAC signature verification for webhook security.

**Priority:** P2 — useful for production deployments, significant effort.

---

### 15. Resource Templates for Per-Provider Data

**Current state:** Resources serve the entire registry (`model://registry/all`) or all current models. There's no resource for a specific provider's models.

**What it does:** Add MCP resource templates like `model://registry/provider/{provider_name}` that return models filtered by provider. This gives MCP clients more granular resource access.

**Effort:** ~2 hours

**Implementation:**
- File: `go-server/cmd/server/main.go` — use `AddResourceTemplate`:
```go
server.AddResourceTemplate(
    &mcp.ResourceTemplate{
        URITemplate: "model://registry/provider/{provider}",
        Name:        "provider-models",
        Description: "JSON dump of models from a specific provider.",
        MIMEType:    "application/json",
    },
    func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
        provider := extractProviderFromURI(req.Params.URI)
        return &mcp.ReadResourceResult{
            Contents: []*mcp.ResourceContents{{
                URI:      req.Params.URI,
                MIMEType: "application/json",
                Text:     resources.ProviderModels(provider),
            }},
        }, nil
    },
)
```
- File: `go-server/internal/resources/resources.go` — add `ProviderModels(provider string) string`.

**Priority:** P2 — nice MCP-native feature, depends on Go SDK template support.

---

## Summary Priority Matrix

| # | Improvement | Priority | Effort | Impact |
|---|-------------|----------|--------|--------|
| 5 | Fix updater Mistral key mismatch | **P0** | 5 min | Fixes broken CI |
| 4 | Deduplicate `formatInt` | **P0** | 10 min | Code quality |
| 3 | Model aliases support | **P0** | 30 min | Fixes most common lookup failures |
| 1 | Enhanced health check | **P0** | 20 min | Ops visibility |
| 2 | "Did you mean?" suggestions | **P0** | 45 min | Major UX improvement |
| 11 | Search by status field | **P1** | 10 min | Completes search coverage |
| 10 | Fix `recencyBonus` staleness | **P1** | 20 min | Fixes time-bomb bug |
| 9 | Deterministic `FindModel` | **P1** | 30 min | Correctness fix |
| 7 | Provider summary tool | **P1** | 1.5 hr | New useful tool |
| 8 | Model changelog resource | **P1** | 2 hr | Keeps clients up to date |
| 6 | Fuzzy search | **P1** | 2 hr | Better typo handling |
| 12 | Batch check_model_status | **P2** | 2-3 hr | Efficiency for bulk checks |
| 15 | Per-provider resource templates | **P2** | 2 hr | MCP-native data access |
| 13 | Capability field expansion | **P2** | 4-6 hr | Structural improvement |
| 14 | Webhook notifications | **P2** | 6-8 hr | Production notification system |

**Recommended implementation order:** Start with #5 (bug fix), then #4, #3, #1, #2 (all P0, completable in a single session), then tackle P1 items starting with #11 and #10 (quick fixes), followed by #9, #7, #8, #6.
