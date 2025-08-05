#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${GREEN}Setting up Task Queue Development Environment${NC}"

# Check prerequisites
check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ $1 found${NC}"
}

echo -e "\n${YELLOW}Checking prerequisites...${NC}"
check_command "go"
check_command "docker"
check_command "docker-compose"
check_command "make"
check_command "protoc"

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_GO_VERSION="1.22"
if [ "$(printf '%s\n' "$REQUIRED_GO_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_GO_VERSION" ]; then
    echo -e "${RED}Go version $REQUIRED_GO_VERSION or higher is required (found $GO_VERSION)${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go version $GO_VERSION${NC}"

# Create necessary directories
echo -e "\n${YELLOW}Creating project directories...${NC}"
mkdir -p "$PROJECT_ROOT"/{bin,tmp,logs}
mkdir -p "$PROJECT_ROOT"/deployments/{prometheus,grafana/{dashboards,datasources}}

# Install Go tools
echo -e "\n${YELLOW}Installing Go tools...${NC}"
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Download Go dependencies
echo -e "\n${YELLOW}Downloading Go dependencies...${NC}"
cd "$PROJECT_ROOT"
go mod download

# Copy environment file
if [ ! -f "$PROJECT_ROOT/.env" ]; then
    echo -e "\n${YELLOW}Creating .env file...${NC}"
    cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/.env"
    echo -e "${GREEN}✓ .env file created${NC}"
fi

# Generate protobuf code
echo -e "\n${YELLOW}Generating protobuf code...${NC}"
make proto-gen

# Setup git hooks (optional)
if [ -d "$PROJECT_ROOT/.git" ]; then
    echo -e "\n${YELLOW}Setting up git hooks...${NC}"
    cat > "$PROJECT_ROOT/.git/hooks/pre-commit" << 'EOF'
#!/bin/bash
make fmt
make lint
EOF
    chmod +x "$PROJECT_ROOT/.git/hooks/pre-commit"
    echo -e "${GREEN}✓ Git hooks configured${NC}"
fi

# Pull Docker images
echo -e "\n${YELLOW}Pulling Docker images...${NC}"
docker-compose -f "$PROJECT_ROOT/deployments/docker-compose.yml" pull

echo -e "\n${GREEN}✅ Setup complete!${NC}"
echo -e "\nNext steps:"
echo -e "  1. Review and update ${YELLOW}.env${NC} file with your settings"
echo -e "  2. Run ${YELLOW}make docker-compose-up${NC} to start infrastructure"
echo -e "  3. Run ${YELLOW}make migrate-up${NC} to apply database migrations"
echo -e "  4. Run ${YELLOW}make run-local${NC} to start all services"
echo -e "\nUseful commands:"
echo -e "  ${YELLOW}make help${NC}         - Show all available commands"
echo -e "  ${YELLOW}make test${NC}         - Run tests"
echo -e "  ${YELLOW}make lint${NC}         - Run linter"
echo -e "  ${YELLOW}make docker-build${NC} - Build Docker images"
