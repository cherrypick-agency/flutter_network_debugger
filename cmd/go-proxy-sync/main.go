package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

const version = "1.0.0"

var (
	// Global flags
	serverURL  = flag.String("server", "http://localhost:9092", "Network Debugger server URL")
	scriptID   = flag.String("script-id", "", "Script ID (required)")
	projectDir = flag.String("dir", ".", "Project directory")
	verbose    = flag.Bool("v", false, "Verbose output")
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	flag.CommandLine.Parse(os.Args[2:])

	// Validate required flags
	if command != "version" && command != "help" && *scriptID == "" {
		fmt.Fprintln(os.Stderr, "Error: --script-id is required")
		os.Exit(1)
	}

	switch command {
	case "upload":
		if err := uploadProject(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "download":
		if err := downloadProject(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "watch":
		if err := watchProject(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "files":
		if err := listFiles(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Printf("go-proxy-sync version %s\n", version)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`go-proxy-sync - Sync script projects with Network Debugger

Usage:
  go-proxy-sync <command> [options]

Commands:
  upload     Upload project directory as ZIP to server
  download   Download project from server
  watch      Watch directory for changes and auto-upload
  files      List files in the script project
  version    Show version information
  help       Show this help message

Options:
  --server string      Server URL (default: http://localhost:9092)
  --script-id string   Script ID (required)
  --dir string         Project directory (default: .)
  -v                   Verbose output

Examples:
  # Upload a Rust project
  go-proxy-sync upload --script-id=abc123 --dir=./my-rust-project

  # Download a project
  go-proxy-sync download --script-id=abc123 --dir=./downloaded-project

  # Watch for changes (auto-upload)
  go-proxy-sync watch --script-id=abc123 --dir=./my-project

  # List project files
  go-proxy-sync files --script-id=abc123

IDE Plugin Integration:
  This tool is designed for IDE plugins (VSCode, IntelliJ, etc.) to sync
  script projects with the Network Debugger for compilation and testing.
`)
}

// uploadProject uploads the project directory as a ZIP file
func uploadProject() error {
	log("Uploading project from: %s", *projectDir)

	// Create ZIP in memory
	zipBuf, err := createProjectZip(*projectDir)
	if err != nil {
		return fmt.Errorf("failed to create ZIP: %w", err)
	}

	log("Created ZIP archive (%d bytes)", zipBuf.Len())

	// Upload to server
	url := fmt.Sprintf("%s/_api/v1/scripts/%s/upload-project", *serverURL, *scriptID)

	// Create multipart request
	body := &bytes.Buffer{}

	req, err := http.NewRequest("POST", url, bytes.NewReader(zipBuf.Bytes()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Write ZIP as form file
	body.Write([]byte("--boundary123\r\n"))
	body.Write([]byte("Content-Disposition: form-data; name=\"project\"; filename=\"project.zip\"\r\n"))
	body.Write([]byte("Content-Type: application/zip\r\n\r\n"))
	body.Write(zipBuf.Bytes())
	body.Write([]byte("\r\n--boundary123--\r\n"))

	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary123")
	req.Body = io.NopCloser(body)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("✓ Upload successful!\n")
	fmt.Printf("  Files uploaded: %v\n", result["fileCount"])
	fmt.Printf("  Total size: %v bytes\n", result["totalSize"])

	if files, ok := result["files"].([]interface{}); ok && len(files) > 0 {
		fmt.Println("\nUploaded files:")
		for _, f := range files {
			fmt.Printf("  - %v\n", f)
		}
	}

	return nil
}

// downloadProject downloads the project as a ZIP and extracts it
func downloadProject() error {
	log("Downloading project to: %s", *projectDir)

	url := fmt.Sprintf("%s/_api/v1/scripts/%s/download-project", *serverURL, *scriptID)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read ZIP into memory
	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	log("Downloaded ZIP archive (%d bytes)", len(zipData))

	// Extract ZIP
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to read ZIP: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(*projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fileCount := 0
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		// Extract file
		outPath := filepath.Join(*projectDir, file.Name)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("failed to create dir for %s: %w", file.Name, err)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", file.Name, err)
		}

		outFile, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create %s: %w", outPath, err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to write %s: %w", outPath, err)
		}

		log("Extracted: %s", file.Name)
		fileCount++
	}

	fmt.Printf("✓ Download successful!\n")
	fmt.Printf("  Files extracted: %d\n", fileCount)
	fmt.Printf("  Output directory: %s\n", *projectDir)

	return nil
}

// watchProject watches the directory for changes and auto-uploads
func watchProject() error {
	fmt.Printf("Watching directory: %s\n", *projectDir)
	fmt.Println("Press Ctrl+C to stop")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Add directory and subdirectories to watcher
	if err := filepath.Walk(*projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden directories and build outputs
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "target" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	// Debounce uploads (don't upload on every single file change)
	var uploadTimer *time.Timer
	debounce := 2 * time.Second

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only watch for Write and Create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// Skip non-source files
				if !isSourceFile(event.Name) {
					continue
				}

				fmt.Printf("File changed: %s\n", filepath.Base(event.Name))

				// Reset timer
				if uploadTimer != nil {
					uploadTimer.Stop()
				}

				uploadTimer = time.AfterFunc(debounce, func() {
					fmt.Println("\nUploading changes...")
					if err := uploadProject(); err != nil {
						fmt.Fprintf(os.Stderr, "Upload failed: %v\n", err)
					} else {
						fmt.Println()
					}
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "Watch error: %v\n", err)
		}
	}
}

// listFiles lists all files in the script project
func listFiles() error {
	url := fmt.Sprintf("%s/_api/v1/scripts/%s/files", *serverURL, *scriptID)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		ScriptID  string `json:"scriptId"`
		FileCount int    `json:"fileCount"`
		Files     []struct {
			Filename string `json:"filename"`
			Size     int    `json:"size"`
			Type     string `json:"type"`
		} `json:"files"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("Script ID: %s\n", result.ScriptID)
	fmt.Printf("Total files: %d\n\n", result.FileCount)

	if len(result.Files) == 0 {
		fmt.Println("No files in project")
		return nil
	}

	fmt.Println("Files:")
	for _, file := range result.Files {
		fmt.Printf("  %-40s %8d bytes  (%s)\n", file.Filename, file.Size, file.Type)
	}

	return nil
}

// createProjectZip creates a ZIP archive of the project directory
func createProjectZip(dir string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip hidden files and build artifacts
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Skip build outputs
		if strings.Contains(path, "/target/") || strings.Contains(path, "/node_modules/") {
			return nil
		}

		// Only include source files
		if !isSourceFile(path) {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Add file to ZIP
		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(zipFile, srcFile)
		return err
	})

	if err != nil {
		return nil, err
	}

	return buf, nil
}

// isSourceFile checks if a file is a valid source file
func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	validExts := []string{
		".rs", ".go", ".ts", ".js", ".dart", ".py", ".zig", ".kt", ".swift",
		".c", ".cpp", ".h", ".hpp",
		".toml", ".json", ".mod", ".sum", ".lock",
	}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}

	return false
}

// log prints verbose output
func log(format string, args ...interface{}) {
	if *verbose {
		fmt.Printf("[debug] "+format+"\n", args...)
	}
}
