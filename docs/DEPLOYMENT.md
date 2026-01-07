# Deployment Guide

This guide covers deploying the Currency Exchange Rate Service to AWS using AWS SAM (Serverless Application Model).

## Prerequisites

### Required Tools

1. **AWS CLI** (v2.x recommended)
   - Install: https://aws.amazon.com/cli/
   - Configure: `aws configure`
   - Verify: `aws sts get-caller-identity`

2. **SAM CLI** (v1.x or v2.x)
   - Install: https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html
   - Verify: `sam --version`

3. **Go** (v1.21 or later)
   - Install: https://golang.org/dl/
   - Verify: `go version`

4. **Make** (optional, for convenience)
   - Windows: Use WSL or Git Bash
   - Mac/Linux: Usually pre-installed

### AWS Account Setup

1. **Create AWS Account** (if you don't have one)
   - Sign up at https://aws.amazon.com/

2. **Configure AWS Credentials**
   ```bash
   aws configure
   # Enter your Access Key ID
   # Enter your Secret Access Key
   # Enter default region (e.g., us-east-1)
   # Enter default output format (json)
   ```

3. **Required IAM Permissions**
   Your AWS user/role needs permissions for:
   - CloudFormation (create/update/delete stacks)
   - Lambda (create/update functions)
   - API Gateway (create/update APIs)
   - DynamoDB (create/update tables)
   - Secrets Manager (create/update secrets)
   - CloudWatch Logs (create log groups)
   - IAM (create roles and policies)

   #### Setting IAM Permissions via AWS Console

   **Option 1: Attach AWS Managed Policy (Quick Setup)**

   For development/testing, you can attach the `PowerUserAccess` managed policy which provides most required permissions (except IAM):

   1. Sign in to the AWS Console: https://console.aws.amazon.com/
   2. Navigate to **IAM** service (search "IAM" in the top search bar)
   3. Click **Users** in the left sidebar
   4. Click on your username (or create a new user if needed)
   5. Click the **Permissions** tab
   6. Click **Add permissions** → **Attach policies directly**
   7. Search for and select:
      - `PowerUserAccess` (provides most permissions)
      - `IAMFullAccess` (for creating roles and policies)
   8. Click **Next** → **Add permissions**

   **Option 2: Create Custom Policy (Recommended for Production)**

   For production, create a custom policy with least privilege:

   1. Sign in to the AWS Console: https://console.aws.amazon.com/
   2. Navigate to **IAM** service
   3. Click **Policies** in the left sidebar
   4. Click **Create policy**
   5. Click the **JSON** tab
   6. Paste the following policy (replace `YOUR_ACCOUNT_ID` and `YOUR_REGION`):

   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Sid": "CloudFormationPermissions",
         "Effect": "Allow",
         "Action": [
           "cloudformation:CreateStack",
           "cloudformation:UpdateStack",
           "cloudformation:DeleteStack",
           "cloudformation:DescribeStacks",
           "cloudformation:DescribeStackEvents",
           "cloudformation:DescribeStackResources",
           "cloudformation:GetTemplate",
           "cloudformation:ValidateTemplate",
           "cloudformation:CreateChangeSet",
           "cloudformation:DescribeChangeSet",
           "cloudformation:ExecuteChangeSet",
           "cloudformation:DeleteChangeSet",
           "cloudformation:ListStacks"
         ],
         "Resource": "*"
       },
       {
         "Sid": "LambdaPermissions",
         "Effect": "Allow",
         "Action": [
           "lambda:CreateFunction",
           "lambda:UpdateFunctionCode",
           "lambda:UpdateFunctionConfiguration",
           "lambda:DeleteFunction",
           "lambda:GetFunction",
           "lambda:ListFunctions",
           "lambda:AddPermission",
           "lambda:RemovePermission",
           "lambda:GetPolicy",
           "lambda:TagResource",
           "lambda:UntagResource"
         ],
         "Resource": "*"
       },
       {
         "Sid": "APIGatewayPermissions",
         "Effect": "Allow",
         "Action": [
           "apigateway:GET",
           "apigateway:POST",
           "apigateway:PUT",
           "apigateway:PATCH",
           "apigateway:DELETE"
         ],
         "Resource": "*"
       },
       {
         "Sid": "DynamoDBPermissions",
         "Effect": "Allow",
         "Action": [
           "dynamodb:CreateTable",
           "dynamodb:UpdateTable",
           "dynamodb:DeleteTable",
           "dynamodb:DescribeTable",
           "dynamodb:ListTables",
           "dynamodb:TagResource",
           "dynamodb:UntagResource"
         ],
         "Resource": "*"
       },
       {
         "Sid": "SecretsManagerPermissions",
         "Effect": "Allow",
         "Action": [
           "secretsmanager:CreateSecret",
           "secretsmanager:UpdateSecret",
           "secretsmanager:DeleteSecret",
           "secretsmanager:DescribeSecret",
           "secretsmanager:GetSecretValue",
           "secretsmanager:PutSecretValue",
           "secretsmanager:ListSecrets",
           "secretsmanager:TagResource"
         ],
         "Resource": "*"
       },
       {
         "Sid": "CloudWatchLogsPermissions",
         "Effect": "Allow",
         "Action": [
           "logs:CreateLogGroup",
           "logs:DeleteLogGroup",
           "logs:DescribeLogGroups",
           "logs:PutRetentionPolicy"
         ],
         "Resource": "*"
       },
       {
         "Sid": "IAMPermissions",
         "Effect": "Allow",
         "Action": [
           "iam:CreateRole",
           "iam:DeleteRole",
           "iam:GetRole",
           "iam:AttachRolePolicy",
           "iam:DetachRolePolicy",
           "iam:PutRolePolicy",
           "iam:DeleteRolePolicy",
           "iam:GetRolePolicy",
           "iam:ListRolePolicies",
           "iam:ListAttachedRolePolicies",
           "iam:PassRole",
           "iam:TagRole",
           "iam:UntagRole",
           "iam:ListRoles"
         ],
         "Resource": "*"
       },
       {
         "Sid": "S3PermissionsForSAM",
         "Effect": "Allow",
         "Action": [
           "s3:CreateBucket",
           "s3:GetBucketLocation",
           "s3:PutObject",
           "s3:GetObject",
           "s3:DeleteObject",
           "s3:ListBucket"
         ],
         "Resource": [
           "arn:aws:s3:::sam-bootstrap-*",
           "arn:aws:s3:::sam-bootstrap-*/*"
         ]
       }
     ]
   }
   ```

   7. Click **Next**
   8. Enter a policy name (e.g., `CurrenseenDeploymentPolicy`)
   9. (Optional) Add description: "Permissions for deploying Currency Exchange Rate Service"
   10. Click **Create policy**
   11. Go back to **Users** → Select your user → **Permissions** tab
   12. Click **Add permissions** → **Attach policies directly**
   13. Search for your newly created policy (`CurrenseenDeploymentPolicy`)
   14. Select it and click **Next** → **Add permissions**

   **Option 3: Using AWS CLI (Alternative)**

   If you prefer using AWS CLI, you can create and attach the policy:

   ```bash
   # Create the policy file
   cat > deployment-policy.json << 'EOF'
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "cloudformation:*",
           "lambda:*",
           "apigateway:*",
           "dynamodb:*",
           "secretsmanager:*",
           "logs:*",
           "iam:CreateRole",
           "iam:DeleteRole",
           "iam:GetRole",
           "iam:AttachRolePolicy",
           "iam:DetachRolePolicy",
           "iam:PutRolePolicy",
           "iam:DeleteRolePolicy",
           "iam:GetRolePolicy",
           "iam:ListRolePolicies",
           "iam:ListAttachedRolePolicies",
           "iam:PassRole",
           "s3:CreateBucket",
           "s3:GetBucketLocation",
           "s3:PutObject",
           "s3:GetObject",
           "s3:DeleteObject",
           "s3:ListBucket"
         ],
         "Resource": "*"
       }
     ]
   }
   EOF

   # Create the policy
   aws iam create-policy \
     --policy-name CurrenseenDeploymentPolicy \
     --policy-document file://deployment-policy.json

   # Attach to user (replace YOUR_USERNAME and YOUR_ACCOUNT_ID)
   aws iam attach-user-policy \
     --user-name YOUR_USERNAME \
     --policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/CurrenseenDeploymentPolicy
   ```

   **Verify Permissions**

   After setting up permissions, verify they work:

   ```bash
   # Check your current identity
   aws sts get-caller-identity

   # Test CloudFormation access
   aws cloudformation list-stacks --max-items 1

   # Test Lambda access
   aws lambda list-functions --max-items 1

   # Test DynamoDB access
   aws dynamodb list-tables
   ```

   **Important Notes:**

   - For production, use **Option 2** (Custom Policy) with resource-specific ARNs instead of `*`
   - The S3 permissions are required for SAM to store deployment artifacts
   - If you encounter permission errors during deployment, check CloudWatch Logs for detailed error messages
   - Some operations may require additional permissions (e.g., CloudWatch Events for Lambda triggers)

## Project Structure

```
go-currenseen/
├── cmd/
│   └── lambda/          # Lambda function entry point
├── infrastructure/
│   └── template.yaml    # SAM template
├── scripts/
│   ├── build-lambda.sh  # Build script
│   └── deploy.sh        # Deployment script
└── Makefile             # Build automation
```

## Build Process

### Manual Build

1. **Build Go binary for Lambda:**
   ```bash
   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap ./cmd/lambda
   chmod +x bootstrap
   ```

2. **Build with SAM:**
   ```bash
   sam build
   ```

### Using Makefile

```bash
# Build for SAM deployment
make sam-build

