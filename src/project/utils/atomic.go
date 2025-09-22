// Package utils provides atomic file operations to prevent corruption during writes
package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// RenameOperation represents a single file rename operation
type RenameOperation struct {
	OldPath string
	NewPath string
}

// AtomicWriteFile writes data to a file atomically using temp file + rename pattern
// This prevents corruption if the process is interrupted during write
func AtomicWriteFile(targetPath string, data []byte, perm os.FileMode) error {
	// Create temp file in same directory to ensure atomic rename
	dir := filepath.Dir(targetPath)
	tempFile, err := os.CreateTemp(dir, "saul_atomic_*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()

	// Cleanup temp file on any error
	defer func() {
		tempFile.Close()
		os.Remove(tempPath)
	}()

	// Write data to temp file
	if _, err := tempFile.Write(data); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}

	// Sync to disk before rename
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %v", err)
	}

	// Close temp file before rename
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// Set correct permissions
	if err := os.Chmod(tempPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %v", err)
	}

	// Atomic rename to final location
	if err := os.Rename(tempPath, targetPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}

// AtomicBatchRename performs multiple file renames atomically with rollback on failure
func AtomicBatchRename(operations []RenameOperation) error {
	if len(operations) == 0 {
		return nil
	}

	completed := make([]RenameOperation, 0, len(operations))

	// Perform all renames, tracking completed operations
	for _, op := range operations {
		if err := os.Rename(op.OldPath, op.NewPath); err != nil {
			// Rollback all completed operations
			rollbackRenames(completed)
			return fmt.Errorf("failed to rename %s to %s: %v", op.OldPath, op.NewPath, err)
		}
		completed = append(completed, op)
	}

	return nil
}

// rollbackRenames reverses completed rename operations
func rollbackRenames(completed []RenameOperation) {
	// Reverse the operations to undo them
	for i := len(completed) - 1; i >= 0; i-- {
		op := completed[i]
		// Swap old/new to reverse the operation
		os.Rename(op.NewPath, op.OldPath)
	}
}