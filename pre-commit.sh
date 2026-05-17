#!/bin/bash
# Pre-commit hook for Matching Service
# Copy to: .git/hooks/pre-commit
# Make executable: chmod +x .git/hooks/pre-commit

set -e

echo "🔍 Running pre-commit checks..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    exit 1
fi

# Format code
echo "📝 Formatting code..."
if ! go fmt ./...; then
    echo "❌ Code formatting failed"
    exit 1
fi

# Run linter if available
if command -v golangci-lint &> /dev/null; then
    echo "🔎 Running linter..."
    if ! golangci-lint run --issues-exit-code=1 ./...; then
        echo "❌ Linting failed"
        exit 1
    fi
fi

# Run tests
echo "🧪 Running tests..."
if ! go test -race -short ./...; then
    echo "❌ Tests failed"
    exit 1
fi

# Check for secrets
echo "🔐 Checking for secrets..."
if grep -r "password\|secret\|token\|key" --include="*.go" --exclude-dir=vendor . 2>/dev/null | grep -v "// password" | grep -v "PASSWORD\|SECRET"; then
    echo "⚠️  Possible secrets found in code (review before committing)"
    echo "Add '// password' or '// secret' comment to suppress warnings"
fi

echo "✅ Pre-commit checks passed!"
exit 0
