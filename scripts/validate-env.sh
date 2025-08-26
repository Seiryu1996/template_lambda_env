#!/bin/bash

# Environment validation script for deployment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Validating deployment environment...${NC}"

# Load .env if exists
if [ -f .env ]; then
    echo -e "${YELLOW}Loading environment from .env file...${NC}"
    export $(grep -v '^#' .env | xargs)
fi

# Validate required environment variables
MISSING_VARS=""

if [ -z "$WEATHER_API_KEY" ]; then
    MISSING_VARS="$MISSING_VARS WEATHER_API_KEY"
fi

if [ -z "$AWS_REGION" ]; then
    echo -e "${YELLOW}Warning: AWS_REGION not set, using default: ap-northeast-1${NC}"
    export AWS_REGION="ap-northeast-1"
fi

if [ -z "$CITY_NAME" ]; then
    echo -e "${YELLOW}Warning: CITY_NAME not set, using default: Tokyo${NC}"
    export CITY_NAME="Tokyo"
fi

# Check for missing required variables
if [ -n "$MISSING_VARS" ]; then
    echo -e "${RED}Error: Missing required environment variables:${NC}"
    for var in $MISSING_VARS; do
        echo -e "${RED}  - $var${NC}"
    done
    echo -e "${YELLOW}"
    echo "Solutions:"
    echo "1. Set environment variables: export WEATHER_API_KEY=your_key"
    echo "2. Create/update .env file with required values"
    echo "3. Get Weather API key from: https://openweathermap.org/api"
    echo -e "${NC}"
    exit 1
fi

# Validate AWS credentials
echo -e "${YELLOW}Checking AWS credentials...${NC}"
if ! docker compose exec dev aws sts get-caller-identity > /dev/null 2>&1; then
    echo -e "${RED}Error: AWS credentials not configured${NC}"
    echo -e "${YELLOW}Configure AWS credentials with: aws configure${NC}"
    exit 1
fi

# Validate Docker environment
echo -e "${YELLOW}Checking Docker environment...${NC}"
if ! docker compose ps | grep -q "dev.*Up"; then
    echo -e "${YELLOW}Starting development environment...${NC}"
    docker compose up -d dev
fi

echo -e "${GREEN}✅ Environment validation passed!${NC}"
echo -e "${GREEN}Ready for deployment with the following configuration:${NC}"
echo -e "  Region: ${AWS_REGION}"
echo -e "  City: ${CITY_NAME}"
echo -e "  API Key: ${WEATHER_API_KEY:0:8}... (masked)"