# Validate SAM template
make sam-validate
```

### Using Build Script

```bash
# Linux/Mac
./scripts/build-lambda.sh

# Windows (using Git Bash or WSL)
bash scripts/build-lambda.sh
```

## Deployment

### Environment Variables

The SAM template configures the following environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `TABLE_NAME` | Auto | DynamoDB table name |
| `AWS_REGION` | Auto | AWS region |
| `LOG_LEVEL` | INFO | Log level (DEBUG, INFO, WARN, ERROR) |
| `LOG_FORMAT` | json | Log format (json, text) |
| `CACHE_TTL` | 1h | Cache TTL duration |
| `EXCHANGE_RATE_API_URL` | (default) | External API URL |
| `EXCHANGE_RATE_API_TIMEOUT` | 10 | HTTP timeout in seconds |
| `EXCHANGE_RATE_API_RETRY_ATTEMPTS` | 3 | Retry attempts |
| `CIRCUIT_BREAKER_FAILURE_THRESHOLD` | 5 | Circuit breaker threshold |
| `CIRCUIT_BREAKER_COOLDOWN_SECONDS` | 30 | Cooldown duration |
| `CIRCUIT_BREAKER_SUCCESS_THRESHOLD` | 1 | Success threshold |
| `SECRETS_MANAGER_SECRET_NAME` | Auto | Secrets Manager secret name |
| `SECRETS_MANAGER_ENABLED` | true | Enable Secrets Manager |
| `SECRETS_MANAGER_CACHE_TTL` | 5m | Secret cache TTL |
| `RATE_LIMIT_REQUESTS_PER_MINUTE` | 100 | Rate limit per minute |
| `RATE_LIMIT_BURST_SIZE` | 10 | Burst size |
| `RATE_LIMIT_ENABLED` | true | Enable rate limiting |

### Deployment Methods

#### Method 1: Using Makefile (Recommended)

```bash
# Deploy to dev (default)
make sam-deploy-dev

