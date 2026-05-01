#!/bin/bash
# MailStone Verifier - WebKit Setup Script
# Fixes WebKit compatibility issues for Wails on Ubuntu 24.04 / Linux Mint 22+

set -e

echo "🔧 MailStone Verifier - WebKit Setup"
echo "====================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo -e "${RED}❌ Please do not run this script as root${NC}"
    echo "Run it as a normal user, it will ask for sudo when needed."
    exit 1
fi

# Step 1: Install WebKit development package
echo "Step 1: Installing WebKit development package..."
echo ""

if dpkg -l | grep -q "libwebkit2gtk-4.1-dev"; then
    echo -e "${GREEN}✅ libwebkit2gtk-4.1-dev is already installed${NC}"
else
    echo "Installing libwebkit2gtk-4.1-dev..."
    sudo apt-get update -qq || true
    sudo apt-get install -y libwebkit2gtk-4.1-dev

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ libwebkit2gtk-4.1-dev installed successfully${NC}"
    else
        echo -e "${RED}❌ Failed to install libwebkit2gtk-4.1-dev${NC}"
        exit 1
    fi
fi

echo ""

# Step 2: Verify installation
echo "Step 2: Verifying pkg-config can find webkit2gtk-4.1..."
echo ""

if pkg-config --exists webkit2gtk-4.1; then
    VERSION=$(pkg-config --modversion webkit2gtk-4.1)
    echo -e "${GREEN}✅ webkit2gtk-4.1 found: version $VERSION${NC}"
else
    echo -e "${RED}❌ webkit2gtk-4.1 not found by pkg-config${NC}"
    exit 1
fi

echo ""

# Step 3: Check if Wails can find webkit2gtk-4.0
echo "Step 3: Checking Wails compatibility..."
echo ""

if pkg-config --exists webkit2gtk-4.0; then
    echo -e "${GREEN}✅ webkit2gtk-4.0 already available (no symlink needed)${NC}"
else
    echo -e "${YELLOW}⚠️  webkit2gtk-4.0 not found (Wails looks for this version)${NC}"
    echo "Creating compatibility symlink..."

    # Find the actual .pc file location
    PC_FILE=$(pkg-config --variable=pcfiledir webkit2gtk-4.1)/webkit2gtk-4.1.pc
    PC_DIR=$(dirname "$PC_FILE")

    if [ -f "$PC_FILE" ]; then
        echo "Found webkit2gtk-4.1.pc at: $PC_FILE"
        echo "Creating symlink: $PC_DIR/webkit2gtk-4.0.pc -> webkit2gtk-4.1.pc"

        sudo ln -sf webkit2gtk-4.1.pc "$PC_DIR/webkit2gtk-4.0.pc"

        if pkg-config --exists webkit2gtk-4.0; then
            echo -e "${GREEN}✅ Symlink created successfully${NC}"
        else
            echo -e "${RED}❌ Symlink creation failed${NC}"
            exit 1
        fi
    else
        echo -e "${RED}❌ Could not find webkit2gtk-4.1.pc file${NC}"
        exit 1
    fi
fi

echo ""

# Step 4: Final verification
echo "Step 4: Final verification..."
echo ""

echo "Checking pkg-config can resolve all webkit dependencies..."
if pkg-config --cflags webkit2gtk-4.0 > /dev/null 2>&1; then
    echo -e "${GREEN}✅ pkg-config can resolve webkit2gtk-4.0${NC}"
else
    echo -e "${RED}❌ pkg-config cannot resolve webkit2gtk-4.0${NC}"
    exit 1
fi

echo ""
echo "====================================="
echo -e "${GREEN}✅ WebKit setup complete!${NC}"
echo "====================================="
echo ""
echo "You can now run the application:"
echo "  cd verifier"
echo "  make dev"
echo ""
