// Package models contains all data types and structures used by the cleaner
package models

import "time"

// CleanupTarget represents a cleanup target
type CleanupTarget struct {
	Name         string
	Path         string
	Description  string
	Size         int64
	Selected     bool
	RequiresSudo bool
	Category     string
	IsCommand    bool   // If true, Path is a command to execute
	Command      string // Custom command to run
}

// BigFile represents a large file found
type BigFile struct {
	Path    string
	Size    int64
	ModTime time.Time
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Hash  string
	Size  int64
	Files []string
}

// OldFile represents an old/unused file
type OldFile struct {
	Path       string
	Size       int64
	LastAccess time.Time
}

// AppMode represents the current application mode
type AppMode int

const (
	ModeCleanup AppMode = iota
	ModeBigFiles
	ModeDuplicates
	ModeOldFiles
)

// State represents the UI state
type State int

const (
	StateMainMenu State = iota
	StateCleanupMenu
	StateScanning
	StateResults
	StateConfirm
	StateCleaning
	StateBigFilesConfig
	StateScanningBigFiles
	StateBigFilesResults
	StateBigFilesConfirm
	StateDuplicatesConfig
	StateScanningDuplicates
	StateDuplicatesResults
	StateDuplicatesConfirm
	StateOldFilesConfig
	StateScanningOldFiles
	StateOldFilesResults
	StateOldFilesConfirm
	StateDone
)

// ScanCompleteMsg is sent when scanning completes
type ScanCompleteMsg struct{}

// CleanCompleteMsg is sent when cleaning completes
type CleanCompleteMsg struct{ Saved int64 }

// CleanErrorMsg is sent when cleaning fails
type CleanErrorMsg struct{ Err error }

// BigFilesScanCompleteMsg is sent when big files scan completes
type BigFilesScanCompleteMsg struct{ Files []BigFile }

// DuplicatesScanCompleteMsg is sent when duplicates scan completes
type DuplicatesScanCompleteMsg struct {
	Groups    []DuplicateGroup
	TotalSize int64
}

// OldFilesScanCompleteMsg is sent when old files scan completes
type OldFilesScanCompleteMsg struct{ Files []OldFile }

// ScanProgressMsg is sent to update scan progress
type ScanProgressMsg struct {
	Status  string
	Percent float64
}
