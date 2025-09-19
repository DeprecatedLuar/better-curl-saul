#!/bin/bash

echo "Building better-curl-saul..."
go build -o saul cmd/main.go

echo "Installing to /usr/local/bin/ (requires sudo)..."
sudo mv saul /usr/local/bin/

echo "Installation complete! You can now use 'saul' from anywhere."
echo "Try: saul version"