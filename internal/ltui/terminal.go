// Package ltui provides a lightweight terminal UI without heavy dependencies
package ltui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"macos-cleaner/internal/models"
)

// Terminal provides simple terminal UI functionality
type Terminal struct {
	Width  int
	Height int
}

// NewTerminal creates a new terminal UI
func NewTerminal() *Terminal {
	return &Terminal{
		Width:  80,
		Height: 24,
	}
}

// Clear clears the terminal screen
func (t *Terminal) Clear() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		fmt.Print("\033[H\033[2J")
	}
}

// MoveCursor moves cursor to position (1-based)
func (t *Terminal) MoveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

// HideCursor hides the cursor
func (t *Terminal) HideCursor() {
	fmt.Print("\033[?25l")
}

// ShowCursor shows the cursor
func (t *Terminal) ShowCursor() {
	fmt.Print("\033[?25h")
}

// SetColor sets text color using ANSI codes
func (t *Terminal) SetColor(color string) {
	codes := map[string]string{
		"red":     "\033[31m",
		"green":   "\033[32m",
		"yellow":  "\033[33m",
		"blue":    "\033[34m",
		"magenta": "\033[35m",
		"cyan":    "\033[36m",
		"white":   "\033[37m",
		"gray":    "\033[90m",
		"reset":   "\033[0m",
		"bold":    "\033[1m",
	}
	if code, ok := codes[color]; ok {
		fmt.Print(code)
	}
}

// Reset resets all formatting
func (t *Terminal) Reset() {
	fmt.Print("\033[0m")
}

// PrintColored prints text with color
func (t *Terminal) PrintColored(color string, text string) {
	t.SetColor(color)
	fmt.Print(text)
	t.Reset()
}

// PrintBold prints bold text
func (t *Terminal) PrintBold(text string) {
	t.SetColor("bold")
	fmt.Print(text)
	t.Reset()
}

// PrintTitle prints a title
func (t *Terminal) PrintTitle(title string) {
	fmt.Println()
	t.PrintColored("cyan", "  🧹 ")
	t.PrintBold(title)
	fmt.Println()
	fmt.Println()
}

// PrintMenu prints the main menu
func (t *Terminal) PrintMenu() string {
	t.Clear()
	t.PrintTitle("macOS Storage Cleaner")

	t.PrintColored("green", "  Choose an option:\n\n")
	fmt.Println("  [1] 🧽 Storage Cleanup - Clean caches, logs, temp files")
	fmt.Println("  [2] 📦 Big Files Finder - Find large files taking up space")
	fmt.Println("  [3] 🔁 Duplicate Finder - Find duplicate files")
	fmt.Println("  [4] 📅 Old Files Finder - Find files not accessed recently")
	fmt.Println()
	t.PrintColored("gray", "  Press 1-4 to select, q to quit")
	fmt.Println()

	return t.ReadKey()
}

// PrintTargets prints cleanup targets for selection
func (t *Terminal) PrintTargets(targets []models.CleanupTarget, cursor int) string {
	t.Clear()
	t.PrintTitle("Storage Cleanup")

	currentCategory := ""
	for i, target := range targets {
		if target.Category != currentCategory {
			currentCategory = target.Category
			fmt.Println()
			t.PrintColored("magenta", "  "+currentCategory+":")
			fmt.Println()
		}

		cursorStr := "  "
		if cursor == i {
			cursorStr = "> "
			t.PrintColored("cyan", cursorStr)
		} else {
			fmt.Print(cursorStr)
		}

		checked := "[ ]"
		if target.Selected {
			checked = "[✓]"
			t.PrintColored("green", checked)
		} else {
			fmt.Print(checked)
		}

		fmt.Printf(" %-28s %s\n", target.Name, target.Description)
	}

	fmt.Println()
	t.PrintColored("gray", "  [↑↓] Navigate  [Space] Toggle  [a] All  [n] None  [s] Scan  [b] Back  [q] Quit")
	fmt.Println()

	return t.ReadKey()
}

// PrintScanning prints scanning status
func (t *Terminal) PrintScanning(status string) {
	t.Clear()
	t.PrintTitle("Scanning...")
	fmt.Println()
	fmt.Printf("  %s\n", status)
}

