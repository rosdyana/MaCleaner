package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"~/.config", filepath.Join(home, ".config")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test existing file
	if !FileExists(tmpFile.Name()) {
		t.Error("FileExists returned false for existing file")
	}

	// Test non-existing file
	if FileExists("/non/existent/path/file.txt") {
		t.Error("FileExists returned true for non-existing file")
	}
}

func TestIsDirectory(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create temp file
	tmpFile, err := os.CreateTemp(tmpDir, "test")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Test directory
	if !IsDirectory(tmpDir) {
		t.Error("IsDirectory returned false for directory")
	}

	// Test file
	if IsDirectory(tmpFile.Name()) {
		t.Error("IsDirectory returned true for file")
	}

	// Test non-existing
	if IsDirectory("/non/existent/path") {
		t.Error("IsDirectory returned true for non-existing path")
	}
}

func TestShortenPath(t *testing.T) {
	tests := []struct {
		path   string
		maxLen int
		want   string
	}{
		{"/short/path", 50, "/short/path"},
		{"/very/long/path/to/some/file.txt", 20, ".../to/some/file.txt"},
		{"/a", 3, "/a"},
		{"", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := ShortenPath(tt.path, tt.maxLen)
			if got != tt.want {
				t.Errorf("ShortenPath(%q, %d) = %q, want %q", tt.path, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFileHash(t *testing.T) {
	// Create temp file with content
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("test content for hashing")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Test hash
	hash1 := FileHash(tmpFile.Name())
	if hash1 == "" {
		t.Error("FileHash returned empty string for existing file")
	}

	// Test consistency
	hash2 := FileHash(tmpFile.Name())
	if hash1 != hash2 {
		t.Error("FileHash returned different hashes for same file")
	}

	// Test non-existing file
	hash3 := FileHash("/non/existent/file")
	if hash3 != "" {
		t.Error("FileHash should return empty string for non-existing file")
	}
}

func TestDirSize(t *testing.T) {
	// Create temp directory with files
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files with known sizes
	sizes := []int64{100, 200, 300}
	var expectedTotal int64

	for i, size := range sizes {
		data := make([]byte, size)
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i)))
		if err := os.WriteFile(filename, data, 0644); err != nil {
			t.Fatal(err)
		}
		expectedTotal += size
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Add file in subdirectory
	subData := make([]byte, 50)
	if err := os.WriteFile(filepath.Join(subDir, "subfile"), subData, 0644); err != nil {
		t.Fatal(err)
	}
	expectedTotal += 50

	// Test DirSize
	size := DirSize(tmpDir)
	if size != expectedTotal {
		t.Errorf("DirSize() = %d, want %d", size, expectedTotal)
	}
}
