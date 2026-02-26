package cleaner

import (
	"os"
	"path/filepath"
	"testing"

	"macos-cleaner/internal/models"
	"macos-cleaner/internal/utils"
)

func TestCleanTarget(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "cleaner_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	content := make([]byte, 1000)
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}
	
	sudoMgr := utils.NewSudoManager()
	cleaner := New(sudoMgr)
	
	target := &models.CleanupTarget{
		Name:    "Test File",
		Path:    testFile,
		Size:    1000,
		Selected: true,
	}
	
	result := cleaner.cleanTarget(target)
	
	if result.Error != nil {
		t.Errorf("cleanTarget() error = %v", result.Error)
	}
	
	if result.Actual != 1000 {
		t.Errorf("cleanTarget() actual = %d, want 1000", result.Actual)
	}
	
	// Verify file was deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File was not deleted")
	}
}

func TestCleanTarget_Directory(t *testing.T) {
	// Create temp directory with files
	tmpDir, err := os.MkdirTemp("", "cleaner_dir_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Add files to directory
	file1 := filepath.Join(testDir, "file1.txt")
	file2 := filepath.Join(testDir, "file2.txt")
	os.WriteFile(file1, make([]byte, 500), 0644)
	os.WriteFile(file2, make([]byte, 700), 0644)
	
	sudoMgr := utils.NewSudoManager()
	cleaner := New(sudoMgr)
	
	target := &models.CleanupTarget{
		Name:    "Test Directory",
		Path:    testDir,
		Size:    1200,
		Selected: true,
	}
	
	result := cleaner.cleanTarget(target)
	
	if result.Error != nil {
		t.Errorf("cleanTarget() error = %v", result.Error)
	}
	
	if result.Actual != 1200 {
		t.Errorf("cleanTarget() actual = %d, want 1200", result.Actual)
	}
	
	// Verify directory was deleted
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("Directory was not deleted")
	}
}

func TestCleanTargets_NotSelected(t *testing.T) {
	// Create temp file
	tmpDir, err := os.MkdirTemp("", "cleaner_skip_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)
	
	sudoMgr := utils.NewSudoManager()
	cleaner := New(sudoMgr)
	
	targets := []models.CleanupTarget{
		{Name: "Test File", Path: testFile, Selected: false}, // Not selected
	}
	
	results, totalSaved := cleaner.CleanTargets(targets, func(status string) {})
	
	// Should not process any targets
	if len(results) != 0 {
		t.Errorf("CleanTargets() processed %d targets, want 0", len(results))
	}
	
	if totalSaved != 0 {
		t.Errorf("CleanTargets() totalSaved = %d, want 0", totalSaved)
	}
	
	// Verify file still exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File should not have been deleted")
	}
}

func TestCleanTargets(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "cleaner_multi_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create multiple files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	file3 := filepath.Join(tmpDir, "file3.txt")
	
	os.WriteFile(file1, make([]byte, 100), 0644)
	os.WriteFile(file2, make([]byte, 200), 0644)
	os.WriteFile(file3, make([]byte, 300), 0644)
	
	targets := []models.CleanupTarget{
		{Name: "File 1", Path: file1, Size: 100, Selected: true},
		{Name: "File 2", Path: file2, Size: 200, Selected: true},
		{Name: "File 3", Path: file3, Size: 300, Selected: false}, // Not selected
	}
	
	sudoMgr := utils.NewSudoManager()
	cleaner := New(sudoMgr)
	
	progressCalled := false
	results, totalSaved := cleaner.CleanTargets(targets, func(status string) {
		progressCalled = true
	})
	
	if !progressCalled {
		t.Error("Progress callback was not called")
	}
	
	if len(results) != 2 {
		t.Errorf("CleanTargets() returned %d results, want 2", len(results))
	}
	
	if totalSaved != 300 {
		t.Errorf("CleanTargets() totalSaved = %d, want 300", totalSaved)
	}
	
	// Verify file1 and file2 were deleted
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("file1 should have been deleted")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("file2 should have been deleted")
	}
	
	// Verify file3 still exists
	if _, err := os.Stat(file3); os.IsNotExist(err) {
		t.Error("file3 should not have been deleted")
	}
}

