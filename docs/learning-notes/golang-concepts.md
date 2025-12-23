# Go Concepts Notebook

## Value Receivers vs Pointer Receivers

### Value Receiver
**Syntax:** `func (t Type) Method()`
**Definition:** Method receives a copy of the value. Use for:
- Small types (primitives, small structs, type aliases like `string`)
- Read-only methods (getters) that don't modify state
- When copying is cheap (e.g., `CurrencyCode` is ~16 bytes)

**Example:**
```go
func (c CurrencyCode) IsValid() bool {
    return currencyCodeRegex.MatchString(string(c))
}
```

**Benefits:**
- Works on both values and pointers automatically
- No mutation concerns
- Idiomatic for small, immutable types

### Pointer Receiver
**Syntax:** `func (t *Type) Method()`
**Definition:** Method receives a pointer to the value. Use for:
- Large structs (to avoid copying overhead)
- Methods that need to mutate the receiver
- Consistency when some methods need pointers

**Example:**
```go
func (e *ExchangeRate) IsExpired(ttl time.Duration) bool {
    expirationTime := e.Timestamp.Add(ttl)
    return time.Now().After(expirationTime)
}
```

**Rule of thumb:** If any method needs a pointer receiver, use pointer receivers for all methods on that type for consistency.

---

## Type Alias vs Struct

### Type Alias
**Syntax:** `type NewType ExistingType`
**Definition:** Creates a new name for an existing type. No additional fields, same memory layout.

**Example:**
```go
type CurrencyCode string
```

**Characteristics:**
- Underlying type remains the same (e.g., `string`)
- No additional memory overhead
- Use for: type safety, validation, domain modeling of primitives
- Can be converted back to underlying type: `string(code)`

### Struct
**Syntax:** `type StructName struct { Field Type }`
**Definition:** Composite type that groups related data fields together.

**Example:**
```go
type ExchangeRate struct {
    Base      CurrencyCode
    Target    CurrencyCode
    Rate      float64
    Timestamp time.Time
    Stale     bool
}
```

**Characteristics:**
- Multiple fields of different types
- Memory size = sum of all fields
- Use for: modeling complex entities with multiple attributes
- Zero value: all fields set to their zero values

**Key Difference:** Type alias = new name for existing type. Struct = new composite type with multiple fields.

---

## If Statement with Initialization

**Syntax:** `if init; condition { }`
**Definition:** Combines variable declaration/assignment with conditional check in a single `if` statement.

**Pattern:**
```go
if err := functionCall(); err != nil {
    return nil, err
}
```

**Components:**
1. **Initialization:** `err := functionCall()` - declares and assigns
2. **Semicolon:** `;` - separates init from condition
3. **Condition:** `err != nil` - checks the error

**Why use it:**
- Keeps error variable scoped to the `if` block
- Prevents `err` from leaking into outer scope
- Idiomatic Go error handling pattern
- Cleaner than separate declaration and check

**Equivalent to:**
```go
err := functionCall()
if err != nil {
    return nil, err
}
```

**Scope:** Variables declared in initialization are only available within the `if` block and its `else` branches.

---

## Constructor Pattern

**Pattern Name:** Constructor/Factory Pattern

**Naming Convention:** Constructor functions start with `New` followed by the type name: `NewTypeName(...)`

**Syntax Patterns:**
```go
// Pattern 1: Returns pointer
func NewType(...params) (*Type, error) {
    return &Type{...}, nil
}

// Pattern 2: Returns value
func NewType(...params) (Type, error) {
    return Type{...}, nil
}
```

**Components:**
1. **Function name:** `NewTypeName` - follows Go naming convention
2. **Return type:** Can be `(*Type, error)` or `(Type, error)` depending on type
3. **Address-of operator:** `&Type{...}` - used when returning pointer

### When to Return Value vs Pointer

**Return value `(Type, error)` when:**
- Type is small (primitives, small structs, type aliases like `string`)
- Type is immutable or doesn't need mutation
- Methods use value receivers
- Example: `func NewCurrencyCode(code string) (CurrencyCode, error)`

**Return pointer `(*Type, error)` when:**
- Type is large (structs with many fields)
- Methods use pointer receivers
- Need to return `nil` to indicate failure
- Need mutability
- Example: `func NewExchangeRate(...) (*ExchangeRate, error)`

