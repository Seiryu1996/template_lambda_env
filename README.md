# Weather Lambda System

Complete serverless weather data collection and history API system built with Go, AWS Lambda, API Gateway, and EventBridge. Includes automated weather data collection with scheduled execution and REST API for querying historical weather patterns.

## 🏗️ Architecture

This project implements a serverless weather data collection system with the following components:

- **Weather Collection Lambda**: Go-based function that fetches weather data from OpenWeatherMap API (scheduled execution every hour via EventBridge)
- **Weather History API**: REST API Lambda function for querying historical weather data with flexible time-based filtering (6h, 24h, custom periods)
- **API Gateway**: RESTful endpoints with CORS support and OpenAPI schema for weather history access
- **EventBridge**: Scheduled rule for automated weather data collection (cron: rate(1 hour))
- **S3 Storage**: Date-organized JSON storage for detailed weather data with versioning and encryption
- **DynamoDB**: High-performance NoSQL storage for weather records with TTL (30 days) and efficient querying
- **CloudFormation/SAM**: Complete Infrastructure as Code with secure IAM roles and least privilege access
- **Docker Development Environment**: Multi-container setup with Go tools, AWS CLI, and SAM CLI for seamless development

## 📁 Project Structure

```
.
├── cmd/
│   ├── weather-lambda/       # Weather data collection Lambda function
│   └── weather-history-api/  # Weather history API Lambda function
├── internal/
│   ├── config/              # Configuration management
│   ├── models/              # Data models
│   ├── services/            # Business logic
│   └── handlers/            # AWS service handlers
├── bin/                     # Weather collection Lambda binary
├── bin-history/             # Weather history API Lambda binary
├── scripts/                 # Build, deployment, and validation scripts
├── tests/                   # Integration tests
├── docker-compose.yml       # Multi-container development environment
├── Dockerfile               # Production Lambda image
├── Dockerfile.dev           # Development image with Go, AWS CLI, SAM CLI
├── template.yaml            # SAM CloudFormation template (dual Lambda + API Gateway)
├── Makefile                # Comprehensive build automation with testing
└── .env.example            # Environment configuration template
```

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- AWS CLI (for deployment)
- Go 1.23+ (if building locally)
- OpenWeatherMap API key (free at https://openweathermap.org/api)

### 1. Environment Setup

```bash
# Copy environment template
make setup-env

# Edit .env with your values
cp .env.example .env
```

Update `.env` with your configuration:
```bash
# AWS Configuration
AWS_REGION=ap-northeast-1
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here

# Application Configuration
S3_BUCKET=weather-data-bucket
DYNAMODB_TABLE=weather-records

# Weather API Configuration
WEATHER_API_KEY=your_weather_api_key_here
CITY_NAME=Tokyo
```

### 2. Development Environment

```bash
# Complete development setup (includes LocalStack and DynamoDB)
make dev-setup

# Or manually:
make docker-dev
make localstack-create-resources
```

### 3. Dependencies and Build

```bash
# Tidy and download Go dependencies (Docker)
make deps-docker

# Build the Lambda function (Docker - recommended)
make build-docker

# Or build locally (requires Go 1.23+)
make deps
make build

# Run tests
make test

# Run integration tests (requires AWS deployment)
make test-aws
```

### 4. AWS Testing

```bash
# Verify AWS credentials
make aws-verify

# Deploy to AWS
make sam-deploy

# Test deployed Lambda function
make sam-invoke-aws
```

## 🛠️ Development

### Using Docker Development Environment

```bash
# Start development environment
make docker-dev

# Enter development container
docker compose exec dev bash

# Inside container, you can run:
go mod tidy
go build ./cmd/weather-lambda
go test ./...
golangci-lint run
```

### Using SAM CLI Environment

```bash
# Start SAM environment
make docker-sam

# Enter SAM container
docker compose exec sam bash

# Inside container:
sam build
sam local invoke WeatherLambdaFunction
```

### Available Make Commands

#### 🏗️ Build Commands
```bash
make build                   # Build Weather Collection Lambda binary (requires Go 1.23+)
make build-docker            # Build Weather Collection Lambda using Docker (recommended)
make build-history           # Build Weather History API Lambda binary (requires Go 1.23+)
make build-history-docker    # Build Weather History API Lambda using Docker (recommended)
make build-all               # Build all Lambda functions using Docker (recommended for deployment)
make docker-build            # Build production Docker image
make sam-build               # Build all functions with SAM (calls build-all automatically)
```

#### 📦 Dependency Management
```bash
make deps                    # Tidy and download Go dependencies (requires Go 1.23+)
make deps-docker             # Tidy dependencies using Docker container (recommended)
make deps-update             # Update all Go dependencies to latest versions
```

#### 🐳 Docker Commands
```bash
make docker-dev              # Start development environment (dev + sam containers)
make docker-sam              # Start SAM CLI environment only
make docker-stop             # Stop all Docker services
```

#### 🧪 Testing Commands
```bash
make test                    # Run Go unit tests
make test-aws                # Run integration tests against AWS resources (requires deployment)
make test-lambda-direct      # Test Weather Collection Lambda using known function name
make test-history-api        # Test Weather History API endpoint (6h period)
make test-aws-cli            # Verify AWS CLI connectivity in container
make lint                    # Run golangci-lint for code quality
make fmt                     # Format Go code
```

#### 🚀 Deployment Commands
```bash
make sam-deploy              # Interactive guided deployment to AWS
make sam-deploy-dev          # Deploy to development environment (requires WEATHER_API_KEY env var)
make sam-deploy-prod         # Deploy to production environment (requires WEATHER_API_KEY env var)
make aws-verify              # Verify AWS credentials and connectivity
make deploy-checklist        # Show pre-deployment requirements checklist
```

#### 🧹 Utility Commands
```bash
make help                    # Show all available commands with descriptions
make clean                   # Clean all build artifacts (bin/, bin-history/, .aws-sam/)
make setup-env               # Copy .env.example to .env for initial setup
make dev-setup               # Complete development environment setup (setup-env + docker-dev + deps-docker)
```

#### 🔍 Monitoring Commands
```bash
make sam-invoke-aws          # Invoke deployed Weather Collection Lambda function
```

## 🚢 Deployment

### Prerequisites for Deployment

1. **Set Weather API Key** (required):
   ```bash
   export WEATHER_API_KEY=your_openweathermap_api_key
   ```
   Or add to `.env` file:
   ```bash
   WEATHER_API_KEY=your_openweathermap_api_key
   ```

2. **Optional Environment Variables**:
   ```bash
   export AWS_REGION=ap-northeast-1    # Default region
   export CITY_NAME=Tokyo              # Default city
   ```

### Development Deployment

```bash
# Method 1: Using Makefile (recommended)
make sam-deploy-dev

# Method 2: Using deployment script
./scripts/deploy.sh --guided

# Method 3: Manual deployment script
./scripts/deploy.sh --environment dev --city "Tokyo"
```

### Production Deployment

```bash
# Using Makefile
make sam-deploy-prod

# Using deployment script
./scripts/deploy.sh --environment prod --city "New York" --stack-name weather-prod
```

### Manual SAM Deployment

```bash
# Environment validation
./scripts/validate-env.sh

# Build and deploy
make sam-build
make sam-deploy              # Guided deployment
make sam-deploy-dev          # Development deployment
make sam-deploy-prod         # Production deployment
```

## 📊 Monitoring and Logs

### CloudWatch Logs

The Lambda function logs to CloudWatch. View logs with:

```bash
# Using AWS CLI
aws logs tail /aws/lambda/weather-lambda-dev-WeatherLambdaFunction --follow

# Using SAM CLI
sam logs --name WeatherLambdaFunction --tail
```

### Data Verification

Check stored data:

```bash
# List S3 objects
aws s3 ls s3://your-weather-bucket/weather-data/ --recursive

# Query DynamoDB
aws dynamodb scan --table-name your-weather-table --limit 5
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/services/
```

### Integration Tests (Requires AWS Resources)

```bash
# Deploy to AWS first
make sam-deploy

# Run tests against real AWS resources
make test-aws

# Or manually:
go test -v -tags=integration ./tests/...
```

### Lambda Function Testing

```bash
# Test deployed function
make sam-invoke-aws

# Or manually:
aws lambda invoke --function-name [FUNCTION-NAME] response.json && cat response.json

## 🌤️ Weather History API

### API Endpoint
```
https://gdsfuvwsae.execute-api.ap-northeast-1.amazonaws.com/dev/weather/history
```

### Supported Query Parameters

| Parameter | Description | Values | Default | Required |
|-----------|-------------|--------|---------|----------|
| `period` | Time range for historical data | `6h`, `24h`, `1d`, or number (1-168 hours) | `6h` | No |
| `city` | City name filter | Any city name | From config (Tokyo) | No |

### Usage Examples

```bash
# Using Makefile (recommended)
make test-history-api                    # Test with default settings (6h)

# Direct API calls
curl -s "API_URL?period=6h"              # Last 6 hours
curl -s "API_URL?period=24h"             # Last 24 hours 
curl -s "API_URL?period=1d"              # Last 1 day (same as 24h)
curl -s "API_URL?period=12"              # Last 12 hours (custom)
curl -s "API_URL?period=6h&city=Tokyo"   # Last 6 hours for Tokyo

# Get API URL dynamically (for scripting)
API_URL=$(aws cloudformation describe-stacks --stack-name weather-lambda-dev --region ap-northeast-1 --query 'Stacks[0].Outputs[?OutputKey==`WeatherHistoryApiUrl`].OutputValue' --output text)
curl -s "$API_URL?period=6h"
```

### CORS Support
The API supports cross-origin requests with the following headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, X-Amz-Date, Authorization, X-Api-Key, X-Amz-Security-Token`

### API Response Format

```json
{
  "statusCode": 200,
  "message": "Weather history retrieved successfully",
  "data": [
    {
      "id": "Tokyo-1756230581",
      "timestamp": "2025-08-26T17:49:41Z",
      "cityName": "Tokyo",
      "temperature": 28.9,
      "description": "clear sky",
      "humidity": 67,
      "pressure": 1009,
      "windSpeed": 6.19,
      "country": "JP",
      "createdAt": "2025-08-26T17:49:41.819072652Z",
      "ttl": 1758822581
    }
  ],
  "count": 1,
  "period": "6h",
  "startTime": "2025-08-26T11:49:50Z",
  "endTime": "2025-08-26T17:49:50Z"
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `statusCode` | number | HTTP status code (200 for success) |
| `message` | string | Response message |
| `data` | array | Array of weather records |
| `count` | number | Number of records returned |
| `period` | string | Requested time period |
| `startTime` | string | Start time of the query range (ISO 8601) |
| `endTime` | string | End time of the query range (ISO 8601) |

#### Weather Record Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique record identifier (City-Timestamp) |
| `timestamp` | string | Weather data timestamp (ISO 8601) |
| `cityName` | string | City name |
| `temperature` | number | Temperature in Celsius |
| `description` | string | Weather condition description |
| `humidity` | number | Humidity percentage |
| `pressure` | number | Atmospheric pressure in hPa |
| `windSpeed` | number | Wind speed in m/s |
| `country` | string | Country code (ISO 3166) |
| `createdAt` | string | Record creation timestamp |
| `ttl` | number | Time to live (Unix timestamp, 30 days from creation) |

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `WEATHER_API_KEY` | OpenWeatherMap API key | - | Yes |
| `CITY_NAME` | City for weather data | Tokyo | No |
| `S3_BUCKET` | S3 bucket name | - | Yes (set by SAM) |
| `DYNAMODB_TABLE` | DynamoDB table name | - | Yes (set by SAM) |
| `AWS_REGION` | AWS region | ap-northeast-1 | No |

### SAM Template Parameters

Configure deployment parameters in `template.yaml`:

- `Environment`: dev, staging, prod
- `WeatherAPIKey`: Your OpenWeatherMap API key
- `CityName`: City name for weather data collection

## 📁 Data Schema

### S3 Storage Format

Weather data is stored in S3 as JSON files with the following structure:

```
s3://bucket-name/weather-data/2024/01-15/Tokyo-1705123456.json
```

### DynamoDB Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` (PK) | String | Unique identifier |
| `timestamp` (SK) | String | ISO 8601 timestamp |
| `cityName` | String | City name |
| `temperature` | Number | Temperature in Celsius |
| `description` | String | Weather description |
| `humidity` | Number | Humidity percentage |
| `pressure` | Number | Atmospheric pressure |
| `windSpeed` | Number | Wind speed |
| `country` | String | Country code |
| `ttl` | Number | Time to live (30 days) |

## 🛡️ Security

- Environment variables are used for sensitive data
- IAM roles with least privilege access
- S3 bucket encryption enabled
- No hardcoded credentials in code
- Security scanning with `make lint`

## 🚨 Troubleshooting

### Common Issues

1. **Build fails**: Run `make deps-docker` or `make deps` to fix Go dependencies
2. **Missing dependencies**: Ensure Docker is running, then run `make deps-docker`
3. **Docker issues**: Check Docker daemon is running with `docker compose ps`
4. **AWS credentials**: Configure AWS credentials with `aws configure` or IAM roles
5. **API key errors**: Set WEATHER_API_KEY environment variable in .env file
6. **AWS permissions**: Check IAM permissions for S3, DynamoDB, Lambda, CloudFormation, and EventBridge
7. **EventBridge unmarshaling errors**: Fixed in latest version - redeploy with `make sam-deploy-dev`
8. **Deployment parameter errors**: Ensure .env file exists with correct WEATHER_API_KEY format
9. **Lambda function not found**: Check function names with `aws lambda list-functions --region ap-northeast-1`
10. **API Gateway CORS errors**: API includes proper CORS headers - check browser network tab

### Recent Fixes (v1.1.0)

- ✅ **EventBridge JSON unmarshaling error**: Lambda function now properly handles EventBridge events using `json.RawMessage`
- ✅ **Makefile environment variable issues**: Fixed deployment commands to properly load .env variables
- ✅ **AWS CLI path issues in containers**: Updated all Makefile commands to use correct shell (`sh` instead of `bash`)
- ✅ **Missing Weather History API testing**: Added comprehensive testing commands and documentation

### Debugging

```bash
# Enable debug logging
export LOG_LEVEL=DEBUG

# Check Docker containers status
docker compose ps

# View container logs
docker compose logs dev
docker compose logs sam

# Test AWS connectivity using Makefile
make test-aws-cli

# Test individual components
make test-lambda-direct          # Test Weather Collection Lambda
make test-history-api           # Test Weather History API

# Check deployed AWS resources
aws cloudformation describe-stacks --stack-name weather-lambda-dev --region ap-northeast-1

# View Lambda function logs (recent entries)
aws logs tail /aws/lambda/weather-lambda-dev-WeatherLambdaFunction --follow --region ap-northeast-1

# Check DynamoDB data
aws dynamodb scan --table-name weather-lambda-dev-weather-records-dev --limit 5 --region ap-northeast-1

# List S3 objects
aws s3 ls s3://weather-lambda-dev-weather-data-dev/weather-data/ --recursive

# Test EventBridge rule
aws events list-rules --region ap-northeast-1 | grep weather

# Manual Lambda invocation for testing
make test-lambda-direct

# Get API Gateway URL
aws cloudformation describe-stacks --stack-name weather-lambda-dev --region ap-northeast-1 \
  --query 'Stacks[0].Outputs[?OutputKey==`WeatherHistoryApiUrl`].OutputValue' --output text
```

### Performance Monitoring

```bash
# Monitor Lambda function metrics
aws logs filter-log-events --log-group-name /aws/lambda/weather-lambda-dev-WeatherLambdaFunction \
  --start-time $(date -d '1 hour ago' +%s)000 --region ap-northeast-1

# Check DynamoDB table metrics
aws dynamodb describe-table --table-name weather-lambda-dev-weather-records-dev --region ap-northeast-1

# Monitor API Gateway usage
aws apigateway get-usage --usage-plan-id YOUR_USAGE_PLAN_ID --key-id YOUR_API_KEY --region ap-northeast-1
```

## 📚 Additional Resources

- [AWS SAM Documentation](https://docs.aws.amazon.com/serverless-application-model/)
- [OpenWeatherMap API](https://openweathermap.org/api)
- [Go AWS SDK](https://aws.amazon.com/sdk-for-go/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [AWS Lambda Go Runtime](https://docs.aws.amazon.com/lambda/latest/dg/lambda-golang.html)
- [AWS API Gateway](https://docs.aws.amazon.com/apigateway/)
- [AWS EventBridge](https://docs.aws.amazon.com/eventbridge/)

## 📋 Release Notes

### v1.1.0 (2025-08-26)
**Major Features:**
- ✅ **Weather History API**: New REST API endpoint for querying historical weather data
- ✅ **Flexible Time Filtering**: Support for 6h, 24h, 1d periods and custom hour ranges (1-168h)
- ✅ **API Gateway Integration**: Complete CORS-enabled REST API with OpenAPI documentation
- ✅ **Comprehensive Testing**: Full test coverage with Makefile automation

**Bug Fixes:**
- ✅ **EventBridge JSON Unmarshaling**: Fixed "cannot unmarshal object into Go struct field Event.detail of type string" error
- ✅ **Makefile Environment Variables**: Fixed deployment commands to properly load .env variables
- ✅ **Docker Container Shell Issues**: Updated all commands to use `sh` instead of `bash`

**Improvements:**
- ✅ **Enhanced Documentation**: Comprehensive API documentation with examples and troubleshooting
- ✅ **Better Error Handling**: Improved Lambda function error handling and logging
- ✅ **Security Enhancements**: Proper IAM roles and secure parameter handling

### v1.0.0 (2025-08-26)
**Initial Release:**
- ✅ **Weather Data Collection**: Automated weather data fetching from OpenWeatherMap API
- ✅ **Dual Storage**: S3 for detailed JSON data, DynamoDB for structured records
- ✅ **EventBridge Scheduling**: Automated hourly execution
- ✅ **Docker Development**: Complete Docker-based development environment
- ✅ **SAM Infrastructure**: CloudFormation-based serverless infrastructure

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.