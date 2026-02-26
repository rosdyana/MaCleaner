// Package cleaner handles all file deletion and cleaning operations
package cleaner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"macos-cleaner/internal/models"
	"macos-cleaner/internal/utils"
)

// Cleaner handles file deletion operations
type Cleaner struct {
	SudoManager *utils.SudoManager
}

// New creates a new Cleaner
func New(sudoMgr *utils.SudoManager) *Cleaner {
	return &Cleaner{
		SudoManager: sudoMgr,
	}
}

// CleanResult represents the result of a cleaning operation
type CleanResult struct {
	Target    string
	Requested int64
	Actual    int64
	Error     error
	Timestamp time.Time
}

// CleanTargets cleans the selected targets and returns actual space freed
func (c *Cleaner) CleanTargets(targets []models.CleanupTarget, progress func(string)) ([]CleanResult, int64) {
	var results []CleanResult
	var totalSaved int64

	// Check if any target needs sudo
	needsSudo := false
	for i := range targets {
		if targets[i].Selected && targets[i].RequiresSudo {
			needsSudo = true
			break
		}
	}

	// Authenticate once if needed
	if needsSudo {
		if err := c.SudoManager.EnsureSudo(); err != nil {
			return results, 0
		}
	}

	for i := range targets {
		if !targets[i].Selected {
			continue
		}

		target := &targets[i]
		progress("Cleaning: " + target.Name)

		result := c.cleanTarget(target)
		results = append(results, result)

		if result.Error == nil {
			totalSaved += result.Actual
			target.Size = 0 // Reset size after successful cleaning
		}
	}

	return results, totalSaved
}

// cleanTarget cleans a single target and returns the actual space freed
func (c *Cleaner) cleanTarget(target *models.CleanupTarget) CleanResult {
	result := CleanResult{
		Target:    target.Name,
		Requested: target.Size,
		Timestamp: time.Now(),
	}

	if target.IsCommand && target.Command != "" {
		if err := c.executeCommand(target.Command); err != nil {
			result.Error = fmt.Errorf("command failed: %w", err)
		}
		// For command-based targets, assume all requested space is freed
		// since we can't easily measure
		result.Actual = target.Size
		return result
	}

	path := utils.ExpandPath(target.Path)
	expandedPath := utils.ExpandPath(target.Path)

	// Check if the path exists before trying to clean
	matches, err := filepath.Glob(expandedPath)
	if err != nil {
		result.Error = fmt.Errorf("invalid path pattern: %w", err)
		return result
	}
	if len(matches) == 0 {
		// No files to clean - this is OK, just means already clean
		result.Actual = 0
		return result
	}

	// Calculate actual size BEFORE deletion
	actualBefore := c.calculateActualSize(path)

	if actualBefore == 0 {
		// Files exist but have no size (maybe permission issue)
		// Try to clean anyway
	}

	// Perform deletion
	if err := c.deletePath(path, target.RequiresSudo); err != nil {
		result.Error = fmt.Errorf("failed to delete %s: %w", target.Path, err)
		return result
	}

	// Wait a moment for filesystem to sync
	time.Sleep(100 * time.Millisecond)

	// Calculate remaining size AFTER deletion
	actualAfter := c.calculateActualSize(path)

	// Actual space freed is the difference
	result.Actual = actualBefore - actualAfter

	return result
}

// calculateActualSize calculates the actual disk space used by a path
// This uses a more robust method that handles wildcards properly
func (c *Cleaner) calculateActualSize(pattern string) int64 {
	// Handle wildcards by finding matching paths
	if strings.Contains(pattern, "*") {
		basePath := strings.Split(pattern, "*")[0]
		var total int64

		// Use find command for more accurate results with wildcards
		cmd := exec.Command("find", basePath, "-type", "f", "-print0")
		output, err := cmd.Output()
		if err != nil {
			// Fallback to walk
			return c.walkCalculateSize(basePath, pattern)
		}

		// Parse null-terminated output
		files := strings.Split(string(output), "\x00")
		for _, file := range files {
			if file == "" {
				continue
			}
			if matched, _ := filepath.Match(pattern, file); matched {
				if info, err := os.Stat(file); err == nil && !info.IsDir() {
					total += info.Size()
				}
			}
		}
		return total
	}

	// No wildcard, use standard calculation
	info, err := os.Stat(pattern)
	if err != nil {
		return 0
	}

	if info.IsDir() {
		return utils.DirSize(pattern)
	}
	return info.Size()
}

