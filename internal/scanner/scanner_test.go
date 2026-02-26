package scanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"macos-cleaner/internal/models"
	"macos-cleaner/internal/utils"
)

func TestCalculateSize(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files with known sizes
	file1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(file1, make([]byte, 1000), 0644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(tmpDir, "file2.txt")
	if err := os.WriteFile(file2, make([]byte, 2000), 0644); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory with file
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	file3 := filepath.Join(subDir, "file3.txt")
	if err := os.WriteFile(file3, make([]byte, 500), 0644); err != nil {
		t.Fatal(err)
	}

	// Test scanner (using direct directory path, not glob)
	sudoMgr := utils.NewSudoManager()
	scanner := New(sudoMgr)

	size := scanner.CalculateSize(tmpDir)
	if size != 3500 {
		t.Errorf("CalculateSize() = %d, want 3500", size)
	}
}

func TestCalculateSizeForTarget(t *testing.T) {
	sudoMgr := utils.NewSudoManager()
	scanner := New(sudoMgr)

	// Create temp file
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, make([]byte, 1000), 0644); err != nil {
		t.Fatal(err)
	}

	target := models.CleanupTarget{
		Name: "Test",
		Path: testFile,
	}

	size := scanner.CalculateSizeForTarget(&target)
	if size != 1000 {
		t.Errorf("CalculateSizeForTarget() = %d, want 1000", size)
	}
}

func TestScanBigFiles(t *testing.T) {
	// Create temp directory structure simulating home
	tmpDir, err := os.MkdirTemp("", "bigfiles_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create Documents directory
	docsDir := filepath.Join(tmpDir, "Documents")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files of different sizes in Documents
	// Small file (should not be found)
	smallFile := filepath.Join(docsDir, "small.txt")
	if err := os.WriteFile(smallFile, make([]byte, 100), 0644); err != nil {
		t.Fatal(err)
	}

	// Large file (should be found)
	largeFile := filepath.Join(docsDir, "large.txt")
	if err := os.WriteFile(largeFile, make([]byte, 2000), 0644); err != nil {
		t.Fatal(err)
	}

	// Very large file (should be found)
	veryLargeFile := filepath.Join(docsDir, "verylarge.txt")
	if err := os.WriteFile(veryLargeFile, make([]byte, 5000), 0644); err != nil {
		t.Fatal(err)
	}

	// Mock home directory temporarily
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	sudoMgr := utils.NewSudoManager()
	scanner := New(sudoMgr)

	progressCalled := false
	files := scanner.ScanBigFiles(1500, func(status string) {
		progressCalled = true
	})

	if !progressCalled {
		t.Error("Progress callback was not called")
	}

	// Should find 2 files (2000 and 5000 bytes)
	if len(files) != 2 {
		t.Errorf("ScanBigFiles() found %d files, want 2", len(files))
	}

	// Check that found files have correct paths
	foundPaths := make(map[string]bool)
	for _, f := range files {
		foundPaths[f.Path] = true
	}

	if !foundPaths[largeFile] {
		t.Error("Large file not found")
	}
	if !foundPaths[veryLargeFile] {
		t.Error("Very large file not found")
	}
	if foundPaths[smallFile] {
		t.Error("Small file should not be found")
	}
}

func TestScanDuplicates(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "duplicates_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	documentsDir := filepath.Join(tmpDir, "Documents")
	if err := os.MkdirAll(documentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create duplicate files (>1MB each to be considered)
	// Since files need to be >1MB, let's create larger content
	content := make([]byte, 2*1024*1024) // 2MB

	file1 := filepath.Join(documentsDir, "file1.bin")
	if err := os.WriteFile(file1, content, 0644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(documentsDir, "file2.bin")
	if err := os.WriteFile(file2, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Create unique file
	uniqueContent := make([]byte, 2*1024*1024)
	uniqueContent[0] = 1 // Make it different
	file3 := filepath.Join(documentsDir, "file3.bin")
	if err := os.WriteFile(file3, uniqueContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create small file (should be ignored - less than 1MB)
	smallContent := []byte("small")
	file4 := filepath.Join(documentsDir, "small.txt")
	if err := os.WriteFile(file4, smallContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Mock home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	sudoMgr := utils.NewSudoManager()
	scanner := New(sudoMgr)

	progressCalled := false
	groups, totalSize := scanner.ScanDuplicates(func(status string) {
		progressCalled = true
	})

	if !progressCalled {
		t.Error("Progress callback was not called")
	}

	// Should find 1 duplicate group (files 1 and 2)
	if len(groups) != 1 {
		t.Errorf("ScanDuplicates() found %d groups, want 1", len(groups))
	}

	if len(groups) > 0 {
		group := groups[0]
		if len(group.Files) != 2 {
			t.Errorf("Duplicate group has %d files, want 2", len(group.Files))
		}

		expectedSavings := int64(len(content)) * 1 // 1 duplicate
		if totalSize != expectedSavings {
			t.Errorf("TotalSize = %d, want %d", totalSize, expectedSavings)
		}
	}
}

func TestScanOldFiles(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "oldfiles_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	documentsDir := filepath.Join(tmpDir, "Documents")
	if err := os.MkdirAll(documentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file with old modification time
	oldFile := filepath.Join(documentsDir, "old.txt")
	if err := os.WriteFile(oldFile, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().AddDate(0, 0, -200) // 200 days ago
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Create a recent file
	newFile := filepath.Join(documentsDir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Mock home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	sudoMgr := utils.NewSudoManager()
	scanner := New(sudoMgr)

	files := scanner.ScanOldFiles(180, func(status string) {})

	// Should find 1 old file
	if len(files) != 1 {
		t.Errorf("ScanOldFiles() found %d files, want 1", len(files))
	}

	if len(files) > 0 {
		if files[0].Path != oldFile {
			t.Errorf("Found wrong file: %s, want %s", files[0].Path, oldFile)
		}
	}
}

func TestParseHomebrewSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"100KB", 100 * 1024},
		{"1.5MB", int64(1.5 * 1024 * 1024)},
		{"2GB", 2 * 1024 * 1024 * 1024},
		{"100", 100},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseHomebrewSize(tt.input)
			// Allow small floating point differences
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 1000 { // Allow 1KB difference due to float precision
				t.Errorf("parseHomebrewSize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
