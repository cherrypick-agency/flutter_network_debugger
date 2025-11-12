#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Linux Packaging Script${NC}"
echo -e "${GREEN}========================================${NC}"

# Configuration
APP_NAME="network-debugger"
VERSION="${VERSION:-1.0.0}"
ARCH="${ARCH:-amd64}"
DIST_DIR="$(pwd)/dist"
PACKAGE_NAME="NetworkDebugger-${VERSION}-linux-${ARCH}"

echo -e "\n${YELLOW}Configuration:${NC}"
echo "  App Name: $APP_NAME"
echo "  Version: $VERSION"
echo "  Architecture: $ARCH"

# Clean previous builds
echo -e "\n${YELLOW}Step 1/6: Cleaning previous builds...${NC}"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Build Go server binary
echo -e "\n${YELLOW}Step 2/6: Building Go server binary...${NC}"
cd "$(dirname "$0")/.."
echo "  Building for Linux $ARCH..."
CGO_ENABLED=1 GOOS=linux GOARCH="$ARCH" \
  go build -ldflags="-s -w" -o "$DIST_DIR/server_linux_$ARCH" \
  ./cmd/network-debugger

if [ ! -f "$DIST_DIR/server_linux_$ARCH" ]; then
  echo -e "${RED}Error: Failed to build Go binary${NC}"
  exit 1
fi
chmod +x "$DIST_DIR/server_linux_$ARCH"
echo -e "${GREEN}  ✓ Go binary built successfully${NC}"

# Build Flutter Linux app
echo -e "\n${YELLOW}Step 3/6: Building Flutter Linux app...${NC}"
cd frontend
flutter clean
flutter pub get
echo "  Building with version: $VERSION"
flutter build linux --release --build-name="$VERSION" --build-number="1" --no-tree-shake-icons

BUILD_DIR="build/linux/x64/release/bundle"
if [ ! -d "$BUILD_DIR" ]; then
  echo -e "${RED}Error: Flutter build failed${NC}"
  exit 1
fi
echo -e "${GREEN}  ✓ Flutter app built successfully${NC}"

# Package app
echo -e "\n${YELLOW}Step 4/6: Packaging app...${NC}"

PACKAGE_DIR="$DIST_DIR/$PACKAGE_NAME"
mkdir -p "$PACKAGE_DIR"

