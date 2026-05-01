#!/bin/bash
# MailStone Verifier - Setup Script
# This script installs Wails CLI and configures your environment

set -e

echo "🔐 MailStone Verifier - Setup Script"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check Go installation
echo "Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    echo "Please install Go 1.21+ from https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✅ Go found: $GO_VERSION${NC}"
echo ""

# Check platform
echo "Detecting platform..."
OS=$(uname -s)
echo "Operating System: $OS"

# Install platform dependencies
if [ "$OS" = "Linux" ]; then
    echo ""
    echo "Installing Linux dependencies..."

    # Detect distribution
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
        VERSION=$VERSION_ID

        echo "Distribution: $DISTRO $VERSION"

        if [ "$DISTRO" = "ubuntu" ] || [ "$DISTRO" = "linuxmint" ] || [ "$DISTRO" = "debian" ]; then
            echo ""
            echo "Installing GTK and WebKit dependencies..."

            # Determine correct WebKit package
            WEBKIT_PKG="libwebkit2gtk-4.0-dev"

            # Use 4.1 for Ubuntu 22+, Mint 22+, Debian 12+
            if [ "$DISTRO" = "ubuntu" ] && [ "${VERSION%%.*}" -ge 22 ]; then
                WEBKIT_PKG="libwebkit2gtk-4.1-dev"
            elif [ "$DISTRO" = "linuxmint" ] && [ "${VERSION%%.*}" -ge 22 ]; then
                WEBKIT_PKG="libwebkit2gtk-4.1-dev"
            elif [ "$DISTRO" = "debian" ] && [ "${VERSION%%.*}" -ge 12 ]; then
                WEBKIT_PKG="libwebkit2gtk-4.1-dev"
            fi

            echo "Using package: $WEBKIT_PKG"
            echo ""
            echo "Note: Some repository errors may appear but can be safely ignored."
            echo ""

            # Update without failing on repo errors
            sudo apt-get update || echo -e "${YELLOW}⚠️  Some repositories failed, continuing...${NC}"

            # Install packages
            if sudo apt-get install -y build-essential libgtk-3-dev $WEBKIT_PKG; then
                echo -e "${GREEN}✅ Dependencies installed${NC}"
            else
                echo -e "${RED}❌ Failed to install dependencies${NC}"
                echo "Please run manually:"
                echo "  sudo apt-get install -y build-essential libgtk-3-dev $WEBKIT_PKG"
                exit 1
            fi
        elif [ "$DISTRO" = "fedora" ]; then
            sudo dnf install -y gtk3-devel webkit2gtk3-devel
            echo -e "${GREEN}✅ Dependencies installed${NC}"
        elif [ "$DISTRO" = "arch" ]; then
            sudo pacman -S --noconfirm gtk3 webkit2gtk
            echo -e "${GREEN}✅ Dependencies installed${NC}"
        else
            echo -e "${YELLOW}⚠️  Unknown distribution. Please install GTK and WebKit manually.${NC}"
        fi
    fi
fi

# Install Wails CLI
echo ""
echo "Installing Wails CLI..."
go install github.com/wailsapp/wails/v2/cmd/wails@latest

GOPATH=$(go env GOPATH)
WAILS_PATH="$GOPATH/bin/wails"

if [ -f "$WAILS_PATH" ]; then
    echo -e "${GREEN}✅ Wails installed: $WAILS_PATH${NC}"
else
    echo -e "${RED}❌ Wails installation failed${NC}"
    exit 1
fi

# Configure PATH
echo ""
echo "Configuring PATH..."

SHELL_RC=""
if [ -n "$ZSH_VERSION" ]; then
    SHELL_RC="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ]; then
    SHELL_RC="$HOME/.bashrc"
fi

if [ -n "$SHELL_RC" ]; then
    PATH_LINE="export PATH=\$PATH:\$(go env GOPATH)/bin"

    if grep -q "go env GOPATH" "$SHELL_RC"; then
        echo -e "${YELLOW}⚠️  PATH already configured in $SHELL_RC${NC}"
    else
        echo "" >> "$SHELL_RC"
        echo "# Added by MailStone Verifier setup" >> "$SHELL_RC"
        echo "$PATH_LINE" >> "$SHELL_RC"
        echo -e "${GREEN}✅ PATH configured in $SHELL_RC${NC}"
        echo "Run: source $SHELL_RC"
    fi
fi

# Add to current session
export PATH=$PATH:$(go env GOPATH)/bin

# Verify installation
echo ""
echo "Verifying installation..."
if command -v wails &> /dev/null; then
    WAILS_VERSION=$(wails version)
    echo -e "${GREEN}✅ Wails is ready: $WAILS_VERSION${NC}"
else
    echo -e "${RED}❌ Wails not found in PATH${NC}"
    echo "Please run: export PATH=\$PATH:\$(go env GOPATH)/bin"
    exit 1
fi

# Download Go dependencies
echo ""
echo "Downloading Go dependencies..."
cd "$(dirname "$0")"
go mod download
go mod tidy
echo -e "${GREEN}✅ Dependencies downloaded${NC}"

# Summary
echo ""
echo "======================================"
echo -e "${GREEN}✅ Setup complete!${NC}"
echo "======================================"
echo ""
echo "Next steps:"
echo "  1. Reload your shell: source $SHELL_RC"
echo "  2. Run in dev mode:    make dev"
echo "  3. Build binary:       make build"
echo ""
echo "Happy verifying! 🔐"
