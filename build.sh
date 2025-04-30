#!/bin/bash

# Check and install Wails
if ! command -v wails &> /dev/null; then
    echo "Installing Wails..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    export PATH="$PATH:$(go env GOPATH)/bin"
fi

# Create output directory
mkdir -p bin

# Build CLI version
cd cmd/cli
go build -o ../../bin/passwordmanager
cd ../..

# Build GUI version
cd cmd/gui
wails build
cd ../..

# Copy the GUI executable to the bin directory
cp cmd/gui/build/bin/passwordmanager-gui bin/

# Clean up build artifacts
echo "Cleaning up build artifacts..."
rm -rf cmd/gui/build
rm -rf cmd/gui/frontend/wailsjs

echo "Build completed. Executables are located in the bin directory:"
echo "- CLI: bin/passwordmanager"
echo "- GUI: bin/passwordmanager-gui"