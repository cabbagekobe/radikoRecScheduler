package internal

import (
	"context"
	"fmt"
	"net/http" // Removed io import
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/briandowns/spinner"
	goradiko "github.com/yyoshiki41/go-radiko"
)

// Helper function to create a mock Radiko client for testing
func createMockRadikoClient(t *testing.T) *goradiko.Client {
	client, err := goradiko.New("dummy_auth_token") // Auth token won't matter for mock server
	if err != nil {
		t.Fatalf("Failed to create mock Radiko client: %v", err)
	}
	return client
}

func TestBulkDownload(t *testing.T) {
	// Save original HTTP client and defer its restoration
	originalHTTPClient := http.DefaultClient
	defer goradiko.SetHTTPClient(originalHTTPClient)

	// Create a temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "bulk-download-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock HTTP server to serve chunks
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".aac") {
			w.WriteHeader(http.StatusOK)
			// Serve a small dummy AAC content
			_, _ = w.Write([]byte("DUMMY AAC CHUNK CONTENT"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Set the global goradiko HTTP client to use the mock server's client
	goradiko.SetHTTPClient(mockServer.Client())

	// Create a mock Radiko client (its internal http.Client is now the mock server's)
	mockClient := createMockRadikoClient(t)

	// Prepare a chunklist with URLs from the mock server
	chunklist := []string{
		fmt.Sprintf("%s/chunk1.aac", mockServer.URL),
		fmt.Sprintf("%s/chunk2.aac", mockServer.URL),
		fmt.Sprintf("%s/chunk3.aac", mockServer.URL),
	}

	ctx := context.Background()
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Mock spinner
	s.Start()
	defer s.Stop()

	downloadedFiles, err := bulkDownload(ctx, mockClient, chunklist, tempDir, s)
	if err != nil {
		t.Fatalf("bulkDownload failed: %v", err)
	}

	// Verify downloads
	if len(downloadedFiles) != len(chunklist) {
		t.Errorf("Expected %d files, got %d", len(chunklist), len(downloadedFiles))
	}

	for _, file := range downloadedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Downloaded file %s does not exist", file)
		}
		content, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("Failed to read downloaded file %s: %v", file, err)
		}
		if string(content) != "DUMMY AAC CHUNK CONTENT" {
			t.Errorf("Downloaded file %s has wrong content: %s", file, string(content))
		}
	}
}

func TestConcatAACFiles(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "concat-aac-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create dummy input AAC files
	inputFiles := make([]string, 3)
	expectedContent := ""
	for i := 0; i < 3; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("input_%d.aac", i))
		content := fmt.Sprintf("CHUNK_%d", i+1)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write dummy input file %s: %v", filePath, err)
		}
		inputFiles[i] = filePath
		expectedContent += content
	}

	outputFile := filepath.Join(tempDir, "output.aac")

	err = concatAACFiles(inputFiles, outputFile)
	if err != nil {
		t.Fatalf("concatAACFiles failed: %v", err)
	}

	// Verify output file
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file %s does not exist", outputFile)
	}

	actualContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file %s: %v", outputFile, err)
	}

	if string(actualContent) != expectedContent {
		t.Errorf("Concatenated content is wrong. Got '%s', want '%s'", string(actualContent), expectedContent)
	}

	// Test case for non-existent input file
	nonExistentInputFiles := []string{filepath.Join(tempDir, "non_existent.aac")}
	err = concatAACFiles(nonExistentInputFiles, filepath.Join(tempDir, "error_output.aac"))
	if err == nil {
		t.Error("concatAACFiles did not return an error for non-existent input file")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to open input file") {
		t.Errorf("concatAACFiles returned wrong error type for non-existent input: %v", err)
	}

	// Test case for output file creation error (e.g., permissions)
	readOnlyDir := filepath.Join(tempDir, "read-only")
	if err := os.Mkdir(readOnlyDir, 0444); err != nil { // Create read-only directory
		t.Fatalf("Failed to create read-only dir: %v", err)
	}
	defer os.RemoveAll(readOnlyDir)
	
	err = concatAACFiles(inputFiles, filepath.Join(readOnlyDir, "output.aac"))
	if err == nil {
		t.Error("concatAACFiles did not return an error for output file creation failure")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to create output file") {
		t.Errorf("concatAACFiles returned wrong error type for output creation failure: %v", err)
	}
}
