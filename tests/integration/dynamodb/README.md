# DynamoDB Integration Tests

This directory contains integration tests for the DynamoDB repository adapter.

## Quick Start

```bash
# 1. Start DynamoDB Local
# Note: In WSL, you may need to use 'sudo' or add your user to docker group
docker run -d -p 8000:8000 --name dynamodb-local amazon/dynamodb-local

# 2. Set environment variables
export AWS_ENDPOINT_URL=http://localhost:8000
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export INTEGRATION_TESTS=true

# 3. Run tests
go test -tags=integration -v ./tests/integration/dynamodb/...
```

### WSL Docker Setup

If you get "permission denied" errors in WSL:

**Option 1: Use sudo (quick fix)**
```bash
sudo docker run -d -p 8000:8000 --name dynamodb-local amazon/dynamodb-local
```

**Option 2: Add user to docker group (recommended)**
```bash
# Add your user to docker group
sudo usermod -aG docker $USER

# Log out and log back in, or run:
newgrp docker

# Verify it works
docker ps
```

**Option 3: Start Docker Desktop**
- If using Docker Desktop on Windows, ensure it's running
- Docker Desktop should automatically work with WSL2

## Prerequisites

### Option 1: DynamoDB Local (Recommended for Local Development)

1. **Download DynamoDB Local**:
   
   **Option A: Using Docker (Recommended)**
   ```bash
   docker run -d -p 8000:8000 --name dynamodb-local amazon/dynamodb-local
   ```
   
   Verify it's running:
   ```bash
   docker ps | grep dynamodb-local
   ```
   
   To stop it later:
   ```bash
   docker stop dynamodb-local
   docker rm dynamodb-local
   ```
   
   **Option B: Download from AWS**
   - Download from: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.DownloadingAndRunning.html
   - Run: `java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -port 8000`

2. **Set Environment Variables**:
   ```bash
   export AWS_ENDPOINT_URL=http://localhost:8000
   export AWS_REGION=us-east-1
   export AWS_ACCESS_KEY_ID=test
   export AWS_SECRET_ACCESS_KEY=test
   export INTEGRATION_TESTS=true
   ```

### Option 2: AWS Test Table

1. **Create a Test Table in AWS**:
   - Use AWS Console or CLI to create a table named `ExchangeRatesTest`
   - Configure with the schema from `docs/DYNAMODB_SCHEMA.md`
   - Ensure GSI `BaseCurrencyIndex` is created

2. **Set Environment Variables**:
   ```bash
   export AWS_REGION=us-east-1
   export INTEGRATION_TESTS=true
   # AWS credentials will be loaded from default chain
   ```

## Running Integration Tests

```bash
# Run all integration tests
INTEGRATION_TESTS=true go test -v ./tests/integration/dynamodb/...

# Run with coverage
INTEGRATION_TESTS=true go test -v -cover ./tests/integration/dynamodb/...

# Run specific test
INTEGRATION_TESTS=true go test -v ./tests/integration/dynamodb/... -run TestDynamoDBRepository_Get_Success
```

## Test Coverage

The integration tests cover:

- ✅ `Get()` - Success and not found scenarios
- ✅ `Save()` - Create and update operations
- ✅ `GetByBase()` - Query by base currency (GSI)
- ✅ `Delete()` - Delete operations
- ✅ `GetStale()` - Stale rate retrieval
- ✅ Context cancellation handling
- ✅ TTL handling

## Notes

- Tests create and manage their own test table (`ExchangeRatesTest`)
- Table is created with GSI `BaseCurrencyIndex` for `GetByBase()` tests
- Tests are skipped unless `INTEGRATION_TESTS=true` is set
- Table cleanup is manual (comment out teardown for debugging)

## Troubleshooting

**Issue**: "permission denied while trying to connect to the docker API"
- **Solution (WSL)**: 
  - Option 1: Use `sudo docker run ...`
  - Option 2: Add user to docker group: `sudo usermod -aG docker $USER` then log out/in
  - Option 3: Ensure Docker Desktop is running (if using Docker Desktop on Windows)
  - Option 4: Start Docker service: `sudo service docker start`

**Issue**: Docker command fails with "pull access denied for d" or similar
- **Solution**: Make sure you copy the entire command on one line:
  ```bash
  docker run -d -p 8000:8000 --name dynamodb-local amazon/dynamodb-local
  ```

**Issue**: Tests fail with "table not found"
- **Solution**: Ensure DynamoDB Local is running (`docker ps | grep dynamodb-local`) or AWS credentials are configured

**Issue**: Tests fail with "resource not found"
- **Solution**: Check that GSI `BaseCurrencyIndex` is created with correct schema

**Issue**: Tests timeout
- **Solution**: Increase timeout in `setupTestTable()` or check network connectivity

**Issue**: "Unable to connect to endpoint"
- **Solution**: Verify DynamoDB Local is accessible:
  ```bash
  curl http://localhost:8000
  # Should return XML response
  ```
