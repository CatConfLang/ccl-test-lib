package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// TODO: Update this URL when ccl-test-data repository is public
	baseURL = "https://raw.githubusercontent.com/tylerbutler/ccl-test-data/main/schemas"

	// Local fallback for development
	localSchemaPath = "../ccl-test-data/schemas"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Printf("Usage: %s [output-dir]\n", os.Args[0])
		fmt.Println("Downloads CCL JSON schemas from ccl-test-data repository")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("  output-dir    Directory to save schemas (default: schemas)")
		os.Exit(0)
	}

	outputDir := "schemas"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	schemas := []string{
		"source-format.json",
		"generated-format.json",
	}

	fmt.Printf("Syncing schemas to %s/\n", outputDir)

	for _, schema := range schemas {
		outputPath := filepath.Join(outputDir, schema)

		// Try local file first (for development)
		localPath := filepath.Join(localSchemaPath, schema)
		if _, err := os.Stat(localPath); err == nil {
			fmt.Printf("  %s (local) -> %s\n", schema, outputPath)
			if err := copyFile(localPath, outputPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error copying local file %s: %v\n", schema, err)
				os.Exit(1)
			}
			continue
		}

		// Fall back to remote download
		url := fmt.Sprintf("%s/%s", baseURL, schema)
		fmt.Printf("  %s (remote) -> %s\n", schema, outputPath)

		if err := downloadFile(url, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", schema, err)
			fmt.Fprintf(os.Stderr, "Tried local path: %s (not found)\n", localPath)
			os.Exit(1)
		}
	}

	fmt.Println("Schema download complete!")
}

func downloadFile(url, outputPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file failed: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file failed: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	return nil
}