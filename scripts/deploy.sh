#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="dev"
STACK_NAME=""
WEATHER_API_KEY=""
CITY_NAME="Tokyo"
GUIDED=false

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -e, --environment ENV    Environment (dev, staging, prod) [default: dev]"
    echo "  -s, --stack-name NAME    Stack name [default: weather-lambda-ENV]"
    echo "  -k, --api-key KEY        Weather API key (required)"
    echo "  -c, --city CITY          City name [default: Tokyo]"
    echo "  -g, --guided             Use guided deployment"
    echo "  -h, --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --guided"
    echo "  $0 -e prod -k YOUR_API_KEY -c \"New York\""
    echo "  $0 --environment staging --api-key YOUR_API_KEY --stack-name my-weather-stack"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -s|--stack-name)
            STACK_NAME="$2"
            shift 2
            ;;
        -k|--api-key)
            WEATHER_API_KEY="$2"
            shift 2
            ;;
        -c|--city)
            CITY_NAME="$2"
            shift 2
            ;;
        -g|--guided)
            GUIDED=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_usage
            exit 1
            ;;
    esac
done

# Set default stack name if not provided
if [ -z "$STACK_NAME" ]; then
    STACK_NAME="weather-lambda-${ENVIRONMENT}"
fi

echo -e "${BLUE}🚀 Starting deployment process...${NC}"
echo -e "${YELLOW}Environment: ${ENVIRONMENT}${NC}"
echo -e "${YELLOW}Stack Name: ${STACK_NAME}${NC}"
echo -e "${YELLOW}City: ${CITY_NAME}${NC}"

# Validate environment
if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|prod)$ ]]; then
    echo -e "${RED}Error: Environment must be dev, staging, or prod${NC}"
    exit 1
fi

# Check if API key is provided (except for guided deployment)
if [ "$GUIDED" = false ]; then
    if [ -z "$WEATHER_API_KEY" ] && [ -z "$WEATHER_API_KEY" ]; then
        echo -e "${RED}Error: Weather API key is required.${NC}"
        echo -e "${YELLOW}Set WEATHER_API_KEY environment variable or use -k option${NC}"
        echo -e "${YELLOW}Tip: You can get a free API key from https://openweathermap.org/api${NC}"
        exit 1
    fi
    # Use command line parameter if provided, otherwise use environment variable
    if [ -n "$1" ] && [[ "$1" == "-k" || "$1" == "--api-key" ]]; then
        WEATHER_API_KEY="$2"
    elif [ -z "$WEATHER_API_KEY" ]; then
        WEATHER_API_KEY="$WEATHER_API_KEY"
    fi
fi

# Build the application
echo -e "${YELLOW}📦 Building application...${NC}"
./scripts/build.sh

# Build with SAM
echo -e "${YELLOW}🔨 Building with SAM...${NC}"
sam build

# Deploy with SAM
if [ "$GUIDED" = true ]; then
    echo -e "${YELLOW}🚀 Starting guided deployment...${NC}"
    docker compose exec dev sam deploy --guided --region ${AWS_REGION:-ap-northeast-1}
else
    echo -e "${YELLOW}🚀 Deploying to AWS...${NC}"
    
    # Get API key from environment if not provided via command line
    if [ -z "$WEATHER_API_KEY" ]; then
        if [ -f .env ]; then
            export $(grep -v '^#' .env | xargs)
        fi
    fi
    
    # Validate API key again
    if [ -z "$WEATHER_API_KEY" ]; then
        echo -e "${RED}Error: WEATHER_API_KEY not found in environment or .env file${NC}"
        exit 1
    fi
    
    # Deploy using Docker
    docker compose exec dev sam deploy \
        --region ${AWS_REGION:-ap-northeast-1} \
        --resolve-s3 \
        --stack-name "$STACK_NAME" \
        --parameter-overrides WeatherAPIKey="$WEATHER_API_KEY" CityName="$CITY_NAME" Environment="$ENVIRONMENT" \
        --capabilities CAPABILITY_IAM \
        --no-confirm-changeset \
        --no-fail-on-empty-changeset
fi

# Get stack outputs
echo -e "${YELLOW}📋 Getting stack outputs...${NC}"
aws cloudformation describe-stacks \
    --stack-name "$STACK_NAME" \
    --query 'Stacks[0].Outputs[*].[OutputKey,OutputValue]' \
    --output table

echo -e "${GREEN}✅ Deployment completed successfully!${NC}"

# Show next steps
echo -e "${BLUE}"
echo "🎉 Next steps:"
echo "1. Check the Lambda function in AWS Console"
echo "2. Monitor CloudWatch logs for execution details"
echo "3. Verify data in S3 bucket and DynamoDB table"
echo "4. The function is scheduled to run every hour automatically"
echo ""
echo "To invoke the function manually:"
echo "aws lambda invoke --function-name ${STACK_NAME}-WeatherLambdaFunction --payload '{}' response.json"
echo -e "${NC}"