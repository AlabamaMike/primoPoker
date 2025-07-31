#!/bin/bash

# PrimoPoker Development Environment Setup Script
# This script installs all necessary tools and services for development

set -e

echo "🃏 Setting up PrimoPoker development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Update package lists
log_info "Updating package lists..."
sudo apt-get update -y

# Install essential development tools
log_info "Installing essential development tools..."
sudo apt-get install -y \
    curl \
    wget \
    git \
    vim \
    nano \
    tree \
    jq \
    unzip \
    build-essential \
    software-properties-common \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release \
    htop \
    postgresql-client \
    redis-tools

# Install Go tools
log_info "Installing Go development tools..."
go install -a github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install -a golang.org/x/tools/cmd/goimports@latest
go install -a golang.org/x/tools/cmd/godoc@latest
go install -a github.com/go-delve/delve/cmd/dlv@latest
go install -a github.com/onsi/ginkgo/v2/ginkgo@latest
go install -a github.com/securecodewarrior/sast-scan/cmd/sast-scan@latest

# Install air for hot reloading (useful for development)
log_info "Installing air for hot reloading..."
go install github.com/air-verse/air@latest

# Install migrate tool for database migrations
log_info "Installing golang-migrate..."
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate

# Install PostgreSQL
log_info "Installing PostgreSQL..."
sudo apt-get install -y postgresql postgresql-contrib
sudo service postgresql start

# Configure PostgreSQL
log_info "Configuring PostgreSQL..."
sudo -u postgres psql -c "CREATE USER primopoker WITH PASSWORD 'primopoker';" || true
sudo -u postgres psql -c "CREATE DATABASE primopoker OWNER primopoker;" || true
sudo -u postgres psql -c "CREATE DATABASE primopoker_test OWNER primopoker;" || true
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE primopoker TO primopoker;" || true
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE primopoker_test TO primopoker;" || true

# Install Redis
log_info "Installing Redis..."
sudo apt-get install -y redis-server
sudo service redis-server start

# Configure Redis
log_info "Configuring Redis..."
sudo sed -i 's/^bind 127.0.0.1/bind 0.0.0.0/' /etc/redis/redis.conf
sudo systemctl restart redis-server

# Install Google Cloud CLI (for GCP deployment)
log_info "Installing Google Cloud CLI..."
curl https://sdk.cloud.google.com | bash
source ~/.bashrc

# Install Terraform (for infrastructure)
log_info "Installing Terraform..."
wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt update && sudo apt install terraform

# Install Docker Compose (additional)
log_info "Installing Docker Compose..."
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install websocket testing tools
log_info "Installing websocket testing tools..."
sudo npm install -g wscat

# Create useful aliases
log_info "Setting up development aliases..."
cat >> ~/.bashrc << 'EOF'

# PrimoPoker Development Aliases
alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'
alias poker-server='go run ./cmd/server'
alias poker-test='go test ./tests/...'
alias poker-build='go build -o bin/server ./cmd/server'
alias poker-bench='go test -bench=. ./tests/...'
alias poker-cover='go test -cover ./tests/...'
alias poker-lint='golangci-lint run'
alias poker-fmt='gofmt -w .'
alias poker-migrate-up='migrate -path ./migrations -database "postgres://primopoker:primopoker@localhost:5432/primopoker?sslmode=disable" up'
alias poker-migrate-down='migrate -path ./migrations -database "postgres://primopoker:primopoker@localhost:5432/primopoker?sslmode=disable" down'
alias poker-logs='journalctl -f -u primopoker'
alias redis-cli='redis-cli -h localhost -p 6379'
alias psql-poker='psql -h localhost -U primopoker -d primopoker'

# Git aliases
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git log --oneline'
alias gd='git diff'

# Development environment info
alias poker-env='echo "🃏 PrimoPoker Development Environment"; echo "Go: $(go version)"; echo "PostgreSQL: $(psql --version)"; echo "Redis: $(redis-server --version)"'
EOF

# Set up Go module cache and workspace
log_info "Setting up Go workspace..."
go mod download
go mod tidy

# Create necessary directories
log_info "Creating project directories..."
mkdir -p logs
mkdir -p tmp
mkdir -p bin
mkdir -p migrations

# Create a sample environment file
log_info "Creating sample environment file..."
cat > .env.example << 'EOF'
# Server Configuration
PORT=8080
HOST=localhost
ENV=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=primopoker
DB_PASSWORD=primopoker
DB_NAME=primopoker
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Game Configuration
MAX_PLAYERS_PER_TABLE=9
SMALL_BLIND=10
BIG_BLIND=20
MAX_TABLES=100

