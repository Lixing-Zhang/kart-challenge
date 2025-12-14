package coupon

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Validator validates coupon codes against multiple coupon files
type Validator struct {
	couponSets []*couponSet
	mu         sync.RWMutex
}

// couponSet represents a set of coupons loaded from a single file
type couponSet struct {
	coupons map[string]bool
	mu      sync.RWMutex
}

// fileLoadResult holds the result of loading a single file
type fileLoadResult struct {
	index   int
	coupons map[string]bool
	err     error
}

// NewValidator creates a new coupon validator
func NewValidator() *Validator {
	return &Validator{
		couponSets: make([]*couponSet, 0),
	}
}

// LoadFromURLs loads coupon data from multiple gzipped URLs concurrently
// Returns error if any file fails to load
func (v *Validator) LoadFromURLs(ctx context.Context, urls []string) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	// Create channels for results and errors
	resultChan := make(chan fileLoadResult, len(urls))

	// Create a WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Launch a goroutine for each URL
	for i, url := range urls {
		wg.Add(1)
		go func(index int, fileURL string) {
			defer wg.Done()

			coupons, err := v.loadFromURL(ctx, fileURL)
			resultChan <- fileLoadResult{
				index:   index,
				coupons: coupons,
				err:     err,
			}
		}(i, url)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results maintaining order
	results := make([]fileLoadResult, len(urls))
	for result := range resultChan {
		results[result.index] = result
	}

	// Check for errors
	for i, result := range results {
		if result.err != nil {
			return fmt.Errorf("failed to load file %d: %w", i+1, result.err)
		}
	}

	// Store the loaded coupon sets
	v.mu.Lock()
	defer v.mu.Unlock()

	v.couponSets = make([]*couponSet, len(results))
	for i, result := range results {
		v.couponSets[i] = &couponSet{
			coupons: result.coupons,
		}
	}

	return nil
}

// loadFromURL downloads and parses a gzipped coupon file from a URL
func (v *Validator) loadFromURL(ctx context.Context, url string) (map[string]bool, error) {
	// Create HTTP client with longer timeout for large files
	client := &http.Client{
		Timeout: 5 * time.Minute, // Large files (~600MB) need more time
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create gzip reader
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Parse coupons from the file
	return parseCoupons(gzReader)
}

// parseCoupons reads coupons from a reader and returns them as a set
func parseCoupons(r io.Reader) (map[string]bool, error) {
	coupons := make(map[string]bool)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			coupons[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return coupons, nil
}

// IsValid checks if a coupon code is valid
// A coupon is valid if:
// 1. It has 8-10 characters
// 2. It appears in at least 2 of the loaded files
func (v *Validator) IsValid(ctx context.Context, code string) bool {
	// Validate length (8-10 characters)
	if len(code) < 8 || len(code) > 10 {
		return false
	}

	v.mu.RLock()
	defer v.mu.RUnlock()

	// If no coupon sets are loaded, the coupon is invalid
	if len(v.couponSets) == 0 {
		return false
	}

	// Use channels for concurrent validation
	foundChan := make(chan bool, len(v.couponSets))

	// Create context with cancellation for early termination
	searchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Search concurrently in all coupon sets
	var wg sync.WaitGroup
	for _, cs := range v.couponSets {
		wg.Add(1)
		go func(couponSet *couponSet) {
			defer wg.Done()

			// Check if context is cancelled
			select {
			case <-searchCtx.Done():
				return
			default:
			}

			couponSet.mu.RLock()
			found := couponSet.coupons[code]
			couponSet.mu.RUnlock()

			if found {
				select {
				case foundChan <- true:
				case <-searchCtx.Done():
				}
			}
		}(cs)
	}

	// Close channel when all searches complete
	go func() {
		wg.Wait()
		close(foundChan)
	}()

	// Count how many files contain the coupon
	count := 0
	for range foundChan {
		count++
		// Early termination: if we found it in 2 files, we can stop
		if count >= 2 {
			cancel() // Cancel remaining searches
			break
		}
	}

	return count >= 2
}

// GetStats returns statistics about loaded coupons
func (v *Validator) GetStats() map[string]interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_files"] = len(v.couponSets)

	fileSizes := make([]int, len(v.couponSets))
	totalCoupons := 0

	for i, cs := range v.couponSets {
		cs.mu.RLock()
		size := len(cs.coupons)
		cs.mu.RUnlock()

		fileSizes[i] = size
		totalCoupons += size
	}

	stats["file_sizes"] = fileSizes
	stats["total_coupons"] = totalCoupons

	return stats
}
