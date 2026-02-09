---
description: Look up the correct API model ID for any AI model. Use when writing code that calls AI APIs, to avoid using outdated or incorrect model IDs.
---

When the user asks about an AI model or you need to write code that calls an AI provider's API, use the model-id-cheatsheet MCP server tools to look up the correct information:

1. If the user mentions a specific model, call `get_model_info` with the model ID to get exact API ID, pricing, context window, and capabilities.
2. If the user asks "what models are available" or wants to browse, call `list_models` with optional filters (provider, status, capability).
3. If the user needs a recommendation (e.g., "cheapest model for coding"), call `recommend_model` with the task and budget.
4. If you're about to write a model ID in code, call `check_model_status` first to verify it's not deprecated.
5. If the user wants to compare options, call `compare_models` with up to 5 model IDs.

Always use the verified API model ID from the tool response, never guess or use outdated IDs from your training data.