**Why return pointers:**
1. **Consistency:** If methods use pointer receivers `(e *ExchangeRate)`, constructor should return pointer
2. **Efficiency:** Avoids copying large structs when passing around
3. **Mutability:** Allows methods to modify the struct if needed
4. **Nil checks:** Can return `nil` to indicate failure (along with error)

**Example:**
```go
func NewExchangeRate(base, target CurrencyCode, rate float64, timestamp time.Time) (*ExchangeRate, error) {
    if err := validateExchangeRate(base, target, rate, timestamp); err != nil {
        return nil, err
    }
    return &ExchangeRate{
        Base:      base,
        Target:    target,
        Rate:      rate,
        Timestamp: timestamp,
        Stale:     false,
    }, nil
}
```

**Memory behavior (pointer return):**
- `&ExchangeRate{...}` creates struct in memory and returns its address
- Caller receives pointer (reference), not a copy
- Methods with pointer receivers can modify the original struct

**Memory behavior (value return):**
- `CurrencyCode{...}` creates value directly
- Caller receives a copy of the value
- Methods with value receivers work on copies

**Summary:**
- **Naming:** Always start with `New` followed by type name
- **Return type:** Choose value or pointer based on type characteristics
- **Pattern:** `func NewTypeName(...) (ReturnType, error)` where `ReturnType` is `Type` or `*Type`

---

## Pointer Operators: Address-of (&) vs Dereference (*)

### Address-of Operator: `&`
**Syntax:** `&value`
**Name:** Address-of operator
**Definition:** Returns the memory address (pointer) of a value.

**Example:**
```go
rate := ExchangeRate{...}  // rate is ExchangeRate (value)
ptr := &rate               // ptr is *ExchangeRate (pointer to rate)
```

**Use:** Get a pointer to a value (create a reference).

### Dereference Operator: `*`
**Syntax:** `*pointer`
**Name:** Dereference operator (also called indirection operator)
**Definition:** Gets the value that a pointer points to.

**Example:**
```go
ptr := &ExchangeRate{...}  // ptr is *ExchangeRate (pointer)
value := *ptr              // value is ExchangeRate (dereferenced value)
```

**Use:** Access the value stored at a memory address.

### Pointer Type: `*Type`
**Syntax:** `*Type` (in type declarations)
**Name:** Pointer type or "pointer to Type"
**Definition:** Declares a type that is a pointer/reference to another type.

**Example:**
```go
func NewExchangeRate(...) (*ExchangeRate, error)
//                        ^^^^^^^^^^^^^^
//                        Pointer type: "pointer to ExchangeRate"
```

**Use:** Declare that a variable, parameter, or return value is a pointer.

### Summary
- `&value` = Address-of operator (gets the address/memory location of a value)
- `*pointer` = Dereference operator (gets the value at an address)
- `*Type` = Pointer type (declares a type that points to Type)

**Relationship:**
```go
value := ExchangeRate{...}  // value
ptr := &value               // get address: ptr is *ExchangeRate
backToValue := *ptr         // dereference: backToValue is ExchangeRate
```

---

## Method Receiver Pattern

**Syntax:** `func (receiver Type) MethodName(params) returnType`
**Definition:** Methods are functions with a receiver that associate the function with a type. The receiver appears between `func` and the method name.

### Components
1. **Receiver:** `(e *ExchangeRate)` - defines which type the method belongs to
2. **Receiver variable:** `e` - name used inside the method to access the receiver
3. **Receiver type:** `*ExchangeRate` - the type this method is associated with

**Example:**
```go
func (e *ExchangeRate) IsExpired(ttl time.Duration) bool {
    expirationTime := e.Timestamp.Add(ttl)
    return time.Now().After(expirationTime)
}
```

### Calling Methods
**Syntax:** `instance.MethodName(args)`

**Example:**
```go
rate := &ExchangeRate{
    Base:      CurrencyCode("USD"),
    Target:    CurrencyCode("EUR"),
    Rate:      1.10,
    Timestamp: time.Now(),
}

expired := rate.IsExpired(5 * time.Minute)
//        ^^^^ instance  ^^^^^^^^^^^^^^^^ method call
```