// walkCalculateSize calculates size by walking directory
func (c *Cleaner) walkCalculateSize(basePath, pattern string) int64 {
	var total int64

	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		// Simple glob match
		if matched, _ := filepath.Match(pattern, path); matched {
			total += info.Size()
		}
		return nil
	})

	return total
}

// deletePath deletes a path, using sudo if required
func (c *Cleaner) deletePath(path string, useSudo bool) error {
	// Handle wildcards by expanding and deleting each match
	if strings.Contains(path, "*") {
		matches, err := filepath.Glob(path)
		if err != nil {
			return fmt.Errorf("glob failed: %w", err)
		}

		if len(matches) == 0 {
			return nil // Nothing to delete
		}

		var lastErr error
		deletedCount := 0
		for _, match := range matches {
			if err := c.deleteSinglePath(match, useSudo); err != nil {
				lastErr = err
				// Continue trying to delete other matches
				continue
			}
			deletedCount++
		}

		if lastErr != nil && deletedCount == 0 {
			return fmt.Errorf("failed to delete any files: %w", lastErr)
		}
		return nil
	}

	return c.deleteSinglePath(path, useSudo)
}

// deleteSinglePath deletes a single file or directory
func (c *Cleaner) deleteSinglePath(path string, useSudo bool) error {
	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil // Already deleted
	}
	if err != nil {
		return fmt.Errorf("cannot access path: %w", err)
	}

	if useSudo {
		if err := c.SudoManager.Run("rm", "-rf", path); err != nil {
			return fmt.Errorf("sudo rm failed: %w", err)
		}
		return nil
	}

	// Try to delete
	if info.IsDir() {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove directory failed: %w", err)
		}
	} else {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove file failed: %w", err)
		}
	}

	return nil
}

// executeCommand executes a shell command for special targets
func (c *Cleaner) executeCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(output))
	}
	return nil
}

// DeleteFiles deletes a list of files and returns total bytes freed
func (c *Cleaner) DeleteFiles(files []string, progress func(string)) (int64, error) {
	var totalDeleted int64

	for _, file := range files {
		progress("Deleting: " + utils.ShortenPath(file, 40))

		// Get size before deletion
		info, err := os.Stat(file)
		if err != nil {
			continue // File might already be gone
		}
		size := info.Size()

		// Determine if sudo is needed (outside home directory)
		needsSudo := !strings.HasPrefix(file, os.Getenv("HOME"))

		var deleteErr error
		if needsSudo {
			// Try without sudo first (in case we have permissions)
			deleteErr = os.Remove(file)
			if deleteErr != nil {
				// Then try with sudo
				deleteErr = c.SudoManager.Run("rm", "-f", file)
			}
		} else {
			deleteErr = os.Remove(file)
		}

		if deleteErr == nil {
			totalDeleted += size
		}
	}

	return totalDeleted, nil
}

// DeleteBigFiles deletes selected big files
func (c *Cleaner) DeleteBigFiles(files []models.BigFile, selected map[int]bool, progress func(string)) int64 {
	var paths []string
	for i, sel := range selected {
		if sel && i < len(files) {
			paths = append(paths, files[i].Path)
		}
	}

	deleted, _ := c.DeleteFiles(paths, progress)
	return deleted
}

// DeleteDuplicates deletes selected duplicate files (keeping one copy)
func (c *Cleaner) DeleteDuplicates(groups []models.DuplicateGroup, selected map[int]bool, progress func(string)) int64 {
	var filesToDelete []string

	for i, sel := range selected {
		if !sel || i >= len(groups) {
			continue
		}

		group := groups[i]
		// Keep the first file, delete the rest
		for j := 1; j < len(group.Files); j++ {
			filesToDelete = append(filesToDelete, group.Files[j])
		}
	}

	deleted, _ := c.DeleteFiles(filesToDelete, progress)
	return deleted
}

// DeleteOldFiles deletes selected old files
func (c *Cleaner) DeleteOldFiles(files []models.OldFile, selected map[int]bool, progress func(string)) int64 {
	var paths []string
	for i, sel := range selected {
		if sel && i < len(files) {
			paths = append(paths, files[i].Path)
		}
	}

	deleted, _ := c.DeleteFiles(paths, progress)
	return deleted
}
