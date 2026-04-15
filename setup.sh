# work for linux

#!/bin/bash

# ZomCleaner Environment Setup
# ---------------------------------------------------------
echo "🧟 Initializing ZomCleaner Environment..."

# 1. Check for Go Compiler
if ! command -v go &> /dev/null; then
    echo "⚠️  Go not found. Attempting to install..."

    # Detect Package Manager
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y golang
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y golang
    elif command -v pacman &> /dev/null; then
        sudo pacman -S --noconfirm go
    elif command -v brew &> /dev/null; then
        brew install go
    else
        echo "❌ Error: No supported package manager found (apt, dnf, pacman, brew)."
        echo "Please install Go manually from https://go.dev/dl/"
        exit 1
    fi
else
    echo "✅ Go is already installed ($(go version | awk '{print $3}'))."
fi

# 2. Verify GCC (for potential CGO dependencies)
if ! command -v gcc &> /dev/null; then
    echo "📦 Installing build-essential tools..."
    if command -v apt-get &> /dev/null; then sudo apt-get install -y build-essential
    elif command -v dnf &> /dev/null; then sudo dnf groupinstall -y "Development Tools"
    elif command -v pacman &> /dev/null; then sudo pacman -S --noconfirm base-devel
    fi
fi

# 3. Final Checks
echo "---"
echo "🚀 Setup Complete."
echo "💡 To run ZomCleaner (Linux):"
echo "   sudo go run Clean0.go"
echo ""
echo "Note: Root privileges are required for /proc filesystem interaction."
