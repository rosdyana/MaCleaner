// Package utils provides utility functions for the cleaner
package utils

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ to the user's home directory
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}
	return path
}

// FileHash computes a fast hash of a file (first 4KB only)
func FileHash(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Only hash first 4KB for speed
	h := md5.New()
	buf := make([]byte, 4096)
	n, _ := file.Read(buf)
	h.Write(buf[:n])

	return fmt.Sprintf("%x", h.Sum(nil))
}

// DirSize calculates the total size of a directory by walking all files
func DirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// FileExists checks if a file or directory exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// SafeGlob performs a safer glob that handles nested directories better
// than standard filepath.Glob for patterns like ~/Library/Caches/*
func SafeGlob(pattern string) ([]string, error) {
	expanded := ExpandPath(pattern)

	// If pattern ends with /*, we need to find all matching directories recursively
	if strings.HasSuffix(pattern, "/*") {
		baseDir := strings.TrimSuffix(expanded, "/*")
		var matches []string
		err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue on error
			}
			// Add the path itself if it's a file or directory in the base
			if path != baseDir {
				matches = append(matches, path)
			}
			return nil
		})
		return matches, err
	}

	return filepath.Glob(expanded)
}

// ShortenPath creates a shortened version of a path for display
func ShortenPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= 3 {
		return "..."
	}
	return "..." + path[len(path)-maxLen+3:]
}
