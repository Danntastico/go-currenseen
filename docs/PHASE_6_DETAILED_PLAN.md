# Phase 6: Cache Strategy & Fallback Logic

## Overview

**Goal**: Implement cache-first strategy with fallback to stale cache when external API is unavailable.

**What We're Building**:
- Enhanced use cases with cache-first logic
- Circuit breaker integration for fallback detection
- Stale cache fallback when circuit is open
- Cache expiration handling
- Updated tests

**Why This Phase**:
- Reduces external API calls (>80% reduction)
- Faster response times (<200ms for cached)
- Graceful degradation when external API is down
- Better user experience (stale data better than no data)

**Dependencies**: 
- Phase 2 (Repository) ✅ Complete
- Phase 3 (DynamoDB Adapter) ✅ Complete
- Phase 5 (Circuit Breaker) ✅ Complete

**Estimated Time**: 4-6 hours

---

## Step 1: Understand Current Implementation

**Objective**: Review current use case implementations

**Why**: Understand what needs to be enhanced

**What to Do**:

1. Review `GetExchangeRateUseCase`:
   - Already has cache-first logic
   - Has basic fallback to stale cache
   - **Missing**: Circuit breaker detection

2. Review `GetAllRatesUseCase`:
   - Already has cache-first logic
   - Has basic fallback to stale cache
   - **Missing**: Circuit breaker detection

3. Review circuit breaker integration:
   - Provider is wrapped with `CircuitBreakerProvider`
   - Returns `ErrCircuitOpen` when circuit is open
   - Need to detect this error in use cases

**Deliverable**: ✅ Understanding of current implementation

**Time**: 15 minutes

---

## Step 2: Enhance GetExchangeRateUseCase

**Objective**: Add circuit breaker detection and proper fallback

**Why**: Graceful degradation when circuit is open

**What to Do**:

### 2.1 Import Circuit Breaker Error

```go
import (
    "github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)
```

### 2.2 Update Execute Method Flow

**Current Flow**:
1. Check cache → return if valid
2. Fetch from provider → save to cache → return
3. Fallback to stale cache if provider fails

**New Flow**:
1. Check cache → return if valid
2. Fetch from provider:
   - If success → save to cache → return
   - If `ErrCircuitOpen` → use `GetStale()` → return stale with warning
   - If other error → fallback to stale cache (if available)
3. If no stale cache available → return error

**Key Changes**:
- Detect `ErrCircuitOpen` specifically
- Use `repository.GetStale()` when circuit is open
- Mark stale rates appropriately
- Preserve existing fallback for other errors

**Deliverable**: ✅ Enhanced GetExchangeRateUseCase

**Time**: 1 hour

---

## Step 3: Enhance GetAllRatesUseCase

**Objective**: Add circuit breaker detection and proper fallback

**Why**: Consistent behavior across use cases

**What to Do**:

Similar to Step 2, but for `GetAllRatesUseCase`:

1. Import circuit breaker error
2. Detect `ErrCircuitOpen` in provider call
3. Use `GetStale()` equivalent (GetByBase already returns stale data)
4. Mark rates as stale when returned from fallback

**Deliverable**: ✅ Enhanced GetAllRatesUseCase

**Time**: 1 hour

---

## Step 4: Update Use Case Tests

**Objective**: Test cache-first and fallback scenarios

**Why**: Ensure correctness

**What to Do**:

### 4.1 GetExchangeRateUseCase Tests

**New Test Cases**:
- `TestGetExchangeRateUseCase_CacheHit_Valid` - Cache hit with valid rate
- `TestGetExchangeRateUseCase_CacheHit_Expired` - Cache hit but expired
- `TestGetExchangeRateUseCase_CacheMiss_FetchSuccess` - Cache miss, fetch succeeds
- `TestGetExchangeRateUseCase_CircuitOpen_FallbackToStale` - Circuit open, use stale cache
- `TestGetExchangeRateUseCase_CircuitOpen_NoStaleCache` - Circuit open, no stale cache
- `TestGetExchangeRateUseCase_ProviderError_FallbackToStale` - Other error, fallback to stale
- `TestGetExchangeRateUseCase_ProviderError_NoStaleCache` - Other error, no stale cache

### 4.2 GetAllRatesUseCase Tests

**New Test Cases**:
- `TestGetAllRatesUseCase_CacheHit_AllValid` - Cache hit, all rates valid
- `TestGetAllRatesUseCase_CacheHit_SomeExpired` - Cache hit, some expired
- `TestGetAllRatesUseCase_CacheMiss_FetchSuccess` - Cache miss, fetch succeeds
- `TestGetAllRatesUseCase_CircuitOpen_FallbackToStale` - Circuit open, use stale cache
- `TestGetAllRatesUseCase_CircuitOpen_NoStaleCache` - Circuit open, no stale cache
- `TestGetAllRatesUseCase_ProviderError_FallbackToStale` - Other error, fallback to stale

**Deliverable**: ✅ Comprehensive tests written

**Time**: 1.5 hours

---

## Step 5: Verify Cache Expiration Logic

**Objective**: Ensure cache expiration is handled correctly

**Why**: Cache TTL must be respected

**What to Do**:

1. Verify `entity.ExchangeRate.IsValid()` is used correctly
2. Verify `entity.ExchangeRate.IsExpired()` is used correctly
3. Verify stale rates are marked with `Stale=true`
4. Verify cache TTL is passed correctly to repository

**Deliverable**: ✅ Cache expiration verified

**Time**: 30 minutes

---

## Step 6: Documentation and Code Review

**Objective**: Document and review implementation

**What to Do**:

1. Update use case documentation
2. Document fallback strategy
3. Code review checklist

**Deliverable**: ✅ Code documented and reviewed

**Time**: 30 minutes

---

## Summary Checklist

Before considering Phase 6 complete:

- [x] GetExchangeRateUseCase enhanced with circuit breaker detection
- [x] GetAllRatesUseCase enhanced with circuit breaker detection
- [x] Fallback to stale cache when circuit is open
- [x] Fallback to stale cache for other provider errors
- [x] Cache expiration logic verified
- [x] Use case tests updated and passing
- [x] Code documented
- [x] Code reviewed

**Status**: ✅ **Phase 6 Complete** - All checklist items verified and completed.

---

## Estimated Total Time

- Step 1: 15 minutes
- Step 2: 1 hour
- Step 3: 1 hour
- Step 4: 1.5 hours
- Step 5: 30 minutes
- Step 6: 30 minutes

**Total**: ~4.5-5 hours

---

## Next Steps

After Phase 6 completion:
- Phase 7: Lambda Handlers & API Gateway Integration
