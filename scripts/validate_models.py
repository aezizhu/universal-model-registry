"""Validate model registry data against live API endpoints.

Usage:
    uv run python scripts/validate_models.py

Set API keys via environment variables:
    OPENAI_API_KEY, ANTHROPIC_API_KEY, GOOGLE_API_KEY, XAI_API_KEY
"""

import asyncio
import os
import sys
from pathlib import Path

import httpx

# Add project root to path
sys.path.insert(0, str(Path(__file__).resolve().parent.parent))
from models_data import MODELS

TIMEOUT = 15.0
RESULTS: list[dict] = []


def report(model_id: str, provider: str, status: str, detail: str = "") -> None:
    icon = {"ok": "[OK]", "warn": "[WARN]", "error": "[ERR]", "skip": "[SKIP]"}[status]
    RESULTS.append({"id": model_id, "provider": provider, "status": status, "detail": detail})
    msg = f"  {icon} {model_id}"
    if detail:
        msg += f" — {detail}"
    print(msg)


async def validate_openai(client: httpx.AsyncClient) -> None:
    key = os.environ.get("OPENAI_API_KEY")
    if not key:
        print("\n--- OpenAI: OPENAI_API_KEY not set, skipping ---")
        for mid, m in MODELS.items():
            if m["provider"] == "OpenAI":
                report(mid, "OpenAI", "skip", "no API key")
        return

    print("\n--- OpenAI: checking model availability ---")
    try:
        resp = await client.get(
            "https://api.openai.com/v1/models",
            headers={"Authorization": f"Bearer {key}"},
            timeout=TIMEOUT,
        )
        resp.raise_for_status()
        available = {m["id"] for m in resp.json()["data"]}
    except Exception as e:
        print(f"  [ERR] Failed to fetch model list: {e}")
        for mid, m in MODELS.items():
            if m["provider"] == "OpenAI":
                report(mid, "OpenAI", "error", f"API error: {e}")
        return

    for mid, m in MODELS.items():
        if m["provider"] != "OpenAI":
            continue
        if mid in available:
            report(mid, "OpenAI", "ok", "model available")
        else:
            report(mid, "OpenAI", "warn", "model ID not in /v1/models list")


async def validate_anthropic(client: httpx.AsyncClient) -> None:
    key = os.environ.get("ANTHROPIC_API_KEY")
    if not key:
        print("\n--- Anthropic: ANTHROPIC_API_KEY not set, skipping ---")
        for mid, m in MODELS.items():
            if m["provider"] == "Anthropic":
                report(mid, "Anthropic", "skip", "no API key")
        return

    print("\n--- Anthropic: checking model availability ---")
    for mid, m in MODELS.items():
        if m["provider"] != "Anthropic":
            continue
        try:
            resp = await client.post(
                "https://api.anthropic.com/v1/messages",
                headers={
                    "x-api-key": key,
                    "anthropic-version": "2023-06-01",
                    "content-type": "application/json",
                },
                json={
                    "model": mid,
                    "max_tokens": 1,
                    "messages": [{"role": "user", "content": "hi"}],
                },
                timeout=TIMEOUT,
            )
            if resp.status_code == 200:
                report(mid, "Anthropic", "ok", "model responds")
            elif resp.status_code == 400 and "model" in resp.text.lower():
                report(mid, "Anthropic", "error", "invalid model ID")
            elif resp.status_code == 401:
                report(mid, "Anthropic", "skip", "auth error — check API key")
                break
            else:
                report(mid, "Anthropic", "warn", f"HTTP {resp.status_code}")
        except Exception as e:
            report(mid, "Anthropic", "error", str(e))


async def validate_google(client: httpx.AsyncClient) -> None:
    key = os.environ.get("GOOGLE_API_KEY")
    if not key:
        print("\n--- Google: GOOGLE_API_KEY not set, skipping ---")
        for mid, m in MODELS.items():
            if m["provider"] == "Google":
                report(mid, "Google", "skip", "no API key")
        return

    print("\n--- Google: checking model availability ---")
    try:
        resp = await client.get(
            f"https://generativelanguage.googleapis.com/v1beta/models?key={key}",
            timeout=TIMEOUT,
        )
        resp.raise_for_status()
        available = {m["name"].split("/")[-1] for m in resp.json().get("models", [])}
    except Exception as e:
        print(f"  [ERR] Failed to fetch model list: {e}")
        for mid, m in MODELS.items():
            if m["provider"] == "Google":
                report(mid, "Google", "error", f"API error: {e}")
        return

    for mid, m in MODELS.items():
        if m["provider"] != "Google":
            continue
        if mid in available:
            report(mid, "Google", "ok", "model available")
        else:
            report(mid, "Google", "warn", "model ID not in API list")


async def validate_xai(client: httpx.AsyncClient) -> None:
    key = os.environ.get("XAI_API_KEY")
    if not key:
        print("\n--- xAI: XAI_API_KEY not set, skipping ---")
        for mid, m in MODELS.items():
            if m["provider"] == "xAI":
                report(mid, "xAI", "skip", "no API key")
        return

    print("\n--- xAI: checking model availability ---")
    try:
        resp = await client.get(
            "https://api.x.ai/v1/models",
            headers={"Authorization": f"Bearer {key}"},
            timeout=TIMEOUT,
        )
        resp.raise_for_status()
        available = {m["id"] for m in resp.json()["data"]}
    except Exception as e:
        print(f"  [ERR] Failed to fetch model list: {e}")
        for mid, m in MODELS.items():
            if m["provider"] == "xAI":
                report(mid, "xAI", "error", f"API error: {e}")
        return

    for mid, m in MODELS.items():
        if m["provider"] != "xAI":
            continue
        if mid in available:
            report(mid, "xAI", "ok", "model available")
        else:
            report(mid, "xAI", "warn", "model ID not in /v1/models list")


async def validate_no_api_providers() -> None:
    """Skip providers without direct API validation (Meta, Mistral, DeepSeek)."""
    for provider in ("Meta", "Mistral", "DeepSeek"):
        models = [(mid, m) for mid, m in MODELS.items() if m["provider"] == provider]
        if models:
            print(f"\n--- {provider}: no live validation (manual check needed) ---")
            for mid, _m in models:
                report(mid, provider, "skip", "no automated validation available")


async def main() -> None:
    print("=" * 60)
    print("Universal Model Registry — Validation Report")
    print("=" * 60)
    print(f"Total models: {len(MODELS)}")
    providers = sorted({m["provider"] for m in MODELS.values()})
    print(f"Providers: {', '.join(providers)}")

    async with httpx.AsyncClient() as client:
        await validate_openai(client)
        await validate_anthropic(client)
        await validate_google(client)
        await validate_xai(client)
        await validate_no_api_providers()

    # Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    ok = sum(1 for r in RESULTS if r["status"] == "ok")
    warn = sum(1 for r in RESULTS if r["status"] == "warn")
    err = sum(1 for r in RESULTS if r["status"] == "error")
    skip = sum(1 for r in RESULTS if r["status"] == "skip")
    print(f"  OK: {ok}  |  WARN: {warn}  |  ERROR: {err}  |  SKIP: {skip}")

    if err > 0:
        print("\nERRORS (model IDs likely invalid):")
        for r in RESULTS:
            if r["status"] == "error":
                print(f"  - {r['id']} ({r['provider']}): {r['detail']}")

    if warn > 0:
        print("\nWARNINGS (model IDs not found in API, may need update):")
        for r in RESULTS:
            if r["status"] == "warn":
                print(f"  - {r['id']} ({r['provider']}): {r['detail']}")

    # Exit with error code if any errors found
    if err > 0:
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