**Key point:** Call on an instance, not the type directly. `rate.IsExpired(ttl)`, not `ExchangeRate.IsExpired(ttl)`.

### Pointer Receiver Behavior
When a method has a pointer receiver `(e *ExchangeRate)`, Go automatically allows calling it on both pointers and values:

```go
ratePtr := &ExchangeRate{...}  // pointer
ratePtr.IsExpired(ttl)         // ✅ works directly

rateValue := ExchangeRate{...}  // value
rateValue.IsExpired(ttl)        // ✅ also works! Go auto-converts to (&rateValue).IsExpired()
```

**Automatic conversion:** Go automatically takes the address when calling a pointer receiver method on a value.

### Method Expression (Advanced)
**Syntax:** `Type.MethodName` creates a function value where the receiver becomes the first parameter.

**Example:**
```go
// Method expression - creates a function value
methodFunc := (*ExchangeRate).IsExpired
expired := methodFunc(ratePtr, ttl)  // receiver passed as first argument
```

**Use case:** Less common, used when you need to pass the method as a function value.

### Summary
- **Method:** Function with a receiver `func (receiver Type) Method()`
- **Receiver:** `(e *ExchangeRate)` - associates method with type
- **Call syntax:** `instance.MethodName(args)` - call on instance, not type
- **Pointer receiver:** Works on both pointers and values (automatic conversion)
- **Receiver variable:** `e` - used inside method to access the receiver's fields/methods

---

## Maps

**Definition:** Maps are key-value collections in Go. They provide fast lookup by key.

**Syntax:** `map[KeyType]ValueType`

**Example:**
```go
map[entity.CurrencyCode]*entity.ExchangeRate
```

**Components:**
1. **`map`** - Go's map type keyword
2. **`[KeyType]`** - Key type in square brackets (must be comparable: string, int, struct with comparable fields, etc.)
3. **`ValueType`** - Value type (can be any type)

### Map Declaration and Initialization

**Declaration:**
```go
var rates map[CurrencyCode]*ExchangeRate  // nil map (cannot be used until initialized)
```

**Initialization:**
```go
// Method 1: Using make
rates := make(map[CurrencyCode]*ExchangeRate)

// Method 2: Using map literal
rates := map[CurrencyCode]*ExchangeRate{
    CurrencyCode("EUR"): &ExchangeRate{...},
    CurrencyCode("GBP"): &ExchangeRate{...},
}

// Method 3: Empty map literal
rates := map[CurrencyCode]*ExchangeRate{}
```

### Map Operations

**Accessing values:**
```go
rate := rates[CurrencyCode("EUR")]  // Returns *ExchangeRate or nil if key doesn't exist
```

**Setting values:**
```go
rates[CurrencyCode("EUR")] = &ExchangeRate{...}
```

**Checking if key exists:**
```go
rate, exists := rates[CurrencyCode("EUR")]
if exists {
    // Key exists, rate is *ExchangeRate
} else {
    // Key doesn't exist, rate is nil
}
```

**Deleting keys:**
```go
delete(rates, CurrencyCode("EUR"))
```

**Iterating:**
```go
for key, value := range rates {
    // key is CurrencyCode
    // value is *ExchangeRate
}
```

### Why Use Pointer Types as Map Values?

**Use pointer types `*Type` instead of value types `Type` when:**

1. **Efficiency:** Avoids copying large structs when storing/retrieving from map
   ```go
   // ❌ Inefficient: copies entire struct
   map[CurrencyCode]ExchangeRate
   
   // ✅ Efficient: stores pointer (8 bytes) instead of full struct
   map[CurrencyCode]*ExchangeRate
   ```

2. **Consistency:** Matches return types from constructors and other functions
   ```go
   // Constructor returns pointer
   func NewExchangeRate(...) (*ExchangeRate, error)
   
   // Map should store pointers for consistency
   map[CurrencyCode]*ExchangeRate
   ```

3. **Mutability:** Allows modifying the struct through the map reference
   ```go
   rate := rates[CurrencyCode("EUR")]
   rate.Stale = true  // Modifies the original struct
   ```

4. **Nil handling:** Can distinguish between "key doesn't exist" and "key exists with nil value"
   ```go
   rate, exists := rates[key]
   if !exists {
       // Key not in map
   } else if rate == nil {
       // Key exists but value is nil
   }
   ```

