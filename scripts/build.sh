#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="weather-lambda"
BUILD_DIR="bin"
SOURCE_DIR="./cmd/weather-lambda"

echo -e "${GREEN}Starting build process for ${BINARY_NAME}...${NC}"

# Create build directory
echo -e "${YELLOW}Creating build directory...${NC}"
mkdir -p ${BUILD_DIR}

# Tidy dependencies
echo -e "${YELLOW}Tidying dependencies...${NC}"
go mod tidy

# Download dependencies
echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod download

# Verify dependencies
echo -e "${YELLOW}Verifying dependencies...${NC}"
go mod verify

# Format code
echo -e "${YELLOW}Formatting code...${NC}"
go fmt ./...

# Vet code
echo -e "${YELLOW}Vetting code...${NC}"
go vet ./...

# Run tests
echo -e "${YELLOW}Running tests...${NC}"
go test ./... -v

# Build for Linux (Lambda runtime)
echo -e "${YELLOW}Building binary for Linux...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o ${BUILD_DIR}/${BINARY_NAME} \
    ${SOURCE_DIR}

# Check if binary was created
if [ -f "${BUILD_DIR}/${BINARY_NAME}" ]; then
    echo -e "${GREEN}Build successful! Binary created at ${BUILD_DIR}/${BINARY_NAME}${NC}"
    
    # Show binary info
    echo -e "${YELLOW}Binary information:${NC}"
    ls -lh ${BUILD_DIR}/${BINARY_NAME}
    file ${BUILD_DIR}/${BINARY_NAME}
else
    echo -e "${RED}Build failed! Binary not found.${NC}"
    exit 1
fi

echo -e "${GREEN}Build process completed successfully!${NC}"