# Deploy to staging
make sam-deploy-staging

# Deploy to production (requires confirmation)
make sam-deploy-prod
```

#### Method 2: Using Deployment Script

```bash
# Deploy to dev
./scripts/deploy.sh dev

# Deploy to staging
./scripts/deploy.sh staging

# Deploy to production
./scripts/deploy.sh prod
```

#### Method 3: Using SAM CLI Directly

```bash
# Build first
sam build

# Deploy (guided mode - interactive)
sam deploy --guided

# Or deploy with parameters
sam deploy \
  --stack-name currenseen-dev \
  --parameter-overrides Environment=dev \
  --capabilities CAPABILITY_IAM \
  --region us-east-1
```

### Deployment Steps

1. **Build the application:**
   ```bash
   make sam-build
   ```

2. **Validate the template:**
   ```bash
   make sam-validate
   # or
   sam validate
   ```

3. **Deploy:**
   ```bash
   make sam-deploy-dev
   ```

4. **Verify deployment:**
   ```bash
   # Get stack outputs
   aws cloudformation describe-stacks \
     --stack-name currenseen-dev \
     --query 'Stacks[0].Outputs' \
     --output table
   ```

## Post-Deployment

### 1. Set API Key in Secrets Manager

After deployment, you need to set the API key in Secrets Manager:

```bash
# Get the secret ARN from stack outputs
SECRET_ARN=$(aws cloudformation describe-stacks \
  --stack-name currenseen-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ApiKeySecretArn`].OutputValue' \
  --output text)

# Set the API key (replace YOUR_API_KEY with your actual key)
aws secretsmanager put-secret-value \
  --secret-id "$SECRET_ARN" \
  --secret-string '{"api-key":"YOUR_API_KEY_HERE"}'
```

### 2. Test the API

```bash
# Get API URL
API_URL=$(aws cloudformation describe-stacks \
  --stack-name currenseen-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ExchangeRateApi`].OutputValue' \
  --output text)

