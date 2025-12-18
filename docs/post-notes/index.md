

# Implementations for a production system 

---
1. `internal\application\usecase\get_all_rates.go`
seen on `getAllRatesUseCase.Execute()`
> Note: This implementation fetches all rates from the provider if cache miss.
In a production system, you might want to check which rates are missing/expired

2. 
