// Package scanner handles all scanning operations for finding files to clean
package scanner

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

// Scanner handles scanning operations
type Scanner struct {
	SudoManager *utils.SudoManager
}

// New creates a new Scanner
func New(sudoMgr *utils.SudoManager) *Scanner {
	return &Scanner{
		SudoManager: sudoMgr,
	}
}

// CalculateSize calculates the total size of files matching a pattern
// This uses SafeGlob for better handling of nested directories
func (s *Scanner) CalculateSize(pattern string) int64 {
	matches, err := utils.SafeGlob(pattern)
	if err != nil {
		return 0
	}

	var total int64
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		if info.IsDir() {
			total += utils.DirSize(match)
		} else {
			total += info.Size()
		}
	}

	return total
}

// CalculateSizeForTarget calculates size for a CleanupTarget
func (s *Scanner) CalculateSizeForTarget(target *models.CleanupTarget) int64 {
	if target.IsCommand {
		return s.calculateCommandSize(target.Command)
	}
	return s.CalculateSize(target.Path)
}

// calculateCommandSize estimates size for command-based targets
func (s *Scanner) calculateCommandSize(command string) int64 {
	// For brew cleanup, use dry-run to estimate size
	if strings.Contains(command, "brew cleanup") {
		return s.getHomebrewCleanupSize()
	}
	if strings.Contains(command, "tmutil deletelocalsnapshots") {
		return s.getTimeMachineSnapshotSize()
	}
	return 0
}

// getHomebrewCleanupSize gets the estimated cleanup size from Homebrew
func (s *Scanner) getHomebrewCleanupSize() int64 {
	// Run brew cleanup -n and parse output for size estimation
	// brew cleanup -n outputs lines like: "Would remove: /path/to/file (1.2MB)"
	cmd := exec.Command("brew", "cleanup", "-n")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	var totalBytes int64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Look for size in parentheses at end of line: "(1.2MB)" or "(1.2 GB)"
		if idx := strings.LastIndex(line, "("); idx != -1 && strings.HasSuffix(line, ")") {
			sizeStr := line[idx+1 : len(line)-1]
			size := parseHomebrewSize(sizeStr)
			totalBytes += size
		}
	}

	return totalBytes
}

// parseHomebrewSize parses size strings like "1.2MB", "3.4GB", "500KB"
func parseHomebrewSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Extract numeric part
	var numStr string
	var unit string
	for i, c := range s {
		if (c >= '0' && c <= '9') || c == '.' {
			numStr = s[:i+1]
		} else {
			unit = s[i:]
			break
		}
	}

	var num float64
	fmt.Sscanf(numStr, "%f", &num)

	unit = strings.ToUpper(strings.TrimSpace(unit))
	switch unit {
	case "KB", "K":
		return int64(num * 1024)
	case "MB", "M":
		return int64(num * 1024 * 1024)
	case "GB", "G":
		return int64(num * 1024 * 1024 * 1024)
	case "TB", "T":
		return int64(num * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(num)
	}
}

// getTimeMachineSnapshotSize gets the size of local Time Machine snapshots
func (s *Scanner) getTimeMachineSnapshotSize() int64 {
	output, err := utils.NewSudoManager().RunWithOutput("tmutil", "listlocalsnapshots", "/")
	if err != nil {
		return 0
	}
	// Count snapshots and estimate size (rough approximation)
	lines := strings.Split(string(output), "\n")
	// Each snapshot is roughly 1-5GB, we'll estimate conservatively
	count := 0
	for _, line := range lines {
		if strings.Contains(line, "com.apple.TimeMachine") {
			count++
		}
	}
	return int64(count) * 1024 * 1024 * 1024 // Estimate 1GB per snapshot
}

