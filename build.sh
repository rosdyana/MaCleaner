#!/bin/bash

# macOS Storage Cleaner Build Script
# Builds optimized binaries for macOS (amd64 and arm64)

set -e

APP_NAME="macos-cleaner"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR="bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building ${APP_NAME} v${VERSION}...${NC}"

# Create build directory
mkdir -p ${BUILD_DIR}

# Build flags for optimization
LDFLAGS="-s -w -buildid="
BUILD_FLAGS="-trimpath"

# Get dependencies
echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod download
go mod tidy

# Build for current architecture
echo -e "${YELLOW}Building for current architecture...${NC}"
CGO_ENABLED=0 go build ${BUILD_FLAGS} -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/${APP_NAME}" .

# Build for amd64 (Intel Macs)
echo -e "${YELLOW}Building for amd64 (Intel Macs)...${NC}"
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/${APP_NAME}-amd64" .

# Build for arm64 (Apple Silicon)
echo -e "${YELLOW}Building for arm64 (Apple Silicon)...${NC}"
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ${BUILD_FLAGS} -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/${APP_NAME}-arm64" .

# Create universal binary (fat binary) if lipo is available
if command -v lipo &> /dev/null; then
    echo -e "${YELLOW}Creating universal binary...${NC}"
    lipo -create -output "${BUILD_DIR}/${APP_NAME}-universal" "${BUILD_DIR}/${APP_NAME}-amd64" "${BUILD_DIR}/${APP_NAME}-arm64"
fi

# Show results
echo ""
echo -e "${GREEN}Build complete!${NC}"
echo ""
echo "Binaries in ${BUILD_DIR}/:"
echo "------------------------"
ls -lh ${BUILD_DIR}/ | tail -n +2 | awk '{printf "  %-25s %8s\n", $9, $5}'
echo ""

# Verify current arch binary
echo "File information:"
echo "------------------------"
file "${BUILD_DIR}/${APP_NAME}"
echo ""

# Optional: Compress with UPX if available
if command -v upx &> /dev/null; then
    echo -e "${YELLOW}Compressing with UPX...${NC}"
    for binary in "${BUILD_DIR}/${APP_NAME}" "${BUILD_DIR}/${APP_NAME}-amd64" "${BUILD_DIR}/${APP_NAME}-arm64"; do
        if [ -f "$binary" ]; then
            upx --best "$binary" 2>/dev/null || echo "UPX compression skipped for $(basename $binary)"
        fi
    done
    echo ""
    echo -e "${GREEN}Compressed binaries:${NC}"
    ls -lh ${BUILD_DIR}/ | tail -n +2 | awk '{printf "  %-25s %8s\n", $9, $5}'
fi

echo ""
echo -e "${GREEN}Done!${NC}"
