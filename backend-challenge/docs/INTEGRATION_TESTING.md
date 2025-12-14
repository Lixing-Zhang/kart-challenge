# Integration Testing with Real S3 Files

## Test Overview

The real S3 coupon files are quite large (~624MB each, total ~1.8GB):
- File 1: https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz
- File 2: https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz  
- File 3: https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz

## Running Integration Tests

### Full Integration Test (1-2 minutes)
```bash
cd backend-challenge
go test -v -timeout=10m -run TestValidator_RealS3Files_Sample ./internal/coupon/
```

This will:
1. Download all 3 gzipped files from S3 (concurrently)
2. Parse and load all coupons into memory
3. Run validation tests with real data
4. Report statistics (total files, total coupons, etc.)

### Manual Testing with curl

Test the running server:

```bash
# Start the server (loads coupon data on startup)
go run cmd/server/*.go

# In another terminal, test validation endpoints:
curl http://localhost:8080/api/coupon/HAPPYHRS
curl http://localhost:8080/api/coupon/FIFTYOFF
curl http://localhost:8080/api/coupon/SUPER100
curl http://localhost:8080/api/coupon/stats
```

## Expected Behavior

### Valid Coupons
Coupons that appear in at least 2 of the 3 files and have 8-10 characters:
```json
{
  "valid": true,
  "coupon": "HAPPYHRS"
}
```

### Invalid Coupons  
Coupons that only appear in 1 file or don't meet length requirements:
```json
{
  "valid": false,
  "coupon": "SUPER100",
  "message": "Coupon not found or invalid"
}
```

### Statistics Endpoint
```json
{
  "total_files": 3,
  "file_sizes": [1000000, 1000000, 1000000],
  "total_coupons": 3000000
}
```

## Performance Notes

- **Parallel Loading**: All 3 files download concurrently (3x speedup)
- **Memory Efficient**: Uses streaming gzip decompression
- **Fast Validation**: O(1) lookups with concurrent searches
- **Load Time**: ~30-60 seconds for 1.8GB of compressed data (depends on network speed)

## Troubleshooting

If the server hangs on startup:
1. Check network connectivity
2. Verify S3 URLs are accessible: `curl -I [URL]`
3. Increase timeout if needed
4. Check logs for specific error messages

If tests timeout:
```bash
go test -v -timeout=15m -run TestValidator_RealS3Files ./internal/coupon/
```
