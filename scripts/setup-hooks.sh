#!/bin/bash
# Setup Git hooks for NCloud Server Controller

echo "🔧 Setting up Git hooks..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Copy pre-commit hook
if [ -f ".githooks/pre-commit" ]; then
    cp .githooks/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    echo "✅ Pre-commit hook installed"
else
    echo "❌ Pre-commit hook not found"
    exit 1
fi

echo "🎉 Git hooks setup complete!"
echo ""
echo "Now pre-commit checks will run automatically before each commit."
echo "To bypass hooks (not recommended), use: git commit --no-verify"
