package internal

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	goradiko "github.com/yyoshiki41/go-radiko" // Alias to avoid conflict with our internal package name
)

// ExecuteJob runs the recording process for a given schedule entry and time.
func ExecuteJob(entry ScheduleEntry, pastTime time.Time) error {
	log.Printf("Starting recording for: %s (%s) for past broadcast at %s", entry.ProgramName, entry.StationID, pastTime.Format("2006-01-02 15:04:05"))

	ctx := context.Background()

	// 1. Authenticate to get the auth token
	radikoClient, err := goradiko.New("") // Initialize with empty token, it will be set by AuthorizeToken
	if err != nil {
		return fmt.Errorf("failed to create Radiko client: %w", err)
	}
	authToken, err := radikoClient.AuthorizeToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to authorize Radiko token: %w", err)
	}
	radikoClient, err = goradiko.New(authToken) // Re-initialize client with obtained token
	if err != nil {
		return fmt.Errorf("failed to create Radiko client with auth token: %w", err)
	}

	// 2. Get M3U8 Playlist URI
	uri, err := radikoClient.TimeshiftPlaylistM3U8(ctx, entry.StationID, pastTime)
	if err != nil {
		return fmt.Errorf("failed to get timeshift M3U8 playlist URI for %s: %w", entry.ProgramName, err)
	}
	log.Printf("Got M3U8 URI: %s", uri)

	// 3. Get Chunklist from M3U8 (from go-radiko package)
	chunklist, err := goradiko.GetChunklistFromM3U8(uri)
	if err != nil {
		return fmt.Errorf("failed to get chunklist from M3U8 for %s: %w", entry.ProgramName, err)
	}
	log.Printf("Found %d audio chunks.", len(chunklist))

	// 4. Create a temporary directory for downloading AAC chunks
	tempDir, err := os.MkdirTemp("", "radigo-chunks-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() {
		log.Printf("Cleaning up temporary directory: %s", tempDir)
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Warning: Failed to remove temporary directory '%s': %v", tempDir, err)
		}
	}()
	log.Printf("Created temporary directory: %s", tempDir)

	// 5. Bulk download AAC files
	downloadedFiles, err := bulkDownload(ctx, radikoClient, chunklist, tempDir)
	if err != nil {
		return fmt.Errorf("failed to bulk download AAC chunks for %s: %w", entry.ProgramName, err)
	}
	log.Printf("Successfully downloaded %d AAC chunks.", len(downloadedFiles))

	// 6. Concatenate AAC files
	// Output directory check - assuming "output" directory in project root
	outputDir := "output" // This should probably be configurable
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
		}
	}
	
	outputFileName := fmt.Sprintf("%s-%s-%s.aac", pastTime.Format("20060102150405"), entry.StationID, entry.ProgramName)
	outputFilePath := filepath.Join(outputDir, outputFileName)

	if err := concatAACFiles(downloadedFiles, outputFilePath); err != nil {
		return fmt.Errorf("failed to concatenate AAC files for %s: %w", entry.ProgramName, err)
	}
	log.Printf("Successfully recorded and saved to: %s", outputFilePath)

	return nil
}

// bulkDownload downloads a list of URLs to a specified directory.
// It returns the list of paths to the downloaded files.
func bulkDownload(ctx context.Context, client *goradiko.Client, urls []string, destDir string) ([]string, error) {
	downloadedFiles := make([]string, 0, len(urls))
	for i, url := range urls {
		fileName := fmt.Sprintf("chunk_%04d.aac", i)
		filePath := filepath.Join(destDir, fileName)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for chunk %d (%s): %w", i, url, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to download chunk %d (%s): %w", i, url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download chunk %d (%s): HTTP status %d", i, url, resp.StatusCode)
		}

		file, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file for chunk %d: %w", i, err)
		}
		defer file.Close()

		if _, err := io.Copy(file, resp.Body); err != nil {
			return nil, fmt.Errorf("failed to save chunk %d to file: %w", i, err)
		}
		downloadedFiles = append(downloadedFiles, filePath)
		log.Printf("Downloaded chunk %d: %s", i, fileName)
	}
	return downloadedFiles, nil
}

// concatAACFiles concatenates multiple AAC files into a single output file.
func concatAACFiles(inputFiles []string, outputFile string) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file '%s': %w", outputFile, err)
	}
	defer outFile.Close()

	for _, inFile := range inputFiles {
		srcFile, err := os.Open(inFile)
		if err != nil {
			return fmt.Errorf("failed to open input file '%s': %w", inFile, err)
		}
		defer srcFile.Close() // Defer inside loop, but be careful with many files

		if _, err := io.Copy(outFile, srcFile); err != nil {
			return fmt.Errorf("failed to concatenate file '%s': %w", inFile, err)
		}
		log.Printf("Concatenated: %s", inFile)
	}
	return nil
}
