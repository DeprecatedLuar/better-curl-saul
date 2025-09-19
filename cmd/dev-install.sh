#!/bin/bash

echo "Building latest version..."
cd ..
go build -o saul cmd/main.go

echo "Updating system installation..."
sudo mv saul /usr/local/bin/

echo "✅ Dev install complete! Test with: saul version"