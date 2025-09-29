package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/config"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/utils"
)

// HistoryResponse represents a stored response with metadata
type HistoryResponse struct {
	Timestamp string      `json:"timestamp"`
	Method    string      `json:"method"`
	URL       string      `json:"url"`
	Status    string      `json:"status"`
	Duration  string      `json:"duration"`
	Headers   interface{} `json:"headers"`
	Body      interface{} `json:"body"`
}

// GetHistoryPath returns the full path to a preset's history directory
func GetHistoryPath(preset string) (string, error) {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return "", err
	}
	return filepath.Join(presetPath, ".history"), nil
}

// CreateHistoryDirectory creates the history directory for a preset
func CreateHistoryDirectory(preset string) error {
	historyPath, err := GetHistoryPath(preset)
	if err != nil {
		return err
	}

	// Create history directory
	err = os.MkdirAll(historyPath, config.DirPermissions)
	if err != nil {
		return fmt.Errorf(display.ErrDirectoryFailed)
	}

	return nil
}

// StoreResponse stores an HTTP response in the history with rotation
func StoreResponse(preset string, response HistoryResponse, historyCount int) error {
	if historyCount <= 0 {
		return nil // History disabled
	}

	// Create history directory if it doesn't exist
	err := CreateHistoryDirectory(preset)
	if err != nil {
		return err
	}

	historyPath, err := GetHistoryPath(preset)
	if err != nil {
		return err
	}

	// Add timestamp to response
	response.Timestamp = time.Now().Format(time.RFC3339)

	// Get existing files and handle rotation
	files, err := getHistoryFiles(historyPath)
	if err != nil {
		return err
	}

	// Determine next file number
	nextNum := 1
	if len(files) > 0 {
		// Rotate files if we're at the limit
		if len(files) >= historyCount {
			// Remove oldest files
			for i := 0; i <= len(files)-historyCount; i++ {
				if i < len(files) {
					os.Remove(filepath.Join(historyPath, files[i]))
				}
			}
			// Renumber remaining files
			err = renumberHistoryFiles(historyPath, historyCount)
			if err != nil {
				return err
			}
			nextNum = historyCount
		} else {
			nextNum = len(files) + 1
		}
	}

	// Save new response
	fileName := fmt.Sprintf("%03d.json", nextNum)
	filePath := filepath.Join(historyPath, fileName)

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}

	return utils.AtomicWriteFile(filePath, jsonData, config.FilePermissions)
}

// ListHistoryResponses returns a list of history responses with metadata
func ListHistoryResponses(preset string) ([]HistoryResponse, error) {
	historyPath, err := GetHistoryPath(preset)
	if err != nil {
		return nil, err
	}

	files, err := getHistoryFiles(historyPath)
	if err != nil {
		return nil, err
	}

	var responses []HistoryResponse
	for _, fileName := range files {
		filePath := filepath.Join(historyPath, fileName)

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip corrupted files
		}

		var response HistoryResponse
		err = json.Unmarshal(data, &response)
		if err != nil {
			continue // Skip corrupted files
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// LoadHistoryResponse loads a specific history response by number (1-based, reverse chronological)
// Number 1 = most recent, 2 = second most recent, etc.
func LoadHistoryResponse(preset string, number int) (*HistoryResponse, error) {
	historyPath, err := GetHistoryPath(preset)
	if err != nil {
		return nil, err
	}

	// Get all available history files to determine reverse indexing
	files, err := getHistoryFiles(historyPath)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no history found for preset '%s'", preset)
	}

	// Validate number range
	if number < 1 || number > len(files) {
		return nil, fmt.Errorf("history response %d not found (available: 1-%d)", number, len(files))
	}

	// Reverse index: 1 = most recent (highest file number), 2 = second most recent, etc.
	actualFileIndex := len(files) - number
	fileName := files[actualFileIndex]
	filePath := filepath.Join(historyPath, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history response %d", number)
	}

	var response HistoryResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse history response %d", number)
	}

	return &response, nil
}

// DeleteHistory removes all history files for a preset
func DeleteHistory(preset string) error {
	historyPath, err := GetHistoryPath(preset)
	if err != nil {
		return err
	}

	// Check if history directory exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return nil // Already gone, success
	}

	// Remove the entire history directory
	return os.RemoveAll(historyPath)
}

// getHistoryFiles returns sorted list of history files
func getHistoryFiles(historyPath string) ([]string, error) {
	entries, err := os.ReadDir(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}

// renumberHistoryFiles renumbers history files to maintain sequence
func renumberHistoryFiles(historyPath string, maxCount int) error {
	files, err := getHistoryFiles(historyPath)
	if err != nil {
		return err
	}

	// Keep only the most recent files
	if len(files) > maxCount {
		files = files[len(files)-maxCount:]
	}

	// Prepare atomic batch rename operations
	var renameOps []utils.RenameOperation
	for i, fileName := range files {
		oldPath := filepath.Join(historyPath, fileName)
		newFileName := fmt.Sprintf("%03d.json", i+1)
		newPath := filepath.Join(historyPath, newFileName)

		if oldPath != newPath {
			renameOps = append(renameOps, utils.RenameOperation{
				OldPath: oldPath,
				NewPath: newPath,
			})
		}
	}

	// Execute all renames atomically with rollback on failure
	return utils.AtomicBatchRename(renameOps)
}