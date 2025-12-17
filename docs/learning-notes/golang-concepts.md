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

## Constructor Pattern (Returning Pointers)

**Pattern Name:** Constructor/Factory Pattern

**Syntax:**
```go
func NewType(...params) (*Type, error) {
    // validation
    if err := validate(...); err != nil {
        return nil, err
    }
    // Address-of operator returns memory address
    return &Type{...}, nil
}
```

**Components:**
1. **Return type:** `(*Type, error)` - returns pointer to struct and error
2. **Pointer type:** `*Type` - indicates function returns a memory reference
3. **Address-of operator:** `&Type{...}` - creates struct literal and returns its address

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

**Memory behavior:**
- `&ExchangeRate{...}` creates struct in memory and returns its address
- Caller receives pointer (reference), not a copy
- Methods with pointer receivers can modify the original struct

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

