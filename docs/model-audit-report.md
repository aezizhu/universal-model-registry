# Model Data Audit Report

**Date:** February 10, 2026
**Scope:** Audit of `go-server/internal/models/data.go` for missing models, pricing accuracy, and status correctness.

---

## 1. Pricing Verification (Spot-Check)

| Model | data.go Input/Output | Verified Price | Status |
|-------|---------------------|---------------|--------|
| Claude Opus 4.6 | $5.00 / $25.00 | $5.00 / $25.00 | CORRECT |
| GPT-5.2 | $1.75 / $14.00 | $1.75 / $14.00 | CORRECT |
| Gemini 2.5 Pro | $1.25 / $10.00 | $1.25 / $10.00 | CORRECT |

All three spot-checked models have accurate pricing. No corrections needed.

**Sources:**
- [Anthropic Pricing](https://platform.claude.com/docs/en/about-claude/pricing)
- [OpenAI Pricing](https://platform.openai.com/docs/pricing)
- [Google Gemini Pricing](https://ai.google.dev/gemini-api/docs/pricing)

---

## 2. Status Verification

### Models Correctly Marked

| Model | Current Status | Assessment |
|-------|---------------|------------|
| gpt-4o | deprecated | CORRECT - Retiring from ChatGPT Feb 13, 2026 |
| gpt-4o-mini | deprecated | CORRECT - Superseded by GPT-4.1 Mini/Nano |
| gpt-4.1 | legacy | CORRECT - Retiring from ChatGPT Feb 13, 2026; still in API |
| claude-opus-4-5 | legacy | CORRECT - Superseded by Opus 4.6 |
| claude-3-7-sonnet-20250219 | deprecated | CORRECT |
| gemini-2.0-flash | deprecated | CORRECT - Retiring March 2026 |
| deepseek-v3 | deprecated | CORRECT - Merged into deepseek-chat V3.2 |

### Potential Status Changes to Consider

| Model | Current Status | Recommendation | Reason |
|-------|---------------|---------------|--------|
| o4-mini | current | current (no change) | Being retired from ChatGPT Feb 13, 2026, but still available in API |
| o3-mini | current | **legacy** | Superseded by o4-mini; may still be available but is predecessor |

**Recommendation:** Consider changing `o3-mini` from "current" to "legacy" since it's described as the "predecessor to o4-mini" in the existing notes and has been fully superseded.

**Source:** [OpenAI Model Retirements](https://openai.com/index/retiring-gpt-4o-and-older-models/)

---

## 3. Missing Providers

### 3A. Amazon Nova / Bedrock

Amazon Nova is a significant omission. It's a major cloud provider's model family available via Amazon Bedrock. Six models recommended for addition.

#### Nova 1.0 Models (Released December 2024)

```go
"amazon-nova-micro": {
    ID:              "amazon-nova-micro",
    DisplayName:     "Amazon Nova Micro",
    Provider:        "Amazon",
    ContextWindow:   128_000,
    MaxOutputTokens: 5_000,
    Vision:          false,
    Reasoning:       false,
    PricingInput:    0.035,
    PricingOutput:   0.14,
    KnowledgeCutoff: "2024-10",
    ReleaseDate:     "2024-12",
    Status:          "current",
    Notes:           "Text-only, lowest latency Nova model, via Amazon Bedrock",
},
"amazon-nova-lite": {
    ID:              "amazon-nova-lite",
    DisplayName:     "Amazon Nova Lite",
    Provider:        "Amazon",
    ContextWindow:   300_000,
    MaxOutputTokens: 5_000,
    Vision:          true,
    Reasoning:       false,
    PricingInput:    0.06,
    PricingOutput:   0.24,
    KnowledgeCutoff: "2024-10",
    ReleaseDate:     "2024-12",
    Status:          "current",
    Notes:           "Multimodal (text, image, video), fast and low-cost, via Amazon Bedrock",
},
"amazon-nova-pro": {
    ID:              "amazon-nova-pro",
    DisplayName:     "Amazon Nova Pro",
    Provider:        "Amazon",
    ContextWindow:   300_000,
    MaxOutputTokens: 5_000,
    Vision:          true,
    Reasoning:       false,
    PricingInput:    0.80,
    PricingOutput:   3.20,
    KnowledgeCutoff: "2024-10",
    ReleaseDate:     "2024-12",
    Status:          "current",
    Notes:           "Multimodal, best balance of accuracy/speed/cost, agentic workflows, via Amazon Bedrock",
},
"amazon-nova-premier": {
    ID:              "amazon-nova-premier",
    DisplayName:     "Amazon Nova Premier",
    Provider:        "Amazon",
    ContextWindow:   1_000_000,
    MaxOutputTokens: 5_000,
    Vision:          true,
    Reasoning:       false,
    PricingInput:    2.50,
    PricingOutput:   12.50,
    KnowledgeCutoff: "2024-10",
    ReleaseDate:     "2025-04",
    Status:          "current",
    Notes:           "Most capable Nova 1.0, 1M context, teacher for distillation, via Amazon Bedrock",
},
```

#### Nova 2 Models (Released December 2025)

```go
"amazon-nova-2-lite": {
    ID:              "amazon-nova-2-lite",
    DisplayName:     "Amazon Nova 2 Lite",
    Provider:        "Amazon",
    ContextWindow:   1_000_000,
    MaxOutputTokens: 65_536,
    Vision:          true,
    Reasoning:       true,
    PricingInput:    1.25,
    PricingOutput:   2.50,
    KnowledgeCutoff: "2025-09",
    ReleaseDate:     "2025-12",
    Status:          "current",
    Notes:           "Fast reasoning model, extended thinking with budget controls, 1M context, via Amazon Bedrock",
},
"amazon-nova-2-pro": {
    ID:              "amazon-nova-2-pro",
    DisplayName:     "Amazon Nova 2 Pro (Preview)",
    Provider:        "Amazon",
    ContextWindow:   1_000_000,
    MaxOutputTokens: 65_536,
    Vision:          true,
    Reasoning:       true,
    PricingInput:    0.30,
    PricingOutput:   10.00,
    KnowledgeCutoff: "2025-09",
    ReleaseDate:     "2025-12",
    Status:          "current",
    Notes:           "Most capable Nova 2, complex agentic tasks, 1M context, preview, via Amazon Bedrock",
},
```

**Sources:**
- [Amazon Nova Pricing](https://aws.amazon.com/nova/pricing/)
- [Amazon Nova Models](https://aws.amazon.com/nova/models/)
- [Amazon Nova 2 Lite Blog](https://aws.amazon.com/blogs/aws/introducing-amazon-nova-2-lite-a-fast-cost-effective-reasoning-model/)

---

### 3B. Cohere

Cohere is a significant enterprise AI provider missing entirely from the registry. The Command model family is widely used for RAG, tool use, and enterprise workflows.

```go
"command-a-03-2025": {
    ID:              "command-a-03-2025",
    DisplayName:     "Command A",
    Provider:        "Cohere",
    ContextWindow:   256_000,
    MaxOutputTokens: 8_000,
    Vision:          false,
    Reasoning:       false,
    PricingInput:    2.50,
    PricingOutput:   10.00,
    KnowledgeCutoff: "2025-01",
    ReleaseDate:     "2025-03",
    Status:          "current",
    Notes:           "Cohere flagship, 111B params, excels at RAG/tool use/agents, runs on 2 GPUs",
},
"command-a-reasoning-08-2025": {
    ID:              "command-a-reasoning-08-2025",
    DisplayName:     "Command A Reasoning",
    Provider:        "Cohere",
    ContextWindow:   256_000,
    MaxOutputTokens: 32_000,
    Vision:          false,
    Reasoning:       true,
    PricingInput:    2.50,
    PricingOutput:   10.00,
    KnowledgeCutoff: "2025-06",
    ReleaseDate:     "2025-08",
    Status:          "current",
    Notes:           "Reasoning variant of Command A, extended output, enterprise agentic workflows",
},
"command-a-vision-07-2025": {
    ID:              "command-a-vision-07-2025",
    DisplayName:     "Command A Vision",
    Provider:        "Cohere",
    ContextWindow:   128_000,
    MaxOutputTokens: 8_000,
    Vision:          true,
    Reasoning:       false,
    PricingInput:    2.50,
    PricingOutput:   10.00,
    KnowledgeCutoff: "2025-05",
    ReleaseDate:     "2025-07",
    Status:          "current",
    Notes:           "Multimodal Command A, 112B params, up to 20 images per request, open weights",
},
"command-r7b-12-2024": {
    ID:              "command-r7b-12-2024",
    DisplayName:     "Command R7B",
    Provider:        "Cohere",
    ContextWindow:   128_000,
    MaxOutputTokens: 4_096,
    Vision:          false,
    Reasoning:       false,
    PricingInput:    0.0375,
    PricingOutput:   0.15,
    KnowledgeCutoff: "2024-10",
    ReleaseDate:     "2024-12",
    Status:          "current",
    Notes:           "Smallest R-series, 7B params, fast tool use, 23 languages, runs on consumer GPUs",
},
```

**Sources:**
- [Cohere Models Overview](https://docs.cohere.com/docs/models)
- [Command A](https://docs.cohere.com/docs/command-a)
- [Command A Reasoning](https://docs.cohere.com/docs/command-a-reasoning)
- [Command A Vision](https://docs.cohere.com/docs/command-a-vision)
- [Cohere Pricing](https://www.metacto.com/blogs/cohere-pricing-explained-a-deep-dive-into-integration-development-costs)

---

### 3C. Perplexity

Perplexity's Sonar models are unique as search-augmented LLMs. They're widely used via API for grounded, citation-backed answers.

```go
"sonar": {
    ID:              "sonar",
    DisplayName:     "Sonar",
    Provider:        "Perplexity",
    ContextWindow:   127_000,
    MaxOutputTokens: 8_000,
    Vision:          false,
    Reasoning:       false,
    PricingInput:    1.00,
    PricingOutput:   1.00,
    KnowledgeCutoff: "2025-02",
    ReleaseDate:     "2025-02",
    Status:          "current",
    Notes:           "Search-augmented LLM, returns answers with citations, cost-effective",
},
"sonar-pro": {
    ID:              "sonar-pro",
    DisplayName:     "Sonar Pro",
    Provider:        "Perplexity",
    ContextWindow:   200_000,
    MaxOutputTokens: 8_000,
    Vision:          false,
    Reasoning:       false,
    PricingInput:    3.00,
    PricingOutput:   15.00,
    KnowledgeCutoff: "2025-02",
    ReleaseDate:     "2025-02",
    Status:          "current",
    Notes:           "Advanced search-augmented LLM, 2x citations vs Sonar, 200K context, multi-step queries",
},
"sonar-reasoning-pro": {
    ID:              "sonar-reasoning-pro",
    DisplayName:     "Sonar Reasoning Pro",
    Provider:        "Perplexity",
    ContextWindow:   128_000,
    MaxOutputTokens: 8_000,
    Vision:          false,
    Reasoning:       true,
    PricingInput:    2.00,
    PricingOutput:   8.00,
    KnowledgeCutoff: "2025-02",
    ReleaseDate:     "2025-03",
    Status:          "current",
    Notes:           "Reasoning model powered by DeepSeek R1 with CoT, search-augmented",
},
```

**Sources:**
- [Perplexity Pricing](https://docs.perplexity.ai/getting-started/pricing)
- [Sonar Pro Docs](https://docs.perplexity.ai/getting-started/models/models/sonar-pro)

---

### 3D. AI21

AI21's Jamba models feature a unique SSM-Transformer hybrid architecture and the longest context window among open models (256K).

```go
"jamba-large-1.7": {
    ID:              "jamba-large-1.7",
    DisplayName:     "Jamba Large 1.7",
    Provider:        "AI21",
    ContextWindow:   256_000,
    MaxOutputTokens: 4_096,
    Vision:          false,
    Reasoning:       false,
    PricingInput:    2.00,
    PricingOutput:   8.00,
    KnowledgeCutoff: "2025-06",
    ReleaseDate:     "2025-08",
    Status:          "current",
    Notes:           "SSM-Transformer hybrid, 256K context, enterprise-focused, available via AI21 API and Bedrock",
},
"jamba-mini-1.7": {
    ID:              "jamba-mini-1.7",
    DisplayName:     "Jamba Mini 1.7",
    Provider:        "AI21",
    ContextWindow:   256_000,
    MaxOutputTokens: 4_096,
    Vision:          false,
    Reasoning:       false,
    PricingInput:    0.20,
    PricingOutput:   0.40,
    KnowledgeCutoff: "2025-06",
    ReleaseDate:     "2025-07",
    Status:          "current",
    Notes:           "Compact SSM-Transformer hybrid, 12B active params, 256K context, cost-efficient",
},
```

**Sources:**
- [AI21 Jamba Docs](https://docs.ai21.com/docs/jamba-foundation-models)
- [AI21 Pricing](https://www.ai21.com/pricing/)
- [Jamba Large 1.7 on HuggingFace](https://huggingface.co/ai21labs/AI21-Jamba-Large-1.7)

---

## 4. Missing Recent Models from Existing Providers

### 4A. DeepSeek R1 (Missing)

The existing `deepseek-reasoner` entry covers DeepSeek V3.2's Thinking Mode, but DeepSeek R1 is a separate standalone reasoning model that's widely used.

```go
"deepseek-r1": {
    ID:              "deepseek-r1",
    DisplayName:     "DeepSeek R1",
    Provider:        "DeepSeek",
    ContextWindow:   128_000,
    MaxOutputTokens: 8_000,
    Vision:          false,
    Reasoning:       true,
    PricingInput:    0.55,
    PricingOutput:   2.19,
    KnowledgeCutoff: "2025-01",
    ReleaseDate:     "2025-01",
    Status:          "current",
    Notes:           "Open-source reasoning model, o1-level performance at fraction of cost, widely deployed",
},
```

**Source:** [DeepSeek Pricing](https://api-docs.deepseek.com/quick_start/pricing/)

### 4B. Mistral Medium 3 (Missing)

```go
"mistral-medium-2505": {
    ID:              "mistral-medium-2505",
    DisplayName:     "Mistral Medium 3",
    Provider:        "Mistral",
    ContextWindow:   131_072,
    MaxOutputTokens: 8_192,
    Vision:          true,
    Reasoning:       false,
    PricingInput:    0.40,
    PricingOutput:   2.00,
    KnowledgeCutoff: "2025-03",
    ReleaseDate:     "2025-05",
    Status:          "current",
    Notes:           "Mid-tier Mistral model, good vision support, strong multilingual",
},
```

**Source:** [Mistral Medium 3](https://mistral.ai/news/mistral-medium-3)

### 4C. Devstral Small 2 (Missing)

```go
"devstral-small-2512": {
    ID:              "devstral-small-2512",
    DisplayName:     "Devstral Small 2",
    Provider:        "Mistral",
    ContextWindow:   256_000,
    MaxOutputTokens: 8_192,
    Vision:          false,
    Reasoning:       false,
    PricingInput:    0.40,
    PricingOutput:   2.00,
    KnowledgeCutoff: "2025-11",
    ReleaseDate:     "2025-12",
    Status:          "current",
    Notes:           "24B coding model, runs on consumer GPUs, Apache 2.0, companion to Devstral 2",
},
```

**Source:** [Devstral Small 2 Analysis](https://artificialanalysis.ai/models/devstral-small-2)

---

## 5. Summary of Recommendations

### Priority 1: Add Missing Providers (20 models)

| Provider | Models to Add | Count |
|----------|--------------|-------|
| Amazon | Nova Micro, Lite, Pro, Premier, Nova 2 Lite, Nova 2 Pro | 6 |
| Cohere | Command A, Command A Reasoning, Command A Vision, Command R7B | 4 |
| Perplexity | Sonar, Sonar Pro, Sonar Reasoning Pro | 3 |
| AI21 | Jamba Large 1.7, Jamba Mini 1.7 | 2 |

### Priority 2: Add Missing Models from Existing Providers (3 models)

| Provider | Models to Add |
|----------|--------------|
| DeepSeek | DeepSeek R1 |
| Mistral | Mistral Medium 3, Devstral Small 2 |

### Priority 3: Status Updates (1 change)

| Model | Current | Recommended | Reason |
|-------|---------|-------------|--------|
| o3-mini | current | legacy | Superseded by o4-mini, described as predecessor in own notes |

### No Changes Needed

- All three spot-checked prices are correct
- All existing deprecated/legacy statuses are appropriate
- No models need to be removed

### Total Impact

- **18 new model entries** recommended for addition
- **1 status update** recommended
- **0 pricing corrections** needed
- New providers would bring total from 7 to 11 providers

---

## 6. Models NOT Recommended for Inclusion

The following were evaluated but not recommended:

- **Ministral 3 (3B/8B/14B):** Very small edge/device models, not typically used via API
- **Cohere Command R / Command R+:** Superseded by Command A family; legacy status would add clutter
- **Perplexity Sonar Deep Research:** Very specialized, more of a workflow than a model
- **Amazon Nova 2 Sonic / Nova 2 Omni:** Speech-to-speech and omnimodal models fall outside the text-focused registry scope
- **Mistral 3 (announced Feb 2026):** Family announcement only, not yet released as distinct API models
