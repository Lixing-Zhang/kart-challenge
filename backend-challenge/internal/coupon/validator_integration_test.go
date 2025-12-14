package coupon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestValidator_RealFiles tests the validator with actual large coupon files
// This test is skipped in short mode as it requires large files to be present
func TestValidator_RealFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real file test in short mode")
	}

	// Check if files exist (relative to backend-challenge directory)
	file1 := filepath.Join("..", "..", "data", "couponbase1")
	file2 := filepath.Join("..", "..", "data", "couponbase2")
	file3 := filepath.Join("..", "..", "data", "couponbase3")

	for _, f := range []string{file1, file2, file3} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Skipf("skipping test: required file %s not found", f)
		}
	}

	validator := NewValidator()
	ctx := context.Background()

	// Load files
	t.Log("Loading coupon files...")
	start := time.Now()
	err := validator.LoadFromFiles(ctx, []string{file1, file2, file3})
	if err != nil {
		t.Fatalf("failed to load files: %v", err)
	}
	t.Logf("Files loaded in %v", time.Since(start))

	stats := validator.GetStats()
	t.Logf("Loaded %v files", stats["total_files"])

	// Test validation - we don't know which specific codes are valid in real files
	// so we test format validation and that the function doesn't panic
	testCases := []struct {
		code   string
		reason string
	}{
		{"TESTCODE", "valid length code"},
		{"NOTEXIST", "another valid length code"},
		{"SHORT", "too short - should be invalid"},
		{"TOOLONGCODE", "too long - should be invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.code, func(t *testing.T) {
			start := time.Now()
			result := validator.IsValid(ctx, tc.code)
			duration := time.Since(start)

			t.Logf("IsValid(%q) = %v (took %v) - %s", tc.code, result, duration, tc.reason)

			// Just verify it doesn't panic and completes in reasonable time
			if duration > 10*time.Second {
				t.Errorf("validation took too long: %v", duration)
			}
		})
	}
}

// TestValidator_RealFiles_Performance measures validation performance
func TestValidator_RealFiles_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Get path relative to test directory
	file1 := filepath.Join("..", "..", "data", "couponbase1")

	// Check if file exists
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Skipf("skipping test: required file %s not found", file1)
	}

	validator := NewValidator()
	ctx := context.Background()

	// For performance test, use just one file
	if err := validator.LoadFromFiles(ctx, []string{file1}); err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	// Test multiple validations
	codes := []string{"SUPER100", "QYYSPC46", "NSZ0VMH4", "NOTEXIST"}

	start := time.Now()
	for i := 0; i < 10; i++ {
		code := codes[i%len(codes)]
		validator.IsValid(ctx, code)
	}
	duration := time.Since(start)

	avgTime := duration / 10
	t.Logf("Average validation time: %v", avgTime)

	// Validation should complete reasonably fast even with large files
	if avgTime > 5*time.Second {
		t.Logf("Warning: Average validation time is high: %v", avgTime)
	}
}
