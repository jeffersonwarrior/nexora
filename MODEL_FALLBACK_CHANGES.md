# Model Fallback Integration Test

This file documents the expected behavior after implementing model fallback and global config changes.

## Changes Summary

1. **Global Config Only**: The system now only uses global config files, no more per-project configs
2. **Model Fallback**: When current models are invalid, system falls back to recent models (up to 5 previous models)
3. **TUI Setup Prompt**: If all attempts fail, shows TUI for model selection
4. **Config Repair**: Attempts to repair invalid configurations

## Expected Order of Operations

When running `nexora -y` or `nexora`:

1. **Try current models**: Validate the currently configured models
2. **Fallback to recent models**: If current models are invalid, try the recent models list (up to 5)
3. **Show TUI setup**: If all fallback options fail, present TUI for model selection
4. **Config repair**: If config format is detected as invalid, attempt repair

## File Changes Made

1. `internal/config/load.go`:
   - Modified `lookupConfigs()` to only use global config paths
   - Added fallback logic in `Load()` function
   - Added config repair attempt for format errors

2. `internal/config/config.go`:
   - Added `modelsNeedSetup` field to track when TUI is needed
   - Added `ValidateAndFallbackModels()` function for fallback logic
   - Added `validateModel()` function to check individual model validity
   - Added `tryRecentModelsFallback()` function for recent model fallback
   - Added `isValidRecentModel()` function to validate recent models
   - Added `RepairConfig()` function to clean up invalid configs
   - Added `ModelsNeedSetup()` getter method

3. `internal/tui/tui.go`:
   - Modified `Init()` to use `ModelsNeedSetup()` instead of `AreModelsConfigured()`

## Testing

Added unit tests in `internal/config/model_fallback_test.go` to verify:
- Model validation logic
- Recent model fallback
- Configuration repair
- ModelsNeedSetup flag behavior