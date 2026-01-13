package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/briandowns/spinner"
)

// MockRadikoClient is a mock implementation of the RadikoClient interface for testing.
type MockRadikoClient struct {
	AuthTokenFn             func(ctx context.Context) (string, error)
	TimeshiftPlaylistM3U8Fn func(ctx context.Context, stationID string, pastTime time.Time) (string, error)
	GetChunklistFromM3U8Fn  func(uri string) ([]string, error)
	DoFn                    func(req *http.Request) (*http.Response, error)
}

func (m *MockRadikoClient) AuthorizeToken(ctx context.Context) (string, error) {
	if m.AuthTokenFn != nil {
		return m.AuthTokenFn(ctx)
	}
	return "mock_auth_token", nil // Default success
}

func (m *MockRadikoClient) TimeshiftPlaylistM3U8(ctx context.Context, stationID string, pastTime time.Time) (string, error) {
	if m.TimeshiftPlaylistM3U8Fn != nil {
		return m.TimeshiftPlaylistM3U8Fn(ctx, stationID, pastTime)
	}
	return "http://mock.m3u8/playlist.m3u8", nil // Default success
}

func (m *MockRadikoClient) GetChunklistFromM3U8(uri string) ([]string, error) {
	if m.GetChunklistFromM3U8Fn != nil {
		return m.GetChunklistFromM3U8Fn(uri)
	}
	return []string{"http://mock.chunk/chunk1.aac", "http://mock.chunk/chunk2.aac"}, nil // Default success
}

func (m *MockRadikoClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFn != nil {
		return m.DoFn(req)
	}
	// Default mock HTTP response for successful download
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("DUMMY AAC CHUNK CONTENT")),
	}, nil
}

