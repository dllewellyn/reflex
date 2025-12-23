package huggingface

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/parquet-go/parquet-go"
)

// Client is a client for the HuggingFace API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new HuggingFace client.
func NewClient(httpClient *http.Client) *Client {
	return &Client{
		baseURL:    "https://huggingface.co",
		httpClient: httpClient,
	}
}

type datasetInfo struct {
	Siblings []struct {
		RFilename string `json:"rfilename"`
	} `json:"siblings"`
}

// DownloadAndRead downloads the dataset file (parquet, jsonl, or json) for the given dataset and split
// and returns a DatasetReader.
func (c *Client) DownloadAndRead(ctx context.Context, datasetID, split string) (DatasetReader, error) {
	// 1. Get dataset info to find the file
	url := fmt.Sprintf("%s/api/datasets/%s", c.baseURL, datasetID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dataset info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch dataset info: status %d", resp.StatusCode)
	}

	var info datasetInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode dataset info: %w", err)
	}

	// 2. Find the file (prioritize extensions)
	var targetFile string
	var fileType string // "parquet" or "jsonl" (covers json too)

	// Priority: .parquet -> .jsonl -> .json
	// Note: huggingface usually has "data/" prefix or similar. We look for the split name in the filename.
	// Simple heuristic: check suffix and if it contains split.

	// Helper to find file by extensions
	findFile := func(exts []string, matchSplit bool) (string, string) {
		for _, s := range info.Siblings {
			for _, ext := range exts {
				if strings.HasSuffix(s.RFilename, ext) {
					if !matchSplit || strings.Contains(s.RFilename, split) {
						return s.RFilename, ext
					}
				}
			}
		}
		return "", ""
	}

	// 1. Try to find precise match (extension + split name)
	targetFile, ext := findFile([]string{".parquet"}, true)
	if targetFile != "" {
		fileType = "parquet"
	} else {
		targetFile, ext = findFile([]string{".jsonl", ".json"}, true)
		if targetFile != "" {
			fileType = "jsonl"
		}
	}

	// 2. Fallback: find any supported file
	if targetFile == "" {
		log.Printf("Warning: no file found matching split %q. Falling back to any supported file.", split)
		targetFile, ext = findFile([]string{".parquet"}, false)
		if targetFile != "" {
			fileType = "parquet"
		} else {
			targetFile, ext = findFile([]string{".jsonl", ".json"}, false)
			if targetFile != "" {
				fileType = "jsonl"
			}
		}
	}

	if targetFile == "" {
		// Log available files for debugging
		var files []string
		for _, s := range info.Siblings {
			files = append(files, s.RFilename)
		}
		log.Printf("Debug: No matching file found. Available files: %v\n", files)
		return nil, fmt.Errorf("no supported file (parquet, jsonl, json) found for split %q in dataset %q", split, datasetID)
	}

	// Log found file for debugging
	// fmt.Printf("Found file: %s (type: %s)\n", targetFile, fileType)

	// 3. Download the file
	fileURL := fmt.Sprintf("%s/datasets/%s/resolve/main/%s", c.baseURL, datasetID, targetFile)
	reqFile, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create file request: %w", err)
	}

	respFile, err := c.httpClient.Do(reqFile)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer respFile.Body.Close()

	if respFile.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: status %d", respFile.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("hf-dataset-*%s", ext))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(tmpFile, respFile.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// Rewind file for reading
	if _, err := tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to seek temp file: %w", err)
	}

	// 4. Create Reader based on type
	if fileType == "parquet" {
		stat, err := tmpFile.Stat()
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return nil, fmt.Errorf("failed to stat temp file: %w", err)
		}

		file, err := parquet.OpenFile(tmpFile, stat.Size())
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return nil, fmt.Errorf("failed to open parquet file: %w", err)
		}

		// Create a reader for all row groups
		rows := parquet.MultiRowGroup(file.RowGroups()...).Rows()

		return &ParquetReader{
			rows: rows,
			file: tmpFile,
		}, nil
	} else {
		// JSON or JSONL
		return NewJSONLReader(tmpFile), nil
	}
}
