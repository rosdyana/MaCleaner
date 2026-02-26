package models

import (
	"testing"
)

func TestGetDefaultTargets(t *testing.T) {
	targets := GetDefaultTargets()

	if len(targets) == 0 {
		t.Error("GetDefaultTargets() returned empty slice")
	}

	// Check that we have targets in different categories
	categories := make(map[string]bool)
	for _, target := range targets {
		categories[target.Category] = true
	}

	expectedCategories := []string{"Cache", "Logs", "Temp", "Trash", "Dev", "Package Manager", "Apps", "System", "Backups", "User"}
	for _, cat := range expectedCategories {
		if !categories[cat] {
			t.Errorf("Missing category: %s", cat)
		}
	}
}

func TestHasSelection(t *testing.T) {
	tests := []struct {
		name     string
		targets  []CleanupTarget
		expected bool
	}{
		{
			name: "no selection",
			targets: []CleanupTarget{
				{Name: "A", Selected: false},
				{Name: "B", Selected: false},
			},
			expected: false,
		},
		{
			name: "one selected",
			targets: []CleanupTarget{
				{Name: "A", Selected: false},
				{Name: "B", Selected: true},
			},
			expected: true,
		},
		{
			name: "all selected",
			targets: []CleanupTarget{
				{Name: "A", Selected: true},
				{Name: "B", Selected: true},
			},
			expected: true,
		},
		{
			name:     "empty slice",
			targets:  []CleanupTarget{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasSelection(tt.targets)
			if result != tt.expected {
				t.Errorf("HasSelection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasBigFilesSelection(t *testing.T) {
	tests := []struct {
		name     string
		selected map[int]bool
		expected bool
	}{
		{
			name:     "no selection",
			selected: map[int]bool{},
			expected: false,
		},
		{
			name:     "one selected",
			selected: map[int]bool{0: false, 1: true},
			expected: true,
		},
		{
			name:     "none selected",
			selected: map[int]bool{0: false, 1: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasBigFilesSelection(tt.selected)
			if result != tt.expected {
				t.Errorf("HasBigFilesSelection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasDuplicateSelection(t *testing.T) {
	tests := []struct {
		name     string
		selected map[int]bool
		expected bool
	}{
		{
			name:     "no selection",
			selected: map[int]bool{},
			expected: false,
		},
		{
			name:     "one selected",
			selected: map[int]bool{0: true},
			expected: true,
		},
		{
			name:     "none selected",
			selected: map[int]bool{0: false, 1: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasDuplicateSelection(tt.selected)
			if result != tt.expected {
				t.Errorf("HasDuplicateSelection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasOldFilesSelection(t *testing.T) {
	tests := []struct {
		name     string
		selected map[int]bool
		expected bool
	}{
		{
			name:     "no selection",
			selected: map[int]bool{},
			expected: false,
		},
		{
			name:     "one selected",
			selected: map[int]bool{2: true},
			expected: true,
		},
		{
			name:     "none selected",
			selected: map[int]bool{0: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasOldFilesSelection(tt.selected)
			if result != tt.expected {
				t.Errorf("HasOldFilesSelection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTargetProperties(t *testing.T) {
	targets := GetDefaultTargets()

	for _, target := range targets {
		t.Run(target.Name, func(t *testing.T) {
			// Every target should have a name
			if target.Name == "" {
				t.Error("Target has no name")
			}

			// Every target should have a category
			if target.Category == "" {
				t.Error("Target has no category")
			}

			// Every target should have a description
			if target.Description == "" {
				t.Error("Target has no description")
			}

			// Non-command targets should have a path
			if !target.IsCommand && target.Path == "" {
				t.Error("Non-command target has no path")
			}

			// Command targets should have a command
			if target.IsCommand && target.Command == "" {
				t.Error("Command target has no command")
			}
		})
	}
}
