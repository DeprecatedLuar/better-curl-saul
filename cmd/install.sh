#!/bin/bash

# Better-Curl-Saul Install Script
# Usage: ./install.sh

set -e

echo "Building saul..."
go build -o saul ./cmd

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Installing to /usr/local/bin/..."
    sudo mv saul /usr/local/bin/
    echo "Installation complete! Test with: saul version"
else
    echo "Build failed!"
    exit 1
fi