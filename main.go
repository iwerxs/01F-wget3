package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/time/rate"
)

func parseRateLimit(limitStr string) (int, error) {
	if limitStr == "" {
		return 0, nil // No limit applied
	}

	validLimits := map[string]int{
		"300k": 300 * 1024,
		"700k": 700 * 1024,
		"2M":   2 * 1024 * 1024,
	}

	limit, exists := validLimits[limitStr]
	if !exists {
		return 0, fmt.Errorf("-rate-limit must be either '300k' or '700k' or '2M'")
	}
	return limit, nil
}

func downloadFile(url string, outputPath string, limit int) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	var reader io.Reader = resp.Body
	if limit > 0 {
		limiter := rate.NewLimiter(rate.Limit(limit), limit)
		reader = &rateLimitedReader{reader: resp.Body, limiter: limiter}
	}

	_, err = io.Copy(outFile, reader)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}

	fmt.Println("Downloaded:", outputPath)
	return nil
}

type rateLimitedReader struct {
	reader  io.Reader
	limiter *rate.Limiter
}

func (r *rateLimitedReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err != nil {
		return n, err
	}
	r.limiter.WaitN(context.Background(), n)
	return n, nil
}


// Main function
func main() {
	var mirrorFlag bool
	var convertLinks bool
	var rejectType string
	var acceptType string
	var recursive bool
	var rateLimit string

	flag.BoolVar(&mirrorFlag, "mirror", false, "Mirror the remote directory structure")
	flag.BoolVar(&convertLinks, "convert-links", false, "Convert links for offline viewing")
	flag.StringVar(&rejectType, "reject", "", "Comma-separated list of file types to reject")
	flag.StringVar(&acceptType, "accept", "", "Comma-separated list of file types to accept")
	flag.BoolVar(&recursive, "recursive", false, "Download directories recursively")
	flag.StringVar(&rateLimit, "rate-limit", "", "Limit download speed (300k or 700k or 2M)")

	flag.Parse()

	// Ensure a URL is provided
	if flag.NArg() < 1 {
		fmt.Println("Usage: wget <URL> [OPTIONS]")
		os.Exit(1)
	}

	url := flag.Arg(0)
	outputPath := filepath.Base(url) // Save with the same filename as URL if no output specified

	// Parse rate-limit value
	limit, err := parseRateLimit(rateLimit)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Download file with rate limiting
	err = downloadFile(url, outputPath, limit)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Download completed:", outputPath)
}