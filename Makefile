.PHONY: help dev build proto clean test certs install-deps docker server client check lint web-dev web-build react-dev react-install

# Default target
help: ## Show this help message
	@echo "🚀 Gentleman MCP Gateway - Development Commands"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev: proto certs ## Start development server with TLS
	@echo "🚀 Starting Gentleman MCP Gateway (development mode)..."
	go run cmd/server/main.go

dev-insecure: proto ## Start development server without TLS
	@echo "⚠️  Starting Gentleman MCP Gateway (INSECURE mode)..."
	go run cmd/server/main.go -insecure

dev-web: proto ## Start development server with gRPC-Web enabled
	@echo "🌐 Starting Gentleman MCP Gateway with gRPC-Web..."
	go run cmd/server/main.go -insecure -enable-web

dev-web-secure: proto ## Start development server with TLS and gRPC-Web
	@echo "🌐 Starting Gentleman MCP Gateway with TLS + gRPC-Web..."
	go run cmd/server/main.go -enable-web

# Build
build: proto ## Build production binary
	@echo "🔨 Building Gentleman MCP Gateway..."
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/gentleman-mcp cmd/server/main.go
	@echo "✅ Binary created: bin/gentleman-mcp"

build-client: proto ## Build client binary
	@echo "🔨 Building client..."
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/mcp-client cmd/client/main.go
	@echo "✅ Client binary created: bin/mcp-client"

build-stream-client: proto ## Build streaming client binary
	@echo "🔨 Building streaming client..."
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/mcp-stream-client cmd/stream-client/main.go
	@echo "✅ Streaming client binary created: bin/mcp-stream-client"

# Protocol Buffers
proto: ## Generate Go code from protobuf definitions
	@echo "📦 Generating protobuf code..."
	buf generate
	@echo "✅ Protobuf code generated"

proto-lint: ## Lint protobuf files
	@echo "🔍 Linting protobuf files..."
	buf lint

# Web Development (using Bun)
web-install: ## Install React app dependencies with Bun
	@echo "🏃‍♂️ Installing React app dependencies with Bun..."
	cd web/react-example && bun install

web-dev: web-install ## Start React development server with Bun
	@echo "🌐 Starting React development server with Bun..."
	cd web/react-example && bun dev

web-build: web-install ## Build React app for production with Bun
	@echo "🔨 Building React app for production with Bun..."
	cd web/react-example && bun run build

web-test: web-install ## Run React app tests with Bun
	@echo "🧪 Running React app tests with Bun..."
	cd web/react-example && bun test --watch=false

web-clean: ## Clean React app build artifacts and dependencies
	@echo "🧹 Cleaning React app..."
	cd web/react-example && bun run clean

web-preview: web-build ## Preview production build locally
	@echo "👀 Starting preview server..."
	cd web/react-example && bun run preview

web-lint: web-install ## Lint React app code
	@echo "🔍 Linting React app..."
	cd web/react-example && bun run lint

web-type-check: web-install ## Type check React app
	@echo "🔍 Type checking React app..."
	cd web/react-example && bun run type-check

# Certificates
certs: ## Generate development TLS certificates
	@echo "🔐 Generating development certificates..."
	./scripts/gen-certs.sh