# Test health endpoint
curl "$API_URL/health"

# Test exchange rate endpoint (requires API key)
curl -H "X-API-Key: YOUR_API_KEY" \
  "$API_URL/rates/USD/EUR"
```

### 3. View Logs

```bash
# Using SAM CLI
make sam-logs

# Or directly
sam logs -n ExchangeRateFunction --stack-name currenseen-dev --tail

# Or using AWS CLI
aws logs tail /aws/lambda/currenseen-dev-ExchangeRateFunction --follow
```

## Environment-Specific Configuration

### Development

- Stack name: `currenseen-dev`
- Environment parameter: `dev`
- Lower rate limits for testing
- More verbose logging

### Staging

- Stack name: `currenseen-staging`
- Environment parameter: `staging`
- Production-like configuration
- Requires confirmation for changes

### Production

- Stack name: `currenseen-prod`
- Environment parameter: `prod`
- Production configuration
- Requires explicit confirmation
- Consider using blue/green deployments

## Infrastructure Components

### Lambda Function

- **Runtime**: `provided.al2023` (custom Go runtime)
- **Handler**: `bootstrap`
- **Memory**: 512 MB
- **Timeout**: 30 seconds
- **Architecture**: x86_64

### API Gateway

- **Type**: REST API
- **Stage**: Environment-specific (dev/staging/prod)
- **Throttling**: 100 requests/second, 200 burst
- **CORS**: Enabled for all origins

### DynamoDB Table

- **Table Name**: `ExchangeRates`
- **Billing**: Pay-per-request
- **TTL**: Enabled (automatic cleanup)
- **GSI**: `BaseCurrencyIndex` for querying by base currency

### Secrets Manager

- **Secret Name**: `{Environment}/currenseen/api-keys`
- **Format**: JSON with `api-key` field
- **Rotation**: Manual (configure rotation if needed)

### CloudWatch Logs

- **Log Group**: `/aws/lambda/{FunctionName}`
- **Retention**: 14 days
- **Format**: JSON (structured logging)

## Troubleshooting

### Build Issues

**Error: "go: command not found"**
- Install Go: https://golang.org/dl/
- Verify: `go version`

**Error: "sam: command not found"**
- Install SAM CLI: https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html
- Verify: `sam --version`

### Deployment Issues

**Error: "Access Denied"**
- Check AWS credentials: `aws sts get-caller-identity`
- Verify IAM permissions
- Ensure you're using the correct AWS account

**Error: "Stack already exists"**
- Update existing stack: `sam deploy --no-confirm-changeset`
- Or delete stack first: `make sam-delete`

**Error: "Template validation failed"**
- Validate template: `sam validate`
- Check YAML syntax
- Verify all required parameters

### Runtime Issues

**Lambda function fails to start**
- Check CloudWatch Logs
- Verify environment variables
- Check IAM permissions

**API Gateway returns 500**
- Check Lambda function logs
- Verify API Gateway integration
- Check request format

**Rate limiting not working**
- Verify `RATE_LIMIT_ENABLED=true`
- Check rate limiter configuration
- Review CloudWatch metrics

## Cleanup

### Delete Stack

```bash
# Using Makefile
make sam-delete

# Or using SAM CLI
sam delete --stack-name currenseen-dev
```

### Manual Cleanup

If automatic cleanup fails, manually delete:
1. Lambda function
2. API Gateway
3. DynamoDB table
4. Secrets Manager secret
5. CloudWatch log groups
6. IAM roles and policies

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - uses: aws-actions/setup-sam@v2
      - uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1
      - run: make sam-build
      - run: make sam-deploy-prod
```

## Best Practices

1. **Always validate before deploying:**
   ```bash
   make sam-validate
   ```

2. **Use environment-specific stacks:**
   - Separate dev, staging, and prod
   - Use different parameter values per environment

3. **Monitor deployments:**
   - Check CloudWatch Logs after deployment
   - Verify API endpoints are working
   - Monitor CloudWatch metrics

4. **Version control:**
   - Commit `template.yaml` to version control
   - Tag releases
   - Document changes

5. **Security:**
   - Never commit API keys
   - Use Secrets Manager for sensitive data
   - Rotate API keys regularly
   - Use least privilege IAM policies

## Additional Resources

- [AWS SAM Documentation](https://docs.aws.amazon.com/serverless-application-model/)
- [Lambda Go Runtime](https://docs.aws.amazon.com/lambda/latest/dg/lambda-golang.html)
- [API Gateway Documentation](https://docs.aws.amazon.com/apigateway/)
- [DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)