// PrintResults prints scan results
func (t *Terminal) PrintResults(targets []models.CleanupTarget, cursor int) string {
	t.Clear()
	t.PrintTitle("Scan Results")

	var totalSize int64
	for _, t := range targets {
		if t.Selected && t.Size > 0 {
			totalSize += t.Size
		}
	}

	fmt.Printf("  Total potential savings: ")
	t.PrintColored("yellow", formatBytes(totalSize))
	fmt.Println()
	fmt.Println()

	currentCategory := ""
	for i, target := range targets {
		if target.Category != currentCategory {
			currentCategory = target.Category
			t.PrintColored("magenta", "  "+currentCategory+":")
			fmt.Println()
		}

		cursorStr := "  "
		if cursor == i {
			cursorStr = "> "
		}

		checked := "[ ]"
		if target.Selected {
			checked = "[✓]"
		}

		sizeStr := formatBytes(target.Size)
		if target.Size == 0 {
			sizeStr = "Empty"
		}

		status := ""
		if target.RequiresSudo && target.Selected {
			status = "⚠ sudo"
		}

		if cursor == i {
			t.PrintColored("cyan", cursorStr+checked)
		} else {
			fmt.Print(cursorStr + checked)
		}
		fmt.Printf(" %-28s %10s %s\n", target.Name, sizeStr, status)
	}

	fmt.Println()
	t.PrintColored("gray", "  [↑↓] Navigate  [Space] Toggle  [c] Clean  [r] Rescan  [b] Back  [q] Quit")
	fmt.Println()

	return t.ReadKey()
}

// PrintConfirm prints confirmation dialog
func (t *Terminal) PrintConfirm(targets []models.CleanupTarget) string {
	t.Clear()
	t.PrintTitle("Confirm Cleanup")

	var totalSize int64
	fmt.Println("  The following will be deleted:")
	fmt.Println()
	for _, target := range targets {
		if target.Selected && target.Size > 0 {
			totalSize += target.Size
			fmt.Printf("    • %s (%s)\n", target.Name, formatBytes(target.Size))
		}
	}

	fmt.Println()
	fmt.Printf("  Total: ")
	t.PrintColored("yellow", formatBytes(totalSize))
	fmt.Println()
	fmt.Println()
	t.PrintColored("red", "  ⚠ This action cannot be undone!")
	fmt.Println()
	fmt.Println()
	t.PrintColored("gray", "  [y] Yes, delete  [n] Cancel")
	fmt.Println()

	return t.ReadKey()
}

// PrintCleaning prints cleaning status
func (t *Terminal) PrintCleaning(status string) {
	t.Clear()
	t.PrintTitle("Cleaning...")
	fmt.Println()
	fmt.Printf("  %s\n", status)
}

// PrintDone prints completion message
func (t *Terminal) PrintDone(totalSaved int64, lastError string) string {
	t.Clear()
	t.PrintTitle("Complete")

	if lastError != "" {
		t.PrintColored("red", "  ❌ Some operations failed:\n")
		fmt.Println()
		lines := strings.Split(lastError, "\n")
		for _, line := range lines {
			fmt.Println("  " + line)
		}
		if totalSaved > 0 {
			fmt.Println()
			t.PrintColored("green", "  ✅ Partial success: ")
			fmt.Printf("%s freed\n", formatBytes(totalSaved))
		}
	} else {
		t.PrintColored("green", "  ✅ Complete!")
		fmt.Println()
		fmt.Println()
		fmt.Printf("  Space freed: ")
		t.PrintColored("yellow", formatBytes(totalSaved))
		fmt.Println()
	}

	fmt.Println()
	t.PrintColored("gray", "  [b] Back to Menu  [q] Quit")
	fmt.Println()

	return t.ReadKey()
}

// PrintBigFilesConfig prints big files configuration
func (t *Terminal) PrintBigFilesConfig() string {
	t.Clear()
	t.PrintTitle("Big Files Finder")

	t.PrintColored("green", "  Find files larger than:\n\n")
	fmt.Println("  [1] 100 MB")
	fmt.Println("  [2] 500 MB")
	fmt.Println("  [3] 1 GB")
	fmt.Println("  [4] 5 GB")
	fmt.Println()
	t.PrintColored("gray", "  Press 1-4 to select, b to go back, q to quit")
	fmt.Println()

	return t.ReadKey()
}

