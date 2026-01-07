#!/bin/bash
# Deployment script for AWS SAM
# Usage: ./scripts/deploy.sh [environment]
# Environment: dev (default), staging, prod

set -e

ENVIRONMENT=${1:-dev}
STACK_NAME="currenseen-${ENVIRONMENT}"

echo "=========================================="
echo "Deploying Currency Exchange Rate Service"
echo "Environment: $ENVIRONMENT"
echo "Stack Name: $STACK_NAME"
echo "=========================================="

# Validate environment
if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|prod)$ ]]; then
    echo "Error: Invalid environment. Must be dev, staging, or prod"
    exit 1
fi

# Check prerequisites
echo "Checking prerequisites..."
command -v sam >/dev/null 2>&1 || { echo "Error: SAM CLI not found. Install it from https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html"; exit 1; }
command -v aws >/dev/null 2>&1 || { echo "Error: AWS CLI not found. Install it from https://aws.amazon.com/cli/"; exit 1; }

# Check AWS credentials
echo "Checking AWS credentials..."
aws sts get-caller-identity >/dev/null 2>&1 || { echo "Error: AWS credentials not configured"; exit 1; }

# Build
echo "Building application..."
make sam-build

# Validate template
echo "Validating SAM template..."
sam validate

# Deploy
echo "Deploying to AWS..."
if [ "$ENVIRONMENT" == "prod" ]; then
    echo "WARNING: Deploying to PRODUCTION!"
    read -p "Are you sure? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Deployment cancelled"
        exit 1
    fi
    CONFIRM_CHANGESET="--confirm-changeset"
else
    CONFIRM_CHANGESET="--no-confirm-changeset"
fi

sam deploy \
    --stack-name "$STACK_NAME" \
    --parameter-overrides "Environment=$ENVIRONMENT" \
    --capabilities CAPABILITY_IAM \
    --region us-east-1 \
    $CONFIRM_CHANGESET \
    --no-fail-on-empty-changeset

# Get outputs
echo ""
echo "=========================================="
echo "Deployment complete!"
echo "=========================================="
echo "Stack outputs:"
aws cloudformation describe-stacks \
    --stack-name "$STACK_NAME" \
    --query 'Stacks[0].Outputs' \
    --output table

echo ""
echo "API URL:"
aws cloudformation describe-stacks \
    --stack-name "$STACK_NAME" \
    --query 'Stacks[0].Outputs[?OutputKey==`ExchangeRateApi`].OutputValue' \
    --output text

