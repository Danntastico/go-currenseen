# DynamoDB Table Schema Design

## Table: ExchangeRates

### Schema Diagram

```mermaid
erDiagram
    ExchangeRates {
        string PK "Partition Key: RATE#{BASE}#{TARGET}"
        string Base "Base currency code"
        string Target "Target currency code"
        number Rate "Exchange rate value"
        number Timestamp "Unix timestamp (seconds)"
        boolean Stale "Stale flag"
        number ttl "TTL timestamp (Unix seconds)"
    }
    
    BaseCurrencyIndex {
        string Base "GSI Partition Key"
        string PK "Original partition key"
        string Target "Target currency code"
        number Rate "Exchange rate value"
        number Timestamp "Unix timestamp (seconds)"
        boolean Stale "Stale flag"
        number ttl "TTL timestamp (Unix seconds)"
    }
    
    ExchangeRates ||--o{ BaseCurrencyIndex : "indexed by"
```

### Primary Key
- **Partition Key (PK)**: `RATE#{BASE}#{TARGET}` (String)
  - Format: `RATE#USD#EUR`, `RATE#GBP#JPY`, etc.
  - Purpose: Enables direct lookup for `Get()` and `Delete()` operations
  - Example: `"RATE#USD#EUR"`

- **Sort Key (SK)**: None (Simple Primary Key)
  - Not needed for Phase 3
  - Can be added later for versioning or time-based ordering

### Attributes

| Attribute Name | DynamoDB Type | Description | Example Value |
|----------------|--------------|-------------|---------------|
| `PK` | String | Partition key (primary key) | `"RATE#USD#EUR"` |
| `Base` | String | Base currency code (ISO 4217) | `"USD"` |
| `Target` | String | Target currency code (ISO 4217) | `"EUR"` |
| `Rate` | Number | Exchange rate value | `0.85` |
| `Timestamp` | Number | Unix timestamp in seconds | `1704067200` |
| `Stale` | Boolean | Whether rate is marked as stale | `false` |
| `ttl` | Number | TTL timestamp (Unix epoch in seconds) | `1704153600` |

**Notes:**
- `Timestamp`: Stored as Unix timestamp (seconds since epoch)
- `ttl`: DynamoDB TTL attribute - items are automatically deleted when TTL expires
- `Base` and `Target`: Stored separately for GSI queries and readability

### Global Secondary Index (GSI)

**Index Name**: `BaseCurrencyIndex`

- **Partition Key**: `Base` (String)
- **Sort Key**: None
- **Projection**: ALL (all attributes projected to GSI)

**Purpose**: Enables efficient `GetByBase()` queries without scanning the entire table.

**Usage**: Query all exchange rates for a specific base currency.

### Example Items

#### Item 1: USD to EUR
```json
{
  "PK": "RATE#USD#EUR",
  "Base": "USD",
  "Target": "EUR",
  "Rate": 0.85,
  "Timestamp": 1704067200,
  "Stale": false,
  "ttl": 1704153600
}
```

#### Item 2: USD to GBP
```json
{
  "PK": "RATE#USD#GBP",
  "Base": "USD",
  "Target": "GBP",
  "Rate": 0.75,
  "Timestamp": 1704067200,
  "Stale": false,
  "ttl": 1704153600
}
```

#### Item 3: EUR to GBP
```json
{
  "PK": "RATE#EUR#GBP",
  "Base": "EUR",
  "Target": "GBP",
  "Rate": 0.88,
  "Timestamp": 1704067200,
  "Stale": false,
  "ttl": 1704153600
}
```

### Access Patterns

| Operation | DynamoDB Operation | Key Used | Index Used |
|-----------|-------------------|---------|------------|
| `Get(base, target)` | `GetItem` | `PK = RATE#{base}#{target}` | Primary table |
| `GetByBase(base)` | `Query` | `Base = {base}` | `BaseCurrencyIndex` (GSI) |
| `Save(rate, ttl)` | `PutItem` | `PK = RATE#{base}#{target}` | Primary table |
| `Delete(base, target)` | `DeleteItem` | `PK = RATE#{base}#{target}` | Primary table |
| `GetStale(base, target)` | `GetItem` | `PK = RATE#{base}#{target}` | Primary table |

### TTL Management

- **TTL Attribute**: `ttl` (Number)
- **Format**: Unix timestamp in seconds
- **Calculation**: `time.Now().Add(ttl).Unix()`
- **Automatic Deletion**: DynamoDB automatically deletes items when `ttl` timestamp is reached
- **Note**: TTL deletion is eventually consistent (may take up to 48 hours)

### Design Decisions

1. **Partition Key Format**: `RATE#{BASE}#{TARGET}`
   - Enables direct lookup for specific currency pairs
   - Unique identifier for each exchange rate
   - Common DynamoDB pattern for composite keys

2. **GSI on Base**: 
   - Enables efficient queries for all rates by base currency
   - Avoids expensive table scans
   - Projection: ALL to avoid additional GetItem calls
   - **Reserved Keyword Handling**: `Base` is a DynamoDB reserved keyword, so queries use `ExpressionAttributeNames` to escape it (e.g., `#base = :base` with `{"#base": "Base"}`)

3. **Separate Base/Target Attributes**:
   - Required for GSI (Base is GSI partition key)
   - Improves query readability
   - Allows filtering/validation

4. **TTL as Separate Attribute**:
   - Leverages DynamoDB's built-in TTL feature
   - Automatic cleanup of expired items
   - Reduces storage costs

---

**Last Updated**: Phase 3 - Step 15  
**Status**: Complete - All documentation updated