5. **Memory efficiency:** Multiple map entries can reference the same struct instance if needed

**When to use value types in maps:**
- Small types (primitives, small structs)
- Immutable types
- When you want independent copies

**Example from codebase:**
```go
FetchAllRates(ctx context.Context, base entity.CurrencyCode) (map[entity.CurrencyCode]*entity.ExchangeRate, error)
//                                                                  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//                                                                  Keys: CurrencyCode, Values: *ExchangeRate (pointers)
```

### Map Characteristics

- **Zero value:** `nil` - cannot be used until initialized with `make()` or literal
- **Key requirement:** Keys must be comparable (no slices, maps, or functions as keys)
- **Reference type:** Maps are reference types (like slices and channels)
- **Not thread-safe:** Use `sync.Map` or mutexes for concurrent access
- **Order:** Iteration order is random (not guaranteed)

### Summary
- **Syntax:** `map[KeyType]ValueType` - key type in brackets, value type after
- **Use pointer values:** For large structs, consistency, mutability, and efficiency
- **Operations:** Access with `map[key]`, check existence with `value, ok := map[key]`
- **Initialization:** Use `make()` or map literal, never use nil map
- **Keys:** Must be comparable types (string, int, struct with comparable fields)

---

## For Loops

**Definition:** Go has one loop construct: `for`. It supports three forms: traditional C-style, range-based, and infinite.

### 1. Traditional C-style For Loop

**Syntax:** `for init; condition; post { }`

**Example:**
```go
for i := 0; i < 10; i++ {
    fmt.Println(i)
}
```

**Components:**
- **Init:** `i := 0` - executed once before first iteration
- **Condition:** `i < 10` - checked before each iteration
- **Post:** `i++` - executed after each iteration

**All parts optional:**
```go
// Infinite loop (all parts omitted)
for {
    // break or return to exit
}

// While-style loop (init omitted)
i := 0
for i < 10 {
    i++
}
```

### 2. For-Range Loop (Most Common)

**Syntax:** `for key, value := range collection { }`

**Used for:** Slices, arrays, maps, strings, channels

#### Range over Slices/Arrays

**Full syntax:**
```go
// With index and value
for i, rate := range rates {
    // i is index (int)
    // rate is *entity.ExchangeRate
}

// Ignore index (use blank identifier)
for _, rate := range rates {
    // rate is *entity.ExchangeRate
    // _ discards the index
}

// Index only
for i := range rates {
    // i is index (int)
}
```

**Example from codebase:**
```go
for _, rate := range rates {
    if rate == nil {
        continue
    }
    // Process rate
}
```

#### Range over Maps

**Full syntax:**
```go
// With key and value
for key, rate := range ratesMap {
    // key is CurrencyCode
    // rate is *entity.ExchangeRate
}

// Ignore key (use blank identifier)
for _, rate := range ratesMap {
    // rate is *entity.ExchangeRate
    // _ discards the key
}

// Key only
for key := range ratesMap {
    // key is CurrencyCode
}
```

**Important:** Map iteration order is random (not guaranteed).

#### Range over Strings

```go
// Returns rune (Unicode code point) and byte index
for i, char := range "Hello" {
    // i is byte index (int)
    // char is rune (int32)
}
```

#### Range over Channels

```go
for value := range channel {
    // Receives values from channel until it's closed
}
```

### 3. Infinite Loop

**Syntax:** `for { }`

**Example:**
```go
for {
    // Loop forever
    if condition {
        break  // Exit loop
    }
}
```

### Special Considerations

#### Blank Identifier (`_`)

Use `_` to ignore values you don't need:

```go
// Ignore index
for _, rate := range rates {
    // Only use rate, not index
}

// Ignore value
for key := range ratesMap {
    // Only use key, not value
}
```

#### Break and Continue

**`break`** - Exits the loop immediately:
```go
for i, rate := range rates {
    if rate == nil {
        break  // Exit loop
    }
}
```

**`continue`** - Skips to next iteration:
```go
for _, rate := range rates {
    if rate == nil {
        continue  // Skip this iteration, go to next
    }
    // Process rate
}
```

#### Modifying During Iteration

