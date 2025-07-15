package repository

// Import
import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PDFRepository interface {
	DownloadAndMerge(urls []string) ([]byte, error)
}

type pdfRepo struct{}

func NewPDFRepository() PDFRepository {
	return &pdfRepo{}
}

func (r *pdfRepo) DownloadAndMerge(urls []string) ([]byte, error) {
	// Create channels for concurrent downloads
	type downloadResult struct {
		index int
		data  []byte
		err   error
	}

	results := make(chan downloadResult, len(urls))
	var wg sync.WaitGroup

	// Start concurrent downloads
	for i, rawURL := range urls {
		wg.Add(1)
		go func(index int, url string) {
			defer wg.Done()

			// First attempt: Try to properly encode the URL
			encodedURL, err := encodeURL(url)
			if err != nil {
				results <- downloadResult{index, nil, fmt.Errorf("invalid URL format %s: %w", url, err)}
				return
			}

			// Try downloading with encoded URL
			pdfData, err := downloadPDF(encodedURL)
			if err != nil {
				// If it's a 400 error, try with manual query parameter encoding
				if isHTTP400Error(err) {
					fmt.Printf("Encoded URL failed with 400, trying manual encoding...\n")
					manualEncodedURL := manualEncodeQueryParams(url)
					fmt.Printf("Manual encoded URL: %s\n", manualEncodedURL)

					pdfData, err = downloadPDF(manualEncodedURL)
					if err != nil {
						// Check if it's a non-PDF content error and skip it
						if isNonPDFError(err) {
							fmt.Printf("Skipping non-PDF content at %s: %v\n", manualEncodedURL, err)
							results <- downloadResult{index, nil, nil} // Skip this URL
							return
						}
						results <- downloadResult{index, nil, fmt.Errorf("download failed for both encoded URLs. Original: %s, Manual: %s - Error: %w", encodedURL, manualEncodedURL, err)}
						return
					}
				} else {
					// Check if it's a non-PDF content error and skip it
					if isNonPDFError(err) {
						fmt.Printf("Skipping non-PDF content at %s: %v\n", encodedURL, err)
						results <- downloadResult{index, nil, nil} // Skip this URL
						return
					}
					fmt.Printf("Downloading \n %s", encodedURL)
					results <- downloadResult{index, nil, fmt.Errorf("download failed for %s: %w", encodedURL, err)}
					return
				}
			}

			results <- downloadResult{index, pdfData, nil}
		}(i, rawURL)
	}

	// Wait for all downloads to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	pdfDataSlice := make([][]byte, 0, len(urls)) // Use slice with capacity but start empty
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		// Skip nil data (non-PDF content that was skipped)
		if result.data != nil {
			pdfDataSlice = append(pdfDataSlice, result.data)
		}
	}

	// Convert to ReadSeeker for merging
	var pdfBuffers []io.ReadSeeker
	for _, data := range pdfDataSlice {
		pdfBuffers = append(pdfBuffers, bytes.NewReader(data))
	}

	// Check if we have any PDFs to merge
	if len(pdfBuffers) == 0 {
		return nil, fmt.Errorf("no valid PDF files found to merge")
	}

	var mergedBuf bytes.Buffer
	conf := model.NewDefaultConfiguration()

	// ✅ In-memory merge using MergeRaw
	if err := api.MergeRaw(pdfBuffers, &mergedBuf, false, conf); err != nil {
		return nil, fmt.Errorf("merge failed: %w", err)
	}

	return mergedBuf.Bytes(), nil
}

// encodeURL properly encodes a URL
func encodeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return parsedURL.String(), nil
}

// manualEncodeQueryParams manually encodes query parameters in the URL
func manualEncodeQueryParams(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL // Return original if parsing fails
	}

	// Parse and re-encode query parameters
	values, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		return rawURL // Return original if query parsing fails
	}

	// Re-encode all query parameters
	parsedURL.RawQuery = values.Encode()
	return parsedURL.String()
}

// downloadPDF downloads and validates a PDF from URL
func downloadPDF(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed for %s: HTTP %d", url, resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/pdf" {
		return nil, fmt.Errorf("invalid content-type for %s: %s", url, ct)
	}

	// Check for PDF header
	header := make([]byte, 5)
	_, err = io.ReadFull(resp.Body, header)
	if err != nil || string(header) != "%PDF-" {
		return nil, fmt.Errorf("not a valid PDF at %s", url)
	}

	var buf bytes.Buffer
	buf.Write(header)
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("error reading PDF from %s: %w", url, err)
	}

	return buf.Bytes(), nil
}

// isHTTP400Error checks if the error is an HTTP 400 error
func isHTTP400Error(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "HTTP 400")
}

// isNonPDFError checks if the error is related to non-PDF content
func isNonPDFError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "invalid content-type") ||
		strings.Contains(errStr, "not a valid PDF")
}
