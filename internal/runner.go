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

	"github.com/briandowns/spinner"            // Import spinner
	goradiko "github.com/yyoshiki41/go-radiko" // Alias to avoid conflict with our internal package name
)

type RadikoClient interface {
	AuthorizeToken(ctx context.Context) (string, error)
	TimeshiftPlaylistM3U8(ctx context.Context, stationID string, pastTime time.Time) (string, error)
	GetChunklistFromM3U8(uri string) ([]string, error)
	Do(req *http.Request) (*http.Response, error) // For bulkDownload
}

// Concrete goradiko client wrapper
type goradikoClient struct {
	client *goradiko.Client
}

func NewGoradikoClient(token string) (RadikoClient, error) {
	client, err := goradiko.New(token)
	if err != nil {
		return nil, err
	}
	return &goradikoClient{client: client}, nil
}

func (g *goradikoClient) AuthorizeToken(ctx context.Context) (string, error) {
	return g.client.AuthorizeToken(ctx)
}

func (g *goradikoClient) TimeshiftPlaylistM3U8(ctx context.Context, stationID string, pastTime time.Time) (string, error) {
	return g.client.TimeshiftPlaylistM3U8(ctx, stationID, pastTime)
}

func (g *goradikoClient) GetChunklistFromM3U8(uri string) ([]string, error) {
	return goradiko.GetChunklistFromM3U8(uri)
}

func (g *goradikoClient) Do(req *http.Request) (*http.Response, error) {
	return g.client.Do(req)
}

// ExecuteJob runs the recording process for a given schedule entry and time.
// It now accepts a RadikoClient interface for dependency injection.
func ExecuteJob(radikoClient RadikoClient, entry ScheduleEntry, pastTime time.Time, outputDir string) error {
	log.Printf("INFO: Starting recording for: %s (%s) for past broadcast at %s", entry.ProgramName, entry.StationID, pastTime.Format("2006-01-02 15:04:05"))

	ctx := context.Background()

	// 1. Authenticate to get the auth token
	log.Println("INFO: Authorizing Radiko token...")
	_, err := radikoClient.AuthorizeToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to authorize Radiko token: %w", err)
	}
	// Re-initialize client with obtained token, if the client supports it.
	// For our goradikoClient wrapper, this step would be handled internally if needed,
	// or more directly by the caller providing a client already authenticated.
	// For the current goradiko.Client design, the token is passed to constructor.
	// We'll assume the provided radikoClient is already capable of using the token or
	// handles internal re-initialization if AuthorizeToken sets internal state.
	// For testing, this allows us to mock the token directly.
	log.Println("INFO: Radiko token authorized successfully.")

	// 2. Get M3U8 Playlist URI
	log.Println("INFO: Getting M3U8 playlist URI...")
	uri, err := radikoClient.TimeshiftPlaylistM3U8(ctx, entry.StationID, pastTime)
	if err != nil {
		return fmt.Errorf("failed to get timeshift M3U8 playlist URI for %s: %w", entry.ProgramName, err)
	}
	log.Printf("INFO: Got M3U8 URI: %s", uri)

	// 3. Get Chunklist from M3U8 (from go-radiko package)
	log.Println("INFO: Getting chunklist from M3U8...")
	chunklist, err := radikoClient.GetChunklistFromM3U8(uri)
	if err != nil {
		return fmt.Errorf("failed to get chunklist from M3U8 for %s: %w", entry.ProgramName, err)
	}
	log.Printf("INFO: Found %d audio chunks.", len(chunklist))

	// 4. Create a temporary directory for downloading AAC chunks
	tempDir, err := os.MkdirTemp("", "radigo-chunks-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() {
		log.Printf("INFO: Cleaning up temporary directory: %s", tempDir)
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("WARNING: Failed to remove temporary directory '%s': %v", tempDir, err)
		}
	}()
	log.Printf("INFO: Created temporary directory: %s", tempDir)

	// 5. Bulk download AAC files with progress spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Build our new spinner
	s.Suffix = fmt.Sprintf(" Downloading %d chunks...", len(chunklist))
	s.Start() // Start the spinner

	downloadedFiles, err := bulkDownload(ctx, radikoClient, chunklist, tempDir, s)
	if err != nil {
		s.Stop() // Stop spinner on error
		return fmt.Errorf("failed to bulk download AAC chunks for %s: %w", entry.ProgramName, err)
	}
	s.Stop() // Stop spinner on success
	log.Printf("INFO: Successfully downloaded %d AAC chunks.", len(downloadedFiles))

	// 6. Concatenate AAC files
	log.Println("INFO: Concatenating AAC files...")
	// Output directory check
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
	log.Printf("INFO: Successfully recorded and saved to: %s", outputFilePath)

	return nil
}

// bulkDownload downloads a list of URLs to a specified directory.
// It returns the list of paths to the downloaded files.
func bulkDownload(ctx context.Context, client RadikoClient, urls []string, destDir string, s *spinner.Spinner) ([]string, error) {
	downloadedFiles := make([]string, 0, len(urls))
	for i, url := range urls {
		s.Suffix = fmt.Sprintf(" Downloading chunk %d/%d...", i+1, len(urls)) // Update spinner suffix
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
			return nil, fmt.Errorf("failed to save chunk %d to file: %s: %w", i, url, err)
		}
		downloadedFiles = append(downloadedFiles, filePath)
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
	}
	log.Printf("INFO: Finished concatenating %d files.", len(inputFiles))
	return nil
}