**Slices/Arrays:**
- Modifying values: ✅ Safe (modifies the element)
- Modifying slice length: ⚠️ Can cause issues (use indices carefully)

**Maps:**
- Modifying values: ✅ Safe
- Adding/deleting keys: ⚠️ Behavior is undefined (don't do it)

**Safe pattern:**
```go
// Collect keys first, then modify
keys := make([]CurrencyCode, 0, len(ratesMap))
for key := range ratesMap {
    keys = append(keys, key)
}
// Now safe to modify map using keys slice
```

#### Range Copies Values

**Important:** Range creates copies of values, not references:

```go
// For value types
for _, rate := range rates {
    rate.Stale = true  // ❌ Doesn't modify original (rate is a copy)
}

// For pointer types
for _, rate := range rates {
    rate.Stale = true  // ✅ Modifies original (rate is pointer, copy of pointer still points to same struct)
}
```

#### Performance Considerations

1. **Pre-allocate slices when possible:**
   ```go
   result := make([]Type, 0, len(input))  // Capacity hint
   for _, item := range input {
       result = append(result, item)
   }
   ```

2. **Use indices when you need to modify:**
   ```go
   for i := range items {
       items[i].Field = value  // Direct modification
   }
   ```

3. **Range is efficient** - no performance penalty vs traditional for loop

### Summary
- **Three forms:** Traditional (`for init; condition; post`), Range (`for key, value := range`), Infinite (`for {}`)
- **Range works with:** Slices, arrays, maps, strings, channels
- **Blank identifier:** Use `_` to ignore values you don't need
- **Break/Continue:** Control flow within loops
- **Range copies values:** Be careful with value types vs pointer types
- **Modification:** Safe to modify values, unsafe to modify collection structure during iteration

---

## Struct Field Tags

**Definition:** Struct field tags are metadata attached to struct fields using backticks. They provide instructions for encoding/decoding, validation, and other reflection-based operations.

**Syntax:** `FieldName FieldType `tag:"value"` // Comment`

**Example:**
```go
type GetRatesRequest struct {
    Base string `json:"base"` // Base currency code (e.g., "USD")
}
```

**Components:**
1. **Field name:** `Base`
2. **Field type:** `string`
3. **Struct tag:** `` `json:"base"` `` (in backticks)
4. **Comment:** `// Base currency code (e.g., "USD")`

### JSON Tags

**Purpose:** Control how struct fields are encoded/decoded to/from JSON.

**Syntax:** `` `json:"field_name"` ``

**Effect:**
- **Encoding (Go → JSON):** Field `Base` becomes `"base"` in JSON
- **Decoding (JSON → Go):** JSON field `"base"` maps to struct field `Base`

**Example:**
```go
type GetRatesRequest struct {
    Base string `json:"base"`
}

// Encoding (Go struct → JSON)
req := GetRatesRequest{Base: "USD"}
jsonBytes, _ := json.Marshal(req)
// Result: {"base":"USD"}

// Decoding (JSON → Go struct)
jsonStr := `{"base":"EUR"}`
var req GetRatesRequest
json.Unmarshal([]byte(jsonStr), &req)
// req.Base = "EUR"
```

**Common JSON tag options:**
```go
Field string `json:"field_name"`           // Custom JSON name
Field string `json:"field_name,omitempty"` // Omit if empty
Field string `json:"-"`                    // Ignore this field in JSON
Field string `json:",omitempty"`          // Use field name, omit if empty
```

### Other Common Tags

**XML tags:**
```go
Field string `xml:"field_name"`
```

**Database tags (GORM, etc.):**
```go
Field string `gorm:"column:field_name"`
```

**Validation tags:**
```go
Field string `validate:"required,min=3"`
```

**Multiple tags:**
```go
Field string `json:"field_name" xml:"field_name" validate:"required"`
```

### Tag Syntax Rules

1. **Backticks required:** Tags must be in backticks (`` ` ``), not quotes
2. **Space-separated:** Multiple tags separated by spaces
3. **Key-value pairs:** Format is `key:"value"`
4. **Quotes in values:** Values are typically quoted strings

**Examples:**
```go
// ✅ Correct
Field string `json:"name"`

// ✅ Multiple tags
Field string `json:"name" xml:"name" validate:"required"`

// ✅ With options
Field string `json:"name,omitempty"`

// ❌ Wrong (quotes instead of backticks)
Field string "json:\"name\""  // ERROR!

// ❌ Wrong (no quotes in value)
Field string `json:name`  // ERROR!
```

### Why Use Tags?

1. **JSON/API serialization:** Control how structs convert to/from JSON
2. **Naming conventions:** Map Go field names (PascalCase) to JSON (camelCase/snake_case)
3. **Field control:** Omit fields, handle empty values, ignore fields
4. **Validation:** Add validation rules for fields
5. **Database mapping:** Map struct fields to database columns

**Example from codebase:**
```go
type GetRateRequest struct {
    Base   string `json:"base"`   // Base currency code (e.g., "USD")
    Target string `json:"target"` // Target currency code (e.g., "EUR")
}
```

**What this means:**
- `Base` field → JSON `"base"` (lowercase)
- `Target` field → JSON `"target"` (lowercase)
- When API receives `{"base":"USD","target":"EUR"}`, it maps to struct fields

### Summary
- **Struct tags:** Metadata in backticks attached to struct fields
- **Syntax:** `` `key:"value"` `` - must use backticks, not quotes
- **JSON tags:** Control JSON encoding/decoding field names
- **Purpose:** Map Go field names to external formats (JSON, XML, DB, etc.)
- **Access:** Tags are accessed via reflection at runtime
- **Multiple tags:** Space-separated tags in same backtick string

---

## Context.Context

**Definition:** `context.Context` is an interface that carries request-scoped values, cancellation signals, and deadlines across API boundaries and between goroutines.

### What Context Does

1. **Cancellation:** Signals when an operation should stop
2. **Timeouts:** Sets deadlines for operations
3. **Request-scoped values:** Carries key-value data through call chain
4. **Propagation:** Passes cancellation/timeouts through function calls

### Why It's Relevant

1. **Prevents resource leaks:** Cancels long-running operations
2. **Timeout control:** Enforces maximum execution time
3. **Graceful shutdown:** Stops operations cleanly
4. **Request tracing:** Carries request IDs and metadata
5. **Standard pattern:** Idiomatic in Go for async operations

### Example from Codebase

```go
func (p *CurrencyAPIProvider) FetchRate(ctx context.Context, base, target CurrencyCode) (*ExchangeRate, error) {
    // Check context before starting operation
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }
    
    // Create request with context (enables cancellation and timeout)
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    // HTTP request respects context timeout/cancellation
}
```

### Creating Contexts

**1. Background context (root):**
```go
ctx := context.Background()  // Never cancelled, no deadline
```

**2. TODO context (placeholder):**
```go
ctx := context.TODO()  // Use when unsure which context to use
```

**3. With cancellation:**
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // Always call cancel to free resources

// Later, cancel the operation:
cancel()  // Signals all operations using this context to stop
```

**4. With timeout:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()  // Always call cancel

// Context automatically cancels after 5 seconds
```

**5. With deadline:**
```go
deadline := time.Now().Add(10 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()
```

**6. With values:**
```go
ctx := context.WithValue(context.Background(), "requestID", "abc123")
// Later retrieve:
requestID := ctx.Value("requestID").(string)
```

### Common Patterns

**1. Check cancellation before operations:**
```go
if ctx.Err() != nil {
    return nil, ctx.Err()
}
```

**2. Pass context to functions (first parameter):**
```go
func FetchRate(ctx context.Context, ...) (*ExchangeRate, error) {
    // ctx is first parameter (Go convention)
}
```

**3. Use with HTTP requests:**
```go
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)  // Respects context timeout/cancellation
```

**4. Use with database operations:**
```go
result, err := db.QueryContext(ctx, "SELECT ...")
// DynamoDB, SQL, etc. all support context
```

**5. Chain contexts:**
```go
// Parent context with timeout
parentCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Child context with shorter timeout
childCtx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()
```

### Context Propagation

**Example call chain:**
```go
// Handler receives context
func Handler(ctx context.Context, event Event) {
    // Pass context down
    result, err := useCase.Execute(ctx, req)
}

// Use case passes to repository
func (uc *UseCase) Execute(ctx context.Context, req Request) {
    rate, err := uc.repository.Get(ctx, base, target)
}

// Repository passes to HTTP client
func (r *Repository) Get(ctx context.Context, ...) {
    resp, err := httpClient.Do(req.WithContext(ctx))
}
```

**Key point:** Context flows down the call chain. If cancelled at the top, all downstream operations are cancelled.

### Context Rules

1. **First parameter:** `context.Context` should be the first parameter
2. **Don't store in structs:** Pass as parameter, don't embed
3. **Don't pass nil:** Use `context.Background()` or `context.TODO()` if needed
4. **Always check:** Check `ctx.Err()` before long operations
5. **Always cancel:** Call `cancel()` from `WithCancel/WithTimeout` to free resources
6. **Immutable:** Contexts are immutable, create new ones with `With*` functions

### Context Errors

```go
// Check if context is done
if ctx.Err() != nil {
    // ctx.Err() returns:
    // - context.Canceled (if cancelled)
    // - context.DeadlineExceeded (if timed out)
    // - nil (if still active)
    return nil, ctx.Err()
}

// Or use select
select {
case <-ctx.Done():
    return nil, ctx.Err()
default:
    // Continue operation
}
```

### When to Use Context

- ✅ HTTP requests (timeouts, cancellation)
- ✅ Database operations (query timeouts)
- ✅ External API calls (prevent hanging)
- ✅ Long-running operations (graceful shutdown)
- ✅ Goroutines (cancellation propagation)
- ✅ Request tracing (carry metadata)

### When NOT to Use Context

- ❌ Internal utility functions (unless needed for cancellation)
- ❌ Synchronous operations that are always fast
- ❌ Functions that don't need cancellation/timeouts

### Summary
- **Purpose:** Cancellation, timeouts, request-scoped values
- **First parameter:** Always first parameter in function signatures
- **Propagation:** Flows down call chain automatically
- **Cancellation:** Check `ctx.Err()` before operations
- **Timeouts:** Use `WithTimeout` to set deadlines
- **Always cancel:** Call `cancel()` from `WithCancel/WithTimeout`
- **Immutable:** Create new contexts with `With*` functions

---

## Defer

**Definition:** `defer` schedules a function call to execute when the surrounding function returns, regardless of how it returns (normal return, panic, or early return).

**Syntax:** `defer functionCall()`

**Example:**
```go
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()  // Always closes, even if function returns early
```

### Key Behaviors

**1. LIFO execution (Last In, First Out):**
```go
defer fmt.Println("A")
defer fmt.Println("B")
defer fmt.Println("C")
// Executes: C, B, A (reverse order)
```

**2. Arguments evaluated immediately:**
```go
i := 0
defer fmt.Println(i)  // Prints 0 (evaluated now, not later)
i++
// Function returns, defer executes: prints 0
```

**3. Always executes, even on panic:**
```go
defer fmt.Println("Cleanup")  // Still executes!
panic("something went wrong")
```

**4. Can modify named return values:**
```go
func example() (result int) {
    defer func() {
        result = 42  // Modifies named return
    }()
    return 0  // Actually returns 42!
}
```

### Common Use Cases

**Resource cleanup:**
```go
file, err := os.Open("data.txt")
if err != nil {
    return err
}
defer file.Close()  // Always closes
```

**Unlocking mutexes:**
```go
mu.Lock()
defer mu.Unlock()  // Always unlocks, even if function panics
```

**Context cancellation:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()  // Always cancels to free resources
```

**HTTP response closing:**
```go
resp, err := http.Get(url)
defer resp.Body.Close()  // Always closes response body
```

### When to Use

- ✅ Closing resources (files, HTTP responses, connections)
- ✅ Unlocking mutexes
- ✅ Recovering from panics
- ✅ Cleanup operations

### When NOT to Use

- ❌ In loops (can cause resource leaks - defer executes at end of function, not loop iteration)
- ❌ For performance-critical code (small overhead)

### Summary
- **Purpose:** Schedule cleanup to run when function returns
- **Execution:** Runs in reverse order (LIFO) when function exits
- **Always runs:** Even on panic or early return
- **Arguments:** Evaluated immediately, execution deferred
- **Common use:** Resource cleanup (close files, unlock mutexes, cancel contexts)
- **Pattern:** `defer resource.Close()` right after acquiring resource

---

