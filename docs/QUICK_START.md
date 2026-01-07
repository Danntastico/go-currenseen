# Quick Start Guide

Get the Currency Exchange Rate Service up and running quickly.

## Prerequisites

- AWS CLI configured
- SAM CLI installed
- Go 1.21+

## Quick Deploy

```bash
# 1. Build
make sam-build

# 2. Deploy to dev
make sam-deploy-dev

# 3. Get API URL
aws cloudformation describe-stacks \
  --stack-name currenseen-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ExchangeRateApi`].OutputValue' \
  --output text

# 4. Set API key (replace YOUR_KEY)
SECRET_ARN=$(aws cloudformation describe-stacks \
  --stack-name currenseen-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ApiKeySecretArn`].OutputValue' \
  --output text)

aws secretsmanager put-secret-value \
  --secret-id "$SECRET_ARN" \
  --secret-string '{"api-key":"YOUR_KEY_HERE"}'

# 5. Test
API_URL=$(aws cloudformation describe-stacks \
  --stack-name currenseen-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ExchangeRateApi`].OutputValue' \
  --output text)

curl -H "X-API-Key: YOUR_KEY_HERE" "$API_URL/rates/USD/EUR"
```

## Common Commands

```bash
# Build
make sam-build

# Validate
make sam-validate

# Deploy
make sam-deploy-dev

# View logs
make sam-logs

# Delete stack
make sam-delete
```

For detailed information, see [DEPLOYMENT.md](./DEPLOYMENT.md).

