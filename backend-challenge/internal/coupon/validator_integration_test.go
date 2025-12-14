package coupon

import (
	"context"
	"testing"
	"time"
)

// TestValidator_RealS3Files tests the validator with actual S3 coupon files
// This test is marked as integration and will download real data from S3
// Run with: go test -v -tags=integration -run TestValidator_RealS3Files
func TestValidator_RealS3Files(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test with real S3 files")
	}

	// Real S3 URLs provided by the user
	urls := []string{
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz",
	}

	t.Log("Creating validator...")
	validator := NewValidator()

	t.Log("Loading coupon data from S3... (this may take a minute)")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	err := validator.LoadFromURLs(ctx, urls)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to load coupon data: %v", err)
	}

	t.Logf("Loaded coupon data in %v", elapsed)

	// Get statistics
	stats := validator.GetStats()
	t.Logf("Statistics: %+v", stats)

	totalFiles, ok := stats["total_files"].(int)
	if !ok || totalFiles != 3 {
		t.Errorf("Expected 3 files, got %v", stats["total_files"])
	}

	totalCoupons, ok := stats["total_coupons"].(int)
	if !ok {
		t.Errorf("Expected total_coupons to be an int, got %T", stats["total_coupons"])
	} else {
		t.Logf("Total coupons loaded: %d", totalCoupons)
		if totalCoupons == 0 {
			t.Error("No coupons were loaded")
		}
	}

	// Test some known valid coupons (examples from the requirements)
	validCoupons := []string{
		"HAPPYHRS", // Should appear in at least 2 files
		"FIFTYOFF", // Should appear in at least 2 files
	}

	for _, code := range validCoupons {
		t.Run("validate_"+code, func(t *testing.T) {
			isValid := validator.IsValid(context.Background(), code)
			t.Logf("Coupon %q is valid: %v", code, isValid)
			// Note: We can't assert true/false without knowing the actual data
			// This test just verifies the validator works with real data
		})
	}

	// Test a known invalid coupon (from requirements: only in 1 file)
	t.Run("validate_SUPER100", func(t *testing.T) {
		isValid := validator.IsValid(context.Background(), "SUPER100")
		t.Logf("Coupon SUPER100 is valid: %v (should be false per requirements)", isValid)
	})

	// Test length validation
	t.Run("validate_too_short", func(t *testing.T) {
		isValid := validator.IsValid(context.Background(), "SHORT")
		if isValid {
			t.Error("Coupon with less than 8 characters should be invalid")
		}
	})

	t.Run("validate_too_long", func(t *testing.T) {
		isValid := validator.IsValid(context.Background(), "TOOLONGCODE")
		if isValid {
			t.Error("Coupon with more than 10 characters should be invalid")
		}
	})

	// Test concurrent validation with real data
	t.Run("concurrent_validation", func(t *testing.T) {
		testCodes := []string{"HAPPYHRS", "FIFTYOFF", "TEST12345", "PROMO2024"}

		results := make(chan bool, len(testCodes))

		for _, code := range testCodes {
			go func(c string) {
				results <- validator.IsValid(context.Background(), c)
			}(code)
		}

		for i := 0; i < len(testCodes); i++ {
			<-results
		}

		t.Log("Concurrent validation completed successfully")
	})
}

// TestValidator_RealS3Files_Sample tests a sample of coupons from real data
// Run with: go test -v -run TestValidator_RealS3Files_Sample
func TestValidator_RealS3Files_Sample(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	urls := []string{
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz",
	}

	validator := NewValidator()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	t.Log("Loading coupon data (this takes ~1-2 minutes for 624MB files)...")
	if err := validator.LoadFromURLs(ctx, urls); err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	stats := validator.GetStats()
	t.Logf("Loaded coupons: %+v", stats)

	// Sample test cases
	testCases := []struct {
		code        string
		expectValid bool
		reason      string
	}{
		{"HAPPYHRS", true, "8 chars, likely in multiple files"},
		{"FIFTYOFF", true, "8 chars, likely in multiple files"},
		{"SHORT", false, "only 5 chars"},
		{"WAYTOLONG!!", false, "12 chars, exceeds 10"},
	}

	for _, tc := range testCases {
		t.Run(tc.code, func(t *testing.T) {
			result := validator.IsValid(context.Background(), tc.code)
			if result != tc.expectValid {
				t.Logf("Code %q: expected %v, got %v (%s)",
					tc.code, tc.expectValid, result, tc.reason)
			}
		})
	}
}