func TestExecuteJob(t *testing.T) {
	mockNow := time.Date(2026, time.January, 13, 10, 0, 0, 0, JST) // Tuesday

	tests := []struct {
		name          string
		mockClient    *MockRadikoClient
		entry         ScheduleEntry
		pastTime      time.Time
		outputDir     string
		expectError   bool
		expectedError string
	}{
		{
			name: "Successful execution",
			mockClient: &MockRadikoClient{
				AuthTokenFn: func(ctx context.Context) (string, error) { return "mock_auth_token", nil },
				TimeshiftPlaylistM3U8Fn: func(ctx context.Context, stationID string, pastTime time.Time) (string, error) {
					return "http://mock.m3u8/playlist.m3u8", nil
				},
				GetChunklistFromM3U8Fn: func(uri string) ([]string, error) {
					return []string{
						"http://mock.chunk/chunk1.aac",
						"http://mock.chunk/chunk2.aac",
					}, nil
				},
				DoFn: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader("DUMMY AAC CHUNK CONTENT")),
					}, nil
				},
			},
			entry: ScheduleEntry{
				ProgramName: "Test Program",
				StationID:   "ST1",
			},
			pastTime:    mockNow.Add(-24 * time.Hour), // Monday
			outputDir:   "output",
			expectError: false,
		},
		{
			name: "Authentication failure",
			mockClient: &MockRadikoClient{
				AuthTokenFn: func(ctx context.Context) (string, error) { return "", fmt.Errorf("auth failed") },
			},
			entry: ScheduleEntry{
				ProgramName: "Test Program",
				StationID:   "ST1",
			},
			pastTime:      mockNow,
			outputDir:     "output",
			expectError:   true,
			expectedError: "failed to authorize Radiko token: auth failed",
		},
		{
			name: "TimeshiftPlaylistM3U8 failure",
			mockClient: &MockRadikoClient{
				AuthTokenFn: func(ctx context.Context) (string, error) { return "mock_auth_token", nil },
				TimeshiftPlaylistM3U8Fn: func(ctx context.Context, stationID string, pastTime time.Time) (string, error) {
					return "", fmt.Errorf("m3u8 failed")
				},
			},
			entry: ScheduleEntry{
				ProgramName: "Test Program",
				StationID:   "ST1",
			},
			pastTime:      mockNow,
			outputDir:     "output",
			expectError:   true,
			expectedError: "failed to get timeshift M3U8 playlist URI for Test Program: m3u8 failed",
		},
		{
			name: "GetChunklistFromM3U8 failure",
			mockClient: &MockRadikoClient{
				AuthTokenFn: func(ctx context.Context) (string, error) { return "mock_auth_token", nil },
				TimeshiftPlaylistM3U8Fn: func(ctx context.Context, stationID string, pastTime time.Time) (string, error) {
					return "http://mock.m3u8/playlist.m3u8", nil
				},
				GetChunklistFromM3U8Fn: func(uri string) ([]string, error) {
					return nil, fmt.Errorf("chunklist failed")
				},
			},
			entry: ScheduleEntry{
				ProgramName: "Test Program",
				StationID:   "ST1",
			},
			pastTime:      mockNow,
			outputDir:     "output",
			expectError:   true,
			expectedError: "failed to get chunklist from M3U8 for Test Program: chunklist failed",
		},
		{
			name: "Bulk download failure (HTTP error)",
			mockClient: &MockRadikoClient{
				AuthTokenFn: func(ctx context.Context) (string, error) { return "mock_auth_token", nil },
				TimeshiftPlaylistM3U8Fn: func(ctx context.Context, stationID string, pastTime time.Time) (string, error) {
					return "http://mock.m3u8/playlist.m3u8", nil
				},
				GetChunklistFromM3U8Fn: func(uri string) ([]string, error) {
					return []string{"http://mock.chunk/chunk1.aac"}, nil
				},
				DoFn: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(strings.NewReader("")),
					}, nil
				},
			},
			entry: ScheduleEntry{
				ProgramName: "Test Program",
				StationID:   "ST1",
			},
			pastTime:      mockNow,
			outputDir:     "output",
			expectError:   true,
			expectedError: "failed to bulk download AAC chunks for Test Program: failed to download chunk 0 (http://mock.chunk/chunk1.aac): HTTP status 500",
		},
		{
			name: "Bulk download failure (network error)",
			mockClient: &MockRadikoClient{
				AuthTokenFn: func(ctx context.Context) (string, error) { return "mock_auth_token", nil },
				TimeshiftPlaylistM3U8Fn: func(ctx context.Context, stationID string, pastTime time.Time) (string, error) {
					return "http://mock.m3u8/playlist.m3u8", nil
				},
				GetChunklistFromM3U8Fn: func(uri string) ([]string, error) {
					return []string{"http://mock.chunk/chunk1.aac"}, nil
				},
				DoFn: func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("network error")
				},
			},
			entry: ScheduleEntry{
				ProgramName: "Test Program",
				StationID:   "ST1",
			},
			pastTime:      mockNow,
			outputDir:     "output",
			expectError:   true,
			expectedError: "failed to bulk download AAC chunks for Test Program: failed to download chunk 0 (http://mock.chunk/chunk1.aac): network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary output directory for each test
			tempOutputDir, err := os.MkdirTemp("", "test-output-")
			if err != nil {
				t.Fatalf("Failed to create temp output dir: %v", err)
			}
			defer os.RemoveAll(tempOutputDir)

			err = ExecuteJob(tt.mockClient, tt.entry, tt.pastTime, tempOutputDir)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error for %s, but got none", tt.name)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("for %s, expected error containing '%s', but got '%v'", tt.name, tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error for %s, but got: %v", tt.name, err)
				}
				// Verify output file exists
				expectedFileName := fmt.Sprintf("%s-%s-%s.aac", tt.pastTime.Format("20060102150405"), tt.entry.StationID, tt.entry.ProgramName)
				outputFilePath := filepath.Join(tempOutputDir, expectedFileName)
				if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
					t.Errorf("expected output file %s to exist, but it did not", outputFilePath)
				}

			}
		})
	}
}

// TestBulkDownload uses MockRadikoClient now
func TestBulkDownload(t *testing.T) {
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

	// Create a MockRadikoClient that uses the mockServer.Client() for DoFn
	mockClient := &MockRadikoClient{
		DoFn: func(req *http.Request) (*http.Response, error) {
			return mockServer.Client().Do(req)
		},
	}

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