certs-clean: ## Remove all certificates
	@echo "🧹 Cleaning certificates..."
	rm -rf certs/*.pem certs/*.srl

# Dependencies
install-deps: ## Install development dependencies
	@echo "📦 Installing dependencies..."
	go mod tidy
	@command -v buf >/dev/null 2>&1 || (echo "Installing buf..." && go install github.com/bufbuild/buf/cmd/buf@latest)
	@command -v grpcurl >/dev/null 2>&1 || (echo "⚠️  Install grpcurl for testing: brew install grpcurl")
	@echo "✅ Dependencies installed"

install-bun: ## Install Bun package manager
	@echo "🏃‍♂️ Installing Bun..."
	./scripts/setup-bun.sh

# Testing
test: ## Run unit tests
	@echo "🧪 Running tests..."
	go test -v ./...

test-integration: dev-insecure & ## Run integration tests (requires server)
	@echo "🧪 Running integration tests..."
	sleep 2  # Wait for server to start
	@echo "Testing handshake..."
	grpcurl -plaintext -d '{"tenant_id":"test","agent_id":"integration","model":"gemma3:4b"}' localhost:50051 mcp.v1.HandshakeService/Register
	@echo "✅ Integration tests completed"
	@pkill -f "cmd/server/main.go" || true

smoke-test: ## Quick smoke test with grpcurl
	@echo "💨 Running smoke test..."
	@echo "1. Testing HandshakeService/Register..."
	grpcurl -plaintext -d '{"tenant_id":"demo","agent_id":"smoke-test","model":"gemma3:4b"}' localhost:50051 mcp.v1.HandshakeService/Register
	@echo "2. Testing reflection..."
	grpcurl -plaintext localhost:50051 list
	@echo "✅ Smoke test completed"

# Client testing
client: build-client ## Run example client
	@echo "🤖 Running example client..."
	./bin/mcp-client

client-insecure: build-client ## Run example client without TLS
	@echo "🤖 Running example client (insecure)..."
	./bin/mcp-client -insecure

stream-client: build-stream-client ## Run streaming chat client
	@echo "🔄 Running streaming chat client..."
	./bin/mcp-stream-client

stream-client-insecure: build-stream-client ## Run streaming chat client without TLS
	@echo "🔄 Running streaming chat client (insecure)..."
	./bin/mcp-stream-client -insecure

# Development utilities
server: dev ## Alias for dev

logs: ## Show server logs (if running as daemon)
	@tail -f /tmp/gentleman-mcp.log 2>/dev/null || echo "No log file found. Run 'make dev' to start server."

ps: ## Show running processes
	@echo "🔍 Gentleman MCP processes:"
	@ps aux | grep -E "(gentleman-mcp|cmd/server)" | grep -v grep || echo "No processes found"

# Code quality
lint: ## Run linter
	@echo "🔍 Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

check: proto-lint lint test ## Run all checks (lint, test, proto-lint)

# Docker
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t gentleman-mcp:latest .

docker-run: docker-build ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 50051:50051 gentleman-mcp:latest

# Cleanup
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf /tmp/gentleman-mcp.log
	go clean

clean-all: clean certs-clean ## Clean everything including certificates
	@echo "🧹 Cleaning everything..."

# Ollama helpers
ollama-setup: ## Setup Ollama with Gemma 3
	@echo "🦙 Setting up Ollama with Gemma 3..."
	@command -v ollama >/dev/null 2>&1 || (echo "❌ Please install Ollama first: https://ollama.com/install" && exit 1)
	ollama pull gemma3:4b
	@echo "✅ Gemma 3 4B model ready"

ollama-start: ## Start Ollama service
	@echo "🦙 Starting Ollama service..."
	ollama serve &
	@echo "✅ Ollama service started"

ollama-test: ## Test Ollama connection
	@echo "🦙 Testing Ollama connection..."
	curl -X POST http://localhost:11434/api/generate \
		-H "Content-Type: application/json" \
		-d '{"model":"gemma3:4b","prompt":"Hello","stream":false}' \
		| jq .response 2>/dev/null || echo "❌ Ollama not responding. Run 'make ollama-start'"

# Full development setup
setup: install-deps certs ollama-setup install-bun ## Complete development setup
	@echo "🎉 Gentleman MCP Gateway setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Start Ollama: make ollama-start"
	@echo "  2. Start server: make dev-web"
	@echo "  3. Start React app: make web-dev (in another terminal)"
	@echo "  4. Open browser: http://localhost:3000"
	@echo "  5. Test streaming: make stream-client (for CLI testing)"

setup-web: install-bun web-install ## Quick web development setup
	@echo "🌐 Web development environment ready!"
	@echo ""
	@echo "🚀 Quick start:"
	@echo "  Terminal 1: make dev-web"
	@echo "  Terminal 2: make web-dev"
	@echo "  Browser:    http://localhost:3000"

# Sprint 2 Demo
sprint2-demo: build build-stream-client ## Demo Sprint 2 streaming capabilities
	@echo "🚀 Sprint 2 Demo: Bidirectional Streaming Chat"
	@echo "1. Start the server with: make dev-insecure"
	@echo "2. In another terminal, run: make stream-client-insecure"
	@echo "3. Type messages to chat with Gemma 3 in real-time!"

# Sprint 3 Demo
sprint3-demo: web-install ## Demo Sprint 3 gRPC-Web + React capabilities
	@echo "🌐 Sprint 3 Demo: gRPC-Web + React Integration"
	@echo ""
	@echo "🚀 Complete Setup (3 terminals):"
	@echo "  Terminal 1: make dev-web"
	@echo "  Terminal 2: make web-dev"
	@echo "  Terminal 3: Open http://localhost:3000"
	@echo ""
	@echo "💡 Quick Test:"
	@echo "  Terminal 1: make dev-web"
	@echo "  Terminal 2: Open http://localhost:8080 (built-in web client)"

# Show status
status: ## Show development environment status
	@echo "📊 Gentleman MCP Gateway Status"
	@echo ""
	@echo "Dependencies:"
	@command -v go >/dev/null 2>&1 && echo "  ✅ Go $(shell go version | cut -d' ' -f3)" || echo "  ❌ Go not installed"
	@command -v buf >/dev/null 2>&1 && echo "  ✅ Buf $(shell buf --version)" || echo "  ❌ Buf not installed"
	@command -v grpcurl >/dev/null 2>&1 && echo "  ✅ grpcurl available" || echo "  ⚠️  grpcurl not installed (optional)"
	@command -v ollama >/dev/null 2>&1 && echo "  ✅ Ollama available" || echo "  ❌ Ollama not installed"
	@echo ""
	@echo "Certificates:"
	@test -f certs/server-cert.pem && echo "  ✅ Server certificate exists" || echo "  ❌ Server certificate missing (run 'make certs')"
	@test -f certs/ca-cert.pem && echo "  ✅ CA certificate exists" || echo "  ❌ CA certificate missing (run 'make certs')"
	@echo ""
	@echo "Build artifacts:"
	@test -f bin/gentleman-mcp && echo "  ✅ Server binary exists" || echo "  ❌ Server binary missing (run 'make build')"
	@test -f bin/mcp-client && echo "  ✅ Client binary exists" || echo "  ❌ Client binary missing (run 'make build-client')"
	@test -f bin/mcp-stream-client && echo "  ✅ Streaming client binary exists" || echo "  ❌ Streaming client binary missing (run 'make build-stream-client')"
	@echo ""
	@echo "Web Development:"
	@command -v bun >/dev/null 2>&1 && echo "  ✅ Bun $(shell bun --version 2>/dev/null || echo 'not found')" || echo "  ❌ Bun not installed (run 'make install-bun')"
	@test -d web/react-example/node_modules && echo "  ✅ React dependencies installed" || echo "  ❌ React dependencies missing (run 'make web-install')"
	@test -f web/generated/mcp/v1/McpServiceClientPb.ts && echo "  ✅ TypeScript gRPC code generated" || echo "  ❌ TypeScript code missing (run 'make proto')"
	@test -f web/react-example/bunfig.toml && echo "  ✅ Bun configuration exists" || echo "  ❌ Bun config missing"
	@echo ""
	@echo "Services:"
	@curl -s http://localhost:11434/api/version >/dev/null 2>&1 && echo "  ✅ Ollama service running" || echo "  ❌ Ollama service not running"
	@nc -z localhost 50051 >/dev/null 2>&1 && echo "  ✅ MCP Gateway running on :50051" || echo "  ❌ MCP Gateway not running"
	@nc -z localhost 8080 >/dev/null 2>&1 && echo "  ✅ gRPC-Web proxy running on :8080" || echo "  ❌ gRPC-Web proxy not running"
	@nc -z localhost 3000 >/dev/null 2>&1 && echo "  ✅ React dev server running on :3000" || echo "  ❌ React dev server not running"

# Bun-specific commands
bun-upgrade: ## Upgrade Bun to latest version
	@echo "🔄 Upgrading Bun..."
	@command -v bun >/dev/null 2>&1 && bun upgrade || echo "❌ Bun not installed (run 'make install-bun')"

bun-clean: ## Clean all Bun cache and lock files
	@echo "🧹 Cleaning Bun cache..."
	cd web/react-example && rm -rf node_modules bun.lockb .bun
	@command -v bun >/dev/null 2>&1 && bun pm cache rm || echo "⚠️  Bun not available"

bun-info: ## Show Bun environment information
	@echo "🏃‍♂️ Bun Environment Information:"
	@command -v bun >/dev/null 2>&1 && bun --version || echo "❌ Bun not installed"
	@command -v bun >/dev/null 2>&1 && bun pm ls || echo "No packages installed"