func TestDeleteFiles(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "delete_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	
	os.WriteFile(file1, make([]byte, 100), 0644)
	os.WriteFile(file2, make([]byte, 200), 0644)
	
	files := []string{file1, file2}
	
	sudoMgr := utils.NewSudoManager()
	cleaner := New(sudoMgr)
	
	progressCalled := false
	deleted, err := cleaner.DeleteFiles(files, func(status string) {
		progressCalled = true
	})
	
	if err != nil {
		t.Errorf("DeleteFiles() error = %v", err)
	}
	
	if !progressCalled {
		t.Error("Progress callback was not called")
	}
	
	if deleted != 300 {
		t.Errorf("DeleteFiles() deleted = %d, want 300", deleted)
	}
	
	// Verify files were deleted
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("file1 should have been deleted")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("file2 should have been deleted")
	}
}

func TestDeleteBigFiles(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "delete_big_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	
	os.WriteFile(file1, make([]byte, 1000), 0644)
	os.WriteFile(file2, make([]byte, 2000), 0644)
	
	files := []models.BigFile{
		{Path: file1, Size: 1000},
		{Path: file2, Size: 2000},
	}
	
	selected := map[int]bool{
		0: true,
		1: false, // Not selected
	}
	
	sudoMgr := utils.NewSudoManager()
	cleaner := New(sudoMgr)
	
	deleted := cleaner.DeleteBigFiles(files, selected, func(status string) {})
	
	if deleted != 1000 {
		t.Errorf("DeleteBigFiles() deleted = %d, want 1000", deleted)
	}
	
	// Verify file1 was deleted, file2 remains
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("file1 should have been deleted")
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("file2 should not have been deleted")
	}
}

func TestDeleteDuplicates(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "delete_dup_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create duplicate files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	file3 := filepath.Join(tmpDir, "file3.txt")
	
	content := []byte("duplicate content")
	os.WriteFile(file1, content, 0644)
	os.WriteFile(file2, content, 0644)
	os.WriteFile(file3, content, 0644)
	
	groups := []models.DuplicateGroup{
		{
			Hash:  "abc123",
			Size:  int64(len(content)),
			Files: []string{file1, file2, file3},
		},
	}
	
	selected := map[int]bool{
		0: true,
	}
	
	sudoMgr := utils.NewSudoManager()
	cleaner := New(sudoMgr)
	
	deleted := cleaner.DeleteDuplicates(groups, selected, func(status string) {})
	
	// Should delete 2 files (file2 and file3), keep file1
	expectedDeleted := int64(len(content)) * 2
	if deleted != expectedDeleted {
		t.Errorf("DeleteDuplicates() deleted = %d, want %d", deleted, expectedDeleted)
	}
	
	// Verify file1 still exists, file2 and file3 deleted
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("file1 should have been kept (first in group)")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("file2 should have been deleted")
	}
	if _, err := os.Stat(file3); !os.IsNotExist(err) {
		t.Error("file3 should have been deleted")
	}
}

func TestDeleteOldFiles(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "delete_old_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	
	os.WriteFile(file1, make([]byte, 100), 0644)
	os.WriteFile(file2, make([]byte, 200), 0644)
	
	files := []models.OldFile{
		{Path: file1, Size: 100},
		{Path: file2, Size: 200},
	}
	
	selected := map[int]bool{
		0: true,
		1: false,
	}
	
	sudoMgr := utils.NewSudoManager()
	cleaner := New(sudoMgr)
	
	deleted := cleaner.DeleteOldFiles(files, selected, func(status string) {})
	
	if deleted != 100 {
		t.Errorf("DeleteOldFiles() deleted = %d, want 100", deleted)
	}
	
	// Verify file1 deleted, file2 remains
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("file1 should have been deleted")
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("file2 should not have been deleted")
	}
}

func TestCalculateActualSize(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "size_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	
	os.WriteFile(file1, make([]byte, 100), 0644)
	os.WriteFile(file2, make([]byte, 200), 0644)
	
	sudoMgr := utils.NewSudoManager()
	cleaner := New(sudoMgr)
	
	// Test single file
	size := cleaner.calculateActualSize(file1)
	if size != 100 {
		t.Errorf("calculateActualSize() = %d, want 100", size)
	}
	
	// Test directory
	size = cleaner.calculateActualSize(tmpDir)
	if size != 300 {
		t.Errorf("calculateActualSize() = %d, want 300", size)
	}
	
	// Test non-existing path
	size = cleaner.calculateActualSize("/non/existent/path")
	if size != 0 {
		t.Errorf("calculateActualSize() = %d, want 0 for non-existing", size)
	}
}
