#!/bin/bash

# Gentleman MCP Gateway - Bun Setup Script
# This script installs Bun and sets up the React development environment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR/.."

echo "🏃‍♂️ Setting up Bun for Gentleman MCP Gateway"
echo ""

# Check if Bun is already installed
if command -v bun >/dev/null 2>&1; then
    echo "✅ Bun is already installed: $(bun --version)"
else
    echo "📦 Installing Bun..."

    # Install Bun
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
        # Windows
        powershell -c "irm bun.sh/install.ps1 | iex"
    else
        # macOS/Linux
        curl -fsSL https://bun.sh/install | bash
    fi

    # Add to PATH for current session
    export PATH="$HOME/.bun/bin:$PATH"

    echo "✅ Bun installed successfully: $(bun --version)"
fi

# Verify Bun installation
echo ""
echo "🔍 Verifying Bun installation..."
if ! command -v bun >/dev/null 2>&1; then
    echo "❌ Bun installation failed or not in PATH"
    echo "💡 Please restart your terminal or run: export PATH=\"\$HOME/.bun/bin:\$PATH\""
    exit 1
fi

echo "✅ Bun version: $(bun --version)"

# Navigate to React project
cd "$PROJECT_ROOT/web/react-example"

# Remove existing npm artifacts
echo ""
echo "🧹 Cleaning existing npm artifacts..."
rm -rf node_modules package-lock.json yarn.lock

# Install dependencies with Bun
echo ""
echo "📦 Installing React dependencies with Bun..."
bun install

# Verify installation
echo ""
echo "🔍 Verifying React setup..."
if [ ! -d "node_modules" ]; then
    echo "❌ Dependencies installation failed"
    exit 1
fi

echo "✅ Dependencies installed successfully"

# Check TypeScript setup
echo ""
echo "🔍 Checking TypeScript setup..."
if bun run type-check 2>/dev/null; then
    echo "✅ TypeScript configuration is valid"
else
    echo "⚠️  TypeScript configuration needs attention"
fi

# Display setup summary
echo ""
echo "🎉 Bun setup completed successfully!"
echo ""
echo "📊 Setup Summary:"
echo "  🏃‍♂️ Bun version: $(bun --version)"
echo "  ⚛️  React project: web/react-example"
echo "  📦 Dependencies: $(ls node_modules | wc -l | tr -d ' ') packages"
echo "  🔧 Configuration: bunfig.toml"
echo ""
echo "🚀 Next steps:"
echo "  1. Start the React dev server:"
echo "     cd web/react-example && bun dev"
echo ""
echo "  2. Or use the Makefile commands:"
echo "     make web-dev          # Start development server"
echo "     make web-build        # Build for production"
echo "     make web-test         # Run tests"
echo "     make web-preview      # Preview production build"
echo ""
echo "💡 Bun commands you can use:"
echo "  bun install            # Install dependencies"
echo "  bun dev                # Start dev server"
echo "  bun run build         # Build production"
echo "  bun test              # Run tests"
echo "  bun run lint          # Lint code"
echo "  bun run type-check    # Check TypeScript"
echo ""
echo "🌐 Development URLs:"
echo "  React App:     http://localhost:3000"
echo "  gRPC-Web:      http://localhost:8080"
echo "  gRPC Server:   localhost:50051"
echo ""
echo "✨ Bun is significantly faster than npm/yarn for:"
echo "  • Installing dependencies (2-10x faster)"
echo "  • Running scripts (faster startup)"
echo "  • Hot reloading (improved dev experience)"
