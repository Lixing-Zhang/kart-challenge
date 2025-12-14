package coupon

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func createGzipResponse(t *testing.T, coupons []string) []byte {
	t.Helper()
	
	var buf strings.Builder
	gz := gzip.NewWriter(&buf)
	
	for _, coupon := range coupons {
		if _, err := gz.Write([]byte(coupon + "\n")); err != nil {
			t.Fatalf("failed to write to gzip: %v", err)
		}
	}
	
	if err := gz.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	
	return []byte(buf.String())
}

func createMockServer(t *testing.T, files map[string][]string) *httptest.Server {
	t.Helper()
	
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		coupons, exists := files[path]
		if !exists {
			http.NotFound(w, r)
			return
		}
		
		data := createGzipResponse(t, coupons)
		w.Header().Set("Content-Type", "application/gzip")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
}

func TestValidator_LoadFromURLs(t *testing.T) {
	t.Run("successful load from multiple files", func(t *testing.T) {
		files := map[string][]string{
			"/file1.gz": {"HAPPYHRS", "FIFTYOFF", "SUPER100"},
			"/file2.gz": {"HAPPYHRS", "FIFTYOFF", "WEEKEND"},
			"/file3.gz": {"HAPPYHRS", "WEEKEND", "SAVE20"},
		}
		server := createMockServer(t, files)
		defer server.Close()
		
		validator := NewValidator()
		urls := []string{
			server.URL + "/file1.gz",
			server.URL + "/file2.gz",
			server.URL + "/file3.gz",
		}
		
		ctx := context.Background()
		err := validator.LoadFromURLs(ctx, urls)
		
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if len(validator.couponSets) != 3 {
			t.Errorf("expected 3 coupon sets, got %d", len(validator.couponSets))
		}
	})
	
	t.Run("empty URLs", func(t *testing.T) {
		validator := NewValidator()
		err := validator.LoadFromURLs(context.Background(), []string{})
		
		if err == nil {
			t.Error("expected error for empty URLs, got nil")
		}
	})
}

func TestValidator_IsValid(t *testing.T) {
	files := map[string][]string{
		"/file1.gz": {"HAPPYHRS", "FIFTYOFF", "SUPER100"},
		"/file2.gz": {"HAPPYHRS", "FIFTYOFF", "WEEKEND"},
		"/file3.gz": {"HAPPYHRS", "WEEKEND", "SAVE20"},
	}
	server := createMockServer(t, files)
	defer server.Close()
	
	validator := NewValidator()
	urls := []string{
		server.URL + "/file1.gz",
		server.URL + "/file2.gz",
		server.URL + "/file3.gz",
	}
	
	ctx := context.Background()
	if err := validator.LoadFromURLs(ctx, urls); err != nil {
		t.Fatalf("failed to load URLs: %v", err)
	}
	
	tests := []struct {
		name     string
		code     string
		expected bool
		reason   string
	}{
		{
			name:     "valid - appears in all 3 files",
			code:     "HAPPYHRS",
			expected: true,
			reason:   "HAPPYHRS appears in all 3 files",
		},
		{
			name:     "valid - appears in 2 files",
			code:     "FIFTYOFF",
			expected: true,
			reason:   "FIFTYOFF appears in file1 and file2",
		},
		{
			name:     "invalid - appears in only 1 file",
			code:     "SUPER100",
			expected: false,
			reason:   "SUPER100 only appears in file1",
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
	files := map[string][]string{
		"/file1.gz": {"HAPPYHRS", "FIFTYOFF"},
		"/file2.gz": {"HAPPYHRS", "WEEKEND"},
		"/file3.gz": {"HAPPYHRS", "SAVE20"},
	}
	server := createMockServer(t, files)
	defer server.Close()
	
	validator := NewValidator()
	urls := []string{
		server.URL + "/file1.gz",
		server.URL + "/file2.gz",
		server.URL + "/file3.gz",
	}
	
	if err := validator.LoadFromURLs(context.Background(), urls); err != nil {
		t.Fatalf("failed to load URLs: %v", err)
	}
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			
			codes := []string{"HAPPYHRS", "FIFTYOFF", "WEEKEND", "NOTEXIST"}
			code := codes[n%len(codes)]
			
			_ = validator.IsValid(context.Background(), code)
		}(i)
	}
	
	wg.Wait()
}

func TestValidator_GetStats(t *testing.T) {
	files := map[string][]string{
		"/file1.gz": {"HAPPYHRS", "FIFTYOFF", "SUPER100"},
		"/file2.gz": {"HAPPYHRS", "FIFTYOFF"},
		"/file3.gz": {"HAPPYHRS"},
	}
	server := createMockServer(t, files)
	defer server.Close()
	
	validator := NewValidator()
	urls := []string{
		server.URL + "/file1.gz",
		server.URL + "/file2.gz",
		server.URL + "/file3.gz",
	}
	
	if err := validator.LoadFromURLs(context.Background(), urls); err != nil {
		t.Fatalf("failed to load URLs: %v", err)
	}
	
	stats := validator.GetStats()
	
	totalFiles, ok := stats["total_files"].(int)
	if !ok || totalFiles != 3 {
		t.Errorf("expected total_files to be 3, got %v", stats["total_files"])
	}
	
	fileSizes, ok := stats["file_sizes"].([]int)
	if !ok {
		t.Fatalf("expected file_sizes to be []int, got %T", stats["file_sizes"])
	}
	
	expectedSizes := []int{3, 2, 1}
	for i, expected := range expectedSizes {
		if fileSizes[i] != expected {
			t.Errorf("expected file %d to have %d coupons, got %d", i, expected, fileSizes[i])
		}
	}
}

func TestValidator_parseCoupons(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
		{
			name:  "single coupon",
			input: "HAPPYHRS\n",
			expected: map[string]bool{
				"HAPPYHRS": true,
			},
		},
		{
			name:  "multiple coupons",
			input: "HAPPYHRS\nFIFTYOFF\nWEEKEND\n",
			expected: map[string]bool{
				"HAPPYHRS": true,
				"FIFTYOFF": true,
				"WEEKEND":  true,
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: map[string]bool{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := parseCoupons(reader)
			
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d coupons, got %d", len(tt.expected), len(result))
			}
			
			for code := range tt.expected {
				if !result[code] {
					t.Errorf("expected coupon %q to be present", code)
				}
			}
		})
	}
}

func BenchmarkValidator_IsValid(b *testing.B) {
	coupons := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		coupons[i] = fmt.Sprintf("COUPON%03d", i)
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		http.NotFound(w, r)
	}))
	defer server.Close()
	
	validator := NewValidator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.IsValid(context.Background(), "COUPON000")
	}
}
