#!/bin/bash
set -euo pipefail

# Determine paths based on script location
SCRIPT_DIR=$(dirname "$0")
HOOK_SOURCE="$SCRIPT_DIR/commit-msg"
HOOK_DEST=".git/hooks/commit-msg"

echo "Configuring git hooks from .github/git-hooks/..."

# Check if .git/hooks directory exists (running from repository root)
if [ ! -d ".git" ]; then
    echo "Error: .git directory not found. Please run this from the repository root."
    exit 1
fi

# Check if source hook file exists
if [ ! -f "$HOOK_SOURCE" ]; then
    echo "Error: Hook source file not found at $HOOK_SOURCE"
    exit 1
fi

# Copy hook file and make it executable
if ! cp "$HOOK_SOURCE" "$HOOK_DEST"; then
    echo "Error: Failed to copy hook to $HOOK_DEST"
    exit 1
fi

if ! chmod +x "$HOOK_DEST"; then
    echo "Error: Failed to make hook executable"
    exit 1
fi

echo "Successfully installed commit-msg hook to $HOOK_DEST"

