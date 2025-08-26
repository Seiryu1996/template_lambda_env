# Weather Lambda Function

Go-based AWS Lambda function that collects weather data and stores it in S3 and DynamoDB using SAM (Serverless Application Model) and Docker for development.

## 🏗️ Architecture

This project implements a serverless weather data collection system with the following components:

- **Weather Collection Lambda**: Go-based function that fetches weather data from OpenWeatherMap API
- **Weather History API**: REST API to query historical weather data with time-based filtering
- **S3 Storage**: Stores detailed weather data as JSON files
- **DynamoDB**: Records weather metrics for fast querying
- **API Gateway**: RESTful API endpoints for accessing weather history
- **CloudFormation/SAM**: Infrastructure as Code for AWS resources
- **Docker**: Development environment with all necessary tools

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
├── scripts/                 # Build and deployment scripts
├── tests/                   # Integration tests
├── docker-compose.yml       # Development environment
├── Dockerfile               # Production Lambda image
├── Dockerfile.dev           # Development image
├── template.yaml            # SAM CloudFormation template
├── Makefile                # Build automation
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

```bash
make help                    # Show all available commands
make deps                    # Tidy and download dependencies (Go 1.23+)
make deps-docker             # Tidy dependencies using Docker
make build                   # Build Lambda binary (Go 1.23+)
make build-docker            # Build Lambda binary using Docker
make build-history           # Build Weather History API Lambda binary
make build-history-docker    # Build Weather History API Lambda using Docker
make build-all               # Build all Lambda functions using Docker
make test                    # Run tests
make lint                    # Run linter
make clean                   # Clean build artifacts
make docker-build            # Build with Docker
make docker-dev              # Start development environment
make sam-build               # Build with SAM
make sam-invoke-aws          # Test deployed Lambda function
make test-history-api        # Test the Weather History API endpoint
make sam-deploy              # Deploy to AWS (guided)
make aws-verify              # Verify AWS credentials
make deploy-checklist        # Show pre-deployment checklist
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

### Weather History API Testing

```bash
# Get API URL from CloudFormation stack
API_URL=$(aws cloudformation describe-stacks --stack-name weather-lambda-dev --region ap-northeast-1 --query 'Stacks[0].Outputs[?OutputKey==`WeatherHistoryApiUrl`].OutputValue' --output text)

# Test different time periods
curl -s "$API_URL?period=6h"   # Last 6 hours
curl -s "$API_URL?period=24h"  # Last 24 hours 
curl -s "$API_URL?period=12"   # Last 12 hours (custom)

# Test with different city (if data available)
curl -s "$API_URL?period=6h&city=Tokyo"

# Using Makefile helper
make test-history-api
```

### API Response Format

```json
{
  "statusCode": 200,
  "message": "Weather history retrieved successfully",
  "data": [
    {
      "id": "Tokyo-1756198688",
      "timestamp": "2025-08-26T08:58:08Z",
      "cityName": "Tokyo",
      "temperature": 31.3,
      "description": "few clouds",
      "humidity": 59,
      "pressure": 1009,
      "windSpeed": 8.2,
      "country": "JP",
      "createdAt": "2025-08-26T08:58:08.614818389Z",
      "ttl": 1758790688
    }
  ],
  "count": 1,
  "period": "6h",
  "startTime": "2025-08-26T03:25:47Z",
  "endTime": "2025-08-26T09:25:47Z"
}
```
```

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

1. **Build fails**: Run `make deps` or `make deps-docker` to fix Go dependencies
2. **Missing dependencies**: Ensure Go 1.23+ and run `go mod tidy`
3. **Docker issues**: Check Docker daemon is running
4. **AWS credentials**: Configure AWS credentials with proper permissions
5. **API key errors**: Set WEATHER_API_KEY environment variable or in .env file
6. **AWS permissions**: Check IAM permissions for S3, DynamoDB, Lambda, and CloudFormation

### Debugging

```bash
# Enable debug logging
export LOG_LEVEL=DEBUG

# Check Docker containers
docker compose ps

# View container logs
docker compose logs dev
docker compose logs sam

# Test AWS connectivity
aws sts get-caller-identity

# Check deployed resources
aws cloudformation describe-stacks --stack-name weather-lambda-dev
```

## 📚 Additional Resources

- [AWS SAM Documentation](https://docs.aws.amazon.com/serverless-application-model/)
- [OpenWeatherMap API](https://openweathermap.org/api)
- [Go AWS SDK](https://aws.amazon.com/sdk-for-go/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.