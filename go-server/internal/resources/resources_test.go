package resources

import (
	"encoding/json"
	"strings"
	"testing"

	"go-server/internal/models"
)

func TestAllModels_ReturnsValidJSON(t *testing.T) {
	result := AllModels()
	var parsed map[string]models.Model
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("AllModels() returned invalid JSON: %v", err)
	}
	if len(parsed) != len(models.Models) {
		t.Errorf("expected %d models in JSON, got %d", len(models.Models), len(parsed))
	}
}

func TestAllModels_ContainsAllModelIDs(t *testing.T) {
	result := AllModels()
	for id := range models.Models {
		if !strings.Contains(result, id) {
			t.Errorf("AllModels() missing model ID %q", id)
		}
	}
}

func TestCurrentModels_ReturnsValidJSON(t *testing.T) {
	result := CurrentModels()
	var parsed map[string]models.Model
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("CurrentModels() returned invalid JSON: %v", err)
	}
	for id, m := range parsed {
		if m.Status != "current" {
			t.Errorf("CurrentModels() contains non-current model %s with status %q", id, m.Status)
		}
	}
}

func TestCurrentModels_ExcludesLegacyAndDeprecated(t *testing.T) {
	result := CurrentModels()
	var parsed map[string]models.Model
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("CurrentModels() returned invalid JSON: %v", err)
	}
	for id := range models.Models {
		m := models.Models[id]
		if m.Status == "current" {
			if _, ok := parsed[id]; !ok {
				t.Errorf("CurrentModels() missing current model %s", id)
			}
		} else {
			if _, ok := parsed[id]; ok {
				t.Errorf("CurrentModels() should not contain %s model %s", m.Status, id)
			}
		}
	}
}

func TestPricingSummary_ReturnsMarkdownTable(t *testing.T) {
	result := PricingSummary()
	if !strings.Contains(result, "Model ID") {
		t.Error("expected 'Model ID' header in pricing summary")
	}
	if !strings.Contains(result, "Provider") {
		t.Error("expected 'Provider' header in pricing summary")
	}
	if !strings.Contains(result, "|") {
		t.Error("expected markdown table format")
	}
}

func TestPricingSummary_OnlyCurrentModels(t *testing.T) {
	result := PricingSummary()
	for id, m := range models.Models {
		if m.Status != "current" {
			if strings.Contains(result, "| "+id+" |") {
				t.Errorf("PricingSummary() should not contain %s model %s", m.Status, id)
			}
		}
	}
}

func TestPricingSummary_SortedByInputPrice(t *testing.T) {
	result := PricingSummary()
	lines := strings.Split(result, "\n")
	var prices []float64
	for _, line := range lines[2:] { // skip header and separator
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Extract price from "| id | provider | $X.XX | $Y.YY | ctx |"
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		priceStr := strings.TrimSpace(parts[3])
		priceStr = strings.TrimPrefix(priceStr, "$")
		var price float64
		if _, err := json.Number(priceStr).Float64(); err == nil {
			price, _ = json.Number(priceStr).Float64()
			prices = append(prices, price)
		}
	}
	for i := 1; i < len(prices); i++ {
		if prices[i] < prices[i-1] {
			t.Errorf("pricing not sorted: $%.2f comes after $%.2f", prices[i], prices[i-1])
		}
	}
}

func TestFormatInt_Resources(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{200000, "200,000"},
	}
	for _, tc := range tests {
		got := formatInt(tc.input)
		if got != tc.want {
			t.Errorf("formatInt(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