// PrintBigFilesResults prints big files results
func (t *Terminal) PrintBigFilesResults(files []models.BigFile, selected map[int]bool, cursor int, minSize int64) string {
	t.Clear()
	t.PrintTitle("Big Files Results")
	fmt.Printf("  (>%s)\n\n", formatBytes(minSize))

	if len(files) == 0 {
		t.PrintColored("green", "  No large files found!")
		fmt.Println()
		fmt.Println()
		t.PrintColored("gray", "  [b] Back  [q] Quit")
		fmt.Println()
		return t.ReadKey()
	} else {
		fmt.Printf("  Found %d large files:\n\n", len(files))

		start := cursor
		if start > len(files)-15 {
			start = len(files) - 15
		}
		if start < 0 {
			start = 0
		}

		end := start + 15
		if end > len(files) {
			end = len(files)
		}

		for i := start; i < end; i++ {
			file := files[i]
			cursorStr := "  "
			if cursor == i {
				cursorStr = "> "
			}

			checked := "[ ]"
			if selected[i] {
				checked = "[✓]"
			}

			shortPath := file.Path
			if len(shortPath) > 50 {
				shortPath = "..." + shortPath[len(shortPath)-47:]
			}

			if cursor == i {
				t.PrintColored("cyan", cursorStr+checked)
			} else if selected[i] {
				t.PrintColored("green", cursorStr+checked)
			} else {
				fmt.Print(cursorStr + checked)
			}
			fmt.Printf(" %10s  %s\n", formatBytes(file.Size), shortPath)
		}

		if len(files) > 15 {
			fmt.Printf("\n  Showing %d-%d of %d files\n", start+1, end, len(files))
		}

		var selectedCount int
		var selectedSize int64
		for i, sel := range selected {
			if sel && i < len(files) {
				selectedCount++
				selectedSize += files[i].Size
			}
		}
		if selectedCount > 0 {
			fmt.Printf("\n  Selected: %d files (", selectedCount)
			t.PrintColored("yellow", formatBytes(selectedSize))
			fmt.Println(")")
		}
	}

	fmt.Println()
	t.PrintColored("gray", "  [↑↓] Navigate  [Space] Toggle  [a] All  [d] Delete  [b] Back  [q] Quit")
	fmt.Println()

	return t.ReadKey()
}

// PrintDuplicatesConfig prints duplicates configuration
func (t *Terminal) PrintDuplicatesConfig() string {
	t.Clear()
	t.PrintTitle("Duplicate Finder")

	fmt.Println("  This will scan your home directory for duplicate files.")
	fmt.Println("  Large directories like ~/Library will be skipped.")
	fmt.Println()
	t.PrintColored("yellow", "  ⚠ This may take several minutes!")
	fmt.Println()
	fmt.Println()
	t.PrintColored("gray", "  [s] Start Scan  [b] Back  [q] Quit")
	fmt.Println()

	return t.ReadKey()
}

// PrintOldFilesConfig prints old files configuration
func (t *Terminal) PrintOldFilesConfig() string {
	t.Clear()
	t.PrintTitle("Old Files Finder")

	t.PrintColored("green", "  Find files not accessed in:\n\n")
	fmt.Println("  [1] 30 days (1 month)")
	fmt.Println("  [2] 90 days (3 months)")
	fmt.Println("  [3] 180 days (6 months)")
	fmt.Println("  [4] 365 days (1 year)")
	fmt.Println()
	t.PrintColored("gray", "  Press 1-4 to select, b to go back, q to quit")
	fmt.Println()

	return t.ReadKey()
}

// PrintDuplicatesResults prints duplicate files results
func (t *Terminal) PrintDuplicatesResults(groups []models.DuplicateGroup, selected map[int]bool, cursor int) string {
	t.Clear()
	t.PrintTitle("Duplicate Files Results")

	if len(groups) == 0 {
		t.PrintColored("green", "  No duplicates found!")
		fmt.Println()
		fmt.Println()
		t.PrintColored("gray", "  [b] Back  [q] Quit")
		fmt.Println()
		return t.ReadKey()
	}

	fmt.Printf("  Found %d duplicate groups:\n\n", len(groups))

	start := cursor
	if start > len(groups)-5 {
		start = len(groups) - 5
	}
	if start < 0 {
		start = 0
	}

	end := start + 5
	if end > len(groups) {
		end = len(groups)
	}

	for i := start; i < end; i++ {
		group := groups[i]
		cursorStr := "  "
		if cursor == i {
			cursorStr = "> "
		}

		checked := "[ ]"
		if selected[i] {
			checked = "[✓]"
		}

		if cursor == i {
			t.PrintColored("cyan", cursorStr+checked)
		} else if selected[i] {
			t.PrintColored("green", cursorStr+checked)
		} else {
			fmt.Print(cursorStr + checked)
		}
		fmt.Printf(" Group %d: %s (%d files)\n", i+1, formatBytes(group.Size), len(group.Files))

		// Show first 3 files
		showCount := 3
		if len(group.Files) < showCount {
			showCount = len(group.Files)
		}
		for j := 0; j < showCount; j++ {
			shortPath := group.Files[j]
			if len(shortPath) > 60 {
				shortPath = "..." + shortPath[len(shortPath)-57:]
			}
			prefix := "    └─"
			if j < showCount-1 || len(group.Files) > showCount {
				prefix = "    ├─"
			}
			fmt.Printf("%s %s\n", prefix, shortPath)
		}
		if len(group.Files) > showCount {
			fmt.Printf("    ... and %d more\n", len(group.Files)-showCount)
		}
		fmt.Println()
	}

	if len(groups) > 5 {
		fmt.Printf("  Showing %d-%d of %d groups\n", start+1, end, len(groups))
	}

	var selectedCount int
	var selectedSize int64
	for i, sel := range selected {
		if sel && i < len(groups) {
			selectedCount++
			selectedSize += groups[i].Size * int64(len(groups[i].Files)-1)
		}
	}
	if selectedCount > 0 {
		fmt.Printf("\n  Selected: %d groups (saves ", selectedCount)
		t.PrintColored("yellow", formatBytes(selectedSize))
		fmt.Println(")")
	}

	fmt.Println()
	t.PrintColored("gray", "  [↑↓] Navigate  [Space] Toggle  [d] Delete Selected  [b] Back  [q] Quit")
	fmt.Println()

	return t.ReadKey()
}

