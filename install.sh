#!/bin/bash
set -e

# Variables
BIN_NAME="promptgen"
INSTALL_DIR="$HOME/.local/bin"

echo "Building $BIN_NAME"

go build -o "$BIN_NAME"

echo "Build successful."

go install

echo "Installed successful"


