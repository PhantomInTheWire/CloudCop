#!/bin/bash
set -e

echo "🚀 CloudCop Developer Environment Setup"
echo "========================================"
echo ""

# Check if running in CloudCop directory
if [ ! -f "README.md" ] || [ ! -d ".github" ]; then
    echo "❌ Error: Please run this script from the CloudCop root directory"
    exit 1
fi

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print status
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

echo "📋 Checking required tools..."
echo ""

# Check for required tools
MISSING_TOOLS=()

if ! command_exists git; then
    MISSING_TOOLS+=("git")
fi

if ! command_exists go; then
    MISSING_TOOLS+=("go (version 1.23+)")
fi

if ! command_exists node; then
    MISSING_TOOLS+=("node (version 20+)")
fi

if ! command_exists npm; then
    MISSING_TOOLS+=("npm")
fi

if ! command_exists python3; then
    MISSING_TOOLS+=("python3")
fi

if ! command_exists docker; then
    MISSING_TOOLS+=("docker")
fi

if [ ${#MISSING_TOOLS[@]} -ne 0 ]; then
    print_error "Missing required tools:"
    for tool in "${MISSING_TOOLS[@]}"; do
        echo "  - $tool"
    done
    echo ""
    echo "Please install missing tools and try again."
    exit 1
fi

print_status "All required tools are installed"
echo ""

# Install pre-commit
echo "📦 Installing pre-commit framework..."
if command_exists pip3; then
    pip3 install --user pre-commit
    print_status "pre-commit installed"
elif command_exists pipx; then
    pipx install pre-commit
    print_status "pre-commit installed via pipx"
else
    print_error "Neither pip3 nor pipx found. Please install pre-commit manually:"
    echo "  pip install pre-commit"
    exit 1
fi
echo ""

# Install Go tools
echo "📦 Installing Go tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
print_status "Go tools installed (golangci-lint, gosec)"
echo ""

# Install Python tools (via uv)
echo "📦 Installing Python tools..."
if ! command_exists uv; then
    print_warning "uv not found, installing..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
fi

cd backend/ai
uv venv
uv pip install ruff mypy bandit pytest
print_status "Python tools installed (ruff, mypy, bandit)"
cd ../..
echo ""

# Install Node.js commitlint
echo "📦 Installing commitlint..."
npm install -g @commitlint/cli @commitlint/config-conventional
print_status "commitlint installed"
echo ""

# Install pre-commit hooks
echo "🔧 Installing pre-commit hooks..."
pre-commit install
pre-commit install --hook-type commit-msg
print_status "Pre-commit hooks installed"
echo ""

# Install frontend dependencies
echo "📦 Installing frontend dependencies..."
cd frontend
npm ci
print_status "Frontend dependencies installed"
cd ..
echo ""

echo "✅ Setup complete!"
echo ""
echo "📝 Next steps:"
echo "  1. Make changes to the code"
echo "  2. Commit with conventional commit format: type(scope): description"
echo "     Examples:"
echo "       - feat(backend): add new API endpoint"
echo "       - fix(frontend): resolve login issue"
echo "       - chore(ci): update pre-commit hooks"
echo "  3. Pre-commit hooks will run automatically on commit"
echo ""
echo "🔍 Useful commands:"
echo "  - Run all hooks manually: pre-commit run --all-files"
echo "  - Update hooks: pre-commit autoupdate"
echo "  - Skip hooks (not recommended): git commit --no-verify"
echo ""