// PrintOldFilesResults prints old files results
func (t *Terminal) PrintOldFilesResults(files []models.OldFile, selected map[int]bool, cursor int, days int) string {
	t.Clear()
	t.PrintTitle("Old Files Results")
	fmt.Printf("  (> %d days)\n\n", days)

	if len(files) == 0 {
		t.PrintColored("green", "  No old files found!")
		fmt.Println()
		fmt.Println()
		t.PrintColored("gray", "  [b] Back  [q] Quit")
		fmt.Println()
		return t.ReadKey()
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	fmt.Printf("  Found %d old files (", len(files))
	t.PrintColored("yellow", formatBytes(totalSize))
	fmt.Println("):")
	fmt.Println()

	start := cursor
	if start > len(files)-15 {
		start = len(files) - 15
	}
	if start < 0 {
		start = 0
	}

	end := start + 15
	if end > len(files) {
		end = len(files)
	}

	for i := start; i < end; i++ {
		file := files[i]
		cursorStr := "  "
		if cursor == i {
			cursorStr = "> "
		}

		checked := "[ ]"
		if selected[i] {
			checked = "[✓]"
		}

		daysAgo := int(time.Since(file.LastAccess).Hours() / 24)
		shortPath := file.Path
		if len(shortPath) > 50 {
			shortPath = "..." + shortPath[len(shortPath)-47:]
		}

		if cursor == i {
			t.PrintColored("cyan", cursorStr+checked)
		} else if selected[i] {
			t.PrintColored("green", cursorStr+checked)
		} else {
			fmt.Print(cursorStr + checked)
		}
		fmt.Printf(" %10s  %4dd  %s\n", formatBytes(file.Size), daysAgo, shortPath)
	}

	if len(files) > 15 {
		fmt.Printf("\n  Showing %d-%d of %d files\n", start+1, end, len(files))
	}

	var selectedCount int
	var selectedSize int64
	for i, sel := range selected {
		if sel && i < len(files) {
			selectedCount++
			selectedSize += files[i].Size
		}
	}
	if selectedCount > 0 {
		fmt.Printf("\n  Selected: %d files (", selectedCount)
		t.PrintColored("yellow", formatBytes(selectedSize))
		fmt.Println(")")
	}

	fmt.Println()
	t.PrintColored("gray", "  [↑↓] Navigate  [Space] Toggle  [a] All  [d] Delete  [b] Back  [q] Quit")
	fmt.Println()

	return t.ReadKey()
}

// ReadKey reads a single keypress
func (t *Terminal) ReadKey() string {
	reader := bufio.NewReader(os.Stdin)

	// Set raw mode
	oldState, err := makeRaw(os.Stdin)
	if err != nil {
		// Fallback to line reading
		line, _ := reader.ReadString('\n')
		return strings.TrimSpace(line)
	}
	defer restoreTerminal(os.Stdin, oldState)

	b, err := reader.ReadByte()
	if err != nil {
		return ""
	}

	// Handle escape sequences (arrow keys)
	if b == '\033' {
		seq := make([]byte, 2)
		reader.Read(seq)
		if seq[0] == '[' {
			switch seq[1] {
			case 'A':
				return "up"
			case 'B':
				return "down"
			}
		}
		return ""
	}

	return string(b)
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
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), units[exp])
}
