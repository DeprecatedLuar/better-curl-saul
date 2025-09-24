#!/bin/bash

# Better-Curl-Saul Install Script
# Usage: curl -sSL https://raw.githubusercontent.com/DeprecatedLuar/better-curl-saul/main/install.sh | bash

set -e

REPO="DeprecatedLuar/better-curl-saul"
BINARY_NAME="saul"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
esac

# Try to download from releases first
echo "Let me see if there is a latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

# If no stable release, try the first release (including prereleases)
if [ -z "$LATEST_RELEASE" ]; then
    echo "No stable release found, but that's for horses anyways. Checking for prereleases..."
    LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases" | grep '"tag_name":' | head -1 | sed -E 's/.*"([^"]+)".*/\1/' || echo "")
fi

if [ -n "$LATEST_RELEASE" ]; then
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/${BINARY_NAME}-${OS}-${ARCH}"

    echo "It's your lucky day, found the release $LATEST_RELEASE"
    echo "Downloading $BINARY_NAME for $OS-$ARCH gimme a sec..."

    if curl -L -o "$BINARY_NAME" "$DOWNLOAD_URL" 2>/dev/null; then
        chmod +x "$BINARY_NAME"
        echo "Download successful! (I think, looks like it)"
    else
        echo "Release download failed. Probably skill issue. I'll build it locally for you..."
        LATEST_RELEASE=""
    fi
fi

# Fallback to local build if no release or download failed
if [ -z "$LATEST_RELEASE" ]; then
    echo "No stable release found, but that's for horses anyways. Building from source..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo "You don't have go installed? you're kidding me right?. Go get Go go, get good or get lost"
        exit 1
    fi

    # Check if we're in the repo directory (local usage)
    if [ -f "go.mod" ] && [ -f "cmd/main.go" ]; then
        echo "Building from current directory..."

        # Get version from git tag or fallback to "dev"
        VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
        LDFLAGS="-X github.com/DeprecatedLuar/better-curl-saul/src/project/utils.Version=${VERSION}"

        go build -ldflags="${LDFLAGS}" -o "$BINARY_NAME" cmd/main.go

        if [ $? -ne 0 ]; then
            echo "Build kinda failed... Probably your fault just saying. Jk I have no idea what happened"
            exit 1
        fi

        echo "The build is done! I hope, try checking it"
    else
        # Remote usage - clone and build
        echo "Cloning repository for improvised build. Buckle up"

        # Check if Git and Go are installed
        if ! command -v git &> /dev/null; then
            echo "HOW?? HOW YOU DON'T HAVE GIT INSTALLED???"
            echo "Do it NOW this is a direct order! Install Git and Go, then try again. >:("
            exit 1
        fi

        # Clone to temporary directory
        TEMP_DIR=$(mktemp -d)
        cd "$TEMP_DIR"

        if ! git clone "https://github.com/$REPO.git" .; then
            echo "Failed to clone repository"
            rm -rf "$TEMP_DIR"
            exit 1
        fi

        echo "Building $BINARY_NAME..."

        # Get version from git tag or fallback to "dev"
        VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
        LDFLAGS="-X github.com/DeprecatedLuar/better-curl-saul/src/project/utils.Version=${VERSION}"

        go build -ldflags="${LDFLAGS}" -o "$BINARY_NAME" cmd/main.go

        if [ $? -ne 0 ]; then
            echo "Build kinda failed... Probably your fault just saying. Jk don't take it personal"
            rm -rf "$TEMP_DIR"
            exit 1
        fi

        # Move binary to original directory
        mv "$BINARY_NAME" "$OLDPWD/"
        cd "$OLDPWD"
        rm -rf "$TEMP_DIR"

        echo "Build successful!"
    fi
fi

# Install binary
echo "Installing to /usr/local/bin/..."
sudo cp "$BINARY_NAME" /usr/local/bin/
sudo chmod +x "/usr/local/bin/$BINARY_NAME"

# Clean up
rm -f "$BINARY_NAME"

echo -e "\n\nInstallation is complete.\n"


cat << 'EOF'
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠛⠉⠁⠀⠀⠀⠀⠀⠹⣿⣿⣟⣿⣿⣿⣿⣿⣿⣿⣿
⣶⣶⣶⣶⣶⣶⣶⣶⡶⠰⠀⠀⠀⢀⣤⣴⣶⣶⣶⣦⡙⢻⢗⣶⣶⣶⣶⣶⣶⣶⣶
⣿⣿⣿⣿⣿⣿⣿⣿⡗⠀⠀⢠⣾⣿⣿⣿⣿⣿⣿⣿⣿⠈⢪⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⡏⠀⠀⢠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠁⠈⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⡄⠀⣿⠋⢀⠉⠙⣿⡟⠛⣋⠙⣿⠀⠀⢿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⡇⣤⣴⣶⡆⣿⣷⣶⣤⣅⣼⠀⠀⣸⣿⣿⣿⣿⣿⣿⣿
⠉⠉⠉⠉⠉⠉⠉⠉⠐⠀⡇⢿⣿⣿⠐⣿⣿⣿⣿⣿⣿⠸⢴⢛⣻⣿⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⠃⢠⡿⢿⡀⢀⣴⣿⣿⣿⣿⡶⢻⠈⢿⣿⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠘⠈⣇⣠⡭⠉⢭⣝⣻⣿⠟⠀⢸⠀⠨⢹⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠿⣿⣿⣿⣿⠟⢋⡄⠀⢸⢀⢠⣼⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⠀⠀⠀⠀⠀⠀⢀⣴⡿⡆⠀⢀⢸⣿⣿⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⣷⣄⡀⣴⣾⣿⣿⢟⣼⡇⠀⠀⠙⠻⢿⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⣠⡄⠀⠘⢿⣿⡦⣥⡝⠱⣿⣿⠃⠀⠀⠀⠀⠀⠀⠉⠛⠻⠿
⠀⠀⠀⠀⠀⠀⠀⠎⢀⡄⠀⠀⠈⢿⣤⡀⣴⠀⣹⡿⠀⠀
⠀⠀⠀⠀⠀⢠⣾⣶⣬⣧⣀⠀⣀⠈⢿⡇⠉⠂⣿⠇⠀⠀
⠀⠀⠀⠀⢀⣾⣿⣝⣛⠿⢿⡿⢿⡗⠈⠃⠀⠀⠸⠀⠀
⠀⠀⠀⠀⢀⣭⣛⣛⣛⠿⠗⠉
⠀⠀⠀⢠⠚⠻⠛⠹⠟⠷⠉⠀
⠀⠀⠀⣄⣄⠀⠀⠀⠀⠀
⠀⠀⠀⢁⠁⢠⠆⣰⠀⠀
⠀⠀⠀⢸⣷⣤⡴⠃⠀
EOF
echo -e "\nTest it with '$BINARY_NAME version'."
echo -e "Or not, I'm not your mom.\n"