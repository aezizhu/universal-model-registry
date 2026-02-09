"""Shared pytest fixtures for the Universal Model Registry test suite."""

import sys
from pathlib import Path

import pytest

# Ensure project root is on sys.path for imports
sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from models_data import MODELS
from registry import (
    check_model_status as _check_model_status,
)
from registry import (
    compare_models as _compare_models,
)
from registry import (
    get_model_info as _get_model_info,
)
from registry import (
    list_models as _list_models,
)
from registry import (
    recommend_model as _recommend_model,
)


@pytest.fixture
def all_models():
    """Return the full MODELS dict."""
    return MODELS


@pytest.fixture
def current_models():
    """Return only current (non-legacy, non-deprecated) models."""
    return {k: v for k, v in MODELS.items() if v["status"] == "current"}


@pytest.fixture
def tool_list_models():
    """Unwrapped list_models tool function."""
    return _list_models.fn


@pytest.fixture
def tool_get_model_info():
    """Unwrapped get_model_info tool function."""
    return _get_model_info.fn


@pytest.fixture
def tool_recommend_model():
    """Unwrapped recommend_model tool function."""
    return _recommend_model.fn


@pytest.fixture
def tool_check_model_status():
    """Unwrapped check_model_status tool function."""
    return _check_model_status.fn


@pytest.fixture
def tool_compare_models():
    """Unwrapped compare_models tool function."""
    return _compare_models.fn


@pytest.fixture
def sample_model_ids():
    """A curated set of model IDs for cross-provider testing."""
    return ["gpt-5", "claude-opus-4-6", "gemini-2.5-pro", "grok-3", "deepseek-reasoner"]
