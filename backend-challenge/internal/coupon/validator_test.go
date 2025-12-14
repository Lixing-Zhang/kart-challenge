package coupon

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// setupTestFiles creates temporary test files and returns their paths
func setupTestFiles(t *testing.T) (string, string, string, func()) {
	t.Helper()

	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "coupons1.txt")
	file2 := filepath.Join(tmpDir, "coupons2.txt")
	file3 := filepath.Join(tmpDir, "coupons3.txt")

	// File 1: VALIDABC, TESTCODE, COUPON01, INVALID1, AAAA1111
	if err := os.WriteFile(file1, []byte("VALIDABC\nTESTCODE\nCOUPON01\nINVALID1\nAAAA1111\n"), 0644); err != nil {
		t.Fatalf("failed to create test file 1: %v", err)
	}

	// File 2: VALIDABC, TESTCODE, SPECIAL9, COUPON02, BBBB2222
	if err := os.WriteFile(file2, []byte("VALIDABC\nTESTCODE\nSPECIAL9\nCOUPON02\nBBBB2222\n"), 0644); err != nil {
		t.Fatalf("failed to create test file 2: %v", err)
	}

	// File 3: VALIDABC, SPECIAL9, COUPON03, CCCC3333, ONLYONE1
	if err := os.WriteFile(file3, []byte("VALIDABC\nSPECIAL9\nCOUPON03\nCCCC3333\nONLYONE1\n"), 0644); err != nil {
		t.Fatalf("failed to create test file 3: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return file1, file2, file3, cleanup
}

func TestValidator_LoadFromFiles(t *testing.T) {
	t.Run("successful load from multiple files", func(t *testing.T) {
		file1, file2, file3, cleanup := setupTestFiles(t)
		defer cleanup()

		validator := NewValidator()
		err := validator.LoadFromFiles(context.Background(), []string{file1, file2, file3})

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		stats := validator.GetStats()
		if stats["total_files"] != 3 {
			t.Errorf("expected 3 files loaded, got %v", stats["total_files"])
		}
	})

	t.Run("empty file paths", func(t *testing.T) {
		validator := NewValidator()
		err := validator.LoadFromFiles(context.Background(), []string{})

		if err == nil {
			t.Error("expected error for empty file paths, got nil")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		validator := NewValidator()
		err := validator.LoadFromFiles(context.Background(), []string{"/non/existent/file.txt"})

		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})
}

func TestValidator_IsValid(t *testing.T) {
	file1, file2, file3, cleanup := setupTestFiles(t)
	defer cleanup()

	validator := NewValidator()
	if err := validator.LoadFromFiles(context.Background(), []string{file1, file2, file3}); err != nil {
		t.Fatalf("failed to load files: %v", err)
	}

	tests := []struct {
		name     string
		code     string
		expected bool
		reason   string
	}{
		{
			name:     "valid - appears in all 3 files",
			code:     "VALIDABC",
			expected: true,
			reason:   "VALIDABC appears in all 3 files",
		},
		{
			name:     "valid - appears in 2 files (1 and 2)",
			code:     "TESTCODE",
			expected: true,
			reason:   "TESTCODE appears in file1 and file2",
		},
		{
			name:     "valid - appears in 2 files (2 and 3)",
			code:     "SPECIAL9",
			expected: true,
			reason:   "SPECIAL9 appears in file2 and file3",
		},
		{
			name:     "invalid - appears in only 1 file",
			code:     "COUPON01",
			expected: false,
			reason:   "COUPON01 only appears in file1",
		},
		{
			name:     "invalid - appears in only 1 file",
			code:     "ONLYONE1",
			expected: false,
			reason:   "ONLYONE1 only appears in file3",
		},
		{
			name:     "invalid - does not exist",
			code:     "NOTEXIST",
			expected: false,
			reason:   "NOTEXIST doesn't appear in any file",
		},
		{
			name:     "invalid - too short",
			code:     "SHORT",
			expected: false,
			reason:   "only 5 characters, minimum is 8",
		},
		{
			name:     "invalid - too long",
			code:     "TOOLONGCODE",
			expected: false,
			reason:   "11 characters, maximum is 10",
		},
		{
			name:     "case insensitive - lowercase valid code",
			code:     "validabc",
			expected: true,
			reason:   "should match VALIDABC (case insensitive)",
		},
		{
			name:     "whitespace handling",
			code:     "  VALIDABC  ",
			expected: true,
			reason:   "should trim whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsValid(context.Background(), tt.code)
			if result != tt.expected {
				t.Errorf("IsValid(%q) = %v, expected %v (reason: %s)",
					tt.code, result, tt.expected, tt.reason)
			}
		})
	}
}

func TestValidator_IsValid_ConcurrentAccess(t *testing.T) {
	file1, file2, file3, cleanup := setupTestFiles(t)
	defer cleanup()

	validator := NewValidator()
	if err := validator.LoadFromFiles(context.Background(), []string{file1, file2, file3}); err != nil {
		t.Fatalf("failed to load files: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent validation requests
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			codes := []string{"VALIDABC", "TESTCODE", "SPECIAL9", "NOTEXIST"}
			code := codes[n%len(codes)]

			result := validator.IsValid(context.Background(), code)

			// Verify expected results for known codes
			switch code {
			case "VALIDABC", "TESTCODE", "SPECIAL9":
				if !result {
					t.Errorf("expected %s to be valid", code)
				}
			case "NOTEXIST":
				if result {
					t.Errorf("expected %s to be invalid", code)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestValidator_IsValid_ContextCancellation(t *testing.T) {
	file1, file2, file3, cleanup := setupTestFiles(t)
	defer cleanup()

	validator := NewValidator()
	if err := validator.LoadFromFiles(context.Background(), []string{file1, file2, file3}); err != nil {
		t.Fatalf("failed to load files: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should handle cancelled context gracefully
	result := validator.IsValid(ctx, "VALIDABC")
	// Since files are small, might complete before cancellation is detected
	// Just verify it doesn't panic
	_ = result
}

func TestValidator_GetStats(t *testing.T) {
	file1, file2, file3, cleanup := setupTestFiles(t)
	defer cleanup()

	validator := NewValidator()

	t.Run("stats before loading", func(t *testing.T) {
		stats := validator.GetStats()
		if stats["total_files"] != 0 {
			t.Errorf("expected 0 files before loading, got %v", stats["total_files"])
		}
	})

	t.Run("stats after loading", func(t *testing.T) {
		if err := validator.LoadFromFiles(context.Background(), []string{file1, file2, file3}); err != nil {
			t.Fatalf("failed to load files: %v", err)
		}

		stats := validator.GetStats()
		if stats["total_files"] != 3 {
			t.Errorf("expected 3 files after loading, got %v", stats["total_files"])
		}

		filePaths, ok := stats["file_paths"].([]string)
		if !ok {
			t.Error("expected file_paths to be []string")
		}
		if len(filePaths) != 3 {
			t.Errorf("expected 3 file paths, got %d", len(filePaths))
		}
	})
}

// TestValidator_LargeFile tests streaming with a larger file
func TestValidator_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large file test in short mode")
	}

	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "large1.txt")
	file2 := filepath.Join(tmpDir, "large2.txt")
	file3 := filepath.Join(tmpDir, "large3.txt")

	// Create files with 10,000 coupons each
	f1, _ := os.Create(file1)
	f2, _ := os.Create(file2)
	f3, _ := os.Create(file3)

	// Add specific test coupons
	_, _ = f1.WriteString("FINDME01\n")
	_, _ = f2.WriteString("FINDME01\n")
	_, _ = f3.WriteString("NOTTHIS1\n")

	// Fill with random data
	for i := 0; i < 9999; i++ {
		_, _ = f1.WriteString("AAAA0000\n")
		_, _ = f2.WriteString("BBBB1111\n")
		_, _ = f3.WriteString("CCCC2222\n")
	}

	f1.Close()
	f2.Close()
	f3.Close()

	validator := NewValidator()
	if err := validator.LoadFromFiles(context.Background(), []string{file1, file2, file3}); err != nil {
		t.Fatalf("failed to load large files: %v", err)
	}

	// Should find FINDME01 (in file1 and file2)
	if !validator.IsValid(context.Background(), "FINDME01") {
		t.Error("expected FINDME01 to be valid")
	}

	// Should not find NOTTHIS1 (only in file3)
	if validator.IsValid(context.Background(), "NOTTHIS1") {
		t.Error("expected NOTTHIS1 to be invalid")
	}
}