# GCP Configuration (for production)
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_APPLICATION_CREDENTIALS=path/to/service-account.json

# Logging
LOG_LEVEL=debug
LOG_FORMAT=text

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100

# Security
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
EOF

# Copy sample env to actual env if it doesn't exist
if [ ! -f .env ]; then
    cp .env.example .env
    log_success "Created .env file from template"
fi

# Create air configuration for hot reloading
log_info "Creating air configuration..."
cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "node_modules", ".git", "logs", "bin"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
EOF

# Create a basic Makefile for common tasks
log_info "Creating Makefile..."
cat > Makefile << 'EOF'
.PHONY: build run test bench cover lint fmt clean dev migrate-up migrate-down setup help

# Build the server
build:
	go build -o bin/server ./cmd/server

# Run the server
run:
	go run ./cmd/server

# Run tests
test:
	go test -v -race ./tests/...

# Run benchmarks
bench:
	go test -bench=. ./tests/...

# Run tests with coverage
cover:
	go test -cover ./tests/...

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	gofmt -w .
	goimports -w .

# Clean build artifacts
clean:
	go clean
	rm -rf bin/
	rm -rf tmp/

# Run in development mode with hot reloading
dev:
	air

# Run database migrations up
migrate-up:
	migrate -path ./migrations -database "postgres://primopoker:primopoker@localhost:5432/primopoker?sslmode=disable" up

# Run database migrations down
migrate-down:
	migrate -path ./migrations -database "postgres://primopoker:primopoker@localhost:5432/primopoker?sslmode=disable" down

# Setup development environment
setup:
	go mod download
	go mod tidy

# Show help
help:
	@echo "PrimoPoker Development Commands:"
	@echo "  build      - Build the server binary"
	@echo "  run        - Run the server"
	@echo "  test       - Run tests"
	@echo "  bench      - Run benchmarks"
	@echo "  cover      - Run tests with coverage"
	@echo "  lint       - Lint the code"
	@echo "  fmt        - Format the code"
	@echo "  clean      - Clean build artifacts"
	@echo "  dev        - Run with hot reloading"
	@echo "  migrate-up - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  setup      - Setup development environment"
	@echo "  help       - Show this help"
EOF

# Set up systemctl services for development
log_info "Setting up development services..."

# Create systemd service for PostgreSQL (ensure it starts on boot)
sudo systemctl enable postgresql

# Create systemd service for Redis (ensure it starts on boot)
sudo systemctl enable redis-server

# Start services
sudo systemctl start postgresql
sudo systemctl start redis-server

# Test database connection
log_info "Testing database connection..."
if PGPASSWORD=primopoker psql -h localhost -U primopoker -d primopoker -c "SELECT 1;" > /dev/null 2>&1; then
    log_success "PostgreSQL connection successful"
else
    log_error "PostgreSQL connection failed"
fi

# Test Redis connection
log_info "Testing Redis connection..."
if redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then
    log_success "Redis connection successful"
else
    log_error "Redis connection failed"
fi

# Install VS Code extensions (if running in codespace)
if [ "$CODESPACE_NAME" ]; then
    log_info "Installing additional VS Code extensions for Codespaces..."
    code --install-extension golang.go
    code --install-extension ms-vscode.vscode-json
    code --install-extension redhat.vscode-yaml
    code --install-extension ms-azuretools.vscode-docker
fi

# Final setup
log_info "Running final setup..."
source ~/.bashrc

# Display environment info
log_success "🎉 PrimoPoker development environment setup complete!"
echo ""
echo "🃏 Environment Information:"
echo "   Go: $(go version)"
echo "   PostgreSQL: $(psql --version | head -n1)"
echo "   Redis: $(redis-server --version)"
echo "   Docker: $(docker --version)"
echo "   Node.js: $(node --version)"
echo ""
echo "🔧 Development Commands:"
echo "   make help          - Show all available commands"
echo "   make dev           - Start development server with hot reload"
echo "   make test          - Run tests"
echo "   make build         - Build the server"
echo "   poker-env          - Show environment info"
echo ""
echo "🗄️  Database:"
echo "   PostgreSQL running on port 5432"
echo "   Redis running on port 6379"
echo "   Database: primopoker (user: primopoker, password: primopoker)"
echo ""
echo "🚀 Ready to start developing! Run 'make dev' to start the server with hot reloading."
