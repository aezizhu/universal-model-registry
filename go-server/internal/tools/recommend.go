package tools

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-server/internal/models"
)

// RecommendModelInput holds parameters for the recommend_model tool.
type RecommendModelInput struct {
	Task   string `json:"task" jsonschema:"Description of the task you need a model for"`
	Budget string `json:"budget,omitempty" jsonschema:"Budget level: cheap, moderate, or expensive"`
}

// RecommendModel scores current models against a task description and budget,
// returning the top 3 recommendations as a markdown list.
func RecommendModel(task, budget string) string {
	if budget == "" {
		budget = "moderate"
	}
	budget = strings.ToLower(budget)
	taskLower := strings.ToLower(task)

	// Collect current models
	var current []models.Model
	for _, m := range models.Models {
		if m.Status == "current" {
			current = append(current, m)
		}
	}

	type scored struct {
		score float64
		model models.Model
	}

	var results []scored
	for _, m := range current {
		score := 0.0

		// ── Task relevance signals ──

		// Coding
		if strings.Contains(taskLower, "coding") ||
			strings.Contains(taskLower, "code") ||
			strings.Contains(taskLower, "programming") {
			if m.Reasoning {
				score += 3
			}
			if m.ContextWindow >= 200_000 {
				score += 1
			}
			if strings.Contains(m.ID, "codestral") || strings.Contains(m.ID, "devstral") {
				score += 2
			}
		}

		// Vision
		if strings.Contains(taskLower, "vision") ||
			strings.Contains(taskLower, "image") ||
			strings.Contains(taskLower, "screenshot") {
			if m.Vision {
				score += 4
			} else {
				score -= 10
			}
		}

		// Reasoning
		if (strings.Contains(taskLower, "reason") ||
			strings.Contains(taskLower, "think") ||
			strings.Contains(taskLower, "math") ||
			strings.Contains(taskLower, "logic")) && m.Reasoning {
			score += 5
		}

		// Long context
		if strings.Contains(taskLower, "long context") ||
			strings.Contains(taskLower, "large document") ||
			strings.Contains(taskLower, "summariz") {
			if m.ContextWindow >= 1_000_000 {
				score += 4
			} else if m.ContextWindow >= 200_000 {
				score += 2
			}
		}

		// Cost-sensitive
		if strings.Contains(taskLower, "cheap") ||
			strings.Contains(taskLower, "batch") ||
			strings.Contains(taskLower, "cost") {
			score += math.Max(0, 5-m.PricingInput)
		}

		// Multilingual
		if strings.Contains(taskLower, "multilingual") ||
			strings.Contains(taskLower, "translat") {
			if m.Provider == "Mistral" {
				score += 2
			}
			if m.ContextWindow >= 128_000 {
				score += 1
			}
		}

		// Open-weight / open-source
		if strings.Contains(taskLower, "open") &&
			(strings.Contains(taskLower, "weight") || strings.Contains(taskLower, "source")) &&
			(m.Provider == "Meta" || m.Provider == "DeepSeek" || m.Provider == "Mistral") {
			score += 3
		}

		// ── Budget modifier ──
		switch budget {
		case "cheap":
			score += math.Max(0, 3-m.PricingInput)
			if m.PricingInput > 5 {
				score -= 5
			}
		case "unlimited", "expensive":
			score += math.Min(m.PricingInput, 5)
		}

		// General quality signal
		score += math.Min(m.PricingInput*0.3, 2)

		// Recency bonus: newer models get a boost (0 to 1.5 points)
		score += recencyBonus(m.ReleaseDate)

		results = append(results, scored{score: score, model: m})
	}

	// Sort descending by score; tie-break by newest release date, then display name
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		if results[i].model.ReleaseDate != results[j].model.ReleaseDate {
			return results[i].model.ReleaseDate > results[j].model.ReleaseDate
		}
		return results[i].model.DisplayName < results[j].model.DisplayName
	})

	top := results
	if len(top) > 3 {
		top = top[:3]
	}

	lines := []string{
		fmt.Sprintf("## Recommendations for: *%s*", task),
		fmt.Sprintf("**Budget:** %s", budget),
		"",
	}
	for i, s := range top {
		var caps []string
		if s.model.Vision {
			caps = append(caps, "vision")
		}
		if s.model.Reasoning {
			caps = append(caps, "reasoning")
		}
		capStr := "standard"
		if len(caps) > 0 {
			capStr = strings.Join(caps, ", ")
		}
		lines = append(lines, fmt.Sprintf(
			"%d. **%s** (`%s`)\n   - Provider: %s | Capabilities: %s\n   - Pricing: $%.2f / $%.2f per 1M tokens\n   - Context: %s tokens\n",
			i+1, s.model.DisplayName, s.model.ID,
			s.model.Provider, capStr,
			s.model.PricingInput, s.model.PricingOutput,
			models.FormatInt(s.model.ContextWindow),
		))
	}

	return strings.Join(lines, "\n")
}

// recencyBonus returns a score bonus (0 to 1.5) based on how recent the model
// release date is. Dates use "YYYY-MM" format. Models released in the last 6
// months get full bonus, decaying to 0 at 18 months.
func recencyBonus(releaseDate string) float64 {
	parts := strings.Split(releaseDate, "-")
	if len(parts) < 2 {
		return 0
	}
	year, err1 := strconv.Atoi(parts[0])
	month, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0
	}

	now := time.Now()
	releaseMonths := year*12 + month
	currentMonths := now.Year()*12 + int(now.Month())
	monthsAgo := float64(currentMonths - releaseMonths)

	bonus := 1.5 * (1.0 - monthsAgo/18.0)
	if bonus < 0 {
		bonus = 0
	}
	if bonus > 1.5 {
		bonus = 1.5
	}
	return bonus
}
