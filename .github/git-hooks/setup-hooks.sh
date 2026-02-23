#!/bin/bash
set -euo pipefail

# Determine paths based on script location
SCRIPT_DIR=$(dirname "$0")
HOOKS_DIR=$(git rev-parse --git-path hooks)

echo "Configuring git hooks from .github/git-hooks/..."

# Check if we're in a git repository
if ! git rev-parse --git-dir >/dev/null 2>&1; then
    echo "Error: Not in a git repository. Please run this from the repository root."
    exit 1
fi

# Function to install a hook
install_hook() {
    local hook_name=$1
    local hook_source="$SCRIPT_DIR/$hook_name"
    local hook_dest="$HOOKS_DIR/$hook_name"

    # Check if source hook file exists
    if [ ! -f "$hook_source" ]; then
        echo "Warning: Hook source file not found at $hook_source, skipping..."
        return 0
    fi

    # Copy hook file and make it executable
    if ! cp "$hook_source" "$hook_dest"; then
        echo "Error: Failed to copy $hook_name to $hook_dest"
        return 1
    fi

    if ! chmod +x "$hook_dest"; then
        echo "Error: Failed to make $hook_name executable"
        return 1
    fi

    echo "Successfully installed $hook_name hook to $hook_dest"
    return 0
}

# Install hooks
install_hook "commit-msg"
install_hook "pre-push"