// ScanBigFiles scans for files larger than the specified size
func (s *Scanner) ScanBigFiles(minSize int64, progress func(status string)) []models.BigFile {
	var files []models.BigFile

	// Scan specific directories instead of entire home to improve performance
	dirs := []string{
		utils.ExpandPath("~/Documents"),
		utils.ExpandPath("~/Desktop"),
		utils.ExpandPath("~/Downloads"),
		utils.ExpandPath("~/Movies"),
		utils.ExpandPath("~/Music"),
		utils.ExpandPath("~/Pictures"),
	}

	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"Library":      true, // Skip Library - it's huge and mostly cache
	}

	scannedCount := 0

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			scannedCount++

			// Update progress periodically
			if scannedCount%500 == 0 {
				progress(fmt.Sprintf("Scanned %d files...", scannedCount))
			}

			// Skip hidden dirs and system dirs
			if info.IsDir() {
				name := info.Name()
				if skipDirs[name] || strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
				return nil
			}

			if info.Size() >= minSize {
				files = append(files, models.BigFile{
					Path:    path,
					Size:    info.Size(),
					ModTime: info.ModTime(),
				})
				progress(fmt.Sprintf("Found: %s (%s)", utils.ShortenPath(info.Name(), 30), formatBytes(info.Size())))
			}

			return nil
		})
	}

	return files
}

// ScanDuplicates scans for duplicate files in the specified directories
func (s *Scanner) ScanDuplicates(progress func(status string)) ([]models.DuplicateGroup, int64) {
	sizeMap := make(map[int64][]string)

	dirs := []string{
		utils.ExpandPath("~/Documents"),
		utils.ExpandPath("~/Desktop"),
		utils.ExpandPath("~/Downloads"),
	}

	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"Library":      true,
	}

	// First pass: group by size
	scannedCount := 0
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				if skipDirs[info.Name()] || strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}

			scannedCount++
			if scannedCount%500 == 0 {
				progress(fmt.Sprintf("Scanned %d files...", scannedCount))
			}

			// Only check files > 1MB to save time
			if info.Size() > 1024*1024 {
				sizeMap[info.Size()] = append(sizeMap[info.Size()], path)
			}

			return nil
		})
	}

	progress(fmt.Sprintf("Found %d files with unique sizes, checking for duplicates...", len(sizeMap)))

	// Second pass: hash files with same size
	hashMap := make(map[string][]string)
	hashCount := 0
	totalPaths := 0
	for _, paths := range sizeMap {
		totalPaths += len(paths)
	}

	for _, paths := range sizeMap {
		if len(paths) < 2 {
			continue
		}

		for _, path := range paths {
			hash := utils.FileHash(path)
			if hash != "" {
				hashMap[hash] = append(hashMap[hash], path)
			}
			hashCount++
			if hashCount%10 == 0 {
				progress(fmt.Sprintf("Hashed %d/%d files...", hashCount, totalPaths))
			}
		}
	}

	// Create duplicate groups
	var groups []models.DuplicateGroup
	var totalSize int64
	for hash, paths := range hashMap {
		if len(paths) > 1 {
			info, _ := os.Stat(paths[0])
			if info != nil {
				groups = append(groups, models.DuplicateGroup{
					Hash:  hash,
					Size:  info.Size(),
					Files: paths,
				})
				totalSize += info.Size() * int64(len(paths)-1)
			}
		}
	}

	return groups, totalSize
}

// ScanOldFiles scans for files not accessed in the specified number of days
func (s *Scanner) ScanOldFiles(days int, progress func(status string)) []models.OldFile {
	var files []models.OldFile
	cutoff := time.Now().AddDate(0, 0, -days)

	dirs := []string{
		utils.ExpandPath("~/Documents"),
		utils.ExpandPath("~/Desktop"),
		utils.ExpandPath("~/Downloads"),
	}

	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				if info != nil && info.IsDir() && skipDirs[info.Name()] {
					return filepath.SkipDir
				}
				return nil
			}

			// Check last access time (using ModTime as approximation)
			if info.ModTime().Before(cutoff) {
				files = append(files, models.OldFile{
					Path:       path,
					Size:       info.Size(),
					LastAccess: info.ModTime(),
				})
			}

			return nil
		})
	}

	return files
}

// formatBytes formats bytes to human-readable string
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return "B"
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return units[exp]
}
