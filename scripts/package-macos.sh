#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}macOS DMG Packaging Script${NC}"
echo -e "${GREEN}========================================${NC}"

# Configuration
APP_NAME="Network Debugger"
BUNDLE_NAME="Network Debugger.app"
VERSION="${VERSION:-1.0.0}"
GO_BINARY_NAME="network-debugger"
DIST_DIR="$(pwd)/dist"
DMG_NAME="NetworkDebugger-${VERSION}-macos-$(uname -m).dmg"

echo -e "\n${YELLOW}Configuration:${NC}"
echo "  App Name: $APP_NAME"
echo "  Version: $VERSION"
echo "  Architecture: $(uname -m)"
echo "  Output: $DMG_NAME"

# Clean previous builds
echo -e "\n${YELLOW}Step 1/5: Cleaning previous builds...${NC}"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Build Go server binary
echo -e "\n${YELLOW}Step 2/5: Building Go server binary...${NC}"
cd "$(dirname "$0")/.."
echo "  Building for macOS $(uname -m)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=$(uname -m | sed 's/x86_64/amd64/;s/arm64/arm64/') \
  go build -ldflags="-s -w" -o "$DIST_DIR/server_darwin_$(uname -m | sed 's/x86_64/amd64/')" \
  ./cmd/network-debugger

if [ ! -f "$DIST_DIR/server_darwin_$(uname -m | sed 's/x86_64/amd64/')" ]; then
  echo -e "${RED}Error: Failed to build Go binary${NC}"
  exit 1
fi
echo -e "${GREEN}  ✓ Go binary built successfully${NC}"

# Build Flutter macOS app
echo -e "\n${YELLOW}Step 3/5: Building Flutter macOS app...${NC}"
cd frontend
flutter clean
flutter pub get
flutter build macos --release

if [ ! -d "build/macos/Build/Products/Release/$BUNDLE_NAME" ]; then
  echo -e "${RED}Error: Flutter build failed${NC}"
  exit 1
fi
echo -e "${GREEN}  ✓ Flutter app built successfully${NC}"

# Copy Go binary to app bundle
echo -e "\n${YELLOW}Step 4/5: Packaging app bundle...${NC}"
APP_BUNDLE="build/macos/Build/Products/Release/$BUNDLE_NAME"
RESOURCES_DIR="$APP_BUNDLE/Contents/Resources"

echo "  Copying Go server binary to app bundle..."
mkdir -p "$RESOURCES_DIR"
cp "$DIST_DIR/server_darwin_$(uname -m | sed 's/x86_64/amd64/')" "$RESOURCES_DIR/"

# Make binary executable
chmod +x "$RESOURCES_DIR/server_darwin_$(uname -m | sed 's/x86_64/amd64/')"

echo "  Verifying app bundle structure..."
if [ ! -f "$RESOURCES_DIR/server_darwin_$(uname -m | sed 's/x86_64/amd64/')" ]; then
  echo -e "${RED}Error: Go binary not found in app bundle${NC}"
  exit 1
fi
echo -e "${GREEN}  ✓ App bundle packaged successfully${NC}"

# Create DMG
echo -e "\n${YELLOW}Step 5/5: Creating DMG image...${NC}"

# Copy app to dist for DMG creation
cp -R "$APP_BUNDLE" "$DIST_DIR/"

# Check if create-dmg is installed
if ! command -v create-dmg &> /dev/null; then
  echo "  create-dmg not found, using hdiutil instead..."

  # Create a temporary directory for DMG contents
  DMG_TEMP="$DIST_DIR/dmg-temp"
  mkdir -p "$DMG_TEMP"
  cp -R "$DIST_DIR/$BUNDLE_NAME" "$DMG_TEMP/"

  # Create DMG using hdiutil
  hdiutil create -volname "$APP_NAME" -srcfolder "$DMG_TEMP" \
    -ov -format UDZO "$DIST_DIR/$DMG_NAME"

  rm -rf "$DMG_TEMP"
else
  echo "  Using create-dmg..."
  create-dmg \
    --volname "$APP_NAME" \
    --window-pos 200 120 \
    --window-size 600 400 \
    --icon-size 100 \
    --icon "$BUNDLE_NAME" 175 120 \
    --hide-extension "$BUNDLE_NAME" \
    --app-drop-link 425 120 \
    "$DIST_DIR/$DMG_NAME" \
    "$DIST_DIR/$BUNDLE_NAME" || {
      echo "  create-dmg failed, falling back to hdiutil..."
      DMG_TEMP="$DIST_DIR/dmg-temp"
      mkdir -p "$DMG_TEMP"
      cp -R "$DIST_DIR/$BUNDLE_NAME" "$DMG_TEMP/"
      hdiutil create -volname "$APP_NAME" -srcfolder "$DMG_TEMP" \
        -ov -format UDZO "$DIST_DIR/$DMG_NAME"
      rm -rf "$DMG_TEMP"
    }
fi

if [ ! -f "$DIST_DIR/$DMG_NAME" ]; then
  echo -e "${RED}Error: Failed to create DMG${NC}"
  exit 1
fi

# Get DMG size
DMG_SIZE=$(du -h "$DIST_DIR/$DMG_NAME" | cut -f1)

echo -e "${GREEN}  ✓ DMG created successfully${NC}"

# Summary
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "\n${YELLOW}Output:${NC}"
echo "  Location: $DIST_DIR/$DMG_NAME"
echo "  Size: $DMG_SIZE"
echo -e "\n${YELLOW}Next steps:${NC}"
echo "  1. Test the DMG: open $DIST_DIR/$DMG_NAME"
echo "  2. Upload to GitHub Releases"
echo ""
