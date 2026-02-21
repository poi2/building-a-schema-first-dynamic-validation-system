#!/bin/bash

# スクリプトの場所を基準にパスを特定
SCRIPT_DIR=$(dirname "$0")
HOOK_SOURCE="$SCRIPT_DIR/commit-msg"
HOOK_DEST=".git/hooks/commit-msg"

echo "Configuring git hooks from .github/git-hooks/..."

# .git/hooks ディレクトリがあるか確認（リポジトリルートで実行されているか）
if [ ! -d ".git" ]; then
    echo "Error: .git directory not found. Please run this from the repository root."
    exit 1
fi

# フックファイルをコピーして実行権限を付与
cp "$HOOK_SOURCE" "$HOOK_DEST"
chmod +x "$HOOK_DEST"

echo "Successfully installed commit-msg hook to $HOOK_DEST"

