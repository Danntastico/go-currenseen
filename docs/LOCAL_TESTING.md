# Local Testing Guide

This guide explains how to test the Currency Exchange Rate Service locally.

## Testing Options

### 1. Unit Tests (Recommended for Development)

Run unit tests to verify handler logic with mocked dependencies:

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run with coverage
make test-coverage
```

**Pros:**
- Fast execution
- No external dependencies (DynamoDB, AWS)
- Tests handlers directly with mock events

**Cons:**
- Doesn't test real AWS integrations
- Requires mocking all dependencies

---

### 2. Local HTTP Server (Recommended for Integration Testing)

Run a local HTTP server that converts HTTP requests to Lambda events:

```bash
# Set required environment variables
export TABLE_NAME=ExchangeRates
export AWS_REGION=us-east-1
export EXCHANGE_RATE_API_URL=https://api.fawazahmed0.currency-api.com/v1

# Optional: Set port (default: 8080)
export PORT=8080

# Build and run
make run-local-server
```

Then test with HTTP requests:

```bash
# Health check
curl http://localhost:8080/health

# Get single rate
curl http://localhost:8080/rates/USD/EUR

# Get all rates for base currency
curl http://localhost:8080/rates/USD
```

**Pros:**
- Tests real HTTP requests
- Uses actual AWS SDK (requires AWS credentials)
- Tests full request/response cycle
- Easy to test with curl or Postman

**Cons:**
- Requires AWS credentials configured
- Requires DynamoDB table (or local DynamoDB)
- Requires external API access

**Requirements:**
- AWS credentials configured (`~/.aws/credentials` or environment variables)
- DynamoDB table exists (or use [DynamoDB Local](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html))
- Network access to external currency API

---

### 3. SAM Local (Recommended for Lambda-Specific Testing)

Run Lambda functions locally using AWS SAM CLI:

```bash
# Prerequisites: Install SAM CLI and Docker
# https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html

# Build for SAM
make sam-build

# Run SAM locally (starts API Gateway locally)
make sam-local
```

SAM Local will:
- Start a local API Gateway on `http://localhost:3000`
- Run Lambda functions in Docker containers
- Simulate AWS Lambda runtime environment

**Pros:**
- Closest to production Lambda environment
- Tests API Gateway integration
- Tests Lambda cold starts and warm starts

**Cons:**
- Requires Docker
- Requires SAM CLI installation
- Slower than unit tests

**Requirements:**
- [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)
- Docker Desktop (or Docker daemon)
- AWS credentials configured

---

## Environment Variables

### Required Variables

```bash
TABLE_NAME=ExchangeRates              # DynamoDB table name
```

### Optional Variables

```bash
AWS_REGION=us-east-1                  # AWS region (default: from AWS config)
CACHE_TTL=1h                          # Cache TTL (default: 1 hour)
EXCHANGE_RATE_API_URL=https://...     # External API URL
EXCHANGE_RATE_API_TIMEOUT=10           # HTTP timeout in seconds
EXCHANGE_RATE_API_RETRY_ATTEMPTS=3     # Retry attempts
CIRCUIT_BREAKER_FAILURE_THRESHOLD=5    # Circuit breaker threshold
CIRCUIT_BREAKER_COOLDOWN_SECONDS=30    # Circuit breaker cooldown
CIRCUIT_BREAKER_SUCCESS_THRESHOLD=1    # Circuit breaker success threshold
SECRETS_MANAGER_ENABLED=false          # Enable Secrets Manager
PORT=8080                              # Local server port (local server only)
```

### Example `.env` File

Create a `.env` file (not committed to git):

```bash
TABLE_NAME=ExchangeRates
AWS_REGION=us-east-1
CACHE_TTL=1h
EXCHANGE_RATE_API_URL=https://api.fawazahmed0.currency-api.com/v1
PORT=8080
```

Load with:

```bash
# Linux/Mac
export $(cat .env | xargs)
make run-local-server

# Windows PowerShell
Get-Content .env | ForEach-Object { $name, $value = $_ -split '=', 2; [Environment]::SetEnvironmentVariable($name, $value) }
make run-local-server
```

---

## Testing with DynamoDB Local

For local testing without AWS, use [DynamoDB Local](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html):

```bash
# Start DynamoDB Local (requires Java)
docker run -p 8000:8000 amazon/dynamodb-local

# Set endpoint override
export AWS_ENDPOINT_URL=http://localhost:8000
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_REGION=us-east-1

# Create table (use AWS CLI or SDK)
aws dynamodb create-table \
  --endpoint-url http://localhost:8000 \
  --table-name ExchangeRates \
  --attribute-definitions \
    AttributeName=PK,AttributeType=S \
    AttributeName=BaseCurrency,AttributeType=S \
  --key-schema \
    AttributeName=PK,KeyType=HASH \
  --global-secondary-indexes \
    IndexName=BaseCurrencyIndex,KeySchema=[{AttributeName=BaseCurrency,KeyType=HASH},{AttributeName=PK,KeyType=RANGE}],Projection={ProjectionType=ALL} \
  --billing-mode PAY_PER_REQUEST
```

**Note:** The current code doesn't support endpoint override. You'd need to modify `config.NewDynamoDBClient()` to support custom endpoints for local development.

---

## Quick Start

### Option 1: Unit Tests (No AWS Required)

```bash
make test
```

### Option 2: Local HTTP Server (Requires AWS Credentials)

```bash
# Set environment variables
export TABLE_NAME=ExchangeRates
export AWS_REGION=us-east-1

# Run server
make run-local-server

# Test in another terminal
curl http://localhost:8080/health
curl http://localhost:8080/rates/USD/EUR
```

### Option 3: SAM Local (Requires Docker + SAM CLI)

```bash
make sam-local

# Test in another terminal
curl http://localhost:3000/health
curl http://localhost:3000/rates/USD/EUR
```

---

## Troubleshooting

### Issue: "Failed to create DynamoDB client"

**Solution:** Ensure AWS credentials are configured:
```bash
aws configure
# Or set environment variables:
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
export AWS_REGION=us-east-1
```

### Issue: "Table not found"

**Solution:** Create the DynamoDB table in AWS (or use DynamoDB Local):
```bash
aws dynamodb create-table --table-name ExchangeRates ...
```

### Issue: "Connection refused" when using local server

**Solution:** Check that the server is running and the port is correct:
```bash
# Check if port is in use
netstat -an | grep 8080  # Linux/Mac
netstat -an | findstr 8080  # Windows
```

### Issue: SAM Local fails to start

**Solution:** 
- Ensure Docker is running
- Check SAM CLI is installed: `sam --version`
- Verify `template.yaml` is valid: `sam validate`

---

## Next Steps

- **Phase 9: Logging & Observability** - Add structured logging for better debugging
- **Integration Tests** - Create comprehensive integration test suite
- **Docker Compose** - Set up local development environment with DynamoDB Local
