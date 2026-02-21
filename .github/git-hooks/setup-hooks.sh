#!/bin/bash

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

# Copy hook file and make it executable
cp "$HOOK_SOURCE" "$HOOK_DEST"
chmod +x "$HOOK_DEST"

echo "Successfully installed commit-msg hook to $HOOK_DEST"

