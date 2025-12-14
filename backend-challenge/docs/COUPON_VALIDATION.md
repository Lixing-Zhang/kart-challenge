# Coupon Validation Feature

## Overview

The coupon validation feature validates promotional coupon codes against three remote gzipped files. A coupon is considered valid if:
1. It has between 8-10 characters
2. It appears in at least 2 of the 3 coupon files

## Architecture

### Components

1. **Validator** (`internal/coupon/validator.go`)
   - Core validation logic with concurrent processing
   - Thread-safe using `sync.RWMutex`
   - Parallel file loading using goroutines and channels
   - Concurrent validation searches with early termination

2. **Handler** (`internal/handlers/coupon_handler.go`)
   - HTTP endpoints for coupon validation
   - Interface-based design for testability

3. **Tests**
   - `internal/coupon/validator_test.go` - Comprehensive validator tests
   - `internal/handlers/coupon_handler_test.go` - Handler tests with mocks

## Performance Optimizations

### Parallel File Loading
- Uses goroutines to download all 3 files concurrently
- 3x faster than sequential loading
- Uses channels to collect results while maintaining order
- WaitGroup ensures all downloads complete before proceeding

```go
// Launch concurrent downloads
for i, url := range urls {
    wg.Add(1)
    go func(index int, fileURL string) {
        defer wg.Done()
        coupons, err := v.loadFromURL(ctx, fileURL)
        resultChan <- fileLoadResult{index: index, coupons: coupons, err: err}
    }(i, url)
}
```

### Concurrent Validation
- Searches all 3 coupon files concurrently
- Early termination when 2 matches are found (no need to check the third file)
- Uses context cancellation to stop remaining searches

```go
// Search concurrently with early termination
for _, cs := range v.couponSets {
    wg.Add(1)
    go func(couponSet *couponSet) {
        defer wg.Done()
        if found {
            foundChan <- true
        }
    }(cs)
}
```

### Thread Safety
- All coupon sets protected with `sync.RWMutex`
- Multiple goroutines can read concurrently
- Safe for high-concurrency environments

## API Endpoints

### Validate Coupon
```
GET /api/coupon/{couponCode}
```

**Response (Valid Coupon - 200 OK):**
```json
{
  "valid": true,
  "coupon": "HAPPYHRS"
}
```

**Response (Invalid Coupon - 404 Not Found):**
```json
{
  "valid": false,
  "coupon": "SUPER100",
  "message": "Coupon not found or invalid"
}
```

### Get Statistics (Debug/Monitoring)
```
GET /api/coupon/stats
```

**Response:**
```json
{
  "total_files": 3,
  "file_sizes": [10000, 10000, 10000],
  "total_coupons": 30000
}
```

## Examples

### Valid Coupons
- `HAPPYHRS` - Appears in all 3 files
- `FIFTYOFF` - Appears in files 1 and 2

### Invalid Coupons
- `SUPER100` - Only in 1 file (needs at least 2)
- `SAVE20` - Only in 1 file
- `SHORT` - Less than 8 characters
- `TOOLONGCODE` - More than 10 characters

## Configuration

Coupon file URLs are configured via environment variables:

```env
COUPON_FILE1_URL=https://fyle-dev-challenge.s3.us-west-2.amazonaws.com/couponbase1.gz
COUPON_FILE2_URL=https://fyle-dev-challenge.s3.us-west-2.amazonaws.com/couponbase2.gz
COUPON_FILE3_URL=https://fyle-dev-challenge.s3.us-west-2.amazonaws.com/couponbase3.gz
```

## Testing

### Run Validator Tests
```bash
go test -v ./internal/coupon/
```

### Run Handler Tests
```bash
go test -v ./internal/handlers/ -run Coupon
```

### Run All Tests
```bash
go test ./...
```

### Test Coverage
- LoadFromURLs: Successful load, empty URLs, failed downloads, context cancellation
- IsValid: Valid/invalid coupons, length validation, concurrent access
- GetStats: File counts and coupon counts
- Concurrent access: 100 goroutines validating simultaneously
- Benchmark: Performance testing with large datasets

## Error Handling

1. **Startup Errors**: If coupon files fail to load, the server exits with an error
2. **Validation Errors**: Invalid coupons return 404 with descriptive message
3. **Network Errors**: Handled during file loading with proper error messages
4. **Context Cancellation**: Supports graceful cancellation via context

## Monitoring

The `/api/coupon/stats` endpoint provides:
- Number of loaded files
- Size of each coupon file
- Total number of unique coupons across all files

Use this for monitoring and debugging in production.

## Future Enhancements

1. **Caching**: Add LRU cache for frequently validated coupons
2. **Periodic Refresh**: Reload coupon files on a schedule
3. **Metrics**: Add Prometheus metrics for validation rates
4. **Distributed Cache**: Use Redis for multi-instance deployments
