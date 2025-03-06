## 01F-wget3

```
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

var rateLimit string

func main() {
	// Define and parse the -rate-limit flag
	flag.StringVar(&rateLimit, "rate-limit", "", "Limit download speed (300KB/s or 700KB/s)")
	flag.Parse()

	// Convert rate limit to bytes per second
	limit, err := parseRateLimit(rateLimit)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Example URL (should be passed as an argument instead)
	url := "https://assets.01-edu.org/wgetDataSamples/20MB.zip"
	outputFile := "output.zip"

	// Start downloading with the rate limiter
	err = downloadFileWithLimit(url, outputFile, limit)
	if err != nil {
		log.Fatal(err)
	}
}

// parseRateLimit validates and converts rate limit from KB/s to bytes per second
func parseRateLimit(limitStr string) (int, error) {
	if limitStr != "300KB/s" && limitStr != "700KB/s" {
		return 0, fmt.Errorf("-rate-limit must be either '300KB/s' or '700KB/s'")
	}

	// Extract numeric part (e.g., "300KB/s" -> "300")
	numericPart := strings.TrimSuffix(limitStr, "KB/s")
	limitKB, err := strconv.Atoi(numericPart)
	if err != nil {
		return 0, fmt.Errorf("invalid rate limit value: %s", limitStr)
	}

	// Convert KB/s to Bytes/s
	return limitKB * 1024, nil
}

// downloadFileWithLimit downloads a file while enforcing the speed limit
func downloadFileWithLimit(url, filename string, limit int) error {
	// Create rate limiter (bytes per second)
	limiter := rate.NewLimiter(rate.Limit(limit), limit)

	// Open HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create output file
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Create a limited reader that respects the rate limit
	reader := &rateLimitedReader{
		reader:  resp.Body,
		limiter: limiter,
	}

	// Copy data with rate limiting
	_, err = io.Copy(outFile, reader)
	if err != nil {
		return err
	}

	log.Println("Download completed:", filename)
	return nil
}

// rateLimitedReader wraps an io.Reader and enforces a rate limit
type rateLimitedReader struct {
	reader  io.Reader
	limiter *rate.Limiter
}

func (r *rateLimitedReader) Read(p []byte) (int, error) {
	// Wait for the limiter to allow reading
	err := r.limiter.WaitN(nil, len(p))
	if err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
```