# Copy Flutter app
echo "  Copying Flutter app files..."
cp -R "$BUILD_DIR"/* "$PACKAGE_DIR/"

# Create resources directory
RESOURCES_DIR="$PACKAGE_DIR/resources"
mkdir -p "$RESOURCES_DIR"

# Copy Go binary
echo "  Copying Go server binary..."
cp "$DIST_DIR/server_linux_$ARCH" "$RESOURCES_DIR/"

# Create launcher script
echo "  Creating launcher script..."
cat > "$PACKAGE_DIR/$APP_NAME" << 'EOF'
#!/bin/bash

# Network Debugger Launcher Script

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Set library path for Flutter
export LD_LIBRARY_PATH="$SCRIPT_DIR/lib:$LD_LIBRARY_PATH"

# Launch the application
cd "$SCRIPT_DIR"
exec "./frontend" "$@"
EOF

chmod +x "$PACKAGE_DIR/$APP_NAME"

# Create install script
echo "  Creating install script..."
cat > "$PACKAGE_DIR/install.sh" << 'EOF'
#!/bin/bash
set -e

echo "========================================"
echo "Network Debugger Installer"
echo "========================================"
echo ""

# Check if running as root
if [ "$EUID" -eq 0 ]; then
  INSTALL_DIR="/opt/network-debugger"
  BIN_DIR="/usr/local/bin"
  DESKTOP_DIR="/usr/share/applications"
  ICON_DIR="/usr/share/icons/hicolor/256x256/apps"
  SUDO_CMD=""
else
  INSTALL_DIR="$HOME/.local/share/network-debugger"
  BIN_DIR="$HOME/.local/bin"
  DESKTOP_DIR="$HOME/.local/share/applications"
  ICON_DIR="$HOME/.local/share/icons/hicolor/256x256/apps"
  SUDO_CMD=""
fi

echo "Installing to: $INSTALL_DIR"
echo ""

# Create directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$BIN_DIR"
mkdir -p "$DESKTOP_DIR"
mkdir -p "$ICON_DIR"

# Copy files
echo "Copying files..."
cp -R * "$INSTALL_DIR/"

# Create symlink
echo "Creating symlink..."
ln -sf "$INSTALL_DIR/network-debugger" "$BIN_DIR/network-debugger"

# Create desktop entry
echo "Creating desktop entry..."
cat > "$DESKTOP_DIR/network-debugger.desktop" << DESKTOP_EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=Network Debugger
Comment=Debug HTTP/WebSocket traffic
Exec=$BIN_DIR/network-debugger
Icon=network-debugger
Terminal=false
Categories=Development;Network;
DESKTOP_EOF

chmod +x "$DESKTOP_DIR/network-debugger.desktop"

# Update desktop database
if command -v update-desktop-database &> /dev/null; then
  echo "Updating desktop database..."
  update-desktop-database "$DESKTOP_DIR" 2>/dev/null || true
fi

echo ""
echo "========================================"
echo "Installation completed successfully!"
echo "========================================"
echo ""
echo "The application has been installed to:"
echo "$INSTALL_DIR"
echo ""
echo "You can now run it by:"
echo "  1. Typing 'network-debugger' in terminal"
echo "  2. Finding 'Network Debugger' in applications menu"
echo ""

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]] && [ "$BIN_DIR" = "$HOME/.local/bin" ]; then
  echo "NOTE: $BIN_DIR is not in your PATH."
  echo "Add this line to your ~/.bashrc or ~/.zshrc:"
  echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
  echo ""
fi
EOF

chmod +x "$PACKAGE_DIR/install.sh"

# Create uninstall script
echo "  Creating uninstall script..."
cat > "$PACKAGE_DIR/uninstall.sh" << 'EOF'
#!/bin/bash

echo "========================================"
echo "Network Debugger Uninstaller"
echo "========================================"
echo ""

# Check installation location
if [ -d "/opt/network-debugger" ]; then
  INSTALL_DIR="/opt/network-debugger"
  BIN_DIR="/usr/local/bin"
  DESKTOP_DIR="/usr/share/applications"
  ICON_DIR="/usr/share/icons/hicolor/256x256/apps"
  NEEDS_SUDO="yes"
else
  INSTALL_DIR="$HOME/.local/share/network-debugger"
  BIN_DIR="$HOME/.local/bin"
  DESKTOP_DIR="$HOME/.local/share/applications"
  ICON_DIR="$HOME/.local/share/icons/hicolor/256x256/apps"
  NEEDS_SUDO="no"
fi

if [ ! -d "$INSTALL_DIR" ]; then
  echo "Network Debugger is not installed."
  exit 1
fi

echo "This will remove Network Debugger from:"
echo "$INSTALL_DIR"
echo ""
read -p "Continue? (y/N) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Uninstall cancelled."
  exit 0
fi

CMD_PREFIX=""
if [ "$NEEDS_SUDO" = "yes" ]; then
  CMD_PREFIX="sudo"
fi

echo "Removing files..."
$CMD_PREFIX rm -rf "$INSTALL_DIR"
$CMD_PREFIX rm -f "$BIN_DIR/network-debugger"
$CMD_PREFIX rm -f "$DESKTOP_DIR/network-debugger.desktop"
$CMD_PREFIX rm -f "$ICON_DIR/network-debugger.png"

# Update desktop database
if command -v update-desktop-database &> /dev/null; then
  update-desktop-database "$DESKTOP_DIR" 2>/dev/null || true
fi

echo ""
echo "========================================"
echo "Uninstallation completed!"
echo "========================================"
echo ""
EOF

chmod +x "$PACKAGE_DIR/uninstall.sh"

# Create README
cat > "$PACKAGE_DIR/README.txt" << EOF
Network Debugger v$VERSION

INSTALLATION
============
1. Extract this archive
2. Run: ./install.sh
3. Follow the on-screen instructions

The installer will:
- Install to ~/.local/share/network-debugger (user) or /opt/network-debugger (system)
- Create a symlink in ~/.local/bin or /usr/local/bin
- Add desktop entry for application menu

RUNNING
=======
After installation, you can run the app by:
- Typing 'network-debugger' in terminal
- Finding 'Network Debugger' in your applications menu

UNINSTALLATION
==============
Run the uninstall script:
  ~/.local/share/network-debugger/uninstall.sh
  or
  /opt/network-debugger/uninstall.sh

SYSTEM REQUIREMENTS
===================
- Linux (64-bit)
- GTK 3.0 or later
- 200 MB disk space
- Internet connection (for updates)

TROUBLESHOOTING
===============
If the app doesn't start:
1. Check if port 9092 is available
2. Check GTK libraries: ldd ./frontend
3. Run from terminal to see error messages: ./network-debugger

For more help, visit:
https://github.com/YOUR_USERNAME/YOUR_REPO
EOF

echo -e "${GREEN}  ✓ Package prepared successfully${NC}"

# Create tar.gz archive
echo -e "\n${YELLOW}Step 5/6: Creating tar.gz archive...${NC}"
cd "$DIST_DIR"
tar -czf "${PACKAGE_NAME}.tar.gz" "$PACKAGE_NAME"

if [ ! -f "${PACKAGE_NAME}.tar.gz" ]; then
  echo -e "${RED}Error: Failed to create tar.gz archive${NC}"
  exit 1
fi

TAR_SIZE=$(du -h "${PACKAGE_NAME}.tar.gz" | cut -f1)
echo -e "${GREEN}  ✓ tar.gz archive created successfully${NC}"

# Create deb package (optional)
echo -e "\n${YELLOW}Step 6/6: Creating deb package (optional)...${NC}"

if command -v dpkg-deb &> /dev/null; then
  DEB_DIR="$DIST_DIR/deb"

  # Normalize version for Debian (must start with a digit)
  DEB_VERSION="$VERSION"
  if [[ ! "$VERSION" =~ ^[0-9] ]]; then
    DEB_VERSION="0.0.0+${VERSION}"
  fi

  DEB_NAME="network-debugger_${DEB_VERSION}_${ARCH}"

  mkdir -p "$DEB_DIR/$DEB_NAME/DEBIAN"
  mkdir -p "$DEB_DIR/$DEB_NAME/opt/network-debugger"
  mkdir -p "$DEB_DIR/$DEB_NAME/usr/share/applications"
  mkdir -p "$DEB_DIR/$DEB_NAME/usr/bin"

  # Copy files
  cp -R "$PACKAGE_DIR"/* "$DEB_DIR/$DEB_NAME/opt/network-debugger/"

  # Create control file
  cat > "$DEB_DIR/$DEB_NAME/DEBIAN/control" << CONTROL_EOF
Package: network-debugger
Version: $DEB_VERSION
Section: devel
Priority: optional
Architecture: $ARCH
Maintainer: Your Name <your.email@example.com>
Description: Network Debugger - HTTP/WebSocket Traffic Inspector
 Network Debugger is a desktop application for debugging HTTP and WebSocket traffic.
 It provides a proxy server with powerful inspection and modification capabilities.
CONTROL_EOF

  # Create postinst script
  cat > "$DEB_DIR/$DEB_NAME/DEBIAN/postinst" << POSTINST_EOF
#!/bin/bash
ln -sf /opt/network-debugger/network-debugger /usr/bin/network-debugger
update-desktop-database /usr/share/applications 2>/dev/null || true
POSTINST_EOF

  chmod +x "$DEB_DIR/$DEB_NAME/DEBIAN/postinst"

  # Build deb
  dpkg-deb --build "$DEB_DIR/$DEB_NAME" "$DIST_DIR/${DEB_NAME}.deb"

  if [ -f "$DIST_DIR/${DEB_NAME}.deb" ]; then
    DEB_SIZE=$(du -h "$DIST_DIR/${DEB_NAME}.deb" | cut -f1)
    echo -e "${GREEN}  ✓ deb package created successfully${NC}"
  else
    echo -e "${YELLOW}  ⚠ deb package creation failed (optional)${NC}"
  fi
else
  echo -e "${YELLOW}  ⚠ dpkg-deb not found, skipping deb package (optional)${NC}"
fi

# Summary
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "\n${YELLOW}Output:${NC}"
echo "  tar.gz: $DIST_DIR/${PACKAGE_NAME}.tar.gz ($TAR_SIZE)"
if [ -f "$DIST_DIR/${DEB_NAME}.deb" ]; then
  echo "  deb: $DIST_DIR/${DEB_NAME}.deb ($DEB_SIZE)"
fi
echo -e "\n${YELLOW}Next steps:${NC}"
echo "  1. Test the installer: tar -xzf ${PACKAGE_NAME}.tar.gz && cd $PACKAGE_NAME && ./install.sh"
echo "  2. Upload to GitHub Releases"
echo ""
